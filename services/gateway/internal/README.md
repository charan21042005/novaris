<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 📦 NOVARIS Gateway Internal Engine (`/services/gateway/internal`)
### Private Architectural Modules, Security Primitives & Middleware Pipelines

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Pattern-Go_Internal_Boundary-00C853?style=flat-square&logo=gitbook&logoColor=white" alt="Pattern" />
  <img src="https://img.shields.io/badge/Scope-Private_To_Module-7C4DFF?style=flat-square&logo=githubactions&logoColor=white" alt="Scope" />
  <img src="https://img.shields.io/badge/Security-Cryptographic_Core-FFA000?style=flat-square&logo=vault&logoColor=white" alt="Security" />
</p>

---

</div>

## 📌 The Purpose of the `internal/` Directory in Go

In the Go programming language, the `internal/` directory is **enforced by the Go compiler itself** (introduced in Go 1.4).

### The Go Compiler Encapsulation Rule:
> *An import of a path containing the element `internal` is disallowed if the importing package is outside the subtree rooted at the parent of the `internal` directory.*

### Why NOVARIS Uses `internal/`:
1. **Zero External API Leakage:** External services, third-party libraries, or other repositories cannot directly import or depend on our private cryptographic utilities, key generation algorithms, or context definitions.
2. **Encapsulation of Security Primitives:** The internal logic for verifying HMAC hashes, checking timestamp drift, and parsing JWT claims remains an implementation detail of the gateway service.
3. **Refactoring Safety:** Internal package APIs can be freely refactored without introducing breaking changes to public consumers.

---

## 🗂️ Internal Package Architecture

```
services/gateway/internal/
├── auth/                 # Cryptographic authentication & security middleware (Day 2-3)
│   ├── jwt.go            # RFC 7519 JWT verification & context claims injection
│   ├── jwt_test.go       # Table-driven unit tests for JWT pipeline
│   ├── hmac.go           # RFC 2104 HMAC-SHA256 request signing & replay defense
│   ├── hmac_test.go      # Unit tests for tampering, replay, & stream restoration
│   └── README.md         # Forensic security documentation for internal/auth
├── openapi/              # OpenAPI 3.0.3 contract & Swagger UI interactive dashboard (Day 12)
│   ├── openapi.go        # Embedded spec handlers & HTTP verb validation
│   ├── openapi.json      # Canonical OpenAPI 3.0.3 machine-readable schema
│   ├── openapi_test.go   # Table-driven tests for MIME types, headers & verbs
│   ├── swagger.html      # Self-contained embedded Swagger UI HTML5 dashboard
│   └── README.md         # OpenAPI embedding & Swagger documentation
├── ratelimit/            # In-memory traffic control & token-bucket rate limiter (Day 4)
│   ├── limiter.go        # Zero-goroutine lazy refill algorithm & sync.RWMutex store
│   ├── limiter_test.go   # Burst, refill, multi-goroutine race tests, and HTTP 429
│   └── README.md         # Rate limiter architecture, math derivation & test guide
├── streaming/            # Distributed event streaming & Kafka KRaft core (Phase 2, Day 15-17)
│   ├── config.go         # Cloud-native Kafka configuration & canonical topic catalog
│   ├── partitioner.go    # Deterministic Murmur2 key routing & FIFO account invariant
│   ├── admin.go          # Pure-Go cluster admin, KRaft metadata & topic provisioning
│   ├── producer.go       # Resilient Go Kafka Producer & TransactionInitiatedEvent model (Day 16)
│   ├── producer_test.go  # Unit tests for event validation, serialization & mock producer
│   ├── consumer.go       # High-performance Kafka Consumer Group & manual offset commits (Day 17)
│   ├── consumer_test.go  # Unit tests for ConsumerConfig, MockConsumer & error handling
│   ├── streaming_test.go # Comprehensive tests for hashing, distribution & integration
│   └── README.md         # Streaming core architecture, Murmur2 math & cluster guide
└── README.md             # This documentation
```

---

## 🔗 How Internal Packages Are Consumed

Internal packages are consumed exclusively by:
* The Gateway root server ([`services/gateway/main.go`](../main.go))
* The Gateway CLI suite ([`services/gateway/cmd/`](../cmd/))

```mermaid
flowchart TD
    subgraph PublicEdge ["Public & Developer Edge"]
        GatewayMain["services/gateway/main.go"]
        GenTokenCLI["cmd/gentoken/main.go"]
        SignRequestCLI["cmd/signrequest/main.go"]
        VerifyKafkaCLI["cmd/verifykafka/main.go"]
        VerifyProducerCLI["cmd/verifyproducer/main.go"]
        VerifyConsumerCLI["cmd/verifyconsumer/main.go"]
        VerifyOpenAPICLI["cmd/verifyopenapi/main.go"]
    end

    subgraph InternalBoundary ["Go Compiler 'internal/' Boundary"]
        AuthPackage["internal/auth (hmac.go, jwt.go)"]
        RateLimitPackage["internal/ratelimit (limiter.go)"]
        OpenAPIPackage["internal/openapi (openapi.go)"]
        StreamingPackage["internal/streaming (config.go, partitioner.go, admin.go)"]
    end

    subgraph ExternalDisallowed ["External Repositories (Disallowed)"]
        ExternalModule["github.com/other/project"]
    end

    GatewayMain -->|"Allowed: imports"| AuthPackage
    GatewayMain -->|"Allowed: imports"| RateLimitPackage
    GatewayMain -->|"Allowed: imports"| OpenAPIPackage
    GatewayMain -->|"Allowed: imports"| StreamingPackage
    GenTokenCLI -->|"Allowed: imports"| AuthPackage
    SignRequestCLI -->|"Allowed: imports"| AuthPackage
    VerifyKafkaCLI -->|"Allowed: imports"| StreamingPackage
    VerifyProducerCLI -->|"Allowed: imports"| StreamingPackage
    VerifyConsumerCLI -->|"Allowed: imports"| StreamingPackage
    VerifyOpenAPICLI -->|"Allowed: imports"| OpenAPIPackage

    ExternalModule -.->|"❌ Rejected by go build (compiler error)"| AuthPackage
    ExternalModule -.->|"❌ Rejected by go build (compiler error)"| RateLimitPackage
    ExternalModule -.->|"❌ Rejected by go build (compiler error)"| OpenAPIPackage
    ExternalModule -.->|"❌ Rejected by go build (compiler error)"| StreamingPackage
```

---

## 🧭 Directory Index

* 👉 **[internal/auth/ README](./auth/README.md):** In-depth documentation covering JWT claims verification, HMAC payload hashing, constant-time equality comparisons, and stream restoration mechanics.
* 👉 **[internal/openapi/ README](./openapi/README.md):** In-depth documentation covering embedded OpenAPI 3.0.3 specification, `//go:embed` hermetic packaging, and interactive Swagger UI dashboard.
* 👉 **[internal/ratelimit/ README](./ratelimit/README.md):** In-depth documentation covering token-bucket algorithm, zero-goroutine lazy mathematical refill, thread safety, and HTTP 429 Retry-After specifications.
* 👉 **[internal/streaming/ README](./streaming/README.md):** In-depth documentation covering Apache Kafka KRaft cluster architecture, deterministic Murmur2 partition routing, topic provisioning, and pure-Go administration.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Architectural Engine
*Go Compiler-Enforced Internal Package Boundary*

</div>
