package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// ============================================================
// COMPLIANCE-AS-CODE RESOURCE TYPES
// ============================================================

// CaCResource represents a single compliance-as-code resource
// parsed from a YAML file. It follows a Kubernetes-like structure
// with apiVersion, kind, metadata, and spec fields.
type CaCResource struct {
	ApiVersion string              `yaml:"apiVersion" json:"api_version"`
	Kind       string              `yaml:"kind" json:"kind"`
	Metadata   CaCResourceMetadata `yaml:"metadata" json:"metadata"`
	Spec       map[string]interface{} `yaml:"spec" json:"spec"`
}

// CaCResourceMetadata holds identifying metadata for a CaC resource.
type CaCResourceMetadata struct {
	Name        string            `yaml:"name" json:"name"`
	UID         string            `yaml:"uid" json:"uid"`
	Namespace   string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
}

// ============================================================
// DATABASE MODEL TYPES
// ============================================================

// CaCRepository represents a connected Git repository for compliance-as-code.
type CaCRepository struct {
	ID              uuid.UUID  `json:"id"`
	OrganizationID  uuid.UUID  `json:"organization_id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Provider        string     `json:"provider"`
	RepoURL         string     `json:"repo_url"`
	Branch          string     `json:"branch"`
	BasePath        string     `json:"base_path"`
	WebhookSecret   string     `json:"-"`
	AccessToken     string     `json:"-"`
	SyncDirection   string     `json:"sync_direction"`
	AutoSync        bool       `json:"auto_sync"`
	RequireApproval bool       `json:"require_approval"`
	LastSyncAt      *time.Time `json:"last_sync_at"`
	LastSyncStatus  *string    `json:"last_sync_status"`
	ResourceCount   int        `json:"resource_count"`
	IsActive        bool       `json:"is_active"`
	Metadata        json.RawMessage `json:"metadata"`
	CreatedBy       *uuid.UUID `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CaCSyncRun represents a single synchronisation run between a
// Git repository and the ComplianceForge platform.
type CaCSyncRun struct {
	ID                 uuid.UUID       `json:"id"`
	OrganizationID     uuid.UUID       `json:"organization_id"`
	RepositoryID       uuid.UUID       `json:"repository_id"`
	Status             string          `json:"status"`
	Direction          string          `json:"direction"`
	TriggerType        string          `json:"trigger_type"`
	CommitSHA          string          `json:"commit_sha"`
	Branch             string          `json:"branch"`
	FilesChanged       int             `json:"files_changed"`
	ResourcesAdded     int             `json:"resources_added"`
	ResourcesUpdated   int             `json:"resources_updated"`
	ResourcesDeleted   int             `json:"resources_deleted"`
	ResourcesUnchanged int             `json:"resources_unchanged"`
	Errors             json.RawMessage `json:"errors"`
	DiffPlanJSON       json.RawMessage `json:"diff_plan"`
	StartedAt          *time.Time      `json:"started_at"`
	CompletedAt        *time.Time      `json:"completed_at"`
	ApprovedBy         *uuid.UUID      `json:"approved_by"`
	ApprovedAt         *time.Time      `json:"approved_at"`
	RejectedBy         *uuid.UUID      `json:"rejected_by"`
	RejectedAt         *time.Time      `json:"rejected_at"`
	RejectionReason    string          `json:"rejection_reason"`
	TriggeredBy        *uuid.UUID      `json:"triggered_by"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	// Joined fields
	RepositoryName     string          `json:"repository_name,omitempty"`
}

// CaCResourceMapping maps a YAML resource to a platform entity.
type CaCResourceMapping struct {
	ID                 uuid.UUID  `json:"id"`
	OrganizationID     uuid.UUID  `json:"organization_id"`
	RepositoryID       uuid.UUID  `json:"repository_id"`
	FilePath           string     `json:"file_path"`
	ApiVersion         string     `json:"api_version"`
	Kind               string     `json:"kind"`
	ResourceName       string     `json:"resource_name"`
	ResourceUID        string     `json:"resource_uid"`
	PlatformEntityType string     `json:"platform_entity_type"`
	PlatformEntityID   *uuid.UUID `json:"platform_entity_id"`
	Status             string     `json:"status"`
	ContentHash        string     `json:"content_hash"`
	LastSyncedAt       *time.Time `json:"last_synced_at"`
	LastSyncRunID      *uuid.UUID `json:"last_sync_run_id"`
	Metadata           json.RawMessage `json:"metadata"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	// Joined fields
	RepositoryName     string     `json:"repository_name,omitempty"`
}

// CaCDriftEvent represents a detected drift between repo and platform state.
type CaCDriftEvent struct {
	ID                uuid.UUID  `json:"id"`
	OrganizationID    uuid.UUID  `json:"organization_id"`
	RepositoryID      uuid.UUID  `json:"repository_id"`
	ResourceMappingID *uuid.UUID `json:"resource_mapping_id"`
	DriftDirection    string     `json:"drift_direction"`
	Status            string     `json:"status"`
	Kind              string     `json:"kind"`
	ResourceName      string     `json:"resource_name"`
	ResourceUID       string     `json:"resource_uid"`
	FieldPath         string     `json:"field_path"`
	RepoValue         string     `json:"repo_value"`
	PlatformValue     string     `json:"platform_value"`
	Description       string     `json:"description"`
	DetectedAt        time.Time  `json:"detected_at"`
	ResolvedAt        *time.Time `json:"resolved_at"`
	ResolvedBy        *uuid.UUID `json:"resolved_by"`
	ResolutionAction  string     `json:"resolution_action"`
	ResolutionNotes   string     `json:"resolution_notes"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	// Joined fields
	RepositoryName    string     `json:"repository_name,omitempty"`
}

// ============================================================
// DIFF / PLAN TYPES
// ============================================================

// DiffPlan describes the set of changes needed to synchronise
// a repository's resources with the platform state.
type DiffPlan struct {
	RepositoryID uuid.UUID    `json:"repository_id"`
	Actions      []DiffAction `json:"actions"`
	Summary      DiffSummary  `json:"summary"`
}

// DiffAction is a single create/update/delete action in a DiffPlan.
type DiffAction struct {
	Action       string      `json:"action"` // create, update, delete, no_change
	Kind         string      `json:"kind"`
	ResourceUID  string      `json:"resource_uid"`
	ResourceName string      `json:"resource_name"`
	FilePath     string      `json:"file_path"`
	Changes      []FieldDiff `json:"changes,omitempty"`
	Resource     *CaCResource `json:"resource,omitempty"`
}

// FieldDiff describes a single field difference between repo and platform.
type FieldDiff struct {
	Field        string `json:"field"`
	RepoValue    string `json:"repo_value"`
	PlatformValue string `json:"platform_value"`
}

// DiffSummary aggregates diff plan statistics.
type DiffSummary struct {
	TotalResources int `json:"total_resources"`
	ToCreate       int `json:"to_create"`
	ToUpdate       int `json:"to_update"`
	ToDelete       int `json:"to_delete"`
	Unchanged      int `json:"unchanged"`
}

// ApplyResult summarises the outcome of applying a DiffPlan.
type ApplyResult struct {
	SyncRunID        uuid.UUID `json:"sync_run_id"`
	ResourcesAdded   int       `json:"resources_added"`
	ResourcesUpdated int       `json:"resources_updated"`
	ResourcesDeleted int       `json:"resources_deleted"`
	Errors           []string  `json:"errors,omitempty"`
}

// ValidationError describes a single validation failure for a CaC resource.
type ValidationError struct {
	ResourceUID string `json:"resource_uid"`
	ResourceName string `json:"resource_name"`
	Kind        string `json:"kind"`
	Field       string `json:"field"`
	Message     string `json:"message"`
	Severity    string `json:"severity"` // error, warning
}

// ============================================================
// REQUEST TYPES
// ============================================================

// CreateRepoRequest is the payload for creating a new CaC repository.
type CreateRepoRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	Provider        string `json:"provider"`
	RepoURL         string `json:"repo_url"`
	Branch          string `json:"branch"`
	BasePath        string `json:"base_path"`
	WebhookSecret   string `json:"webhook_secret"`
	AccessToken     string `json:"access_token"`
	SyncDirection   string `json:"sync_direction"`
	AutoSync        bool   `json:"auto_sync"`
	RequireApproval bool   `json:"require_approval"`
}

// UpdateRepoRequest is the payload for updating a CaC repository.
type UpdateRepoRequest struct {
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	Branch          *string `json:"branch"`
	BasePath        *string `json:"base_path"`
	WebhookSecret   *string `json:"webhook_secret"`
	AccessToken     *string `json:"access_token"`
	SyncDirection   *string `json:"sync_direction"`
	AutoSync        *bool   `json:"auto_sync"`
	RequireApproval *bool   `json:"require_approval"`
	IsActive        *bool   `json:"is_active"`
}

// TriggerSyncRequest is the payload for triggering a manual sync.
type TriggerSyncRequest struct {
	CommitSHA string `json:"commit_sha"`
	Branch    string `json:"branch"`
}

// ResolveDriftRequest is the payload for resolving a drift event.
type ResolveDriftRequest struct {
	ResolutionAction string `json:"resolution_action"`
	ResolutionNotes  string `json:"resolution_notes"`
}

// ValidateYAMLRequest wraps raw YAML content for validation.
type ValidateYAMLRequest struct {
	Content string `json:"content"`
}

// PlanRequest requests a diff plan for given YAML content.
type PlanRequest struct {
	RepositoryID uuid.UUID `json:"repository_id"`
	Content      string    `json:"content"`
}

// ApplyRequest requests application of a sync run's diff plan.
type ApplyRequest struct {
	SyncRunID uuid.UUID `json:"sync_run_id"`
}

// CaCRepoFilters holds query parameters for listing repositories.
type CaCRepoFilters struct {
	Provider string
	IsActive *bool
	Search   string
	Page     int
	PageSize int
}

// CaCDriftFilters holds query parameters for listing drift events.
type CaCDriftFilters struct {
	RepositoryID *uuid.UUID
	Direction    string
	Status       string
	Kind         string
	Page         int
	PageSize     int
}

// CaCMappingFilters holds query parameters for listing resource mappings.
type CaCMappingFilters struct {
	RepositoryID *uuid.UUID
	Kind         string
	Status       string
	Search       string
	Page         int
	PageSize     int
}

// ============================================================
// VALID RESOURCE KINDS
// ============================================================

var validCaCKinds = map[string]bool{
	"ControlImplementation": true,
	"Policy":                true,
	"RiskAcceptance":        true,
	"EvidenceConfig":        true,
	"Framework":             true,
	"RiskTreatment":         true,
	"AssetClassification":   true,
	"AuditSchedule":         true,
	"IncidentPlaybook":      true,
	"VendorAssessment":      true,
}

// kindToPlatformEntity maps CaC resource kinds to platform entity types.
var kindToPlatformEntity = map[string]string{
	"ControlImplementation": "control_implementation",
	"Policy":                "policy",
	"RiskAcceptance":        "risk_acceptance",
	"EvidenceConfig":        "evidence_config",
	"Framework":             "framework",
	"RiskTreatment":         "risk_treatment",
	"AssetClassification":   "asset_classification",
	"AuditSchedule":         "audit_schedule",
	"IncidentPlaybook":      "incident_playbook",
	"VendorAssessment":      "vendor_assessment",
}

// ============================================================
// CAC ENGINE
// ============================================================

// CaCEngine provides the compliance-as-code service layer,
// handling YAML parsing, validation, diffing, syncing, and
// drift detection between Git repositories and the platform.
type CaCEngine struct {
	pool *pgxpool.Pool
}

// NewCaCEngine creates a new CaCEngine backed by the given pool.
func NewCaCEngine(pool *pgxpool.Pool) *CaCEngine {
	return &CaCEngine{pool: pool}
}

// ============================================================
// YAML PARSING & VALIDATION
// ============================================================

// ParseYAML parses and validates a single compliance-as-code
// YAML document. It supports multi-document YAML by returning
// the first document found.
func (e *CaCEngine) ParseYAML(content []byte) (*CaCResource, error) {
	var resource CaCResource
	if err := yaml.Unmarshal(content, &resource); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	if resource.ApiVersion == "" {
		return nil, fmt.Errorf("missing required field: apiVersion")
	}
	if resource.Kind == "" {
		return nil, fmt.Errorf("missing required field: kind")
	}
	if resource.Metadata.Name == "" {
		return nil, fmt.Errorf("missing required field: metadata.name")
	}

	if !validCaCKinds[resource.Kind] {
		return nil, fmt.Errorf("unsupported resource kind: %s", resource.Kind)
	}

	// Generate UID from kind + name if not provided
	if resource.Metadata.UID == "" {
		resource.Metadata.UID = fmt.Sprintf("%s/%s",
			strings.ToLower(resource.Kind),
			strings.ToLower(strings.ReplaceAll(resource.Metadata.Name, " ", "-")))
	}

	return &resource, nil
}

// ParseMultiYAML parses a multi-document YAML byte stream and
// returns all successfully parsed CaC resources.
func (e *CaCEngine) ParseMultiYAML(content []byte) ([]CaCResource, []error) {
	docs := strings.Split(string(content), "\n---")
	var resources []CaCResource
	var errs []error

	for _, doc := range docs {
		trimmed := strings.TrimSpace(doc)
		if trimmed == "" || trimmed == "---" {
			continue
		}
		res, err := e.ParseYAML([]byte(trimmed))
		if err != nil {
			errs = append(errs, err)
			continue
		}
		resources = append(resources, *res)
	}

	return resources, errs
}

// ValidateResources performs cross-resource validation on a set
// of CaC resources, checking for duplicate UIDs, missing cross-
// references, and other inter-resource constraints.
func (e *CaCEngine) ValidateResources(resources []CaCResource) []ValidationError {
	var errs []ValidationError
	uidSeen := make(map[string]int)

	for i, r := range resources {
		uid := r.Metadata.UID
		if uid == "" {
			uid = fmt.Sprintf("%s/%s",
				strings.ToLower(r.Kind),
				strings.ToLower(strings.ReplaceAll(r.Metadata.Name, " ", "-")))
		}

		// Check for duplicate UIDs
		if prevIdx, exists := uidSeen[uid]; exists {
			errs = append(errs, ValidationError{
				ResourceUID:  uid,
				ResourceName: r.Metadata.Name,
				Kind:         r.Kind,
				Field:        "metadata.uid",
				Message:      fmt.Sprintf("duplicate resource UID %q (first seen at index %d)", uid, prevIdx),
				Severity:     "error",
			})
		}
		uidSeen[uid] = i

		// Validate required fields per kind
		kindErrors := validateKindSpec(r)
		errs = append(errs, kindErrors...)
	}

	return errs
}

// validateKindSpec checks kind-specific required fields in the spec.
func validateKindSpec(r CaCResource) []ValidationError {
	var errs []ValidationError

	uid := r.Metadata.UID
	if uid == "" {
		uid = fmt.Sprintf("%s/%s",
			strings.ToLower(r.Kind),
			strings.ToLower(strings.ReplaceAll(r.Metadata.Name, " ", "-")))
	}

	switch r.Kind {
	case "ControlImplementation":
		if _, ok := r.Spec["control_code"]; !ok {
			errs = append(errs, ValidationError{
				ResourceUID:  uid,
				ResourceName: r.Metadata.Name,
				Kind:         r.Kind,
				Field:        "spec.control_code",
				Message:      "ControlImplementation requires spec.control_code",
				Severity:     "error",
			})
		}
		if _, ok := r.Spec["status"]; !ok {
			errs = append(errs, ValidationError{
				ResourceUID:  uid,
				ResourceName: r.Metadata.Name,
				Kind:         r.Kind,
				Field:        "spec.status",
				Message:      "ControlImplementation requires spec.status",
				Severity:     "error",
			})
		}
	case "Policy":
		if _, ok := r.Spec["title"]; !ok {
			errs = append(errs, ValidationError{
				ResourceUID:  uid,
				ResourceName: r.Metadata.Name,
				Kind:         r.Kind,
				Field:        "spec.title",
				Message:      "Policy requires spec.title",
				Severity:     "error",
			})
		}
	case "RiskAcceptance":
		if _, ok := r.Spec["risk_id"]; !ok {
			errs = append(errs, ValidationError{
				ResourceUID:  uid,
				ResourceName: r.Metadata.Name,
				Kind:         r.Kind,
				Field:        "spec.risk_id",
				Message:      "RiskAcceptance requires spec.risk_id",
				Severity:     "error",
			})
		}
		if _, ok := r.Spec["accepted_by"]; !ok {
			errs = append(errs, ValidationError{
				ResourceUID:  uid,
				ResourceName: r.Metadata.Name,
				Kind:         r.Kind,
				Field:        "spec.accepted_by",
				Message:      "RiskAcceptance requires spec.accepted_by",
				Severity:     "error",
			})
		}
	case "EvidenceConfig":
		if _, ok := r.Spec["evidence_type"]; !ok {
			errs = append(errs, ValidationError{
				ResourceUID:  uid,
				ResourceName: r.Metadata.Name,
				Kind:         r.Kind,
				Field:        "spec.evidence_type",
				Message:      "EvidenceConfig requires spec.evidence_type",
				Severity:     "error",
			})
		}
		if _, ok := r.Spec["collection_method"]; !ok {
			errs = append(errs, ValidationError{
				ResourceUID:  uid,
				ResourceName: r.Metadata.Name,
				Kind:         r.Kind,
				Field:        "spec.collection_method",
				Message:      "EvidenceConfig requires spec.collection_method",
				Severity:     "error",
			})
		}
	}

	return errs
}

// ============================================================
// DIFF & APPLY
// ============================================================

// DiffWithPlatform compares a set of parsed CaC resources against
// the current platform state and produces a DiffPlan describing
// the actions needed to synchronise them.
func (e *CaCEngine) DiffWithPlatform(ctx context.Context, orgID, repoID uuid.UUID, resources []CaCResource) (*DiffPlan, error) {
	// Fetch existing resource mappings for this repo
	rows, err := e.pool.Query(ctx, `
		SELECT id, resource_uid, kind, resource_name, content_hash, status, file_path
		FROM cac_resource_mappings
		WHERE organization_id = $1 AND repository_id = $2`,
		orgID, repoID)
	if err != nil {
		return nil, fmt.Errorf("query existing mappings: %w", err)
	}
	defer rows.Close()

	type existingMapping struct {
		ID           uuid.UUID
		ResourceUID  string
		Kind         string
		ResourceName string
		ContentHash  string
		Status       string
		FilePath     string
	}
	existing := make(map[string]existingMapping)
	for rows.Next() {
		var m existingMapping
		if err := rows.Scan(&m.ID, &m.ResourceUID, &m.Kind, &m.ResourceName, &m.ContentHash, &m.Status, &m.FilePath); err != nil {
			return nil, fmt.Errorf("scan mapping: %w", err)
		}
		existing[m.ResourceUID] = m
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mappings: %w", err)
	}

	plan := &DiffPlan{
		RepositoryID: repoID,
	}

	// Track which existing resources are seen
	seen := make(map[string]bool)

	for _, r := range resources {
		uid := r.Metadata.UID
		if uid == "" {
			uid = fmt.Sprintf("%s/%s",
				strings.ToLower(r.Kind),
				strings.ToLower(strings.ReplaceAll(r.Metadata.Name, " ", "-")))
		}
		seen[uid] = true

		contentHash := computeResourceHash(r)

		if ex, exists := existing[uid]; exists {
			// Resource exists — check for changes
			if ex.ContentHash == contentHash {
				plan.Actions = append(plan.Actions, DiffAction{
					Action:       "no_change",
					Kind:         r.Kind,
					ResourceUID:  uid,
					ResourceName: r.Metadata.Name,
					FilePath:     ex.FilePath,
					Resource:     &r,
				})
				plan.Summary.Unchanged++
			} else {
				plan.Actions = append(plan.Actions, DiffAction{
					Action:       "update",
					Kind:         r.Kind,
					ResourceUID:  uid,
					ResourceName: r.Metadata.Name,
					FilePath:     ex.FilePath,
					Changes: []FieldDiff{
						{
							Field:         "content_hash",
							RepoValue:     contentHash,
							PlatformValue: ex.ContentHash,
						},
					},
					Resource: &r,
				})
				plan.Summary.ToUpdate++
			}
		} else {
			// New resource
			plan.Actions = append(plan.Actions, DiffAction{
				Action:       "create",
				Kind:         r.Kind,
				ResourceUID:  uid,
				ResourceName: r.Metadata.Name,
				FilePath:     buildFilePath(r),
				Resource:     &r,
			})
			plan.Summary.ToCreate++
		}
	}

	// Find orphaned resources (in platform but not in repo)
	for uid, ex := range existing {
		if !seen[uid] {
			plan.Actions = append(plan.Actions, DiffAction{
				Action:       "delete",
				Kind:         ex.Kind,
				ResourceUID:  uid,
				ResourceName: ex.ResourceName,
				FilePath:     ex.FilePath,
			})
			plan.Summary.ToDelete++
		}
	}

	plan.Summary.TotalResources = len(plan.Actions)

	return plan, nil
}

// ApplyChanges applies a DiffPlan's actions within a sync run,
// creating/updating/deleting resource mappings as needed.
func (e *CaCEngine) ApplyChanges(ctx context.Context, orgID, syncRunID uuid.UUID, plan *DiffPlan) (*ApplyResult, error) {
	result := &ApplyResult{SyncRunID: syncRunID}

	tx, err := e.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set RLS context
	if _, err := tx.Exec(ctx, "SET LOCAL app.current_org = $1", orgID.String()); err != nil {
		return nil, fmt.Errorf("set RLS context: %w", err)
	}

	for _, action := range plan.Actions {
		switch action.Action {
		case "create":
			entityType := kindToPlatformEntity[action.Kind]
			contentHash := ""
			if action.Resource != nil {
				contentHash = computeResourceHash(*action.Resource)
			}
			now := time.Now().UTC()
			_, err := tx.Exec(ctx, `
				INSERT INTO cac_resource_mappings
					(organization_id, repository_id, file_path, api_version, kind,
					 resource_name, resource_uid, platform_entity_type, status,
					 content_hash, last_synced_at, last_sync_run_id)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'synced', $9, $10, $11)`,
				orgID, plan.RepositoryID, action.FilePath,
				"complianceforge.io/v1", action.Kind,
				action.ResourceName, action.ResourceUID, entityType,
				contentHash, now, syncRunID)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("create %s: %v", action.ResourceUID, err))
				continue
			}
			result.ResourcesAdded++

		case "update":
			contentHash := ""
			if action.Resource != nil {
				contentHash = computeResourceHash(*action.Resource)
			}
			now := time.Now().UTC()
			_, err := tx.Exec(ctx, `
				UPDATE cac_resource_mappings
				SET content_hash = $1, status = 'synced', last_synced_at = $2, last_sync_run_id = $3
				WHERE organization_id = $4 AND repository_id = $5 AND resource_uid = $6`,
				contentHash, now, syncRunID,
				orgID, plan.RepositoryID, action.ResourceUID)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("update %s: %v", action.ResourceUID, err))
				continue
			}
			result.ResourcesUpdated++

		case "delete":
			_, err := tx.Exec(ctx, `
				UPDATE cac_resource_mappings
				SET status = 'orphaned'
				WHERE organization_id = $1 AND repository_id = $2 AND resource_uid = $3`,
				orgID, plan.RepositoryID, action.ResourceUID)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("delete %s: %v", action.ResourceUID, err))
				continue
			}
			result.ResourcesDeleted++
		}
	}

	// Update sync run with results
	errJSON, _ := json.Marshal(result.Errors)
	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `
		UPDATE cac_sync_runs
		SET status = 'completed', resources_added = $1, resources_updated = $2,
		    resources_deleted = $3, resources_unchanged = $4, errors = $5, completed_at = $6
		WHERE id = $7 AND organization_id = $8`,
		result.ResourcesAdded, result.ResourcesUpdated,
		result.ResourcesDeleted, plan.Summary.Unchanged,
		errJSON, now, syncRunID, orgID)
	if err != nil {
		return nil, fmt.Errorf("update sync run: %w", err)
	}

	// Update repository resource count and last sync
	_, err = tx.Exec(ctx, `
		UPDATE cac_repositories
		SET resource_count = (
			SELECT COUNT(*) FROM cac_resource_mappings
			WHERE repository_id = $1 AND organization_id = $2 AND status != 'orphaned'
		), last_sync_at = $3, last_sync_status = 'completed'
		WHERE id = $1 AND organization_id = $2`,
		plan.RepositoryID, orgID, now)
	if err != nil {
		log.Warn().Err(err).Msg("failed to update repository resource count")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return result, nil
}

// ============================================================
// DRIFT DETECTION
// ============================================================

// DetectDrift compares the current resource mappings against
// the platform state and creates drift events for any discrepancies.
func (e *CaCEngine) DetectDrift(ctx context.Context, orgID, repoID uuid.UUID) ([]CaCDriftEvent, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT id, resource_uid, kind, resource_name, content_hash, status
		FROM cac_resource_mappings
		WHERE organization_id = $1 AND repository_id = $2 AND status = 'synced'`,
		orgID, repoID)
	if err != nil {
		return nil, fmt.Errorf("query resource mappings: %w", err)
	}
	defer rows.Close()

	var driftEvents []CaCDriftEvent

	for rows.Next() {
		var mappingID uuid.UUID
		var resourceUID, kind, resourceName, contentHash, status string
		if err := rows.Scan(&mappingID, &resourceUID, &kind, &resourceName, &contentHash, &status); err != nil {
			log.Warn().Err(err).Msg("scan mapping for drift detection")
			continue
		}

		// Check if the platform entity still exists and matches
		// For now, we create drift events for resources in 'synced' state
		// that haven't been checked recently
		var lastDriftCheck time.Time
		err := e.pool.QueryRow(ctx, `
			SELECT COALESCE(MAX(detected_at), '1970-01-01'::timestamptz)
			FROM cac_drift_events
			WHERE organization_id = $1 AND repository_id = $2
			  AND resource_uid = $3 AND status = 'open'`,
			orgID, repoID, resourceUID).Scan(&lastDriftCheck)
		if err != nil {
			log.Warn().Err(err).Str("uid", resourceUID).Msg("check last drift")
			continue
		}

		// Skip if there's already an open drift event for this resource
		if lastDriftCheck.After(time.Date(1970, 1, 2, 0, 0, 0, 0, time.UTC)) {
			continue
		}

		// Simulate platform-side check: compare content hash with latest
		// (in production this would compare the actual platform entity state)
		drift := CaCDriftEvent{
			OrganizationID:    orgID,
			RepositoryID:      repoID,
			ResourceMappingID: &mappingID,
			DriftDirection:    "platform_ahead",
			Status:            "open",
			Kind:              kind,
			ResourceName:      resourceName,
			ResourceUID:       resourceUID,
			Description:       fmt.Sprintf("Platform state may have diverged from repository for %s/%s", kind, resourceName),
			DetectedAt:        time.Now().UTC(),
		}
		driftEvents = append(driftEvents, drift)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mappings: %w", err)
	}

	// Persist drift events
	for i := range driftEvents {
		var id uuid.UUID
		err := e.pool.QueryRow(ctx, `
			INSERT INTO cac_drift_events
				(organization_id, repository_id, resource_mapping_id, drift_direction,
				 status, kind, resource_name, resource_uid, field_path,
				 repo_value, platform_value, description, detected_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			RETURNING id`,
			driftEvents[i].OrganizationID, driftEvents[i].RepositoryID,
			driftEvents[i].ResourceMappingID, driftEvents[i].DriftDirection,
			driftEvents[i].Status, driftEvents[i].Kind,
			driftEvents[i].ResourceName, driftEvents[i].ResourceUID,
			driftEvents[i].FieldPath, driftEvents[i].RepoValue,
			driftEvents[i].PlatformValue, driftEvents[i].Description,
			driftEvents[i].DetectedAt).Scan(&id)
		if err != nil {
			log.Warn().Err(err).Str("uid", driftEvents[i].ResourceUID).Msg("persist drift event")
			continue
		}
		driftEvents[i].ID = id
	}

	return driftEvents, nil
}

