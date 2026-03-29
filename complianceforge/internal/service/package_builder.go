package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// ============================================================
// PACKAGE BUILDER
// Exports an organisation's control library, policies, or
// templates as a marketplace-ready package. Strips organisation-
// specific data (org IDs, user references, internal notes) and
// produces a portable PackageData payload.
// ============================================================

// PackageBuilder exports organisation assets as marketplace packages.
type PackageBuilder struct {
	pool *pgxpool.Pool
}

// NewPackageBuilder creates a new PackageBuilder.
func NewPackageBuilder(pool *pgxpool.Pool) *PackageBuilder {
	return &PackageBuilder{pool: pool}
}

// ExportConfig defines what to export and how to package it.
type ExportConfig struct {
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	PackageType       string    `json:"package_type"`       // control_pack, policy_bundle, etc.
	Category          string    `json:"category"`
	FrameworkIDs      []uuid.UUID `json:"framework_ids"`    // which adopted frameworks to export
	ControlIDs        []uuid.UUID `json:"control_ids"`      // specific controls (empty = all for frameworks)
	IncludeMappings   bool      `json:"include_mappings"`
	IncludeGuidance   bool      `json:"include_guidance"`
	IncludeEvidence   bool      `json:"include_evidence_templates"`
	StripInternalNotes bool     `json:"strip_internal_notes"`
	Regions           []string  `json:"regions"`
	Industries        []string  `json:"industries"`
	Tags              []string  `json:"tags"`
}

// PackageData is the portable, marketplace-ready export payload.
type PackageData struct {
	SchemaVersion string                   `json:"schema_version"`
	PackageType   string                   `json:"package_type"`
	ExportedAt    string                   `json:"exported_at"`
	Controls      []ExportedControl        `json:"controls,omitempty"`
	Policies      []ExportedPolicy         `json:"policies,omitempty"`
	Templates     []ExportedTemplate       `json:"templates,omitempty"`
	Metadata      ExportMetadata           `json:"metadata"`
	Hash          string                   `json:"hash"`
	SizeBytes     int64                    `json:"size_bytes"`
}

// ExportedControl represents a single control in the export package.
type ExportedControl struct {
	Code               string            `json:"code"`
	Title              string            `json:"title"`
	Description        string            `json:"description"`
	Category           string            `json:"category"`
	ControlType        string            `json:"control_type"`
	ImplementationType string            `json:"implementation_type"`
	Guidance           string            `json:"guidance,omitempty"`
	TestProcedure      string            `json:"test_procedure,omitempty"`
	EvidenceReqs       []string          `json:"evidence_requirements,omitempty"`
	Mappings           map[string][]string `json:"mappings,omitempty"`
}

// ExportedPolicy represents a policy document in the export package.
type ExportedPolicy struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Version     string `json:"version"`
	Content     string `json:"content,omitempty"`
}

// ExportedTemplate represents a template in the export package.
type ExportedTemplate struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Content     string `json:"content"`
}

// ExportMetadata provides context about the export.
type ExportMetadata struct {
	TotalControls int      `json:"total_controls"`
	TotalPolicies int      `json:"total_policies"`
	Frameworks    []string `json:"frameworks"`
	Categories    []string `json:"categories"`
}

