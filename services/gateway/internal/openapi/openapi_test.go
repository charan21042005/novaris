// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// OPENAPI SERVING UNIT TEST SUITE (Day 12)
// FILE: services/gateway/internal/openapi/openapi_test.go
// ==============================================================================
//
// 📚 COMPUTER SCIENCE & GO HTTP TESTING PRINCIPLES FOR BEGINNERS:
//
// 1. WHAT IS net/http/httptest?
//    In Go, you don't need to spin up a live TCP network socket to test HTTP handlers!
//    The standard library `httptest` package provides:
//      - `httptest.NewRequest`: Creates an in-memory `http.Request` struct.
//      - `httptest.NewRecorder`: An in-memory implementation of `http.ResponseWriter`
//        that captures status codes, response headers, and body bytes in a buffer.
//    This allows unit tests to execute at microsecond speeds (~0.001s) without port collisions.
//
// 2. WHY t.Fatalf VS t.Errorf?
//    - `t.Fatalf`: Stops execution of the current test function immediately. Used when
//      a failure makes subsequent assertions impossible or dangerous (e.g., if JSON
//      parsing fails, trying to inspect JSON fields will cause a nil pointer panic!).
//    - `t.Errorf`: Records a failure and continues testing, allowing developers to see
//      multiple field mismatch errors in a single test run.
//
// 3. TABLE-DRIVEN SUBTESTS (t.Run):
//    `t.Run` spawns isolated, named subtests. This produces readable CLI output:
//      --- PASS: TestHandlers_MethodNotAllowed/SpecHandler_POST (0.00s)
//      --- PASS: TestHandlers_MethodNotAllowed/SwaggerUIHandler_DELETE (0.00s)
// ==============================================================================

package openapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestSpecHandler_ReturnsValidOpenAPIJSON asserts that GET /openapi.json returns
// HTTP 200 OK, Content-Type application/json, and a valid OpenAPI 3.0 schema containing
// all mandatory NOVARIS ingress endpoints.
func TestSpecHandler_ReturnsValidOpenAPIJSON(t *testing.T) {
	// Construct a synthetic in-memory HTTP GET request targeting /openapi.json:
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	
	// Create an in-memory response recorder to capture the handler's output:
	rec := httptest.NewRecorder()

	// Instantiate the handler under test and dispatch the request:
	handler := SpecHandler()
	handler.ServeHTTP(rec, req)

	// Assertion 1: Verify the HTTP status code is 200 OK:
	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK, got %d", rec.Code)
	}

	// Assertion 2: Verify the Content-Type header starts with application/json:
	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	// Assertion 3: Verify the response body parses as valid JSON:
	var schema map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &schema); err != nil {
		t.Fatalf("failed to parse returned bytes as valid JSON: %v", err)
	}

	// Assertion 4: Validate fundamental OpenAPI 3.0 specification version string:
	if version, ok := schema["openapi"].(string); !ok || !strings.HasPrefix(version, "3.0") {
		t.Errorf("expected openapi 3.0.x version, got %v", schema["openapi"])
	}

	// Assertion 5: Verify the top-level "paths" map exists:
	paths, ok := schema["paths"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected paths object in OpenAPI schema")
	}

	// Assertion 6: Verify all 5 mandatory NOVARIS routes exist in the contract:
	expectedPaths := []string{"/healthz", "/", "/openapi.json", "/docs", "/v1/payments"}
	for _, p := range expectedPaths {
		if _, exists := paths[p]; !exists {
			t.Errorf("expected path %s in OpenAPI schema, but was missing", p)
		}
	}
}

// TestSwaggerUIHandler_ReturnsInteractiveHTML asserts that GET /docs returns
// HTTP 200 OK, Content-Type text/html, and renders the Swagger UI bootstrap script.
func TestSwaggerUIHandler_ReturnsInteractiveHTML(t *testing.T) {
	// Construct synthetic HTTP GET request for /docs:
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()

	// Dispatch request to SwaggerUIHandler:
	handler := SwaggerUIHandler()
	handler.ServeHTTP(rec, req)

	// Assertion 1: HTTP 200 OK:
	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK, got %d", rec.Code)
	}

	// Assertion 2: Header Content-Type contains text/html:
	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("expected Content-Type text/html, got %s", contentType)
	}

	// Assertion 3: Verify the HTML body contains the SwaggerUIBundle constructor:
	body := rec.Body.String()
	if !strings.Contains(body, "SwaggerUIBundle") {
		t.Errorf("expected HTML body to contain SwaggerUIBundle initialization")
	}

	// Assertion 4: Verify the HTML body specifies the spec URL pointing to /openapi.json:
	if !strings.Contains(body, "/openapi.json") {
		t.Errorf("expected HTML body to reference /openapi.json URL")
	}
}

// TestHandlers_MethodNotAllowed asserts that mutating HTTP verbs (POST, PUT, DELETE, PATCH)
// sent to documentation routes are rejected with HTTP 405 Method Not Allowed.
func TestHandlers_MethodNotAllowed(t *testing.T) {
	// Table of unsupported mutating HTTP methods:
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		// Subtest for SpecHandler (/openapi.json):
		t.Run("SpecHandler_"+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/openapi.json", nil)
			rec := httptest.NewRecorder()
			SpecHandler().ServeHTTP(rec, req)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected HTTP 405 Method Not Allowed on %s, got %d", method, rec.Code)
			}
		})

		// Subtest for SwaggerUIHandler (/docs):
		t.Run("SwaggerUIHandler_"+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/docs", nil)
			rec := httptest.NewRecorder()
			SwaggerUIHandler().ServeHTTP(rec, req)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected HTTP 405 Method Not Allowed on %s, got %d", method, rec.Code)
			}
		})
	}
}
