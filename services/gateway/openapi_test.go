// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// GATEWAY OPENAPI & SWAGGER UI END-TO-END CONTRACT TESTS (Day 12)
// FILE: services/gateway/openapi_test.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & INTEGRATION TESTING PRINCIPLES FOR BEGINNERS:
//
// 1. WHAT IS END-TO-END CONTRACT INTEGRATION TESTING?
//    While unit tests verify individual functions in isolation (like `SpecHandler`),
//    integration tests boot the real `SetupServerMux` router on an ephemeral TCP socket
//    using `httptest.NewServer(handler)`.
//    This tests the complete HTTP stack:
//      Client TCP Socket -> HTTP Router (Mux) -> Embedded Spec Handler -> Network Wire -> HTTP Client
//
// 2. EPHEMERAL TCP PORT ALLOCATION:
//    Notice `cfg.Port = "0"` and `httptest.NewServer`:
//    The operating system assigns a random available ephemeral TCP port (e.g., 54621).
//    This prevents "address already in use" errors and allows test suites to execute
//    concurrently across multiple CI/CD worker threads.
// ==============================================================================

package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/charan21042005/novaris/services/gateway/internal/ratelimit"
)

// TestGateway_OpenAPISpecServing boots the live gateway server and verifies that
// GET /openapi.json delivers a compliant, well-formed OpenAPI 3.0 specification.
func TestGateway_OpenAPISpecServing(t *testing.T) {
	// Construct test gateway server configuration:
	cfg := Config{
		Port:           "0",
		ServiceName:    "novaris-gateway-test",
		ServiceVersion: "0.12.0-test",
		Environment:    "test",
	}

	// Instantiate mock in-memory stores for rate limiting and idempotency:
	bucketStore := ratelimit.NewBucketStore(10, 2)
	idempStore := NewInMemoryIdempotencyStore()
	
	// Initialize the real production server mux router:
	handler := SetupServerMux(cfg, nil, bucketStore, idempStore)
	
	// Boot the live test server on an ephemeral loopback TCP socket:
	server := httptest.NewServer(handler)
	// Ensure the server TCP listener is cleanly closed when the test completes:
	defer server.Close()

	// --------------------------------------------------------------------------
	// 1. Dispatch real HTTP GET request over network wire to /openapi.json
	// --------------------------------------------------------------------------
	resp, err := server.Client().Get(server.URL + "/openapi.json")
	if err != nil {
		t.Fatalf("failed to GET /openapi.json: %v", err)
	}
	defer resp.Body.Close()

	// Assertion 1: Status code must be HTTP 200 OK:
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK from /openapi.json, got %d", resp.StatusCode)
	}

	// Assertion 2: Content-Type must be application/json:
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	// Assertion 3: Read response body bytes:
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	// Assertion 4: Parse response body as JSON:
	var spec map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &spec); err != nil {
		t.Fatalf("failed to decode openapi.json: %v", err)
	}

	// Assertion 5: Verify 'info' block exists and contains platform title:
	info, ok := spec["info"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing 'info' block in openapi spec")
	}

	if title, ok := info["title"].(string); !ok || !strings.Contains(title, "NOVARIS") {
		t.Errorf("expected title to contain NOVARIS, got %v", info["title"])
	}

	// Assertion 6: Verify 'paths' block exists and documents /v1/payments:
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing 'paths' block in openapi spec")
	}

	if _, ok := paths["/v1/payments"]; !ok {
		t.Errorf("missing '/v1/payments' route in openapi paths")
	}
}

// TestGateway_SwaggerUIServing verifies that GET /docs serves the dark-themed
// interactive Swagger UI HTML developer dashboard over the network.
func TestGateway_SwaggerUIServing(t *testing.T) {
	cfg := Config{
		Port:           "0",
		ServiceName:    "novaris-gateway-test",
		ServiceVersion: "0.12.0-test",
		Environment:    "test",
	}

	bucketStore := ratelimit.NewBucketStore(10, 2)
	idempStore := NewInMemoryIdempotencyStore()
	handler := SetupServerMux(cfg, nil, bucketStore, idempStore)
	server := httptest.NewServer(handler)
	defer server.Close()

	// --------------------------------------------------------------------------
	// 2. Dispatch HTTP GET request to /docs
	// --------------------------------------------------------------------------
	resp, err := server.Client().Get(server.URL + "/docs")
	if err != nil {
		t.Fatalf("failed to GET /docs: %v", err)
	}
	defer resp.Body.Close()

	// Assertion 1: HTTP 200 OK:
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK from /docs, got %d", resp.StatusCode)
	}

	// Assertion 2: Content-Type must be text/html:
	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("expected Content-Type text/html, got %s", contentType)
	}

	// Assertion 3: Verify the HTML body contains Swagger UI initialization:
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	body := string(bodyBytes)

	if !strings.Contains(body, "SwaggerUIBundle") {
		t.Errorf("expected HTML body to contain SwaggerUIBundle initialization")
	}

	// Assertion 4: Verify the HTML links to /openapi.json:
	if !strings.Contains(body, "/openapi.json") {
		t.Errorf("expected HTML body to link to /openapi.json")
	}
}

// TestGateway_RootDiscoveryContainsDocsLinks verifies that the root discovery
// endpoint (GET /) advertises both /docs and /openapi.json to client callers.
func TestGateway_RootDiscoveryContainsDocsLinks(t *testing.T) {
	cfg := Config{
		Port:           "0",
		ServiceName:    "novaris-gateway-test",
		ServiceVersion: "0.12.0-test",
		Environment:    "test",
	}

	bucketStore := ratelimit.NewBucketStore(10, 2)
	idempStore := NewInMemoryIdempotencyStore()
	handler := SetupServerMux(cfg, nil, bucketStore, idempStore)
	server := httptest.NewServer(handler)
	defer server.Close()

	// --------------------------------------------------------------------------
	// 3. Dispatch HTTP GET request to /
	// --------------------------------------------------------------------------
	resp, err := server.Client().Get(server.URL + "/")
	if err != nil {
		t.Fatalf("failed to GET /: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK from /, got %d", resp.StatusCode)
	}

	var info InfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		t.Fatalf("failed to parse JSON from /: %v", err)
	}

	if info.Day < 12 {
		t.Errorf("expected at least Day 12 in root discovery, got %d", info.Day)
	}

	// Assertion: Ensure docs and openapi hypermedia links are present in discovery banner:
	if _, ok := info.Endpoints["openapi_spec"]; !ok {
		t.Errorf("missing openapi_spec in root discovery endpoints map")
	}

	if _, ok := info.Endpoints["swagger_docs"]; !ok {
		t.Errorf("missing swagger_docs in root discovery endpoints map")
	}
}
