<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../../../../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 📑 NOVARIS OpenAPI & Swagger UI Engine (`/internal/openapi`)
### Embedded OAS 3.0.3 Specification, Zero-Filesystem Dependency & Interactive API Documentation

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/Language-Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Specification-OpenAPI_3.0.3-6BA539?style=flat-square&logo=openapiinitiative&logoColor=white" alt="OpenAPI" />
  <img src="https://img.shields.io/badge/Embedding-//go:embed_Compiler_Directives-00C853?style=flat-square&logo=webassembly&logoColor=white" alt="Embed" />
  <img src="https://img.shields.io/badge/Interactive-Swagger_UI_v5-85EA2D?style=flat-square&logo=swagger&logoColor=white" alt="Swagger" />
  <img src="https://img.shields.io/badge/Caching-RFC_9110_max--age=3600-7C4DFF?style=flat-square&logo=cloudflare&logoColor=white" alt="Caching" />
</p>

---

</div>

## 📌 Package Overview

The `internal/openapi` package is the self-contained contract documentation engine of the NOVARIS API Gateway. It serves the machine-readable **OpenAPI 3.0.3 Specification** and hosts an embedded, customized **Swagger UI Interactive Dashboard** directly from the compiled Go binary.

By leveraging Go's native compiler directive `//go:embed`, all assets (JSON schema, Swagger HTML, CSS styling, and client-side JavaScript) are baked directly into the executable binary during compilation, guaranteeing **zero runtime filesystem dependencies** in minimal Docker scratch and distroless containers.

```
                          INCOMING HTTP REQUEST
                                    │
                  ┌─────────────────┴─────────────────┐
                  ▼                                   ▼
        GET /openapi.json                          GET /docs
  ┌───────────────────────────────┐   ┌───────────────────────────────┐
  │ openapi.SpecHandler()         │   │ openapi.SwaggerUIHandler()    │
  │                               │   │                               │
  │ * Emits RFC 8259 JSON         │   │ * Emits HTML5 Dashboard       │
  │ * Content-Type: application/json│ │ * Content-Type: text/html     │
  │ * Cache-Control: max-age=3600 │   │ * Cache-Control: max-age=3600 │
  │ * RFC 9110 HEAD support       │   │ * RFC 9110 HEAD support       │
  │ * Embedded openapi.json bytes │   │ * Embedded swagger.html bytes │
  └───────────────┬───────────────┘   └───────────────┬───────────────┘
                  │                                   │
                  ▼                                   ▼
        Tooling, SDK Generators,           Developer Browser Screen
       CI Validators & Postman             (Interactive Try-It-Out)
```

---

## 🏛️ Architecture & Computer Science Principles

### 1. Compile-Time Binary Embedding (`//go:embed`)

In standard web architectures (Node.js, Python, traditional Go), serving static HTML templates or JSON schema files requires reading files from the host operating system filesystem at runtime. If the application is packaged into a minimal Docker container (such as `FROM scratch` or `distroless`) where filesystem paths differ or assets are accidentally omitted, the service crashes at runtime with `ENOENT` (File Not Found).

NOVARIS solves this with Go 1.16+ compiler directives:

```go
package openapi

import (
    _ "embed" // Blank import registers the compiler embed extension
    "net/http"
)

// Embedded directly into the data segment of the compiled Go executable:
//go:embed openapi.json
var OpenAPISpecJSON []byte

//go:embed swagger.html
var SwaggerUIHTML []byte
```

* **Memory Footprint:** The embedded assets live in the read-only static data segment of the process. They require zero heap allocations (`runtime.mallocgc`) on each incoming HTTP request.
* **Hermetic Packaging:** A single static binary contains the complete API server, database migration verifier, OpenAPI contract, and interactive web dashboard.

---

### 2. RFC 9110 HTTP Verb & Caching Semantics

Both handlers enforce strict HTTP protocol compliance:

