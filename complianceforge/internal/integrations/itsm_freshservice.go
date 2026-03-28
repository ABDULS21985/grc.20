package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ============================================================
// FRESHSERVICE ITSM ADAPTER
// ============================================================

// FreshserviceConfig holds credentials for the Freshservice REST API.
type FreshserviceConfig struct {
	Domain string `json:"domain"`    // e.g. "mycompany" (becomes mycompany.freshservice.com)
	APIKey string `json:"api_key"`   // Freshservice API key
}

// FreshserviceTicket represents a Freshservice ticket.
type FreshserviceTicket struct {
	ID          int64    `json:"id,omitempty"`
	Subject     string   `json:"subject"`
	Description string   `json:"description,omitempty"`
	Status      int      `json:"status,omitempty"`       // 2=Open, 3=Pending, 4=Resolved, 5=Closed
	Priority    int      `json:"priority,omitempty"`      // 1=Low, 2=Medium, 3=High, 4=Urgent
	Type        string   `json:"type,omitempty"`          // Incident, Service Request, etc.
	Source      int      `json:"source,omitempty"`        // 1=Email, 2=Portal, 3=Phone, etc.
	Tags        []string `json:"tags,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

// FreshserviceAdapter implements the integration adapter for Freshservice ITSM.
type FreshserviceAdapter struct {
	config     FreshserviceConfig
	httpClient *http.Client
	baseURL    string
}

func NewFreshserviceAdapter() *FreshserviceAdapter {
	return &FreshserviceAdapter{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (f *FreshserviceAdapter) Type() string {
	return "itsm_freshservice"
}

func (f *FreshserviceAdapter) Connect(_ context.Context, config map[string]interface{}) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal Freshservice config: %w", err)
	}
	if err := json.Unmarshal(data, &f.config); err != nil {
		return fmt.Errorf("failed to parse Freshservice config: %w", err)
	}

	if f.config.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if f.config.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}

	// Build base URL — support both full URLs and just the domain name
	domain := strings.TrimRight(f.config.Domain, "/")
	if strings.HasPrefix(domain, "https://") || strings.HasPrefix(domain, "http://") {
		f.baseURL = strings.TrimRight(domain, "/")
	} else {
		f.baseURL = fmt.Sprintf("https://%s.freshservice.com", domain)
	}

	return nil
}

func (f *FreshserviceAdapter) Disconnect(_ context.Context) error {
	f.config = FreshserviceConfig{}
	f.baseURL = ""
	return nil
}

// HealthCheck verifies connectivity to the Freshservice API.
func (f *FreshserviceAdapter) HealthCheck(ctx context.Context) (*HealthResult, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		f.baseURL+"/api/v2/agents/me", nil)
	if err != nil {
		return &HealthResult{Status: "unhealthy", Message: err.Error()}, nil
	}
	f.setAuth(req)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return &HealthResult{
			Status:  "unhealthy",
			Message: fmt.Sprintf("Freshservice unreachable: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	latency := time.Since(start).Milliseconds()

	if resp.StatusCode == http.StatusUnauthorized {
		return &HealthResult{
			Status:  "unhealthy",
			Message: "Authentication failed — check API key",
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &HealthResult{
			Status:  "degraded",
			Message: fmt.Sprintf("Freshservice returned HTTP %d: %s", resp.StatusCode, string(body)),
		}, nil
	}

	log.Info().Int64("latency_ms", latency).Msg("Freshservice health check passed")
	return &HealthResult{
		Status:  "healthy",
		Message: fmt.Sprintf("Freshservice API accessible (%dms)", latency),
	}, nil
}

// Sync fetches open tickets from Freshservice.
func (f *FreshserviceAdapter) Sync(ctx context.Context) (*SyncResultAdapter, error) {
	result := &SyncResultAdapter{}

	// Fetch open and pending tickets
	tickets, err := f.ListTickets(ctx, "open", 100)
	if err != nil {
		return nil, fmt.Errorf("failed to list Freshservice tickets: %w", err)
	}

	result.RecordsProcessed = len(tickets)
	result.RecordsCreated = len(tickets)

	log.Info().
		Int("tickets_found", len(tickets)).
		Msg("Freshservice sync completed")

	return result, nil
}

// ============================================================
// CRUD OPERATIONS
// ============================================================

// CreateTicket creates a new ticket in Freshservice.
func (f *FreshserviceAdapter) CreateTicket(ctx context.Context, ticket FreshserviceTicket) (*FreshserviceTicket, error) {
	if ticket.Status == 0 {
		ticket.Status = 2 // Open
	}
	if ticket.Priority == 0 {
		ticket.Priority = 2 // Medium
	}
	if ticket.Source == 0 {
		ticket.Source = 2 // Portal
	}

	payload, err := json.Marshal(ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ticket: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		f.baseURL+"/api/v2/tickets", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	f.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create ticket request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Freshservice returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var created struct {
		Ticket FreshserviceTicket `json:"ticket"`
	}
	if err := json.Unmarshal(body, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created ticket: %w", err)
	}

	log.Info().Int64("id", created.Ticket.ID).Str("subject", created.Ticket.Subject).Msg("Freshservice ticket created")
	return &created.Ticket, nil
}

// GetTicket fetches a Freshservice ticket by ID.
func (f *FreshserviceAdapter) GetTicket(ctx context.Context, ticketID int64) (*FreshserviceTicket, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/api/v2/tickets/%d", f.baseURL, ticketID), nil)
	if err != nil {
		return nil, err
	}
	f.setAuth(req)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get ticket request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Freshservice returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Ticket FreshserviceTicket `json:"ticket"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ticket: %w", err)
	}

	return &result.Ticket, nil
}

// UpdateTicket updates fields on an existing Freshservice ticket.
func (f *FreshserviceAdapter) UpdateTicket(ctx context.Context, ticketID int64, updates map[string]interface{}) (*FreshserviceTicket, error) {
	payload, err := json.Marshal(updates)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal updates: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		fmt.Sprintf("%s/api/v2/tickets/%d", f.baseURL, ticketID),
		bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	f.setAuth(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("update ticket request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Freshservice returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Ticket FreshserviceTicket `json:"ticket"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse updated ticket: %w", err)
	}

	log.Info().Int64("id", ticketID).Msg("Freshservice ticket updated")
	return &result.Ticket, nil
}

// ListTickets fetches tickets filtered by status.
func (f *FreshserviceAdapter) ListTickets(ctx context.Context, filter string, perPage int) ([]FreshserviceTicket, error) {
	if perPage <= 0 {
		perPage = 30
	}

	queryURL := fmt.Sprintf("%s/api/v2/tickets?per_page=%d", f.baseURL, perPage)
	if filter != "" {
		queryURL += fmt.Sprintf("&filter=%s", filter)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, err
	}
	f.setAuth(req)

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list tickets request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Freshservice returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Tickets []FreshserviceTicket `json:"tickets"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tickets: %w", err)
	}

	return result.Tickets, nil
}

// setAuth sets Freshservice API key authentication (basic auth with key:X).
func (f *FreshserviceAdapter) setAuth(req *http.Request) {
	req.SetBasicAuth(f.config.APIKey, "X")
	req.Header.Set("Accept", "application/json")
}
