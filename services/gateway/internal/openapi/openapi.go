// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// OPENAPI & INTERACTIVE SWAGGER UI SERVING SUBSYSTEM (Day 12)
// FILE: services/gateway/internal/openapi/openapi.go
// ==============================================================================

package openapi

import (
	_ "embed"
	"net/http"
)

// OpenAPISpecJSON embeds the machine-readable OpenAPI 3.0.3 specification directly
// into the compiled binary, eliminating external filesystem runtime dependencies.
//
//go:embed openapi.json
var OpenAPISpecJSON []byte

// SwaggerUIHTML embeds the standalone, customized HTML/CSS/JS dashboard that loads
// Swagger UI and renders the /openapi.json specification in the developer's browser.
//
//go:embed swagger.html
var SwaggerUIHTML []byte

// SpecHandler returns an HTTP handler that serves the raw OpenAPI 3.0.3 JSON schema.
// Publicly accessible without authentication for tooling, SDK generation, and Postman import.
func SpecHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(OpenAPISpecJSON)
	}
}

// SwaggerUIHandler returns an HTTP handler that serves the interactive Swagger UI page.
// Publicly accessible at GET /docs.
func SwaggerUIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=3600")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(SwaggerUIHTML)
	}
}