* **GET & HEAD Verb Enforcement:** Endpoints explicitly reject mutating verbs (`POST`, `PUT`, `DELETE`, `PATCH`) with `405 Method Not Allowed`. For `HEAD` requests, Go's `net/http` stack automatically emits headers while suppressing socket payload writes.
* **Client-Side & CDN Edge Caching:**
  ```http
  Cache-Control: public, max-age=3600
  ```
  Instructs client browsers, edge load balancers, and CDN proxies to cache the 18 KB JSON specification and HTML shell for 1 hour, reducing latency to $<1\text{ms}$ for repeat visits.

---

## 📁 Package Structure

```
services/gateway/internal/openapi/
├── openapi.go          # SpecHandler & SwaggerUIHandler HTTP handler functions
├── openapi.json        # Canonical OpenAPI 3.0.3 machine-readable schema (18 KB)
├── openapi_test.go     # Table-driven unit test suite for handlers, headers & verbs
├── swagger.html        # Embedded interactive Swagger UI HTML5 dashboard
└── README.md           # This comprehensive documentation
```

---

## 📑 Contract Specifications & Endpoints

| Endpoint | HTTP Verbs | Content-Type | Access Control | Description |
| :--- | :---: | :---: | :---: | :--- |
| `/openapi.json` | `GET`, `HEAD` | `application/json` | Public | Emits the complete, machine-readable OpenAPI 3.0.3 specification describing all gateway routes, request schemas, headers, and status codes. |
| `/docs` | `GET`, `HEAD` | `text/html; charset=utf-8` | Public | Serves the interactive Swagger UI dashboard allowing developers to explore routes, inspect models, and execute authenticated test requests. |

---

## 💻 Go Implementation & Usage

### Registering Handlers in Gateway Ingress

The handlers are instantiated and attached to the standard `http.ServeMux` router in [`services/gateway/main.go`](../../main.go):

```go
package main

import (
    "net/http"
    "github.com/charan21042005/novaris/services/gateway/internal/openapi"
)

func registerRoutes(mux *http.ServeMux) {
    // Public OpenAPI 3.0.3 machine-readable contract:
    mux.HandleFunc("/openapi.json", openapi.SpecHandler())

    // Public Interactive Swagger UI Documentation Dashboard:
    mux.HandleFunc("/docs", openapi.SwaggerUIHandler())
}
```

---

## 🧪 Testing & Verification

The package includes table-driven unit tests verifying MIME types, caching directives, HTTP verb rejection, and body payload integrity:

```powershell
# Run all openapi package unit tests
Set-Location "services/gateway"
go test -v ./internal/openapi/...
```

### Test Coverage Breakdown:
1. `TestSpecHandler_GET`: Verifies HTTP 200, `Content-Type: application/json`, `Cache-Control: public, max-age=3600`, and non-empty JSON body.
2. `TestSpecHandler_HEAD`: Verifies headers are emitted with zero body bytes written.
3. `TestSpecHandler_DisallowedMethods`: Asserts `POST`, `PUT`, `DELETE` return HTTP 405.
4. `TestSwaggerUIHandler_GET`: Verifies HTTP 200, `Content-Type: text/html; charset=utf-8`, and presence of Swagger UI initialization script.
5. `TestSwaggerUIHandler_HEAD`: Verifies HEAD compliance.
6. `TestOpenAPISpec_Validity`: Validates that `openapi.json` contains valid JSON syntax and parses into valid schema structures.

---

## 🔗 Related Components

* 🛡️ **[internal/auth](../auth/README.md):** Cryptographic JWT and HMAC authentication layer documented in the OpenAPI specification.
* ⏱️ **[internal/ratelimit](../ratelimit/README.md):** Token-bucket rate limiter documented with HTTP 429 response schemas.
* 🔍 **[cmd/verifyopenapi](../../cmd/verifyopenapi/README.md):** Automated CI validator ensuring `openapi.json` conforms to platform standards.

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../../../../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

### 🎓 NOVARIS API Gateway Documentation Subsystem
*Hermetic Binary Embedding • OpenAPI 3.0.3 • Interactive Swagger UI*

</div>
