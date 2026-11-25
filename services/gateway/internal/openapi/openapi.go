// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// OPENAPI & INTERACTIVE SWAGGER UI SERVING SUBSYSTEM (Day 12)
// FILE: services/gateway/internal/openapi/openapi.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & GO SYSTEM CONCEPTS FOR BEGINNERS:
//
// 1. WHAT IS //go:embed?
//    Introduced in Go 1.16, `//go:embed` is a compiler directive (an instruction to
//    the Go compiler) that takes a file on your disk and directly bundles its raw
//    bytes into the compiled binary executable during `go build`.
//
//    Why is this revolutionary?
//    In traditional languages (like Python, Node.js, or older Go), if an application
//    needs to serve an HTML template or a JSON config, that file MUST exist on the
//    host filesystem at runtime. If the file is moved, deleted, or missing in a
//    Docker scratch image, the server crashes with "file not found"!
//    With `//go:embed`, the resulting binary is completely self-contained (zero host
//    file path dependencies).
//
// 2. THE BLANK IMPORT `_ "embed"`:
//    To use `//go:embed`, Go requires importing the standard library `"embed"` package.
//    Because our code doesn't call functions like `embed.FS` directly (it uses the
//    compiler directive syntax), we use the blank identifier `_ "embed"` to satisfy
//    Go's strict unused import compiler checks.
//
// 3. HTTP HEAD METHOD SUPPORT:
//    RFC 9110 specifies that any endpoint supporting GET SHOULD also support HEAD.
//    A HEAD request returns the exact same HTTP headers as GET (such as Content-Type
//    and Content-Length) but omits the response body, allowing health probes, caches,
//    and CDN edge proxies to check document availability with zero network payload.
// ==============================================================================

package openapi

import (
	// Blank import required by the Go compiler to activate the //go:embed directive:
	_ "embed"
	"net/http"
)

// OpenAPISpecJSON embeds the machine-readable OpenAPI 3.0.3 specification directly
// into the compiled binary memory. At runtime, this variable is a read-only slice of bytes
// containing the exact JSON file content bundled at build time.
//
//go:embed openapi.json
var OpenAPISpecJSON []byte

// SwaggerUIHTML embeds the standalone, customized HTML/CSS/JS dashboard that loads
// Swagger UI and renders the /openapi.json specification in the developer's browser.
//
//go:embed swagger.html
var SwaggerUIHTML []byte

// SpecHandler constructs and returns a standard Go http.HandlerFunc that serves the raw
// OpenAPI 3.0.3 JSON schema to requesting clients.
//
// Publicly accessible without authentication for tooling, SDK generators, and Postman.
func SpecHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enforce HTTP verb constraints: only GET and HEAD are permitted on documentation resources:
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set the standard MIME type indicating the payload is valid RFC 8259 JSON:
		w.Header().Set("Content-Type", "application/json")

		// Instruct downstream HTTP caches and developer browsers to cache the spec for up to 1 hour (3600s),
		// dramatically reducing redundant payload transfer on repeated Swagger UI page loads:
		w.Header().Set("Cache-Control", "public, max-age=3600")

		// Emit HTTP 200 OK header status code:
		w.WriteHeader(http.StatusOK)

		// For GET requests, write the raw embedded JSON bytes to the client socket.
		// For HEAD requests, Go's net/http automatically suppresses the body bytes.
		_, _ = w.Write(OpenAPISpecJSON)
	}
}

// SwaggerUIHandler constructs and returns a standard Go http.HandlerFunc that serves the
// interactive Swagger UI HTML page at GET /docs.
//
// Publicly accessible to any developer visiting the gateway port in their web browser.
func SwaggerUIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enforce HTTP verb constraints: only GET and HEAD are permitted:
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Set the MIME type indicating HTML document formatted in UTF-8:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		// Enable 1-hour client-side caching for the static HTML dashboard shell:
		w.Header().Set("Cache-Control", "public, max-age=3600")

		// Emit HTTP 200 OK header status code:
		w.WriteHeader(http.StatusOK)

		// Write the embedded HTML bytes to the HTTP response writer:
		_, _ = w.Write(SwaggerUIHTML)
	}
}
