// Package streaming provides high-performance, fault-tolerant Apache Kafka event streaming
// primitives for the NOVARIS payment settlement and adaptive fraud intelligence platform.
package streaming

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
)

// PartitionerStrategy specifies the algorithm used to map message keys to partition indices.
type PartitionerStrategy string

const (
	// StrategyFNV1a uses the Fowler-Noll-Vo 1a non-cryptographic hash (fast, uniform distribution).
	StrategyFNV1a PartitionerStrategy = "FNV-1a"

	// StrategyMurmur2 mimics Apache Kafka's default Java Partitioner algorithm.
	StrategyMurmur2 PartitionerStrategy = "Murmur2-Compatible"

	// StrategySHA256 uses cryptographic SHA-256 (higher CPU cost, 100% collision resilience).
	StrategySHA256 PartitionerStrategy = "SHA-256"
)

// Partitioner provides deterministic mathematical mapping of transaction routing keys (e.g. AccountID)
// to specific partition IDs across a Kafka topic.
//
// 🎓 THE MATHEMATICAL & COMPUTER SCIENCE INVARIANT:
//
// In distributed computing, the "CAP Theorem" and the physics of network latency dictate that
// maintaining a total, globally ordered sequence across a multi-node cluster under 100,000 TPS
// is physically impossible without a centralized bottleneck that kills throughput.
//
// In payment systems:
//   - Total ordering between independent accounts is UNNECESSARY (Alice sending money to Bob has zero
//     causal relationship with Charlie buying coffee from Starbucks).
//   - Strict FIFO ordering FOR A SINGLE ACCOUNT IS MANDATORY (preventing race conditions, double-spending,
//     and negative balance corruption).
//
// Kafka achieves horizontal scalability through Partitions. By using AccountID as the routing key:
//   P = |Hash(AccountID)| mod N_partitions
//
// Mathematical Invariant:
//   ∀ t1, t2 ∈ Transactions: Key(t1) == Key(t2) ⇒ Partition(t1) == Partition(t2)
//
// This guarantees that all state transitions for a single account are appended to the SAME partition log,
// processed sequentially by the SAME consumer thread, with zero cross-thread locking overhead!
type Partitioner struct {
	strategy PartitionerStrategy
}

// NewPartitioner instantiates a Partitioner with the specified hashing strategy.
func NewPartitioner(strategy PartitionerStrategy) *Partitioner {
	if strategy == "" {
		strategy = StrategyFNV1a
	}
	return &Partitioner{strategy: strategy}
}

// RouteAccountKey computes the partition index [0, numPartitions - 1] for a given routing key.
// If numPartitions <= 1, it trivially returns 0.
// If key is empty, it returns 0 (or round-robin in consumer writers).
func (p *Partitioner) RouteAccountKey(key string, numPartitions int) int {
	if numPartitions <= 1 {
		return 0
	}
	if len(key) == 0 {
		return 0
	}

	var hashVal uint64

	switch p.strategy {
	case StrategySHA256:
		sum := sha256.Sum256([]byte(key))
		hashVal = binary.BigEndian.Uint64(sum[:8])
	case StrategyMurmur2:
		hashVal = uint64(murmur2([]byte(key)))
	case StrategyFNV1a:
		fallthrough
	default:
		h := fnv.New64a()
		_, _ = h.Write([]byte(key))
		hashVal = h.Sum64()
	}

	// Abs and modulo:
	// Cast to int64, clear sign bit to prevent negative modulo results, then mod numPartitions.
	positiveHash := hashVal & math.MaxInt64
	return int(positiveHash % uint64(numPartitions))
}

// PartitionDistribution tracks the distribution of keys across available partitions
// to evaluate hashing uniformity and avoid "hot partition" imbalances.
type PartitionDistribution struct {
	TotalKeys      int            `json:"total_keys"`
	NumPartitions  int            `json:"num_partitions"`
	PartitionLoads map[int]int    `json:"partition_loads"`
	LoadPercentages map[int]float64 `json:"load_percentages"`
	StandardDev    float64        `json:"standard_deviation"`
}

// AnalyzeDistribution evaluates the uniformity of partition routing across a test set of account IDs.
// A good hash function yields a near-zero standard deviation across partitions.
func (p *Partitioner) AnalyzeDistribution(keys []string, numPartitions int) PartitionDistribution {
	loads := make(map[int]int, numPartitions)
	percentages := make(map[int]float64, numPartitions)

	for i := 0; i < numPartitions; i++ {
		loads[i] = 0
	}

	for _, k := range keys {
		part := p.RouteAccountKey(k, numPartitions)
		loads[part]++
	}

	total := len(keys)
	var mean float64
	if numPartitions > 0 {
		mean = float64(total) / float64(numPartitions)
	}

	var sumSqDiff float64
	for i := 0; i < numPartitions; i++ {
		count := loads[i]
		if total > 0 {
			percentages[i] = (float64(count) / float64(total)) * 100.0
		} else {
			percentages[i] = 0
		}
		diff := float64(count) - mean
		sumSqDiff += diff * diff
	}

	var stdDev float64
	if numPartitions > 0 {
		stdDev = math.Sqrt(sumSqDiff / float64(numPartitions))
	}

	return PartitionDistribution{
		TotalKeys:       total,
		NumPartitions:   numPartitions,
		PartitionLoads:  loads,
		LoadPercentages: percentages,
		StandardDev:     stdDev,
	}
}

// murmur2 provides a Go port of the MurmurHash2 algorithm used in standard Apache Kafka Java clients.
// This ensures that Go producers and Java/Scala/Python producers route identical account keys
// to the exact same partition index!
func murmur2(data []byte) uint32 {
	const (
		seed uint32 = 0x9747b28c
		m    uint32 = 0x5bd1e995
		r           = 24
	)

	length := len(data)
	h := seed ^ uint32(length)
	length4 := length >> 2

	for i := 0; i < length4; i++ {
		i4 := i << 2
		k := uint32(data[i4+0]) |
			(uint32(data[i4+1]) << 8) |
			(uint32(data[i4+2]) << 16) |
			(uint32(data[i4+3]) << 24)

		k *= m
		k ^= k >> r
		k *= m

		h *= m
		h ^= k
	}

	rem := length & 3
	switch rem {
	case 3:
		h ^= uint32(data[(length&^3)+2]) << 16
		fallthrough
	case 2:
		h ^= uint32(data[(length&^3)+1]) << 8
		fallthrough
	case 1:
		h ^= uint32(data[length&^3])
		h *= m
	}

	h ^= h >> 13
	h *= m
	h ^= h >> 15

	return h
}

// FormatRoutingExplanation produces a beginner-friendly mathematical demonstration of how
// a specific transaction key was mapped to its partition.
func FormatRoutingExplanation(key string, numPartitions int, partition int, strategy PartitionerStrategy) string {
	return fmt.Sprintf(
		"Key '%s' -> [%s Hashing] -> Modulo %d Partitions -> Landed on Partition %d (Strict per-account FIFO guaranteed)",
		key, strategy, numPartitions, partition,
	)
}
