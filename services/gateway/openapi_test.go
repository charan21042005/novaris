// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// GATEWAY OPENAPI & SWAGGER UI END-TO-END CONTRACT TESTS (Day 12)
// FILE: services/gateway/openapi_test.go
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

func TestGateway_OpenAPISpecServing(t *testing.T) {
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

	// 1. Test GET /openapi.json
	resp, err := server.Client().Get(server.URL + "/openapi.json")
	if err != nil {
		t.Fatalf("failed to GET /openapi.json: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK from /openapi.json, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var spec map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &spec); err != nil {
		t.Fatalf("failed to decode openapi.json: %v", err)
	}

	// Verify info block
	info, ok := spec["info"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing 'info' block in openapi spec")
	}

	if title, ok := info["title"].(string); !ok || !strings.Contains(title, "NOVARIS") {
		t.Errorf("expected title to contain NOVARIS, got %v", info["title"])
	}

	// Verify paths
	paths, ok := spec["paths"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing 'paths' block in openapi spec")
	}

	if _, ok := paths["/v1/payments"]; !ok {
		t.Errorf("missing '/v1/payments' route in openapi paths")
	}
}

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

	// 2. Test GET /docs
	resp, err := server.Client().Get(server.URL + "/docs")
	if err != nil {
		t.Fatalf("failed to GET /docs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK from /docs, got %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("expected Content-Type text/html, got %s", contentType)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	bodyStr := string(bodyBytes)
	if !strings.Contains(bodyStr, "SwaggerUIBundle") {
		t.Errorf("expected /docs HTML to contain SwaggerUIBundle initialization")
	}

	if !strings.Contains(bodyStr, "NOVARIS Distributed Gateway") {
		t.Errorf("expected /docs HTML to contain NOVARIS branding")
	}
}

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

	resp, err := server.Client().Get(server.URL + "/")
	if err != nil {
		t.Fatalf("failed to GET /: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK from root discovery, got %d", resp.StatusCode)
	}

	var info InfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		t.Fatalf("failed to decode root response: %v", err)
	}

	if info.Day != 12 {
		t.Errorf("expected Day 12 in root discovery, got %d", info.Day)
	}

	if _, ok := info.Endpoints["openapi_spec"]; !ok {
		t.Errorf("missing openapi_spec in root discovery endpoints map")
	}

	if _, ok := info.Endpoints["swagger_docs"]; !ok {
		t.Errorf("missing swagger_docs in root discovery endpoints map")
	}
}
