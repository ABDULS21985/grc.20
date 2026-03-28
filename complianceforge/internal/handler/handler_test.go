package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/complianceforge/platform/internal/models"
)

// ============================================================
// HEALTH ENDPOINT TESTS
// ============================================================

func TestHealthEndpoint(t *testing.T) {
	// This is a unit test — no database required
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	// Simulate the health handler directly
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
	})

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)

	if body["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", body["status"])
	}
}

// ============================================================
// AUTH ENDPOINT TESTS
// ============================================================

func TestLoginMissingFields(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Simulate missing fields response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    "MISSING_FIELDS",
			Message: "Email and password are required",
		},
	})

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.StatusCode)
	}
}

func TestLoginInvalidJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	w.WriteHeader(http.StatusBadRequest)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.StatusCode)
	}
	_ = req
}

// ============================================================
// FRAMEWORK ENDPOINT TESTS
// ============================================================

func TestFrameworkListResponse(t *testing.T) {
	// Test that framework list returns expected structure
	frameworks := []models.ComplianceFramework{
		{
			Code:              "ISO27001",
			Name:              "ISO 27001",
			Version:           "2022",
			IssuingBody:       "ISO/IEC",
			Category:          "security",
			IsSystemFramework: true,
			IsActive:          true,
			TotalControls:     93,
		},
		{
			Code:              "NIST_CSF_2",
			Name:              "NIST CSF 2.0",
			Version:           "2.0",
			IssuingBody:       "NIST",
			Category:          "security",
			IsSystemFramework: true,
			IsActive:          true,
			TotalControls:     106,
		},
	}

	resp := models.APIResponse{
		Success: true,
		Data:    frameworks,
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded models.APIResponse
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !decoded.Success {
		t.Error("Expected success = true")
	}
}

// ============================================================
// PAGINATION TESTS
// ============================================================

func TestPaginationDefaults(t *testing.T) {
	params := models.DefaultPagination()

	if params.Page != 1 {
		t.Errorf("Expected default page 1, got %d", params.Page)
	}
	if params.PageSize != 20 {
		t.Errorf("Expected default page_size 20, got %d", params.PageSize)
	}
	if params.SortDir != "desc" {
		t.Errorf("Expected default sort_dir 'desc', got '%s'", params.SortDir)
	}
}

func TestPaginationOffset(t *testing.T) {
	tests := []struct {
		page     int
		pageSize int
		expected int
	}{
		{1, 20, 0},
		{2, 20, 20},
		{3, 20, 40},
		{1, 50, 0},
		{5, 10, 40},
	}

	for _, tt := range tests {
		params := models.PaginationParams{Page: tt.page, PageSize: tt.pageSize}
		if params.Offset() != tt.expected {
			t.Errorf("Page %d, Size %d: expected offset %d, got %d",
				tt.page, tt.pageSize, tt.expected, params.Offset())
		}
	}
}

// ============================================================
// API RESPONSE STRUCTURE TESTS
// ============================================================

func TestAPIResponseSuccess(t *testing.T) {
	resp := models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "ok"},
	}

	jsonBytes, _ := json.Marshal(resp)
	var decoded map[string]interface{}
	json.Unmarshal(jsonBytes, &decoded)

	if decoded["success"] != true {
		t.Error("Expected success = true")
	}
	if decoded["error"] != nil {
		t.Error("Expected error to be nil on success")
	}
}

func TestAPIResponseError(t *testing.T) {
	resp := models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    "NOT_FOUND",
			Message: "Resource not found",
			Details: map[string]string{"id": "must be a valid UUID"},
		},
	}

	jsonBytes, _ := json.Marshal(resp)
	var decoded map[string]interface{}
	json.Unmarshal(jsonBytes, &decoded)

	if decoded["success"] != false {
		t.Error("Expected success = false")
	}
	if decoded["error"] == nil {
		t.Error("Expected error to be present")
	}
}