// ExportAsPackage exports an organisation's controls as a marketplace package.
// It queries the org's control implementations, strips org-specific data,
// builds cross-framework mappings, and produces a portable PackageData.
func (b *PackageBuilder) ExportAsPackage(ctx context.Context, orgID uuid.UUID, config ExportConfig) (*PackageData, error) {
	if config.Name == "" {
		return nil, fmt.Errorf("name is required in export config")
	}
	if config.PackageType == "" {
		config.PackageType = "control_pack"
	}

	log.Info().
		Str("org_id", orgID.String()).
		Str("name", config.Name).
		Str("type", config.PackageType).
		Msg("starting package export")

	pkg := &PackageData{
		SchemaVersion: "1.0",
		PackageType:   config.PackageType,
		ExportedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	// Export controls
	controls, frameworks, categories, err := b.exportControls(ctx, orgID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to export controls: %w", err)
	}
	pkg.Controls = controls

	pkg.Metadata = ExportMetadata{
		TotalControls: len(controls),
		TotalPolicies: 0,
		Frameworks:    frameworks,
		Categories:    categories,
	}

	// Calculate hash and size
	dataBytes, err := json.Marshal(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to serialise package: %w", err)
	}

	hash := sha256.Sum256(dataBytes)
	pkg.Hash = fmt.Sprintf("%x", hash)
	pkg.SizeBytes = int64(len(dataBytes))

	// Enforce 10 MB limit
	if pkg.SizeBytes > maxPackageDataSize {
		return nil, fmt.Errorf("exported package exceeds 10 MB limit (%d bytes); reduce scope", pkg.SizeBytes)
	}

	log.Info().
		Str("org_id", orgID.String()).
		Int("controls", len(controls)).
		Int64("size_bytes", pkg.SizeBytes).
		Str("hash", pkg.Hash).
		Msg("package export completed")

	return pkg, nil
}

// exportControls queries the organisation's control implementations and
// associated framework controls, strips org-specific data, and optionally
// includes cross-framework mappings and guidance.
func (b *PackageBuilder) exportControls(ctx context.Context, orgID uuid.UUID, config ExportConfig) ([]ExportedControl, []string, []string, error) {
	// Build query based on config
	query := `
		SELECT DISTINCT
			fc.code, fc.title, fc.description,
			COALESCE(fc.control_type, 'preventive') AS control_type,
			COALESCE(fc.implementation_type, 'technical') AS implementation_type,
			fc.guidance,
			fd.name AS domain_name,
			f.code AS framework_code
		FROM control_implementations ci
		JOIN framework_controls fc ON ci.framework_control_id = fc.id
		JOIN framework_domains fd ON fc.domain_id = fd.id
		JOIN frameworks f ON fd.framework_id = f.id
		JOIN organization_frameworks of2 ON ci.org_framework_id = of2.id
		WHERE ci.organization_id = $1 AND ci.deleted_at IS NULL`

	args := []interface{}{orgID}
	argIdx := 2

	// Filter by specific framework IDs if provided
	if len(config.FrameworkIDs) > 0 {
		query += fmt.Sprintf(" AND of2.framework_id = ANY($%d)", argIdx)
		args = append(args, config.FrameworkIDs)
		argIdx++
	}

	// Filter by specific control IDs if provided
	if len(config.ControlIDs) > 0 {
		query += fmt.Sprintf(" AND ci.id = ANY($%d)", argIdx)
		args = append(args, config.ControlIDs)
		argIdx++
	}

	query += " ORDER BY fc.code"

	rows, err := b.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to query controls: %w", err)
	}
	defer rows.Close()

	var controls []ExportedControl
	frameworkSet := map[string]bool{}
	categorySet := map[string]bool{}

	for rows.Next() {
		var (
			code, title, desc, ctrlType, implType string
			guidance, domainName, fwCode          string
		)
		if err := rows.Scan(&code, &title, &desc, &ctrlType, &implType, &guidance, &domainName, &fwCode); err != nil {
			return nil, nil, nil, fmt.Errorf("scan failed: %w", err)
		}

		ctrl := ExportedControl{
			Code:               code,
			Title:              title,
			Description:        desc,
			Category:           domainName,
			ControlType:        ctrlType,
			ImplementationType: implType,
		}

		if config.IncludeGuidance && guidance != "" {
			ctrl.Guidance = guidance
		}

		frameworkSet[fwCode] = true
		categorySet[domainName] = true

		controls = append(controls, ctrl)
	}

	// Optionally load cross-framework mappings
	if config.IncludeMappings && len(controls) > 0 {
		controls, err = b.enrichWithMappings(ctx, controls)
		if err != nil {
			log.Warn().Err(err).Msg("failed to enrich controls with mappings, continuing without")
		}
	}

	// Collect unique framework codes and categories
	frameworks := make([]string, 0, len(frameworkSet))
	for fw := range frameworkSet {
		frameworks = append(frameworks, fw)
	}
	categories := make([]string, 0, len(categorySet))
	for cat := range categorySet {
		categories = append(categories, cat)
	}

	return controls, frameworks, categories, nil
}

// enrichWithMappings loads cross-framework mappings for exported controls
// and populates the Mappings field on each ExportedControl.
func (b *PackageBuilder) enrichWithMappings(ctx context.Context, controls []ExportedControl) ([]ExportedControl, error) {
	// Build a set of control codes for lookup
	codes := make([]string, len(controls))
	codeIndex := map[string]int{}
	for i, c := range controls {
		codes[i] = c.Code
		codeIndex[c.Code] = i
	}

	rows, err := b.pool.Query(ctx, `
		SELECT
			src.code AS source_code,
			tgt.code AS target_code,
			f.code AS target_framework
		FROM control_mappings cm
		JOIN framework_controls src ON cm.source_control_id = src.id
		JOIN framework_controls tgt ON cm.target_control_id = tgt.id
		JOIN framework_domains fd ON tgt.domain_id = fd.id
		JOIN frameworks f ON fd.framework_id = f.id
		WHERE src.code = ANY($1)`,
		codes)
	if err != nil {
		return controls, fmt.Errorf("failed to query mappings: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var srcCode, tgtCode, tgtFW string
		if err := rows.Scan(&srcCode, &tgtCode, &tgtFW); err != nil {
			continue
		}
		if idx, ok := codeIndex[srcCode]; ok {
			if controls[idx].Mappings == nil {
				controls[idx].Mappings = map[string][]string{}
			}
			controls[idx].Mappings[tgtFW] = append(controls[idx].Mappings[tgtFW], tgtCode)
		}
	}

	return controls, nil
}
