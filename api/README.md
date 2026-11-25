<div align="center">

<!-- Top Banner -->
<p align="center">
  <img src="../assets/banner.svg" alt="NOVARIS Banner" width="100%" />
</p>

# 📑 NOVARIS OpenAPI 3.0.3 Contract Vault
### Formal Machine-Readable API Specifications, Schema Definitions & Interactive Documentation

<p align="center">
  <img src="https://img.shields.io/badge/Phase_1_Capstone-COMPLETED_(100%25_Verified)-00C853?style=flat-square&logo=checkbox&logoColor=white" alt="Phase 1 Capstone" />
  <img src="https://img.shields.io/badge/OpenAPI-3.0.3_OAS-00E5FF?style=flat-square&logo=openapiinitiative&logoColor=white" alt="OpenAPI" />
  <img src="https://img.shields.io/badge/Contract-First_Standard-00C853?style=flat-square&logo=swagger&logoColor=white" alt="Contract" />
  <img src="https://img.shields.io/badge/Formats-YAML_%26_JSON-7C4DFF?style=flat-square&logo=json&logoColor=white" alt="Formats" />
  <img src="https://img.shields.io/badge/Interactive_Docs-Swagger_UI_%26_Redoc-FFA000?style=flat-square&logo=postman&logoColor=white" alt="Docs" />
</p>

---

</div>

## 📌 Purpose of This Directory

This directory houses the formal, versioned **OpenAPI 3.0.3 Specification** for the NOVARIS Distributed Payment Settlement Gateway.

In contract-first fintech engineering:
1. **Machine-Readable Truth:** Both backend engineers and external merchants share a single source of truth defining parameter schemas, mandatory headers, and error codes.
2. **Automated SDK Generation:** Client libraries for TypeScript/React, Go, and Python can be compiled directly using `openapi-generator-cli`.
3. **Mocking & Contract Testing:** QA teams can spin up automated mock servers using Stoplight Prism (`prism mock openapi.yaml`) without needing a running database.
4. **Live Interactive Exploration:** The Gateway microservice natively serves an interactive Swagger UI dashboard at `GET /docs`.

---

## 🗺️ Specification Artifacts

| File | Format | Description |
| :--- | :---: | :--- |
| **[`openapi.yaml`](./openapi.yaml)** | YAML | Human-readable canonical OpenAPI 3.0.3 specification with rich descriptions and examples. |
| **[`openapi.json`](./openapi.json)** | JSON | Minified/pretty machine-parseable JSON specification for tooling, Postman import, and programmatic consumption. |

---

## ⚡ How to Consume & Render the Specification

### 1. Interactive Swagger UI in Local Browser
The gateway microservice serves interactive API documentation out-of-the-box. Start the gateway and visit:
* **Interactive UI:** `http://localhost:8080/docs`
* **Raw JSON Spec:** `http://localhost:8080/openapi.json`

### 2. Rendering Headlessly with Redocly CLI
```bash
# Preview via Redocly live development server
npx @redocly/cli preview-docs api/openapi.yaml --port 8081

# Validate specification conformance
npx @redocly/cli lint api/openapi.yaml
```

### 3. Running an Instant Mock Server (Prism)
```bash
# Spin up an HTTP mock server validating requests against openapi.yaml
npx @stoplight/prism-cli mock api/openapi.yaml -p 4010
```

### 4. Generating Client SDKs (TypeScript / Go / Python)
```bash
# Generate TypeScript Axios client
npx @openapitools/openapi-generator-cli generate \
  -i api/openapi.yaml \
  -g typescript-axios \
  -o sdk/typescript

# Generate Go API client
npx @openapitools/openapi-generator-cli generate \
  -i api/openapi.yaml \
  -g go \
  -o sdk/go
```

---

<div align="center">

<!-- Bottom Banner -->
<p align="center">
  <img src="../assets/footer.svg" alt="NOVARIS Footer" width="100%" />
</p>

<p align="center">
  <strong>NOVARIS API Specifications • Maintained by <a href="https://github.com/charan21042005">Tokachichu Sri Charan</a></strong>
</p>

</div>