// ============================================================
// EXPORT
// ============================================================

// ExportAsYAML exports all synced resource mappings for an organisation
// as a slice of CaCResource structs suitable for YAML serialisation.
func (e *CaCEngine) ExportAsYAML(ctx context.Context, orgID uuid.UUID) ([]CaCResource, error) {
	rows, err := e.pool.Query(ctx, `
		SELECT api_version, kind, resource_name, resource_uid, metadata
		FROM cac_resource_mappings
		WHERE organization_id = $1 AND status IN ('synced', 'pending')
		ORDER BY kind, resource_name`,
		orgID)
	if err != nil {
		return nil, fmt.Errorf("query resource mappings: %w", err)
	}
	defer rows.Close()

	var resources []CaCResource
	for rows.Next() {
		var apiVersion, kind, name, uid string
		var metaJSON json.RawMessage
		if err := rows.Scan(&apiVersion, &kind, &name, &uid, &metaJSON); err != nil {
			log.Warn().Err(err).Msg("scan mapping for export")
			continue
		}

		var spec map[string]interface{}
		if len(metaJSON) > 2 { // not just "{}"
			_ = json.Unmarshal(metaJSON, &spec)
		}
		if spec == nil {
			spec = make(map[string]interface{})
		}

		resources = append(resources, CaCResource{
			ApiVersion: apiVersion,
			Kind:       kind,
			Metadata: CaCResourceMetadata{
				Name: name,
				UID:  uid,
			},
			Spec: spec,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mappings: %w", err)
	}

	return resources, nil
}

// ============================================================
// REPOSITORY CRUD
// ============================================================

// ListRepositories returns a paginated list of CaC repositories for an org.
func (e *CaCEngine) ListRepositories(ctx context.Context, orgID uuid.UUID, filters CaCRepoFilters) ([]CaCRepository, int, error) {
	where := "WHERE organization_id = $1"
	args := []interface{}{orgID}
	argIdx := 2

	if filters.Provider != "" {
		where += fmt.Sprintf(" AND provider = $%d", argIdx)
		args = append(args, filters.Provider)
		argIdx++
	}
	if filters.IsActive != nil {
		where += fmt.Sprintf(" AND is_active = $%d", argIdx)
		args = append(args, *filters.IsActive)
		argIdx++
	}
	if filters.Search != "" {
		where += fmt.Sprintf(" AND (name ILIKE $%d OR repo_url ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filters.Search+"%")
		argIdx++
	}

	var total int
	err := e.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM cac_repositories "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count repositories: %w", err)
	}

	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 20
	}
	offset := (filters.Page - 1) * filters.PageSize
	query := fmt.Sprintf(`
		SELECT id, organization_id, name, description, provider, repo_url, branch,
		       base_path, sync_direction, auto_sync, require_approval,
		       last_sync_at, last_sync_status, resource_count, is_active,
		       metadata, created_by, created_at, updated_at
		FROM cac_repositories %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, filters.PageSize, offset)

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query repositories: %w", err)
	}
	defer rows.Close()

	var repos []CaCRepository
	for rows.Next() {
		var r CaCRepository
		if err := rows.Scan(
			&r.ID, &r.OrganizationID, &r.Name, &r.Description,
			&r.Provider, &r.RepoURL, &r.Branch, &r.BasePath,
			&r.SyncDirection, &r.AutoSync, &r.RequireApproval,
			&r.LastSyncAt, &r.LastSyncStatus, &r.ResourceCount,
			&r.IsActive, &r.Metadata, &r.CreatedBy,
			&r.CreatedAt, &r.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan repository: %w", err)
		}
		repos = append(repos, r)
	}

	return repos, total, nil
}

// CreateRepository creates a new CaC repository record.
func (e *CaCEngine) CreateRepository(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, req CreateRepoRequest) (*CaCRepository, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.RepoURL == "" {
		return nil, fmt.Errorf("repo_url is required")
	}
	if req.Provider == "" {
		req.Provider = "github"
	}
	if req.Branch == "" {
		req.Branch = "main"
	}
	if req.BasePath == "" {
		req.BasePath = "/"
	}
	if req.SyncDirection == "" {
		req.SyncDirection = "pull"
	}

	var repo CaCRepository
	err := e.pool.QueryRow(ctx, `
		INSERT INTO cac_repositories
			(organization_id, name, description, provider, repo_url, branch,
			 base_path, webhook_secret, access_token, sync_direction,
			 auto_sync, require_approval, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, organization_id, name, description, provider, repo_url,
		          branch, base_path, sync_direction, auto_sync, require_approval,
		          resource_count, is_active, metadata, created_by, created_at, updated_at`,
		orgID, req.Name, req.Description, req.Provider, req.RepoURL,
		req.Branch, req.BasePath, req.WebhookSecret, req.AccessToken,
		req.SyncDirection, req.AutoSync, req.RequireApproval, userID,
	).Scan(
		&repo.ID, &repo.OrganizationID, &repo.Name, &repo.Description,
		&repo.Provider, &repo.RepoURL, &repo.Branch, &repo.BasePath,
		&repo.SyncDirection, &repo.AutoSync, &repo.RequireApproval,
		&repo.ResourceCount, &repo.IsActive, &repo.Metadata,
		&repo.CreatedBy, &repo.CreatedAt, &repo.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert repository: %w", err)
	}

	log.Info().
		Str("repo_id", repo.ID.String()).
		Str("name", repo.Name).
		Str("provider", repo.Provider).
		Msg("CaC repository created")

	return &repo, nil
}

// UpdateRepository updates an existing CaC repository.
func (e *CaCEngine) UpdateRepository(ctx context.Context, orgID, repoID uuid.UUID, req UpdateRepoRequest) (*CaCRepository, error) {
	setClauses := []string{}
	args := []interface{}{orgID, repoID}
	argIdx := 3

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.Branch != nil {
		setClauses = append(setClauses, fmt.Sprintf("branch = $%d", argIdx))
		args = append(args, *req.Branch)
		argIdx++
	}
	if req.BasePath != nil {
		setClauses = append(setClauses, fmt.Sprintf("base_path = $%d", argIdx))
		args = append(args, *req.BasePath)
		argIdx++
	}
	if req.WebhookSecret != nil {
		setClauses = append(setClauses, fmt.Sprintf("webhook_secret = $%d", argIdx))
		args = append(args, *req.WebhookSecret)
		argIdx++
	}
	if req.AccessToken != nil {
		setClauses = append(setClauses, fmt.Sprintf("access_token = $%d", argIdx))
		args = append(args, *req.AccessToken)
		argIdx++
	}
	if req.SyncDirection != nil {
		setClauses = append(setClauses, fmt.Sprintf("sync_direction = $%d", argIdx))
		args = append(args, *req.SyncDirection)
		argIdx++
	}
	if req.AutoSync != nil {
		setClauses = append(setClauses, fmt.Sprintf("auto_sync = $%d", argIdx))
		args = append(args, *req.AutoSync)
		argIdx++
	}
	if req.RequireApproval != nil {
		setClauses = append(setClauses, fmt.Sprintf("require_approval = $%d", argIdx))
		args = append(args, *req.RequireApproval)
		argIdx++
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf(`
		UPDATE cac_repositories SET %s
		WHERE organization_id = $1 AND id = $2
		RETURNING id, organization_id, name, description, provider, repo_url,
		          branch, base_path, sync_direction, auto_sync, require_approval,
		          last_sync_at, last_sync_status, resource_count, is_active,
		          metadata, created_by, created_at, updated_at`,
		strings.Join(setClauses, ", "))

	var repo CaCRepository
	err := e.pool.QueryRow(ctx, query, args...).Scan(
		&repo.ID, &repo.OrganizationID, &repo.Name, &repo.Description,
		&repo.Provider, &repo.RepoURL, &repo.Branch, &repo.BasePath,
		&repo.SyncDirection, &repo.AutoSync, &repo.RequireApproval,
		&repo.LastSyncAt, &repo.LastSyncStatus, &repo.ResourceCount,
		&repo.IsActive, &repo.Metadata, &repo.CreatedBy,
		&repo.CreatedAt, &repo.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("repository not found")
		}
		return nil, fmt.Errorf("update repository: %w", err)
	}

	return &repo, nil
}

// DeleteRepository soft-deletes a CaC repository by deactivating it.
func (e *CaCEngine) DeleteRepository(ctx context.Context, orgID, repoID uuid.UUID) error {
	tag, err := e.pool.Exec(ctx, `
		UPDATE cac_repositories SET is_active = false
		WHERE organization_id = $1 AND id = $2 AND is_active = true`,
		orgID, repoID)
	if err != nil {
		return fmt.Errorf("delete repository: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("repository not found")
	}
	return nil
}

// GetRepository returns a single repository by ID.
func (e *CaCEngine) GetRepository(ctx context.Context, orgID, repoID uuid.UUID) (*CaCRepository, error) {
	var r CaCRepository
	err := e.pool.QueryRow(ctx, `
		SELECT id, organization_id, name, description, provider, repo_url,
		       branch, base_path, sync_direction, auto_sync, require_approval,
		       last_sync_at, last_sync_status, resource_count, is_active,
		       metadata, created_by, created_at, updated_at
		FROM cac_repositories
		WHERE organization_id = $1 AND id = $2`,
		orgID, repoID).Scan(
		&r.ID, &r.OrganizationID, &r.Name, &r.Description,
		&r.Provider, &r.RepoURL, &r.Branch, &r.BasePath,
		&r.SyncDirection, &r.AutoSync, &r.RequireApproval,
		&r.LastSyncAt, &r.LastSyncStatus, &r.ResourceCount,
		&r.IsActive, &r.Metadata, &r.CreatedBy,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("repository not found")
		}
		return nil, fmt.Errorf("get repository: %w", err)
	}
	return &r, nil
}

// ============================================================
// SYNC RUNS
// ============================================================

// TriggerSync creates a new sync run for a repository.
func (e *CaCEngine) TriggerSync(ctx context.Context, orgID, repoID, userID uuid.UUID, req TriggerSyncRequest) (*CaCSyncRun, error) {
	// Verify repository exists and is active
	repo, err := e.GetRepository(ctx, orgID, repoID)
	if err != nil {
		return nil, err
	}
	if !repo.IsActive {
		return nil, fmt.Errorf("repository is not active")
	}

	now := time.Now().UTC()
	branch := req.Branch
	if branch == "" {
		branch = repo.Branch
	}

	var run CaCSyncRun
	err = e.pool.QueryRow(ctx, `
		INSERT INTO cac_sync_runs
			(organization_id, repository_id, status, direction, trigger_type,
			 commit_sha, branch, started_at, triggered_by)
		VALUES ($1, $2, 'pending', $3, 'manual', $4, $5, $6, $7)
		RETURNING id, organization_id, repository_id, status, direction,
		          trigger_type, commit_sha, branch, files_changed,
		          resources_added, resources_updated, resources_deleted,
		          resources_unchanged, errors, diff_plan, started_at,
		          completed_at, approved_by, approved_at, rejected_by,
		          rejected_at, rejection_reason, triggered_by,
		          created_at, updated_at`,
		orgID, repoID, repo.SyncDirection, req.CommitSHA, branch, now, userID,
	).Scan(
		&run.ID, &run.OrganizationID, &run.RepositoryID,
		&run.Status, &run.Direction, &run.TriggerType,
		&run.CommitSHA, &run.Branch, &run.FilesChanged,
		&run.ResourcesAdded, &run.ResourcesUpdated,
		&run.ResourcesDeleted, &run.ResourcesUnchanged,
		&run.Errors, &run.DiffPlanJSON, &run.StartedAt,
		&run.CompletedAt, &run.ApprovedBy, &run.ApprovedAt,
		&run.RejectedBy, &run.RejectedAt, &run.RejectionReason,
		&run.TriggeredBy, &run.CreatedAt, &run.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create sync run: %w", err)
	}

	log.Info().
		Str("sync_run_id", run.ID.String()).
		Str("repo_id", repoID.String()).
		Msg("CaC sync run triggered")

	return &run, nil
}

// GetSyncRuns returns a paginated list of sync runs for an org.
func (e *CaCEngine) GetSyncRuns(ctx context.Context, orgID uuid.UUID, repoID *uuid.UUID, page, pageSize int) ([]CaCSyncRun, int, error) {
	where := "WHERE sr.organization_id = $1"
	args := []interface{}{orgID}
	argIdx := 2

	if repoID != nil {
		where += fmt.Sprintf(" AND sr.repository_id = $%d", argIdx)
		args = append(args, *repoID)
		argIdx++
	}

	var total int
	err := e.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM cac_sync_runs sr "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count sync runs: %w", err)
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
		SELECT sr.id, sr.organization_id, sr.repository_id, sr.status, sr.direction,
		       sr.trigger_type, sr.commit_sha, sr.branch, sr.files_changed,
		       sr.resources_added, sr.resources_updated, sr.resources_deleted,
		       sr.resources_unchanged, sr.errors, sr.diff_plan, sr.started_at,
		       sr.completed_at, sr.approved_by, sr.approved_at, sr.rejected_by,
		       sr.rejected_at, sr.rejection_reason, sr.triggered_by,
		       sr.created_at, sr.updated_at,
		       COALESCE(r.name, '') as repository_name
		FROM cac_sync_runs sr
		LEFT JOIN cac_repositories r ON r.id = sr.repository_id
		%s
		ORDER BY sr.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query sync runs: %w", err)
	}
	defer rows.Close()

	var runs []CaCSyncRun
	for rows.Next() {
		var r CaCSyncRun
		if err := rows.Scan(
			&r.ID, &r.OrganizationID, &r.RepositoryID,
			&r.Status, &r.Direction, &r.TriggerType,
			&r.CommitSHA, &r.Branch, &r.FilesChanged,
			&r.ResourcesAdded, &r.ResourcesUpdated,
			&r.ResourcesDeleted, &r.ResourcesUnchanged,
			&r.Errors, &r.DiffPlanJSON, &r.StartedAt,
			&r.CompletedAt, &r.ApprovedBy, &r.ApprovedAt,
			&r.RejectedBy, &r.RejectedAt, &r.RejectionReason,
			&r.TriggeredBy, &r.CreatedAt, &r.UpdatedAt,
			&r.RepositoryName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan sync run: %w", err)
		}
		runs = append(runs, r)
	}

	return runs, total, nil
}

// GetSyncRun returns a single sync run by ID.
func (e *CaCEngine) GetSyncRun(ctx context.Context, orgID, syncRunID uuid.UUID) (*CaCSyncRun, error) {
	var r CaCSyncRun
	err := e.pool.QueryRow(ctx, `
		SELECT sr.id, sr.organization_id, sr.repository_id, sr.status, sr.direction,
		       sr.trigger_type, sr.commit_sha, sr.branch, sr.files_changed,
		       sr.resources_added, sr.resources_updated, sr.resources_deleted,
		       sr.resources_unchanged, sr.errors, sr.diff_plan, sr.started_at,
		       sr.completed_at, sr.approved_by, sr.approved_at, sr.rejected_by,
		       sr.rejected_at, sr.rejection_reason, sr.triggered_by,
		       sr.created_at, sr.updated_at,
		       COALESCE(rp.name, '') as repository_name
		FROM cac_sync_runs sr
		LEFT JOIN cac_repositories rp ON rp.id = sr.repository_id
		WHERE sr.organization_id = $1 AND sr.id = $2`,
		orgID, syncRunID).Scan(
		&r.ID, &r.OrganizationID, &r.RepositoryID,
		&r.Status, &r.Direction, &r.TriggerType,
		&r.CommitSHA, &r.Branch, &r.FilesChanged,
		&r.ResourcesAdded, &r.ResourcesUpdated,
		&r.ResourcesDeleted, &r.ResourcesUnchanged,
		&r.Errors, &r.DiffPlanJSON, &r.StartedAt,
		&r.CompletedAt, &r.ApprovedBy, &r.ApprovedAt,
		&r.RejectedBy, &r.RejectedAt, &r.RejectionReason,
		&r.TriggeredBy, &r.CreatedAt, &r.UpdatedAt,
		&r.RepositoryName,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("sync run not found")
		}
		return nil, fmt.Errorf("get sync run: %w", err)
	}
	return &r, nil
}

// ApproveSyncRun marks a sync run as approved.
func (e *CaCEngine) ApproveSyncRun(ctx context.Context, orgID, syncRunID, userID uuid.UUID) (*CaCSyncRun, error) {
	now := time.Now().UTC()
	tag, err := e.pool.Exec(ctx, `
		UPDATE cac_sync_runs
		SET status = 'approved', approved_by = $1, approved_at = $2
		WHERE organization_id = $3 AND id = $4 AND status = 'awaiting_approval'`,
		userID, now, orgID, syncRunID)
	if err != nil {
		return nil, fmt.Errorf("approve sync run: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("sync run not found or not awaiting approval")
	}
	return e.GetSyncRun(ctx, orgID, syncRunID)
}

// RejectSyncRun marks a sync run as rejected.
func (e *CaCEngine) RejectSyncRun(ctx context.Context, orgID, syncRunID, userID uuid.UUID, reason string) (*CaCSyncRun, error) {
	now := time.Now().UTC()
	tag, err := e.pool.Exec(ctx, `
		UPDATE cac_sync_runs
		SET status = 'rejected', rejected_by = $1, rejected_at = $2, rejection_reason = $3
		WHERE organization_id = $4 AND id = $5 AND status = 'awaiting_approval'`,
		userID, now, reason, orgID, syncRunID)
	if err != nil {
		return nil, fmt.Errorf("reject sync run: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, fmt.Errorf("sync run not found or not awaiting approval")
	}
	return e.GetSyncRun(ctx, orgID, syncRunID)
}

// ============================================================
// DRIFT MANAGEMENT
// ============================================================

// ListDriftEvents returns a paginated list of drift events for an org.
func (e *CaCEngine) ListDriftEvents(ctx context.Context, orgID uuid.UUID, filters CaCDriftFilters) ([]CaCDriftEvent, int, error) {
	where := "WHERE de.organization_id = $1"
	args := []interface{}{orgID}
	argIdx := 2

	if filters.RepositoryID != nil {
		where += fmt.Sprintf(" AND de.repository_id = $%d", argIdx)
		args = append(args, *filters.RepositoryID)
		argIdx++
	}
	if filters.Direction != "" {
		where += fmt.Sprintf(" AND de.drift_direction = $%d", argIdx)
		args = append(args, filters.Direction)
		argIdx++
	}
	if filters.Status != "" {
		where += fmt.Sprintf(" AND de.status = $%d", argIdx)
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.Kind != "" {
		where += fmt.Sprintf(" AND de.kind = $%d", argIdx)
		args = append(args, filters.Kind)
		argIdx++
	}

	var total int
	err := e.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM cac_drift_events de "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count drift events: %w", err)
	}

	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 20
	}
	offset := (filters.Page - 1) * filters.PageSize

	query := fmt.Sprintf(`
		SELECT de.id, de.organization_id, de.repository_id, de.resource_mapping_id,
		       de.drift_direction, de.status, de.kind, de.resource_name,
		       de.resource_uid, de.field_path, de.repo_value, de.platform_value,
		       de.description, de.detected_at, de.resolved_at, de.resolved_by,
		       de.resolution_action, de.resolution_notes,
		       de.created_at, de.updated_at,
		       COALESCE(r.name, '') as repository_name
		FROM cac_drift_events de
		LEFT JOIN cac_repositories r ON r.id = de.repository_id
		%s
		ORDER BY de.detected_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, filters.PageSize, offset)

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query drift events: %w", err)
	}
	defer rows.Close()

	var events []CaCDriftEvent
	for rows.Next() {
		var ev CaCDriftEvent
		if err := rows.Scan(
			&ev.ID, &ev.OrganizationID, &ev.RepositoryID,
			&ev.ResourceMappingID, &ev.DriftDirection, &ev.Status,
			&ev.Kind, &ev.ResourceName, &ev.ResourceUID,
			&ev.FieldPath, &ev.RepoValue, &ev.PlatformValue,
			&ev.Description, &ev.DetectedAt, &ev.ResolvedAt,
			&ev.ResolvedBy, &ev.ResolutionAction, &ev.ResolutionNotes,
			&ev.CreatedAt, &ev.UpdatedAt,
			&ev.RepositoryName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan drift event: %w", err)
		}
		events = append(events, ev)
	}

	return events, total, nil
}

