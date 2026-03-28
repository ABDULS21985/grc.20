// Package testutil provides test helpers for integration and unit tests.
package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/config"
	"github.com/complianceforge/platform/internal/database"
	"github.com/complianceforge/platform/internal/middleware"
)

// TestDB creates a database connection for integration tests.
// It reads from TEST_DATABASE_URL env var or defaults to a local test database.
func TestDB(t *testing.T) *database.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://complianceforge:complianceforge@localhost:5432/complianceforge_test?sslmode=disable"
	}

	cfg := config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "complianceforge",
		Password:        "complianceforge",
		Name:            "complianceforge_test",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 60,
		ConnMaxIdleTime: 30,
	}

	if os.Getenv("TEST_DB_HOST") != "" {
		cfg.Host = os.Getenv("TEST_DB_HOST")
	}

	db, err := database.New(cfg)
	if err != nil {
		t.Skipf("Skipping integration test — database not available: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// TestOrgID returns a fixed UUID for test organization.
func TestOrgID() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000099")
}

// TestUserID returns a fixed UUID for test user.
func TestUserID() uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-000000000088")
}

// TestJWTSecret is the JWT secret used in tests.
const TestJWTSecret = "test-jwt-secret-for-integration-tests-only"

// GenerateTestToken creates a valid JWT token for test requests.
func GenerateTestToken(t *testing.T, userID, orgID uuid.UUID, roles []string) string {
	t.Helper()

	now := time.Now()
	claims := middleware.Claims{
		UserID:         userID,
		OrganizationID: orgID,
		Email:          "test@complianceforge.io",
		Roles:          roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "complianceforge-test",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(TestJWTSecret))
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}
	return signed
}

// TestClient wraps httptest for making authenticated API requests.
type TestClient struct {
	t      *testing.T
	server *httptest.Server
	token  string
}

// NewTestClient creates a test client with a pre-configured JWT token.
func NewTestClient(t *testing.T, handler http.Handler) *TestClient {
	t.Helper()
	server := httptest.NewServer(handler)
	token := GenerateTestToken(t, TestUserID(), TestOrgID(), []string{"org_admin"})

	t.Cleanup(func() {
		server.Close()
	})

	return &TestClient{t: t, server: server, token: token}
}

// Get makes an authenticated GET request.
func (c *TestClient) Get(path string) *http.Response {
	c.t.Helper()
	req, err := http.NewRequest("GET", c.server.URL+path, nil)
	if err != nil {
		c.t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Request failed: %v", err)
	}
	return resp
}

// Post makes an authenticated POST request with JSON body.
func (c *TestClient) Post(path string, body interface{}) *http.Response {
	c.t.Helper()
	jsonBody, err := json.Marshal(body)
	if err != nil {
		c.t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest("POST", c.server.URL+path, bytes.NewReader(jsonBody))
	if err != nil {
		c.t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Request failed: %v", err)
	}
	return resp
}

// Put makes an authenticated PUT request with JSON body.
func (c *TestClient) Put(path string, body interface{}) *http.Response {
	c.t.Helper()
	jsonBody, err := json.Marshal(body)
	if err != nil {
		c.t.Fatalf("Failed to marshal body: %v", err)
	}

	req, err := http.NewRequest("PUT", c.server.URL+path, bytes.NewReader(jsonBody))
	if err != nil {
		c.t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Request failed: %v", err)
	}
	return resp
}

// Delete makes an authenticated DELETE request.
func (c *TestClient) Delete(path string) *http.Response {
	c.t.Helper()
	req, err := http.NewRequest("DELETE", c.server.URL+path, nil)
	if err != nil {
		c.t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatalf("Request failed: %v", err)
	}
	return resp
}

// GetPublic makes an unauthenticated GET request.
func (c *TestClient) GetPublic(path string) *http.Response {
	c.t.Helper()
	resp, err := http.Get(c.server.URL + path)
	if err != nil {
		c.t.Fatalf("Request failed: %v", err)
	}
	return resp
}

// DecodeJSON decodes a response body into the target struct.
func DecodeJSON(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
}

// AssertStatus checks that the response has the expected status code.
func AssertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		t.Errorf("Expected status %d, got %d", expected, resp.StatusCode)
	}
}

// SeedTestOrg creates a test organization in the database.
func SeedTestOrg(t *testing.T, db *database.DB) {
	t.Helper()
	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO organizations (id, name, slug, industry, country_code, status, tier, timezone, default_language)
		VALUES ($1, 'Test Organisation', 'test-org', 'technology', 'GB', 'active', 'enterprise', 'Europe/London', 'en')
		ON CONFLICT (id) DO NOTHING`,
		TestOrgID())
	if err != nil {
		t.Fatalf("Failed to seed test org: %v", err)
	}
}

// SeedTestUser creates a test user in the database.
func SeedTestUser(t *testing.T, db *database.DB) {
	t.Helper()
	ctx := context.Background()
	_, err := db.Pool.Exec(ctx, `
		INSERT INTO users (id, organization_id, email, password_hash, first_name, last_name, status)
		VALUES ($1, $2, 'test@complianceforge.io', '$2a$12$placeholder', 'Test', 'User', 'active')
		ON CONFLICT DO NOTHING`,
		TestUserID(), TestOrgID())
	if err != nil {
		t.Fatalf("Failed to seed test user: %v", err)
	}
}

// CleanupTestData removes test data after tests.
func CleanupTestData(t *testing.T, db *database.DB) {
	t.Helper()
	ctx := context.Background()
	tables := []string{
		"risk_control_mappings", "risk_indicator_values", "risk_indicators",
		"risk_treatments", "risk_assessments", "risks",
		"control_test_results", "control_evidence", "control_implementations",
		"organization_frameworks", "policy_attestations", "policy_attestation_campaigns",
		"policy_exceptions", "policy_versions", "policies",
		"audit_findings", "audits", "incidents", "assets", "vendors",
		"user_entity_permissions", "user_roles", "user_sessions", "user_mfa",
	}
	for _, table := range tables {
		db.Pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE organization_id = $1", table), TestOrgID())
	}
}
