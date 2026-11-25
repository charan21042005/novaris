<div align="center">

<!-- Floating Top Banner -->
<p align="center">
  <img src="../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 📖 NOVARIS THEORY TEXTBOOK • DAY 03
## Cryptographic HMAC-SHA256 Request Signing & Replay Attack Defense

<p align="center">
  <img src="https://img.shields.io/badge/Volume-01_Foundations_%26_Security-00E5FF?style=flat-square&logo=gitbook&logoColor=white" alt="Volume" />
  <img src="https://img.shields.io/badge/Day-03_of_84-00C853?style=flat-square&logo=googlecalendar&logoColor=white" alt="Day" />
  <img src="https://img.shields.io/badge/Level-First_Principles_to_Production-7C4DFF?style=flat-square&logo=academia&logoColor=white" alt="Level" />
  <img src="https://img.shields.io/badge/Standard-Industrial_FinTech_Grade-FFA000?style=flat-square&logo=standards&logoColor=white" alt="Standard" />
</p>

---

</div>

## 📑 Detailed Table of Contents

1. [Preface: Sealing the Envelope — The Gap Day 2 Left Open](#1-preface-sealing-the-envelope--the-gap-day-2-left-open)
2. [Chapter 0: The Plain-English Cryptography & Integrity Lexicon](#2-chapter-0-the-plain-english-cryptography--integrity-lexicon)
   - 0.1 Message Integrity vs. Authentication vs. Confidentiality
   - 0.2 Cryptographic Foundations (Hashing vs. Encryption vs. HMAC)
   - 0.3 Attack Vectors (Tampering, Man-in-the-Middle, Replay Attacks, Timing Side-Channels)
   - 0.4 Systems Architecture (Canonicalization, Stream Draining & Restoration, Defense in Depth)
3. [Chapter 0.5: Complete Software Components, Tooling & Dependency Directory](#3-chapter-05-complete-software-components-tooling--dependency-directory)
   - 0.5.1 Core Cryptographic Primitives (`crypto/hmac`, `crypto/sha256`, `crypto/subtle`, `encoding/hex`)
   - 0.5.2 Go Standard Library Modules (`io`, `bytes`, `strconv`, `net/http`, `context`)
   - 0.5.3 CLI & Probing Utilities (`cmd/signrequest`, `cmd/gentoken`, `curl`, PowerShell)
   - 0.5.4 Standards & RFC Specifications (RFC 2104, FIPS PUB 198-1, RFC 6234)
4. [Chapter 1: The Theoretical Foundations of Request Signing in FinTech](#4-chapter-1-the-theoretical-foundations-of-request-signing-in-fintech)
   - 1.1 The Man-in-the-Middle Vulnerability: Why JWT Authentication Alone is Insufficient
   - 1.2 The Dollar Amount Rewrite Exploit ($50.00 → $50,000.00)
   - 1.3 The Two Pillars of API Defense: Identity (JWT) + Integrity (HMAC)
   - 1.4 How Real-World Gateways Do It (Stripe Webhooks, AWS SigV4, GitHub Webhooks)
5. [Chapter 2: The Mathematical Architecture of HMAC-SHA256 (RFC 2104)](#5-chapter-2-the-mathematical-architecture-of-hmac-sha256-rfc-2104)
   - 2.1 The Mathematics of HMAC: Inner and Outer Padding ($ipad$, $opad$)
   - 2.2 Why Plain Hashing $H(K \parallel M)$ Fails: The Length Extension Attack
   - 2.3 The Cryptographic Avalanche Effect in Payment Tamper Detection
   - 2.4 Symmetric HMAC vs. Asymmetric Digital Signatures (RSA/ECDSA)
6. [Chapter 3: Critical Engineering Obstacles & Defense Science](#6-chapter-3-critical-engineering-obstacles--defense-science)
   - 3.1 The Canonicalization Problem: Why Byte-Accuracy Matters in JSON
   - 3.2 The Stream Draining Trap in Go HTTP Servers (`io.ReadAll` and `io.NopCloser`)
   - 3.3 The Side-Channel Timing Attack: Why `==` is a Critical Vulnerability and `hmac.Equal` is Mandatory
   - 3.4 Replay Attacks: The $\pm 300$-Second Timestamp Window Defense
   - 3.5 Beyond Timestamps: Nonce Deduping in Redis (The Horizon to Week 3-4)
   - 3.6 Fail-Fast Optimization: Cheap Replay Checking Before Crypto Math
7. [Chapter 4: Defense in Depth & Middleware Stacking Architecture](#7-chapter-4-defense-in-depth--middleware-stacking-architecture)
   - 4.1 The Ensemble Security Model: Why Layered Defenses Outperform Monolithic Checks
   - 4.2 Middleware Stacking Order: Why `requireJWT(requireHMAC(paymentsHandler))` Matters
   - 4.3 Separation of Concerns: Outer Gateways vs. Inner Ledger Engines
8. [Chapter 5: The Data Science → Full-Stack Integrity Rosetta Stone](#8-chapter-5-the-data-science--full-stack-integrity-rosetta-stone)
9. [Chapter 6: Architectural Trade-Off Analysis ("Why This and Not That")](#9-chapter-6-architectural-trade-off-analysis-why-this-and-not-that)
   - 6.1 HMAC-SHA256 vs. RSA-SHA256 (Speed, Key Management, and Operational Complexity)
   - 6.2 Timestamp Window Duration (300s vs 60s vs 15m: Clock Skew vs Replay Vulnerability)
   - 6.3 Hex Encoding vs. Base64 Encoding for HTTP Signature Headers
   - 6.4 Raw Network Stream Signing vs. Normalized JSON Signing
10. [Chapter 7: Line-by-Line Code Masterclass (Complete Dissections)](#10-chapter-7-line-by-line-code-masterclass-complete-dissections)
    - 7.1 `services/gateway/internal/auth/hmac.go` (The Complete HMAC Engine)
    - 7.2 `services/gateway/cmd/signrequest/main.go` (The Developer Request Signer CLI)
    - 7.3 `services/gateway/main.go` (Stacked Route Pipeline & Restored Body Handling)
    - 7.4 `services/gateway/internal/auth/hmac_test.go` (Unit Test Suite)
    - 7.5 `services/gateway/README.md` (Updated Gateway Specs)
11. [Chapter 8: Verification Laboratory: Execution Proof & Log Dissection](#11-chapter-8-verification-laboratory-execution-proof--log-dissection)
    - 8.1 Complete Automated Unit Test Suite (10 passing tests)
    - 8.2 Real-World HTTP Verification Matrix: 7 Concrete Probe Scenarios
12. [Chapter 9: The Interview Defense Compendium (10 High-Stakes Tier-1 Q&As)](#12-chapter-9-the-interview-defense-compendium-10-high-stakes-tier-1-qas)
13. [Chapter 10: The Roadmap Horizon (Day 4 Preview & Hands-On Exercises)](#13-chapter-10-the-roadmap-horizon-day-4-preview--hands-on-exercises)

---

## 1. Preface: Sealing the Envelope — The Gap Day 2 Left Open

On **Day 2**, we implemented JSON Web Token (JWT) authentication. Our gateway learned how to answer the first fundamental question of financial engineering: **"Who are you?"**

Every request arriving at `/v1/payments` without a valid, cryptographically verified Bearer token is immediately rejected with HTTP `401 Unauthorized`. The gateway extracts the merchant identity (`merchant_id`) and permissions (`role`) and injects them into the Go request context.

However, Day 2 left a massive, subtle security vulnerability wide open:
**A valid JWT proves WHO is sending the request, but it says absolutely NOTHING about WHAT was sent in the request body!**

```
       WHAT DAY 2 VERIFIED:                   WHAT DAY 2 COMPLETELY IGNORED:
       
┌──────────────────────────────────────┐     ┌────────────────────────────────────────────────────────┐
│ Header: Bearer <JWT>                 │     │ Body:                                                  │
│ Proves: "The caller is Merchant_123" │     │   { "amount_cents": 5000, "recipient": "supplier" }   │
│ Gateway says: "Identity approved!"   │     │                                                        │
└──────────────────────────────────────┘     │ PROBLEM: An attacker sitting in the middle can rewrite │
                                             │ amount_cents to $5,000,000, and Day 2's gateway will   │
                                             │ STILL PROCESS IT because the JWT is valid!             │
                                             └────────────────────────────────────────────────────────┘
```

### The Physical Analogy: The Wax Seal on an Envelope
* **JWT (Day 2)** is like a company badge worn around an employee's neck. The guard looks at the badge and says: *"Yes, you work here. You may enter."*
* **HMAC Request Signing (Day 3)** is the **crimson wax seal stamped over the envelope's flap**. Even if an employee carries the letter across town, the recipient checks the wax seal. If anyone sliced open the envelope, swapped out a \$50 check for a \$50,000 check, and taped it back together, the wax seal is broken, and the bank immediately burns the check!

**Today on Day 3**, we seal the envelope. We implement **Cryptographic HMAC-SHA256 Request Body Signing** combined with **Replay Attack Defense** (`X-Novaris-Timestamp` with a $\pm 300$-second window) and **Constant-Time Verification** (`hmac.Equal`).

---

## 2. Chapter 0: The Plain-English Cryptography & Integrity Lexicon

To ensure anyone from a computer science student to a senior data scientist can follow without friction, here are the core concepts defined from absolute first principles:

### 0.1 Message Integrity vs. Authentication vs. Confidentiality

The "CIA Triad" is the holy trinity of information security. Beginner engineers frequently confuse these three pillars:

```
               THE THREE PILLARS OF FINANCIAL NETWORK SECURITY
               
      CONFIDENTIALITY             AUTHENTICATION                 INTEGRITY
     (Secrecy / Privacy)        (Identity Verification)      (Tamper Prevention)
   ┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐
   │ TLS / HTTPS / AES   │     │ JWT / Bearer Tokens │     │ HMAC-SHA256 Signing │
   │                     │     │                     │     │                     │
   │ "Nobody between us  │     │ "I can mathematically│    │ "Not a single byte  │
   │ can eavesdrop or    │     │ prove who is talking│     │ was altered since   │
   │ read this payload." │     │ to me right now."   │     │ the sender sent it."│
   └─────────────────────┘     └─────────────────────┘     └─────────────────────┘
```

1. **Confidentiality:** Scrambling data so eavesdroppers cannot read it (provided by TLS 1.3 / HTTPS).
2. **Authentication (Day 2):** Proving the identity of the sender (provided by JWT).
3. **Integrity (Day 3):** Proving that the payload arrived **100% byte-for-byte identical** to what the sender generated, with zero modification in transit (provided by HMAC).

---

### 0.2 Cryptographic Foundations

#### 1. What is "Hashing" (e.g., SHA-256)?
* **Plain English:** A digital fingerprint. You feed in any input (a 10-character password or a 10-gigabyte database backup), and it produces a fixed 64-character hexadecimal digest.
* **Property:** It is **unkeyed** and public. Anyone who has the data can compute the exact same hash.
* **Why Hashing Alone Cannot Prove Authenticity:** If Alice sends Bob `Message` + `SHA256(Message)`, an attacker in the middle (Mallory) can intercept it, change `Message` to `ForgedMessage`, calculate `SHA256(ForgedMessage)`, and send both to Bob. Bob hashes the forged message, sees it matches Mallory's hash, and naively trusts it!

#### 2. What is "HMAC" (Hash-based Message Authentication Code)?
* **Plain English:** A **keyed cryptographic fingerprint**. It is a mathematical function that combines the message $M$ with a secret key $K$ before running the hash algorithm.
* **Why HMAC Solves Mallory's Attack:** Because Mallory does **not know** the secret key $K$, she cannot compute a valid HMAC for her forged message. When Bob receives the forged message, he hashes it with $K$. The signature does not match, and Bob immediately rejects the request.

#### 3. What is a "Canonical String"?
* **Plain English:** An agreed-upon, standardized recipe for formatting different pieces of information before feeding them into a cryptographic hash.
* **Why it Matters:** A hash function does not understand "JSON objects" or "headers". It only sees a stream of raw binary bytes. If the client signs `Timestamp + "." + Body` while the server hashes `Body + "." + Timestamp`, the hashes will be completely different! Both sides must construct the **exact same canonical sequence**.

---

### 0.3 Attack Vectors

#### 4. What is a "Man-in-the-Middle (MitM) Tampering Attack"?
* **Plain English:** An adversary situated on a network router, proxy, or compromised VPN who intercepts data traveling between a merchant and our API Gateway, alters the payload (e.g., increasing payout amounts or changing bank account numbers), and forwards it to the server.
* **In NOVARIS:** Defended against by HMAC-SHA256 signing. Any alteration of even a single bit in the JSON body invalidates the signature.

#### 5. What is a "Replay Attack"?
* **Plain English:** An attacker intercepts a completely valid, authentic request and sends (replays) it to the server multiple times to duplicate an action.
* **Real-World Analogy:** You give a friend a signed permission slip: *"Please give John $100 from my desk."* John takes the slip, gets \$100. Tomorrow, John Photocopies the same slip and asks for another \$100. The signature is genuine, but the request was replayed!
* **In NOVARIS:** We include an `X-Novaris-Timestamp` header inside the signed canonical string. If the timestamp is older than 300 seconds (5 minutes) or in the future, the gateway rejects it immediately.

#### 6. What is a "Side-Channel Timing Attack"?
* **Plain English:** An attack where an adversary measures the **nanosecond latency** of a server's response to guess secret cryptographic keys or signatures byte by byte.
* **How it Works with Naive `==`:** Standard string comparisons in C, Python, and Go return `false` as soon as they encounter the first mismatched character. A signature starting with the correct first letter takes ~10 nanoseconds longer to fail than one with a wrong first letter. An attacker fires 100,000 requests, graphs the response latency, and discovers the signature character by character!
* **The Defense (`hmac.Equal`):** A **constant-time comparison** algorithm that always takes the exact same number of CPU cycles regardless of where or whether characters match.

---

### 0.4 Systems Architecture

#### 7. What is "Stream Draining" in HTTP Servers?
* **Plain English:** In programming languages like Go, the HTTP request body (`r.Body`) is an **I/O stream** (`io.ReadCloser`), not an in-memory string. It is like water flowing through a garden hose.
* **The Trap:** When our HMAC middleware reads the stream to compute the signature (`io.ReadAll(r.Body)`), the stream is drained (the hose is empty!). If you don't refill the stream, when downstream business logic (`paymentsHandler`) tries to read the JSON payload, it receives 0 bytes (EOF)!
* **The Fix:** We save the read bytes and restore the body using `r.Body = io.NopCloser(bytes.NewReader(bodyBytes))`.

#### 8. What is "Defense in Depth"?
* **Plain English:** Layering multiple independent security mechanisms so that if any single mechanism fails or has a blind spot, the remaining layers still prevent catastrophe.
* **Ensemble Analogy:** In machine learning, a Random Forest combines hundreds of decision trees. Even if one tree makes an error, the ensemble output is robust. In NOVARIS:
  * Layer 1: JWT verifies **Identity**.
  * Layer 2: HMAC verifies **Payload Integrity & Replay Freshness**.
  * Layer 3: Rate Limiting (Day 4) protects against **Flooding & DoS**.

---

## 3. Chapter 0.5: Complete Software Components, Tooling & Dependency Directory

Here is the exhaustive directory of software tools, libraries, and specifications utilized during Day 3:

### 0.5.1 Core Cryptographic Primitives

| Component / Package | Module | Primary Purpose in Day 3 |
| :--- | :--- | :--- |
| **`crypto/hmac`** | Go Standard Library | Implements Hash-based Message Authentication Codes (FIPS PUB 198-1). Provides `hmac.New` and `hmac.Equal` (constant-time comparison). |
| **`crypto/sha256`** | Go Standard Library | Implements the SHA-256 cryptographic hash function (FIPS 180-4). Generates 256-bit (32-byte) digests. |
| **`crypto/subtle`** | Go Standard Library | Powers constant-time byte comparisons (`ConstantTimeCompare`), preventing side-channel timing attacks. |
| **`encoding/hex`** | Go Standard Library | Encodes raw 32-byte binary HMAC digests into 64-character URL-safe lowercase hexadecimal strings. |

### 0.5.2 Go Standard Library Modules

| Package | Purpose & Functionality |
| :--- | :--- |
| **`io`** | Provides `io.ReadAll` (reads raw body stream) and `io.NopCloser` (wraps in-memory byte buffers into `io.ReadCloser`). |
| **`bytes`** | Provides `bytes.NewReader` to create a fresh readable stream for `r.Body` after middleware inspection. |
| **`strconv`** | High-performance integer string parsing (`strconv.ParseInt`, `strconv.FormatInt`) for UNIX timestamps. |
| **`net/http`** | Web server muxing, handler composition (`http.HandlerFunc`), and HTTP status codes (`401 Unauthorized`, `400 Bad Request`). |
| **`context`** | Carries pre-verified claims across middleware boundaries. |
| **`encoding/json`** | Decodes verified request payloads into strongly-typed `PaymentRequest` structs. |
| **`flag`** | CLI argument parsing for test utilities (`cmd/signrequest/main.go`). |

### 0.5.3 CLI & Probing Utilities

| Tool | Purpose in Day 3 |
| :--- | :--- |
| **`cmd/signrequest`** | Developer CLI utility: takes raw JSON and an optional timestamp offset, computes valid HMAC signatures, and prints ready-to-run curl commands. |
| **`cmd/gentoken`** | Developer CLI utility from Day 2: mints valid Bearer JWTs to accompany signed requests. |
| **PowerShell Test Harness** | Automated script compiling `gateway.exe`, booting it, and verifying 7 distinct security scenarios (tampering, replay, missing headers). |
| **`go test -v ./...`** | Runs 10 comprehensive unit tests covering JWT, HMAC, tampering, stale timestamps, and middleware pipelines. |

### 0.5.4 Standards & RFC Specifications

| Standard | Full Title | Why NOVARIS Adheres to It |
| :--- | :--- | :--- |
| **RFC 2104** | HMAC: Keyed-Hashing for Message Authentication | Governs cryptographic message integrity verification used across global payment APIs. |
| **FIPS PUB 198-1** | The Keyed-Hash Message Authentication Code | Federal standard defining the exact padding and XOR math of HMAC. |
| **RFC 6234** | US Secure Hash Algorithms (SHA and SHA-based HMAC) | Specifies SHA-256 mathematical behavior, preventing length extension vulnerabilities. |

---

## 4. Chapter 1: The Theoretical Foundations of Request Signing in FinTech

### 1.1 The Man-in-the-Middle Vulnerability: Why JWT Alone Fails

Consider an API that uses **only JWT** for security:
1. Merchant Acme Corp logs in and receives a valid JWT token signed by NOVARIS.
2. Acme sends an HTTP request:
   ```http
   POST /v1/payments HTTP/1.1
   Host: api.novaris.io
   Authorization: Bearer eyJhbGciOiJIUzI1Ni... (Valid JWT for Acme Corp)
   Content-Type: application/json

   {
     "amount_cents": 5000,
     "recipient": "supplier_factory_456"
   }
   ```
3. An adversary (sitting on a corporate WiFi router or intermediate VPN) intercepts this packet.
4. The adversary **does not need to crack the JWT**. The JWT is untouched!
5. The adversary modifies the HTTP body bytes:
   ```json
   {
     "amount_cents": 5000000,
     "recipient": "attacker_bank_account_999"
   }
   ```
6. The packet arrives at the API Gateway.
7. **What happens?**
   * The gateway checks the `Authorization: Bearer <JWT>` header.
   * Signature is valid! Expiration is in the future!
   * Gateway says: *"Acme Corp is authenticated. Let's process the payment!"*
   * **Acme Corp just lost \$50,000 to an attacker!**

```
                  THE CATASTROPHIC FAILURE OF JWT-ONLY SECURITY
                  
  Merchant Acme                     Malicious Attacker                     API Gateway
┌──────────────┐                   ┌──────────────────┐                   ┌───────────┐
│ Valid JWT    │──[ $50 Payment ]─►│ Intercepts packet│──[ $50,000 Mod ]─►│ Checks JWT│
│ Signed by    │                   │ Leaves JWT intact│                   │ Valid! ✅  │
│ Gateway      │                   │ Modifies Body!   │                   │ Settles   │
└──────────────┘                   └──────────────────┘                   │ $50,000! ❌│
                                                                          └───────────┘
```

### 1.2 The Two Pillars of API Defense: Identity + Integrity

To prevent this existential failure, enterprise payment gateways implement **two complementary security layers**:

```
                       THE TWO PILLARS OF API DEFENSE
                       
           PILLAR 1: IDENTITY                     PILLAR 2: INTEGRITY
              (Built Day 2)                          (Built Day 3)
         ┌─────────────────────┐                ┌─────────────────────┐
         │     Bearer JWT      │                │     HMAC-SHA256     │
         ├─────────────────────┤                ├─────────────────────┤
         │ Question:           │                │ Question:           │
         │ "Who is sending?"   │                │ "Was this altered?" │
         ├─────────────────────┤                ├─────────────────────┤
         │ Covers:             │                │ Covers:             │
         │ Merchant ID, Role   │                │ Exact Request Body, │
         │ Expiration Date     │                │ Timestamp           │
         └─────────────────────┘                └─────────────────────┘
```

Both checks run on every single payment request. Both must pass.

### 1.3 How Industry Leaders Implement Request Signing
* **Stripe Webhooks:** Stripe includes a `Stripe-Signature` header on every outgoing event payload containing `t=<timestamp>,v1=<hmac_sha256>`. Your server recomputes the signature before processing charges.
* **AWS Signature Version 4 (SigV4):** Every call to AWS S3, DynamoDB, or SQS requires an `Authorization: AWS4-HMAC-SHA256 Credential=..., SignedHeaders=..., Signature=...` header.
* **GitHub Webhooks:** GitHub sends `X-Hub-Signature-256` computed over the JSON webhook payload using a shared repository secret.

---

## 5. Chapter 2: The Mathematical Architecture of HMAC-SHA256 (RFC 2104)

### 2.1 The Mathematics of HMAC

HMAC is not simply `Hash(Key + Message)`. It is a mathematically proven construction designed to prevent **length-extension attacks**.

Given:
* $H$: A cryptographic hash function (SHA-256).
* $K$: The shared secret key.
* $M$: The message to be signed (our canonical payload).
* $B$: The block size of the hash function (64 bytes for SHA-256).
* $ipad$: The byte `0x36` repeated $B$ times (inner padding).
* $opad$: The byte `0x5c` repeated $B$ times (outer padding).

$$\text{HMAC}(K, M) = H\Big((K \oplus opad) \parallel H\big((K \oplus ipad) \parallel M\big)\Big)$$

```
                               HMAC MATHEMATICAL PIPELINE
                               
                      Secret Key (K)
                      ┌─────┴─────┐
                      │           │
                 XOR with      XOR with
                   ipad          opad
                  (0x36)        (0x5c)
                      │           │
                      ▼           ▼
                   [ K_in ]    [ K_out ]
                      │           │
    Message (M) ──────┤           │
                      ▼           │
            SHA-256( K_in || M )  │
                      │           │
                      ▼           │
                Inner Hash ───────┤
                                  ▼
                        SHA-256( K_out || Inner Hash )
                                  │
                                  ▼
                         Final HMAC Signature (32 bytes)
```

1. **Step 1:** Key $K$ is XORed with the inner padding constant (`0x36`) to produce $K_{in}$.
2. **Step 2:** $K_{in}$ is prepended to message $M$, and hashed with SHA-256. This is the **Inner Digest**.
3. **Step 3:** Key $K$ is XORed with the outer padding constant (`0x5c`) to produce $K_{out}$.
4. **Step 4:** $K_{out}$ is prepended to the Inner Digest, and hashed again. This produces the **Final HMAC**.

### 2.2 Why Naive Hashing $H(K \parallel M)$ Fails: Length Extension Attacks

Beginners often think: *"Why not just do `SHA256(Secret + Body)`?"*
* **The Catastrophic Flaw:** Merkle-Damgård hash functions (like SHA-1, SHA-256, and MD5) process data in sequential blocks. The internal state of the hash after processing `Secret + Body` is identical to the final output hash!
* **The Exploit:** An attacker who sees `Message` and `SHA256(Secret + Message)` can append `; transfer $1,000,000 to attacker` to the message and calculate the new valid hash **without ever knowing what the Secret was!**
* **The Solution:** HMAC's nested two-pass architecture ($K_{out}$ wrapping $K_{in}$) mathematically eliminates length extension attacks.

---

## 6. Chapter 3: Critical Engineering Obstacles & Defense Science

Implementing HMAC in production requires solving several subtle engineering traps that don't appear in basic tutorials.

### 3.1 The Canonicalization Problem

HMAC operates on **raw binary bytes**, not semantic JSON meaning.
Consider these two JSON bodies:

```json
// Payload A:
{"amount_cents":5000,"recipient":"merchant_456"}

// Payload B:
{
  "amount_cents": 5000,
  "recipient": "merchant_456"
}
```

* To a human, Payload A and Payload B are identical.
* To `json.Unmarshal`, both produce the exact same Go struct.
* **To SHA-256, they produce completely different hashes!** Payload B has spaces, line breaks, and indentation bytes that Payload A does not have.

**The Golden Rule of Payment Request Signing:**
> **The server MUST read and sign the exact, raw, untouched byte slice received from the network socket BEFORE any JSON parser touches it.**

In our gateway:
```go
// Correct: Read exact wire bytes directly
bodyBytes, _ := io.ReadAll(r.Body)

// Construct canonical string using raw bytes:
canonical := BuildCanonicalPayload(timestamp, bodyBytes)
```

---

### 3.2 The Stream Draining Trap in Go HTTP Servers

In Go, `r.Body` is an `io.ReadCloser`. When you call `io.ReadAll(r.Body)`:
1. The bytes are read from the socket buffer into memory.
2. The internal stream pointer reaches **EOF (End of File)**.
3. If you do not restore `r.Body`, when the next middleware or handler calls `json.NewDecoder(r.Body).Decode(&req)`:
   * It attempts to read from `r.Body`.
   * It immediately reads 0 bytes and hits EOF!
   * The handler throws an error: `unexpected end of JSON input`.

**The Restoration Idiom:**
```go
// 1. Read the bytes
bodyBytes, err := io.ReadAll(r.Body)

// 2. Immediately replace r.Body with an io.NopCloser wrapping a fresh bytes.Reader
r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
```
Now, downstream handlers can read the body as if the middleware had never touched it!

---

### 3.3 The Side-Channel Timing Attack: Why `==` is Fatal

Consider standard string equality in Go or C:
```go
// ❌ CRITICAL SECURITY VULNERABILITY (DO NOT WRITE THIS)
if expectedSig == providedSig {
    // Valid!
}
```

**How the CPU executes `==`:**
```
Index:    0    1    2    3    4  ...
Expected: a    f    8    b    2  ...
Provided: a    f    9    0    0  ...
Match?   YES  YES   NO! ──► Returns FALSE immediately (short-circuit)
```
* If character 0 does not match: returns in **2 nanoseconds**.
* If characters 0 and 1 match: returns in **4 nanoseconds**.
* If characters 0, 1, and 2 match: returns in **6 nanoseconds**.

An attacker sitting on low-latency infrastructure sends millions of requests, measures the response latency distribution, and discovers the signature character by character!

**The Defense (`hmac.Equal`):**
```go
// ✅ CRYPTOGRAPHICALLY SECURE CONSTANT-TIME EQUALITY
if !hmac.Equal([]byte(expectedSig), []byte(providedSig)) {
    return false, errors.New("invalid signature")
}
```
`hmac.Equal` compares all bytes using bitwise OR operations without branching, taking the exact same execution time whether the match fails on byte 0 or byte 63.

---

### 3.4 Replay Attacks: The $\pm 300$-Second Timestamp Window

What happens if an attacker captures a valid packet with a valid JWT and a valid HMAC signature?
* They cannot modify the body (HMAC catches it).
* They cannot modify the token (JWT catches it).
* **BUT they can resend the exact same packet 100 times!**

**The Defense:**
We include the client timestamp in the canonical string:
$$\text{Canonical} = \texttt{timestamp} + \texttt{"."} + \texttt{rawBody}$$

When the gateway receives a request:
```go
now := time.Now().Unix()
age := now - timestamp
if age > 300 || age < -300 {
    // Reject immediately: stale or clock-skewed request!
    return false, errors.New("timestamp outside allowed window")
}
```
* If an attacker replays the packet 10 minutes later, the gateway rejects it because the timestamp is stale.
* If the attacker tries to change the timestamp header to `now`, the HMAC signature fails because the timestamp was sealed inside the canonical hash!

---

## 7. Chapter 4: Defense in Depth & Middleware Stacking Architecture

### 4.1 The Ensemble Security Model

In NOVARIS, security is designed like a bank vault with multiple sequential doors:

```
                          INCOMING PAYMENT REQUEST
                                     │
                                     ▼
        ┌─────────────────────────────────────────────────────────┐
        │             LAYER 1: JWT AUTHENTICATION                 │
        │                                                         │
        │  * Checks Authorization: Bearer <token>                 │
        │  * Validates HMAC-SHA256 signature                      │
        │  * Validates expiration (exp) timestamp                 │
        │  * Injects Claims (merchant_id, role) into Context      │
        │                                                         │
        │  ❌ Fails? ──► 401 Unauthorized ("missing/invalid token")│
        └────────────────────────────┬────────────────────────────┘
                                     │ Passes
                                     ▼
        ┌─────────────────────────────────────────────────────────┐
        │              LAYER 2: HMAC REQUEST SIGNING              │
        │                                                         │
        │  * Checks X-Novaris-Timestamp within ±300s window       │
        │  * Reads raw body bytes from socket                     │
        │  * Restores r.Body via io.NopCloser                     │
        │  * Recomputes HMAC-SHA256 over canonical string         │
        │  * Compares using constant-time hmac.Equal              │
        │                                                         │
        │  ❌ Fails? ──► 401 Unauthorized ("invalid signature")   │
        └────────────────────────────┬────────────────────────────┘
                                     │ Passes
                                     ▼
        ┌─────────────────────────────────────────────────────────┐
        │              LAYER 3: BUSINESS LOGIC                    │
        │                                                         │
        │  * paymentsHandler reads verified Claims & intact Body  │
        │  * Passes settlement event to PostgreSQL Ledger         │
        │  * Emits event to Kafka settlement topic                │
        │  * Returns HTTP 200 OK                                  │
        └─────────────────────────────────────────────────────────┘
```

### 4.2 Middleware Stacking Order: Why JWT Runs Before HMAC

Notice the composition in `main.go`:
```go
mux.HandleFunc("/v1/payments", auth.JWTMiddleware(auth.HMACMiddleware(paymentsHandler)))
```

**Why JWT (Outer) wraps HMAC (Inner):**
1. **Computational Efficiency:** Checking a JWT header is computationally cheap (verifying a small 200-byte token string).
2. **Fail-Fast Defense:** If an unauthenticated bot floods the gateway, `JWTMiddleware` rejects the request immediately. The gateway **never wastes CPU reading the request body or running HMAC hashes**.

---

## 8. Chapter 5: The Data Science → Full-Stack Integrity Rosetta Stone

If you come from Data Science, Machine Learning, or Python, you already understand cryptographic integrity. Here is the direct translation:

| Data Science / Python Concept | Go Gateway Equivalent | Why They Are the Same Fundamental Idea |
| :--- | :--- | :--- |
| `sha256sum dataset.csv` | `auth.ComputeHMACSignature(body, ts, secret)` | Both compute a cryptographic fingerprint to confirm data has not been corrupted or tampered with. |
| Git commit SHA (`git log`) | HMAC signature of raw request body | A Git commit hash is a SHA-1/SHA-256 digest of tree bytes. If you change a single comma, the commit SHA completely changes (avalanche effect). |
| `np.allclose(a, b)` vs. `a == b` | `hmac.Equal(expected, provided)` vs. `==` | Both replace naive comparison (`==`) with specialized functions to avoid subtle bugs (floating point rounding vs. side-channel timing attacks). |
| AWS `boto3` signing S3 requests | `cmd/signrequest` helper utility | Every time you call `boto3.client('s3').upload_file()`, boto3 builds a canonical string with timestamps and computes an HMAC-SHA256 signature in headers under the hood. |
| Python generator exhaustion | Go `io.ReadAll(r.Body)` stream draining | Once you loop through a Python generator (`list(gen)`), it is empty. In Go, once you read `r.Body`, it reaches EOF and must be re-instantiated. |

---

## 9. Chapter 6: Architectural Trade-Off Analysis ("Why This and Not That")

### 6.1 HMAC-SHA256 vs. RSA-SHA256 (Asymmetric Digital Signatures)

* **Symmetric HMAC:**
  * *Pros:* **10x to 50x faster** than RSA. Minimal CPU footprint at 10,000 requests per second.
  * *Cons:* Both client and server must store the shared secret.
* **Asymmetric RSA/ECDSA:**
  * *Pros:* Client keeps private key; server only holds public key.
  * *Cons:* Slower mathematical verification (modular exponentiation / elliptic curve operations).
* **NOVARIS Decision:** We use **HMAC-SHA256** for Day 3. Each merchant is provisioned an API secret key during onboarding (the standard Stripe and AWS model), delivering sub-millisecond gateway processing.

### 6.2 Timestamp Replay Window: Why Exactly 300 Seconds?

* If window is **too long** (e.g., 24 hours): An intercepted request can be replayed by an attacker for an entire day!
* If window is **too short** (e.g., 5 seconds): Legitimate requests will be rejected due to normal cloud network latency or slight **clock drift (NTP skew)** between client and server.
* **FinTech Standard:** $\pm 300$ seconds ($\pm 5$ minutes) strikes the perfect operational balance, completely neutralizing persistent replay attacks while accommodating distributed server clock variations.

---

## 10. Chapter 7: Line-by-Line Code Masterclass (Complete Dissections)

### 7.1 `services/gateway/internal/auth/hmac.go`

Here is the complete cryptographic implementation of our HMAC engine:

```go
package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	HeaderNovarisSignature = "X-Novaris-Signature"
	HeaderNovarisTimestamp = "X-Novaris-Timestamp"
	ReplayWindowSeconds    int64 = 300
	DefaultHMACSecret      = "novaris_insecure_dev_hmac_secret_change_in_production_2026"
)

func GetHMACSecret() []byte {
	secret := os.Getenv("HMAC_SECRET")
	if secret == "" {
		return []byte(DefaultHMACSecret)
	}
	return []byte(secret)
}

func BuildCanonicalPayload(timestamp int64, rawBody []byte) []byte {
	tsStr := strconv.FormatInt(timestamp, 10)
	buf := make([]byte, 0, len(tsStr)+1+len(rawBody))
	buf = append(buf, tsStr...)
	buf = append(buf, '.')
	buf = append(buf, rawBody...)
	return buf
}

func ComputeHMACSignature(rawBody []byte, timestamp int64, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	canonical := BuildCanonicalPayload(timestamp, rawBody)
	mac.Write(canonical)
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyHMACSignature(rawBody []byte, timestampStr string, providedSig string, secret []byte) (bool, error) {
	if timestampStr == "" {
		return false, errors.New("missing timestamp header")
	}
	if providedSig == "" {
		return false, errors.New("missing signature header")
	}

	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Replay defense
	now := time.Now().Unix()
	age := now - timestamp
	if age > ReplayWindowSeconds || age < -ReplayWindowSeconds {
		return false, fmt.Errorf("request timestamp outside allowed window (age: %d seconds, allowed: ±%d)", age, ReplayWindowSeconds)
	}

	expectedSig := ComputeHMACSignature(rawBody, timestamp, secret)

	// Constant-time comparison defense against side-channel timing attacks
	if !hmac.Equal([]byte(expectedSig), []byte(providedSig)) {
		return false, errors.New("invalid signature: payload hash mismatch")
	}

	return true, nil
}

func HMACMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sig := r.Header.Get(HeaderNovarisSignature)
		tsStr := r.Header.Get(HeaderNovarisTimestamp)

		if sig == "" || tsStr == "" {
			respondHMACUnauthorized(w, "missing required signature headers (X-Novaris-Signature, X-Novaris-Timestamp)")
			return
		}

		// Cheap check first before reading body stream
		timestamp, err := strconv.ParseInt(tsStr, 10, 64)
		if err != nil {
			respondHMACUnauthorized(w, "invalid X-Novaris-Timestamp format: must be integer UNIX epoch")
			return
		}

		now := time.Now().Unix()
		age := now - timestamp
		if age > ReplayWindowSeconds || age < -ReplayWindowSeconds {
			respondHMACUnauthorized(w, fmt.Sprintf("request timestamp outside allowed window (age: %ds, max: ±%ds)", age, ReplayWindowSeconds))
			return
		}

		// Canonicalization: Read raw body bytes
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"bad_request","message":"unable to read request body: %v"}`+"\n", err)
			return
		}

		// Stream restoration for downstream handlers
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		secret := GetHMACSecret()
		valid, err := VerifyHMACSignature(bodyBytes, tsStr, sig, secret)
		if err != nil || !valid {
			respondHMACUnauthorized(w, fmt.Sprintf("HMAC signature verification failed: %v", err))
			return
		}

		next.ServeHTTP(w, r)
	}
}

func respondHMACUnauthorized(w http.ResponseWriter, reason string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	fmt.Fprintf(w, `{"error":"unauthorized","auth_type":"hmac_sha256","message":%q}`+"\n", reason)
}
```

#### Detailed Dissection of `hmac.go`:
1. **Lines 28–39 (Headers & Configuration):** Declares `X-Novaris-Signature` and `X-Novaris-Timestamp`, defining a standard $\pm 300$-second replay window and dev fallback secret.
2. **Lines 50–57 (`BuildCanonicalPayload`):** Converts the integer timestamp to a string and appends a single dot `.` followed by the raw body bytes. This exact recipe binds the timestamp and body together.
3. **Lines 61–66 (`ComputeHMACSignature`):** Hashes the canonical bytes using `hmac.New(sha256.New, secret)` and hex-encodes the resulting 32-byte hash.
4. **Lines 70–103 (`VerifyHMACSignature`):**
   - Parses the integer timestamp and executes the replay window freshness check.
   - Recomputes the expected signature over the raw body.
   - Crucially, performs `hmac.Equal([]byte(expectedSig), []byte(providedSig))` to prevent side-channel timing attacks.
5. **Lines 108–160 (`HMACMiddleware`):**
   - Checks presence of headers.
   - Fails fast on invalid timestamps before touching the body stream.
   - Reads `r.Body` using `io.ReadAll`.
   - **Restores `r.Body` using `io.NopCloser(bytes.NewReader(bodyBytes))`** so downstream business handlers can still decode the payload!
   - Invokes `VerifyHMACSignature` and passes execution to `next.ServeHTTP`.

---

### 7.2 `services/gateway/cmd/signrequest/main.go`

A developer CLI utility to mint valid, custom, or deliberately tampered/stale HMAC signatures for local testing:

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
)

func main() {
	bodyFlag := flag.String("body", `{"amount_cents":5000,"recipient":"merchant_456"}`, "Raw JSON request body to sign")
	offsetFlag := flag.Int64("offset", 0, "Timestamp offset in seconds (e.g. -600 for a 10m stale request)")
	secretFlag := flag.String("secret", "", "Custom HMAC secret (defaults to HMAC_SECRET env or dev fallback)")
	tamperFlag := flag.Bool("tamper", false, "Simulate tampering by altering the body after computing signature")
	flag.Parse()

	// Determine secret key
	secret := auth.GetHMACSecret()
	if *secretFlag != "" {
		secret = []byte(*secretFlag)
	}

	// Calculate timestamp with requested offset
	targetTime := time.Now().Unix() + *offsetFlag
	timestampStr := strconv.FormatInt(targetTime, 10)

	// Compute HMAC signature on original body
	originalBodyBytes := []byte(*bodyFlag)
	signature := auth.ComputeHMACSignature(originalBodyBytes, targetTime, secret)

	outputBody := *bodyFlag
	if *tamperFlag {
		outputBody = `{"amount_cents":5000000,"recipient":"attacker_tampered_payload"}`
	}

	fmt.Println("==============================================================================")
	fmt.Println("NOVARIS Developer HMAC-SHA256 Request Signing Utility")
	fmt.Println("==============================================================================")
	fmt.Printf("Raw Body:    %s\n", outputBody)
	fmt.Printf("Timestamp:   %s\n", timestampStr)
	fmt.Printf("Signature:   %s\n", signature)
	if *tamperFlag {
		fmt.Println("[WARNING] -tamper active: signature was generated for original body, not tampered body!")
	}
	if *offsetFlag != 0 {
		fmt.Printf("[INFO] -offset active: timestamp is %d seconds offset from current server time.\n", *offsetFlag)
	}
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Println("Example Headers:")
	fmt.Printf("  X-Novaris-Timestamp: %s\n", timestampStr)
	fmt.Printf("  X-Novaris-Signature: %s\n", signature)
	fmt.Println("------------------------------------------------------------------------------")
	fmt.Println("Ready-to-Run curl command (assuming $TOKEN contains valid JWT):")
	fmt.Printf("curl -i -X POST http://localhost:8080/v1/payments \\\n"+
		"  -H \"Authorization: Bearer $TOKEN\" \\\n"+
		"  -H \"X-Novaris-Timestamp: %s\" \\\n"+
		"  -H \"X-Novaris-Signature: %s\" \\\n"+
		"  -H \"Content-Type: application/json\" \\\n"+
		"  -d '%s'\n", timestampStr, signature, outputBody)
	fmt.Println("==============================================================================")

	if os.Getenv("NOVARIS_CLI_RAW") == "1" {
		fmt.Printf("%s|%s|%s\n", timestampStr, signature, outputBody)
	}
}
```

#### Detailed Dissection of `cmd/signrequest/main.go`:
* **Lines 18–22 (`flag` parsing):** Allows dynamic generation of valid signatures for any JSON body (`-body`), intentional replay testing (`-offset -600`), and simulated in-flight tampering (`-tamper`).
* **Lines 31–36 (Timestamp Arithmetic & Signature Computation):** Applies the offset to `time.Now().Unix()` and computes the HMAC signature over the exact un-tampered body.
* **Lines 38–41 (Tamper Simulation):** If `-tamper` is passed, alters the body string *after* the signature was computed, outputting a mismatched payload to test gateway rejection.

---

### 7.3 `services/gateway/main.go`

Our API Gateway router, now composed with the stacked defense pipeline:

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charan21042005/novaris/services/gateway/internal/auth"
)

// PaymentRequest defines the expected payload for initiating a settlement transaction.
type PaymentRequest struct {
	AmountCents int64  `json:"amount_cents"`
	Recipient   string `json:"recipient"`
	Currency    string `json:"currency,omitempty"`
}

// PaymentResponse is returned by /v1/payments upon successful multi-layer verification.
type PaymentResponse struct {
	Status      string    `json:"status"`
	MerchantID  string    `json:"merchant_id"`
	Role        string    `json:"role"`
	AmountCents int64     `json:"amount_cents,omitempty"`
	Recipient   string    `json:"recipient,omitempty"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

// paymentsHandler is the protected business settlement endpoint.
// It retrieves verified JWT claims from the context and decodes the intact body.
func paymentsHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.GetClaimsFromContext(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, `{"error":"unauthorized","message":"missing context claims"}`)
		return
	}

	// Read and decode the restored request body (verified intact by auth.HMACMiddleware)
	var req PaymentRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	payload := PaymentResponse{
		Status:      "authorized",
		MerchantID:  claims.MerchantID,
		Role:        claims.Role,
		AmountCents: req.AmountCents,
		Recipient:   req.Recipient,
		Message:     "Payment request verified successfully: JWT identity confirmed and HMAC body integrity preserved.",
		Timestamp:   time.Now().UTC(),
	}

	_ = json.NewEncoder(w).Encode(payload)
}

func main() {
	mux := http.NewServeMux()

	// 1. Public Routes
	mux.HandleFunc("/healthz", healthzHandler(cfg))
	mux.HandleFunc("/", rootHandler(cfg))

	// 2. Protected Routes (Stacked Middleware: JWTMiddleware -> HMACMiddleware -> paymentsHandler)
	mux.HandleFunc("/v1/payments", auth.JWTMiddleware(auth.HMACMiddleware(paymentsHandler)))

	// Server lifecycle & graceful termination...
}
```

#### Detailed Dissection of `main.go`:
1. **Lines 44–69 (`paymentsHandler`):** Notice that `paymentsHandler` does not touch HMAC math or token validation. It trusts that if execution reached this line, the merchant is authentic and the body was not tampered with. It decodes `r.Body` normally because `HMACMiddleware` restored the stream via `io.NopCloser`.
2. **Line 81 (`auth.JWTMiddleware(auth.HMACMiddleware(paymentsHandler))`):** Demonstrates functional middleware composition in Go standard library. Request flow travels outside-in: `JWTMiddleware` $\rightarrow$ `HMACMiddleware` $\rightarrow$ `paymentsHandler`.

---

### 7.4 `services/gateway/internal/auth/hmac_test.go`

Our automated cryptographic unit test suite:

```go
package auth

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestComputeAndVerifyHMAC_Success(t *testing.T) {
	secret := []byte("test_secret_key_12345")
	body := []byte(`{"amount_cents":5000,"recipient":"merchant_456"}`)
	now := time.Now().Unix()

	sig := ComputeHMACSignature(body, now, secret)
	valid, err := VerifyHMACSignature(body, strconv.FormatInt(now, 10), sig, secret)
	if err != nil || !valid {
		t.Fatalf("expected verification to succeed: %v", err)
	}
}

func TestVerifyHMAC_TamperedBody(t *testing.T) {
	secret := []byte("test_secret_key_12345")
	originalBody := []byte(`{"amount_cents":5000,"recipient":"merchant_456"}`)
	tamperedBody := []byte(`{"amount_cents":5000000,"recipient":"attacker"}`)
	now := time.Now().Unix()

	sig := ComputeHMACSignature(originalBody, now, secret)
	valid, err := VerifyHMACSignature(tamperedBody, strconv.FormatInt(now, 10), sig, secret)
	if err == nil || valid {
		t.Fatal("expected validation to fail for tampered body")
	}
}

func TestVerifyHMAC_StaleTimestamp(t *testing.T) {
	secret := []byte("test_secret_key_12345")
	body := []byte(`{"amount_cents":100}`)
	staleTime := time.Now().Unix() - 600 // 10 minutes ago

	sig := ComputeHMACSignature(body, staleTime, secret)
	valid, err := VerifyHMACSignature(body, strconv.FormatInt(staleTime, 10), sig, secret)
	if err == nil || valid {
		t.Fatal("expected validation to fail for stale timestamp")
	}
}
```

---

### 7.5 `services/gateway/README.md`

We updated the gateway service documentation with Day 3 endpoint contracts, curl verification examples, and developer tooling instructions.

---

## 11. Chapter 8: Verification Laboratory: Execution Proof & Log Dissection

### 8.1 Automated Unit Test Suite Output

We executed our test suite across all gateway packages:
```powershell
PS C:\novaris\services\gateway> go test -v ./...
```

**Real Console Execution Output:**
```
=== RUN   TestComputeAndVerifyHMAC_Success
--- PASS: TestComputeAndVerifyHMAC_Success (0.00s)
=== RUN   TestVerifyHMAC_TamperedBody
--- PASS: TestVerifyHMAC_TamperedBody (0.00s)
=== RUN   TestVerifyHMAC_StaleTimestamp
--- PASS: TestVerifyHMAC_StaleTimestamp (0.00s)
=== RUN   TestVerifyHMAC_FutureTimestamp
--- PASS: TestVerifyHMAC_FutureTimestamp (0.00s)
=== RUN   TestVerifyHMAC_WrongSecret
--- PASS: TestVerifyHMAC_WrongSecret (0.00s)
=== RUN   TestHMACMiddleware
--- PASS: TestHMACMiddleware (0.00s)
=== RUN   TestGenerateAndValidateToken
--- PASS: TestGenerateAndValidateToken (0.00s)
=== RUN   TestExpiredToken
--- PASS: TestExpiredToken (0.00s)
=== RUN   TestTamperedToken
--- PASS: TestTamperedToken (0.00s)
=== RUN   TestJWTMiddleware
--- PASS: TestJWTMiddleware (0.00s)
PASS
ok      github.com/charan21042005/novaris/services/gateway/internal/auth        1.938s
```
*10 out of 10 automated unit tests passing cleanly.*

---

### 8.2 Real-World HTTP Verification Matrix: 7 Concrete Probe Scenarios

Using `test_probe_day3.ps1`, we booted `gateway.exe` and verified 7 distinct security scenarios:

```
┌──────┬────────────────────────┬────────────────────────────────────────────────────────┬──────────────┬────────┐
│ Case │ Endpoint               │ Authentication & Integrity Condition                   │ Expected HTTP│ Result │
├──────┼────────────────────────┼────────────────────────────────────────────────────────┼──────────────┼────────┤
│  1   │ GET /healthz           │ No authentication headers (Public route)               │ 200 OK       │  PASS  │
│  2   │ POST /v1/payments      │ Missing All Auth Headers (No JWT, No HMAC)             │ 401 Unauth   │  PASS  │
│  3   │ POST /v1/payments      │ Valid JWT, but Missing HMAC Headers                    │ 401 Unauth   │  PASS  │
│  4   │ POST /v1/payments      │ Valid JWT + Valid HMAC + Valid Timestamp               │ 200 OK       │  PASS  │
│  5   │ POST /v1/payments      │ Valid JWT + Original HMAC + Tampered Body ($50k)       │ 401 Unauth   │  PASS  │
│  6   │ POST /v1/payments      │ Valid JWT + Valid HMAC + Stale Timestamp (-10m)        │ 401 Unauth   │  PASS  │
│  7   │ POST /v1/payments      │ Valid JWT + Valid HMAC + Future Timestamp (+10m)       │ 401 Unauth   │  PASS  │
└──────┴────────────────────────┴────────────────────────────────────────────────────────┴──────────────┴────────┘
```

#### Live Console Verification Log:
```
=== LIVE PROBE 1: Public /healthz ===
Status: ok Service: novaris-gateway

=== GENERATING TEST CREDENTIALS ===
Token: eyJhbGciOiJIUzI1NiIsInR5cCI6Ik ...
Timestamp: 1795417328
Signature: dc9771ca0151c046cf9ebefbd69ecb37f7a9c0af1ab9304886145056876e842b

=== LIVE PROBE 2: Missing All Auth Headers ===
Expected 401 Caught: 401

=== LIVE PROBE 3: JWT Only, Missing HMAC Headers ===
Expected 401 Caught: 401

=== LIVE PROBE 4: Fully Valid JWT + HMAC Signed Request ===
Response 200 OK:
{
  "status": "authorized",
  "merchant_id": "merchant_stripe_live",
  "role": "merchant",
  "amount_cents": 5000,
  "recipient": "merchant_456",
  "message": "Payment request verified successfully: JWT identity confirmed and HMAC body integrity preserved.",
  "timestamp": "2026-11-23T07:02:08.4554153Z"
}

=== LIVE PROBE 5: Tampered Body ($50.00 changed to $50,000.00) ===
Expected 401 Tampered Body Caught: 401

=== LIVE PROBE 6: Stale Replay Timestamp (10m in the past) ===
Expected 401 Stale Timestamp Caught: 401

=== LIVE PROBE 7: Future Timestamp (+10m ahead) ===
Expected 401 Future Timestamp Caught: 401
```

---

### 8.3 Complete Step-by-Step CLI Command Runbook (Bash & PowerShell)

Here is the exact copy-pasteable command runbook to verify all Day 3 HMAC request signing, canonicalization, and replay defense in any terminal:

#### 1. Run Automated Unit Tests (All 10 tests across JWT & HMAC):
```bash
cd services/gateway
go test -v ./...
```

#### 2. Start the Gateway Service (Terminal 1):
```bash
cd services/gateway
go run main.go
```

#### 3. Mint Test Credentials via Developer CLIs (Terminal 2):
```bash
# Generate a valid Bearer JWT
TOKEN=$(go run cmd/gentoken/main.go -merchant merchant_stripe_live -role merchant -mins 15)

# Generate valid HMAC signature and timestamp for standard payload
# (Or run cmd/signrequest/main.go to view the interactive output)
BODY='{"amount_cents":5000,"recipient":"merchant_456"}'
TS=$(date +%s)
SECRET="novaris_insecure_dev_hmac_secret_change_in_production_2026"
CANONICAL="$TS.$BODY"
SIG=$(echo -n "$CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')
```
*PowerShell Alternative:*
```powershell
$token = (go run cmd/gentoken/main.go -merchant merchant_stripe_live -role merchant -mins 15)
$body = '{"amount_cents":5000,"recipient":"merchant_456"}'
$ts = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds().ToString()
$canonical = $ts + '.' + $body
$secret = "novaris_insecure_dev_hmac_secret_change_in_production_2026"
$hmacObj = New-Object System.Security.Cryptography.HMACSHA256
$hmacObj.Key = [System.Text.Encoding]::UTF8.GetBytes($secret)
$sig = [BitConverter]::ToString($hmacObj.ComputeHash([System.Text.Encoding]::UTF8.GetBytes($canonical))).Replace('-','').ToLower()
```

#### 4. Execute the 7 Verification Probes via `curl`:
```bash
# Probe 1: Public Health Probe -> Expect 200 OK
curl -i http://localhost:8080/healthz

# Probe 2: Missing All Auth Headers -> Expect 401 Unauthorized
curl -i -X POST http://localhost:8080/v1/payments \
  -H "Content-Type: application/json" \
  -d "$BODY"

# Probe 3: JWT Only, Missing HMAC Headers -> Expect 401 Unauthorized
curl -i -X POST http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "$BODY"

# Probe 4: Fully Valid JWT + HMAC Signed Request -> Expect 200 OK with restored body
curl -i -X POST http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Novaris-Timestamp: $TS" \
  -H "X-Novaris-Signature: $SIG" \
  -H "Content-Type: application/json" \
  -d "$BODY"

# Probe 5: Tampered Request Body ($50 changed to $50,000) -> Expect 401 Unauthorized
curl -i -X POST http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Novaris-Timestamp: $TS" \
  -H "X-Novaris-Signature: $SIG" \
  -H "Content-Type: application/json" \
  -d '{"amount_cents":5000000,"recipient":"attacker"}'

# Probe 6: Stale Replay Timestamp (10m in past) -> Expect 401 Unauthorized ("outside allowed window")
STALE_TS=$(($TS - 600))
STALE_CANONICAL="$STALE_TS.$BODY"
STALE_SIG=$(echo -n "$STALE_CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')
curl -i -X POST http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Novaris-Timestamp: $STALE_TS" \
  -H "X-Novaris-Signature: $STALE_SIG" \
  -H "Content-Type: application/json" \
  -d "$BODY"

# Probe 7: Future Timestamp (+10m ahead) -> Expect 401 Unauthorized
FUTURE_TS=$(($TS + 600))
FUTURE_CANONICAL="$FUTURE_TS.$BODY"
FUTURE_SIG=$(echo -n "$FUTURE_CANONICAL" | openssl dgst -sha256 -hmac "$SECRET" | sed 's/^.* //')
curl -i -X POST http://localhost:8080/v1/payments \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Novaris-Timestamp: $FUTURE_TS" \
  -H "X-Novaris-Signature: $FUTURE_SIG" \
  -H "Content-Type: application/json" \
  -d "$BODY"
```

#### 5. Execute the 7 Verification Probes via Windows PowerShell:
```powershell
$headersValid = @{
    "Authorization"       = "Bearer " + $token
    "X-Novaris-Timestamp" = $ts
    "X-Novaris-Signature" = $sig
}

# Probe 1: Public Health -> Expect 200 OK
Invoke-RestMethod -Uri "http://localhost:8080/healthz"

# Probe 2: Missing All Headers -> Expect 401
try { Invoke-RestMethod -Uri "http://localhost:8080/v1/payments" -Method Post -Body $body -ContentType "application/json" } catch { $_.Exception.Response.StatusCode.value__ }

# Probe 3: JWT Only -> Expect 401
try { Invoke-RestMethod -Uri "http://localhost:8080/v1/payments" -Method Post -Body $body -ContentType "application/json" -Headers @{ Authorization = "Bearer " + $token } } catch { $_.Exception.Response.StatusCode.value__ }

# Probe 4: Fully Valid Signed Request -> Expect 200 OK
Invoke-RestMethod -Uri "http://localhost:8080/v1/payments" -Method Post -Body $body -ContentType "application/json" -Headers $headersValid

# Probe 5: Tampered Body -> Expect 401
try { Invoke-RestMethod -Uri "http://localhost:8080/v1/payments" -Method Post -Body '{"amount_cents":5000000,"recipient":"attacker"}' -ContentType "application/json" -Headers $headersValid } catch { $_.Exception.Response.StatusCode.value__ }
```

---

## 12. Chapter 9: The Interview Defense Compendium (10 High-Stakes Tier-1 Q&As)

### Q1: "Why did you implement HMAC request signing if you already had JWT authentication in place?"
> **Model Answer:** "JWT and HMAC solve two completely distinct security problems. JWT establishes Identity: it proves *who* is calling by verifying the bearer token. However, a JWT says nothing about the HTTP request payload. If an adversary intercepts a valid JWT packet on the wire, they could alter the payment amount from \$50 to \$50,000, and the JWT would still pass verification! HMAC-SHA256 establishes Integrity and Authenticity: it proves that the specific request body was not altered by a single byte since the merchant signed it."

### Q2: "What is a canonical string, and why can't we simply deserialize JSON to a struct before signing?"
> **Model Answer:** "A cryptographic hash operates strictly on raw bytes, not semantic meaning. If the client signs `{'a':1,'b':2}` and the server deserializes it and re-serializes it as `{'b':2,'a':1}` or adds indentation spaces, the byte sequence changes and the signature fails even though the JSON data is logically identical. This is the canonicalization problem. In NOVARIS, our middleware signs the raw, untouched byte slice directly from the network stream before any JSON parser touches it."

### Q3: "What is the stream draining trap in Go HTTP handlers, and how does your code resolve it?"
> **Model Answer:** "In Go, `r.Body` is an `io.ReadCloser` backed by a network socket. Calling `io.ReadAll(r.Body)` consumes the stream to EOF. If the middleware does not restore the stream, subsequent handlers will receive an empty body when attempting to parse JSON. We resolve this by capturing `bodyBytes` and immediately restoring `r.Body = io.NopCloser(bytes.NewReader(bodyBytes))`."

### Q4: "Why is using `==` to compare cryptographic signatures an exploitable security vulnerability?"
> **Model Answer:** "Standard string comparison operators (`==`) short-circuit, returning `false` on the first non-matching byte. This leaks nanosecond timing information: a signature that matches for 3 characters takes slightly longer to fail than one that fails on the first byte. An attacker can measure this statistical difference to deduce valid signatures byte by byte (a timing side-channel attack). We use `hmac.Equal`, which executes in constant time regardless of where or whether bytes differ."

### Q5: "How does NOVARIS protect against Replay Attacks without querying a database?"
> **Model Answer:** "We include an `X-Novaris-Timestamp` header bound into the canonical string: `timestamp + '.' + rawBody`. The gateway verifies that the timestamp is within a strict $\pm 300$-second ($\pm 5$ minutes) window against server time. If an attacker intercepts the request and attempts to replay it 10 minutes later, the gateway rejects it as stale. If the attacker modifies the timestamp header to circumvent the check, the HMAC signature breaks because the timestamp is sealed inside the hash."

### Q6: "Why did you choose a $\pm 300$-second replay window instead of 5 seconds or 1 hour?"
> **Model Answer:** "It represents the optimal engineering balance between security and distributed operational tolerance. A 5-second window causes false rejections due to normal internet latency fluctuations and slight NTP clock drift between distributed servers. A 1-hour window leaves a massive vulnerability surface for replay attacks. A 5-minute window is the established industry standard across Stripe, AWS, and PayPal."

### Q7: "What is the difference between HMAC and digital signatures like RSA or ECDSA?"
> **Model Answer:** "HMAC is symmetric: the same shared secret key is used by the merchant to sign and by the gateway to verify. It is extremely fast (~10x to 50x faster than RSA) and has minimal CPU overhead. RSA and ECDSA are asymmetric: signing requires a private key, while verification uses a public key. For server-to-server merchant API ingestion where a shared API secret is provisioned during onboarding, symmetric HMAC is the optimal high-throughput choice."

### Q8: "In your middleware pipeline, why does `JWTMiddleware` run before `HMACMiddleware`?"
> **Model Answer:** "For fail-fast computational efficiency. Verifying a JWT header is an in-memory string check that is computationally cheaper than reading an entire request body from the network stream and computing HMAC-SHA256 hashes. Running JWT first allows the gateway to reject unauthenticated bot traffic immediately without wasting CPU cycles on body buffering or cryptographic hashing."

### Q9: "Why is HMAC immune to length extension attacks that affect plain `SHA256(Secret + Message)`?"
> **Model Answer:** "Merkle-Damgård hash functions process data in sequential blocks where the internal state after hashing `Secret + Message` matches the final output hash. In naive hashing, an attacker can append data and compute a valid hash without knowing the secret. HMAC uses a nested, two-pass construction: $H((K \oplus opad) \parallel H((K \oplus ipad) \parallel M))$. The outer hash encrypts the output of the inner hash with $K_{out}$, mathematically blocking length extension."

### Q10: "Explain in two sentences what JWT and HMAC are each responsible for on `/v1/payments`."
> **Model Answer:** "**JWT authenticates the caller** by verifying their identity and session credentials without a database lookup. **HMAC verifies the message** by cryptographically proving that the specific payment payload arrived completely untampered and within a valid time window."

---

## 13. Chapter 10: The Roadmap Horizon (Day 4 Preview & Hands-On Exercises)

### 13.1 Looking Ahead: Day 4 Preview (Distributed Token-Bucket Rate Limiter)

With **Identity (Day 2)** and **Integrity (Day 3)** secured, our gateway can still be brought down by a legitimate merchant who accidentally triggers an infinite loop or fires 50,000 requests per second.

```
       DAY 2 (JWT)          DAY 3 (HMAC)               DAY 4 (RATE LIMITING)
     "Who are you?"     "Was this altered?"    "Are you sending too many, too fast?"
           │                    │                                │
           ▼                    ▼                                ▼
     Identity Proof       Payload Integrity           Resource Protection & DoS
```

On **Day 4**, we build a **Distributed Sliding Token-Bucket Rate Limiter**:
1. Enforce per-merchant transaction rate limits (e.g., 100 requests per second with burst capacity up to 150).
2. Return standardized HTTP `429 Too Many Requests` with `Retry-After` headers when limits are exceeded.
3. Prepare the Redis key structure for distributed cluster rate-limiting.

---

### 13.2 Hands-On Mastery Exercises for the Reader

1. **Tampering Experiment:**
   * Run `go run cmd/signrequest/main.go` and copy the output curl command.
   * Run the curl command. Observe the `200 OK` response echoing the payment.
   * Now edit the curl command, changing `"amount_cents":5000` to `"amount_cents":5001`. Run it.
   * Observe the immediate transition to `401 Unauthorized` with `payload hash mismatch`!
2. **Replay Experiment:**
   * Run `go run cmd/signrequest/main.go -offset -400`.
   * Execute the resulting curl command.
   * Observe the `401 Unauthorized` error: `request timestamp outside allowed window (age: 400s, max: ±300s)`.
3. **Payload Restoration Experiment:**
   * In `services/gateway/internal/auth/hmac.go`, temporarily comment out the line:
     `r.Body = io.NopCloser(bytes.NewReader(bodyBytes))`
   * Recompile and run. Notice how `paymentsHandler` fails to decode JSON because the body stream was drained!
   * Uncomment the line and observe it working again.

---

<div align="center">

<!-- Floating Bottom Banner -->
<p align="center">
  <img src="../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS Ledger Engineering Curriculum • Day 03 Complete
*Built with mathematical rigor for distributed financial infrastructure.*

</div>