// ResolveDrift marks a drift event as resolved.
func (e *CaCEngine) ResolveDrift(ctx context.Context, orgID, driftID, userID uuid.UUID, req ResolveDriftRequest) error {
	now := time.Now().UTC()
	tag, err := e.pool.Exec(ctx, `
		UPDATE cac_drift_events
		SET status = 'resolved', resolved_at = $1, resolved_by = $2,
		    resolution_action = $3, resolution_notes = $4
		WHERE organization_id = $5 AND id = $6 AND status IN ('open', 'acknowledged')`,
		now, userID, req.ResolutionAction, req.ResolutionNotes, orgID, driftID)
	if err != nil {
		return fmt.Errorf("resolve drift: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("drift event not found or already resolved")
	}
	return nil
}

// ============================================================
// RESOURCE MAPPINGS
// ============================================================

// ListResourceMappings returns a paginated list of resource mappings.
func (e *CaCEngine) ListResourceMappings(ctx context.Context, orgID uuid.UUID, filters CaCMappingFilters) ([]CaCResourceMapping, int, error) {
	where := "WHERE rm.organization_id = $1"
	args := []interface{}{orgID}
	argIdx := 2

	if filters.RepositoryID != nil {
		where += fmt.Sprintf(" AND rm.repository_id = $%d", argIdx)
		args = append(args, *filters.RepositoryID)
		argIdx++
	}
	if filters.Kind != "" {
		where += fmt.Sprintf(" AND rm.kind = $%d", argIdx)
		args = append(args, filters.Kind)
		argIdx++
	}
	if filters.Status != "" {
		where += fmt.Sprintf(" AND rm.status = $%d", argIdx)
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.Search != "" {
		where += fmt.Sprintf(" AND (rm.resource_name ILIKE $%d OR rm.resource_uid ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+filters.Search+"%")
		argIdx++
	}

	var total int
	err := e.pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM cac_resource_mappings rm "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count resource mappings: %w", err)
	}

	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 || filters.PageSize > 100 {
		filters.PageSize = 20
	}
	offset := (filters.Page - 1) * filters.PageSize

	query := fmt.Sprintf(`
		SELECT rm.id, rm.organization_id, rm.repository_id, rm.file_path,
		       rm.api_version, rm.kind, rm.resource_name, rm.resource_uid,
		       rm.platform_entity_type, rm.platform_entity_id, rm.status,
		       rm.content_hash, rm.last_synced_at, rm.last_sync_run_id,
		       rm.metadata, rm.created_at, rm.updated_at,
		       COALESCE(r.name, '') as repository_name
		FROM cac_resource_mappings rm
		LEFT JOIN cac_repositories r ON r.id = rm.repository_id
		%s
		ORDER BY rm.kind, rm.resource_name
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)
	args = append(args, filters.PageSize, offset)

	rows, err := e.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query resource mappings: %w", err)
	}
	defer rows.Close()

	var mappings []CaCResourceMapping
	for rows.Next() {
		var m CaCResourceMapping
		if err := rows.Scan(
			&m.ID, &m.OrganizationID, &m.RepositoryID, &m.FilePath,
			&m.ApiVersion, &m.Kind, &m.ResourceName, &m.ResourceUID,
			&m.PlatformEntityType, &m.PlatformEntityID, &m.Status,
			&m.ContentHash, &m.LastSyncedAt, &m.LastSyncRunID,
			&m.Metadata, &m.CreatedAt, &m.UpdatedAt,
			&m.RepositoryName,
		); err != nil {
			return nil, 0, fmt.Errorf("scan resource mapping: %w", err)
		}
		mappings = append(mappings, m)
	}

	return mappings, total, nil
}

// ============================================================
// WEBHOOK SIGNATURE VERIFICATION
// ============================================================

// ValidateWebhookSignature verifies a GitHub-style webhook signature.
// The signature should be in the format "sha256=<hex>".
func (e *CaCEngine) ValidateWebhookSignature(payload []byte, secret, signature string) bool {
	if secret == "" || signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSig := "sha256=" + hex.EncodeToString(expectedMAC)

	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// ============================================================
// INTERNAL HELPERS
// ============================================================

// computeResourceHash returns a deterministic SHA-256 hash of a CaC resource
// for use in content comparison / change detection.
func computeResourceHash(r CaCResource) string {
	data, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// buildFilePath constructs a default file path for a new CaC resource.
func buildFilePath(r CaCResource) string {
	namespace := r.Metadata.Namespace
	if namespace == "" {
		namespace = "default"
	}
	name := strings.ToLower(strings.ReplaceAll(r.Metadata.Name, " ", "-"))
	kind := strings.ToLower(r.Kind)
	return fmt.Sprintf("%s/%s/%s.yaml", namespace, kind, name)
}
