// ==============================================================================
// NOVARIS: Distributed Payment Settlement & Adaptive Fraud Intelligence Platform
// OPENAPI SERVING UNIT TEST SUITE (Day 12)
// FILE: services/gateway/internal/openapi/openapi_test.go
// ==============================================================================

package openapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSpecHandler_ReturnsValidOpenAPIJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()

	handler := SpecHandler()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &schema); err != nil {
		t.Fatalf("failed to parse returned bytes as valid JSON: %v", err)
	}

	// Validate fundamental OpenAPI 3.0 fields
	if version, ok := schema["openapi"].(string); !ok || !strings.HasPrefix(version, "3.0") {
		t.Errorf("expected openapi 3.0.x version, got %v", schema["openapi"])
	}

	paths, ok := schema["paths"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected paths object in OpenAPI schema")
	}

	expectedPaths := []string{"/healthz", "/", "/openapi.json", "/docs", "/v1/payments"}
	for _, p := range expectedPaths {
		if _, exists := paths[p]; !exists {
			t.Errorf("expected path %s in OpenAPI schema, but was missing", p)
		}
	}
}

func TestSwaggerUIHandler_ReturnsInteractiveHTML(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()

	handler := SwaggerUIHandler()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("expected Content-Type text/html, got %s", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "SwaggerUIBundle") {
		t.Errorf("expected HTML body to contain SwaggerUIBundle initialization")
	}

	if !strings.Contains(body, "/openapi.json") {
		t.Errorf("expected HTML body to reference /openapi.json URL")
	}
}

func TestHandlers_MethodNotAllowed(t *testing.T) {
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run("SpecHandler_"+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/openapi.json", nil)
			rec := httptest.NewRecorder()
			SpecHandler().ServeHTTP(rec, req)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("expected HTTP 405 Method Not Allowed on %s, got %d", method, rec.Code)
			}
		})

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
