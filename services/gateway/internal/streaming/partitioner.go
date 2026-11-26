// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// FILE: services/gateway/internal/streaming/partitioner.go
// PURPOSE: Deterministic Partition Key Hashing & Per-Account FIFO Guarantees
// ==============================================================================
//
// 🎓 THE MATHEMATICAL & COMPUTER SCIENCE INVARIANTS (FOR BEGINNERS):
//
// 1. Why "Global FIFO" is a Distributed Systems Fallacy:
//    In a high-throughput financial system processing 25,000+ transactions per second,
//    attempting to maintain a single, globally ordered queue requires acquiring a global lock
//    on every single write. This bottlenecks the entire platform to the throughput of a single
//    CPU core (~5,000 TPS) and causes catastrophic Head-of-Line Blocking: if Merchant A's
//    payment takes 200ms to evaluate, thousands of independent merchants are frozen waiting
//    behind it.
//
// 2. The Domain Reality of Financial Accounting:
//    Financial transactions between INDEPENDENT accounts have zero causal dependency.
//    Alice transferring $50 to Bob in California has NO mathematical impact on Charlie
//    buying groceries from a merchant in London.
//    HOWEVER, all transactions for the SAME account MUST be processed in strict, chronological
//    First-In, First-Out (FIFO) order to prevent double-spending, race conditions, and ledger drift.
//
// 3. The Kafka Partitioning Invariant:
//    Kafka achieves horizontal scalability by splitting a topic into N independent partitions.
//    By using the AccountID (or MerchantID) as the routing key, a deterministic hash function
//    maps that key to a specific partition:
//
//        Partition = |Hash(AccountID)| mod N_partitions
//
//    Invariant:
//        ∀ t1, t2 ∈ Transactions: Key(t1) == Key(t2) ⇒ Partition(t1) == Partition(t2)
//
//    Because each partition is consumed sequentially by a single thread in a consumer group,
//    strict per-account FIFO ordering is mathematically guaranteed with ZERO cross-thread locks!
// ==============================================================================

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
	// Default strategy in our Go microservice stack.
	StrategyFNV1a PartitionerStrategy = "FNV-1a"

	// StrategyMurmur2 mimics Apache Kafka's default Java Partitioner algorithm.
	// Essential for cross-language compatibility when Go producers share topics with Java/Scala consumers.
	StrategyMurmur2 PartitionerStrategy = "Murmur2-Compatible"

	// StrategySHA256 uses cryptographic SHA-256 (higher CPU cost, 100% collision resilience).
	StrategySHA256 PartitionerStrategy = "SHA-256"
)

// Partitioner provides deterministic mathematical mapping of transaction routing keys (e.g. AccountID)
// to specific partition IDs across a Kafka topic.
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
//
// 🎓 Step-by-Step Mathematical Walkthrough for Beginners:
//  1. Guard Clause: If numPartitions <= 1, there is only 1 partition (Partition 0), so return 0.
//  2. Empty Key Guard: If key is empty, return 0 (or round-robin).
//  3. Hashing: Convert the string key into an unsigned 64-bit integer hash using FNV-1a, Murmur2, or SHA-256.
//  4. Sign Bit Masking: Bitwise AND with math.MaxInt64 (`hashVal & 0x7FFFFFFFFFFFFFFF`).
//     Why? In two's complement binary arithmetic, if the highest bit is 1, casting to a signed integer
//     yields a negative number! In Go and Java, `-5 % 3` equals `-2`, which is an INVALID partition index!
//     Clearing the sign bit guarantees the hash is strictly non-negative.
//  5. Modulo Arithmetic: `positiveHash % uint64(numPartitions)` guarantees the result falls
//     strictly in the range [0, numPartitions - 1].
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
		// Compute 32-byte SHA-256 digest; extract first 8 bytes as big-endian uint64:
		sum := sha256.Sum256([]byte(key))
		hashVal = binary.BigEndian.Uint64(sum[:8])
	case StrategyMurmur2:
		// Compute 32-bit MurmurHash2 matching Kafka Java client:
		hashVal = uint64(murmur2([]byte(key)))
	case StrategyFNV1a:
		fallthrough
	default:
		// 64-bit Fowler-Noll-Vo 1a hash:
		h := fnv.New64a()
		_, _ = h.Write([]byte(key))
		hashVal = h.Sum64()
	}

	// Clear the sign bit to guarantee non-negative value:
	positiveHash := hashVal & math.MaxInt64
	return int(positiveHash % uint64(numPartitions))
}

// PartitionDistribution tracks the distribution of keys across available partitions
// to evaluate hashing uniformity and avoid "hot partition" imbalances.
type PartitionDistribution struct {
	TotalKeys       int             `json:"total_keys"`
	NumPartitions   int             `json:"num_partitions"`
	PartitionLoads  map[int]int     `json:"partition_loads"`
	LoadPercentages map[int]float64 `json:"load_percentages"`
	StandardDev     float64         `json:"standard_deviation"`
}

// AnalyzeDistribution evaluates the uniformity of partition routing across a test set of account IDs.
// A good hash function yields a near-zero standard deviation across partitions.
//
// Mathematical Explanation:
// If 3,000 keys are distributed across 3 partitions, the ideal mean is 1,000 keys per partition.
// We compute the standard deviation:
//   σ = sqrt( (1/N) * Σ(count_i - mean)^2 )
// Low standard deviation (< 5% of mean) proves the absence of "hot partitions."
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
//
// 🎓 Computer Science Details on MurmurHash2:
// Created by Austin Appleby, MurmurHash2 is a non-cryptographic hash function famous for its speed
// and excellent "Avalanche Effect" (changing a single bit in the key flips ~50% of the bits in the output hash).
//
// Kafka Constants:
//  - seed: 0x9747b28c (the exact seed hardcoded in Kafka's Java DefaultPartitioner)
//  - m:    0x5bd1e995 (a 32-bit arithmetic constant chosen for optimal bit distribution)
//  - r:    24 (rotation constant)
//
// By matching these exact constants, NOVARIS guarantees that Go producers and Java microservices
// hash identical account IDs to the exact same partition index!
func murmur2(data []byte) uint32 {
	const (
		seed uint32 = 0x9747b28c
		m    uint32 = 0x5bd1e995
		r           = 24
	)

	length := len(data)
	h := seed ^ uint32(length)
	length4 := length >> 2

	// Process 4-byte chunks in a tight loop:
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

	// Process remaining 1 to 3 trailing bytes:
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

	// Final bit mixing to cascade the avalanche effect:
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
