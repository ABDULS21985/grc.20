package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// ============================================================
// YAML PARSING TESTS
// ============================================================

// TestParseYAML_ControlImplementation verifies parsing of a
// ControlImplementation resource with all required fields.
func TestParseYAML_ControlImplementation(t *testing.T) {
	content := []byte(`
apiVersion: complianceforge.io/v1
kind: ControlImplementation
metadata:
  name: access-control-policy-enforcement
  uid: controlimplementation/access-control-policy-enforcement
  labels:
    framework: ISO27001
    domain: access-control
  annotations:
    owner: security-team
spec:
  control_code: A.9.1.1
  framework: ISO27001
  status: implemented
  implementation_type: technical
  description: |
    Access control policy is enforced through centralised IAM
    with MFA enabled for all privileged accounts.
  evidence:
    - type: screenshot
      path: evidence/iam-mfa-config.png
    - type: log_export
      path: evidence/access-audit-log.csv
  owner: security-team@example.com
  review_frequency: quarterly
`)

	engine := NewCaCEngine(nil)
	res, err := engine.ParseYAML(content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if res.ApiVersion != "complianceforge.io/v1" {
		t.Errorf("expected apiVersion 'complianceforge.io/v1', got %q", res.ApiVersion)
	}
	if res.Kind != "ControlImplementation" {
		t.Errorf("expected kind 'ControlImplementation', got %q", res.Kind)
	}
	if res.Metadata.Name != "access-control-policy-enforcement" {
		t.Errorf("expected name 'access-control-policy-enforcement', got %q", res.Metadata.Name)
	}
	if res.Metadata.UID != "controlimplementation/access-control-policy-enforcement" {
		t.Errorf("expected UID 'controlimplementation/access-control-policy-enforcement', got %q", res.Metadata.UID)
	}
	if res.Metadata.Labels["framework"] != "ISO27001" {
		t.Errorf("expected label framework=ISO27001, got %q", res.Metadata.Labels["framework"])
	}

	// Verify spec fields
	if code, ok := res.Spec["control_code"]; !ok || code != "A.9.1.1" {
		t.Errorf("expected spec.control_code='A.9.1.1', got %v", res.Spec["control_code"])
	}
	if status, ok := res.Spec["status"]; !ok || status != "implemented" {
		t.Errorf("expected spec.status='implemented', got %v", res.Spec["status"])
	}
}

// TestParseYAML_Policy verifies parsing of a Policy resource.
func TestParseYAML_Policy(t *testing.T) {
	content := []byte(`
apiVersion: complianceforge.io/v1
kind: Policy
metadata:
  name: information-security-policy
  uid: policy/information-security-policy
  labels:
    category: governance
spec:
  title: Information Security Policy
  version: "3.2"
  status: published
  owner: ciso@example.com
  review_date: "2026-06-15"
  classification: internal
  sections:
    - title: Purpose
      content: This policy establishes the framework for information security.
    - title: Scope
      content: Applies to all employees, contractors, and third parties.
`)

	engine := NewCaCEngine(nil)
	res, err := engine.ParseYAML(content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if res.Kind != "Policy" {
		t.Errorf("expected kind 'Policy', got %q", res.Kind)
	}
	if res.Metadata.Name != "information-security-policy" {
		t.Errorf("expected name 'information-security-policy', got %q", res.Metadata.Name)
	}
	if title, ok := res.Spec["title"]; !ok || title != "Information Security Policy" {
		t.Errorf("expected spec.title='Information Security Policy', got %v", res.Spec["title"])
	}
	if version, ok := res.Spec["version"]; !ok || version != "3.2" {
		t.Errorf("expected spec.version='3.2', got %v", res.Spec["version"])
	}
}

// TestParseYAML_RiskAcceptance verifies parsing of a RiskAcceptance resource.
func TestParseYAML_RiskAcceptance(t *testing.T) {
	content := []byte(`
apiVersion: complianceforge.io/v1
kind: RiskAcceptance
metadata:
  name: legacy-system-migration-risk
  uid: riskacceptance/legacy-system-migration-risk
spec:
  risk_id: RISK-2024-047
  risk_title: Legacy ERP system lacks modern encryption
  risk_level: medium
  accepted_by: cto@example.com
  accepted_date: "2026-01-15"
  expiry_date: "2026-07-15"
  justification: |
    Migration to new ERP system is planned for Q3 2026.
    Compensating controls are in place including network segmentation.
  compensating_controls:
    - Network segmentation isolating legacy system
    - Enhanced monitoring and logging
    - Restricted access to authorised personnel only
  conditions:
    - Must be reviewed monthly
    - Migration project must remain on track
`)

	engine := NewCaCEngine(nil)
	res, err := engine.ParseYAML(content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if res.Kind != "RiskAcceptance" {
		t.Errorf("expected kind 'RiskAcceptance', got %q", res.Kind)
	}
	if riskID, ok := res.Spec["risk_id"]; !ok || riskID != "RISK-2024-047" {
		t.Errorf("expected spec.risk_id='RISK-2024-047', got %v", res.Spec["risk_id"])
	}
	if acceptedBy, ok := res.Spec["accepted_by"]; !ok || acceptedBy != "cto@example.com" {
		t.Errorf("expected spec.accepted_by='cto@example.com', got %v", res.Spec["accepted_by"])
	}
}

// TestParseYAML_EvidenceConfig verifies parsing of an EvidenceConfig resource.
func TestParseYAML_EvidenceConfig(t *testing.T) {
	content := []byte(`
apiVersion: complianceforge.io/v1
kind: EvidenceConfig
metadata:
  name: aws-cloudtrail-evidence
  uid: evidenceconfig/aws-cloudtrail-evidence
  labels:
    provider: aws
spec:
  evidence_type: automated_log
  collection_method: api_pull
  source_system: AWS CloudTrail
  frequency: daily
  retention_days: 365
  controls:
    - A.12.4.1
    - A.12.4.3
  format: json
  validation_rules:
    - field: eventSource
      required: true
    - field: eventTime
      required: true
`)

	engine := NewCaCEngine(nil)
	res, err := engine.ParseYAML(content)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if res.Kind != "EvidenceConfig" {
		t.Errorf("expected kind 'EvidenceConfig', got %q", res.Kind)
	}
	if evType, ok := res.Spec["evidence_type"]; !ok || evType != "automated_log" {
		t.Errorf("expected spec.evidence_type='automated_log', got %v", res.Spec["evidence_type"])
	}
	if method, ok := res.Spec["collection_method"]; !ok || method != "api_pull" {
		t.Errorf("expected spec.collection_method='api_pull', got %v", res.Spec["collection_method"])
	}
}

// TestParseYAML_InvalidKind verifies that an unsupported kind is rejected.
func TestParseYAML_InvalidKind(t *testing.T) {
	content := []byte(`
apiVersion: complianceforge.io/v1
kind: UnsupportedResource
metadata:
  name: test-resource
spec:
  field: value
`)

	engine := NewCaCEngine(nil)
	_, err := engine.ParseYAML(content)
	if err == nil {
		t.Fatal("expected error for unsupported kind, got nil")
	}
	if got := err.Error(); got != "unsupported resource kind: UnsupportedResource" {
		t.Errorf("expected 'unsupported resource kind: UnsupportedResource', got %q", got)
	}
}

// TestParseYAML_MissingRequired verifies that missing required fields
// are correctly reported as errors.
func TestParseYAML_MissingRequired(t *testing.T) {
	tests := []struct {
		name    string
		content string
		errMsg  string
	}{
		{
			name:    "missing apiVersion",
			content: "kind: Policy\nmetadata:\n  name: test\nspec:\n  title: Test",
			errMsg:  "missing required field: apiVersion",
		},
		{
			name:    "missing kind",
			content: "apiVersion: complianceforge.io/v1\nmetadata:\n  name: test\nspec:\n  title: Test",
			errMsg:  "missing required field: kind",
		},
		{
			name:    "missing metadata.name",
			content: "apiVersion: complianceforge.io/v1\nkind: Policy\nmetadata:\n  uid: test\nspec:\n  title: Test",
			errMsg:  "missing required field: metadata.name",
		},
	}

	engine := NewCaCEngine(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.ParseYAML([]byte(tt.content))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.errMsg {
				t.Errorf("expected error %q, got %q", tt.errMsg, err.Error())
			}
		})
	}
}

// ============================================================
// VALIDATION TESTS
// ============================================================

// TestValidateResources_Valid verifies that valid resources pass validation.
func TestValidateResources_Valid(t *testing.T) {
	resources := []CaCResource{
		{
			ApiVersion: "complianceforge.io/v1",
			Kind:       "ControlImplementation",
			Metadata:   CaCResourceMetadata{Name: "ctrl-a", UID: "controlimplementation/ctrl-a"},
			Spec:       map[string]interface{}{"control_code": "A.5.1", "status": "implemented"},
		},
		{
			ApiVersion: "complianceforge.io/v1",
			Kind:       "Policy",
			Metadata:   CaCResourceMetadata{Name: "pol-a", UID: "policy/pol-a"},
			Spec:       map[string]interface{}{"title": "Test Policy"},
		},
		{
			ApiVersion: "complianceforge.io/v1",
			Kind:       "RiskAcceptance",
			Metadata:   CaCResourceMetadata{Name: "risk-a", UID: "riskacceptance/risk-a"},
			Spec:       map[string]interface{}{"risk_id": "RISK-001", "accepted_by": "cto@example.com"},
		},
		{
			ApiVersion: "complianceforge.io/v1",
			Kind:       "EvidenceConfig",
			Metadata:   CaCResourceMetadata{Name: "ev-a", UID: "evidenceconfig/ev-a"},
			Spec:       map[string]interface{}{"evidence_type": "screenshot", "collection_method": "manual"},
		},
	}

	engine := NewCaCEngine(nil)
	errs := engine.ValidateResources(resources)
	if len(errs) != 0 {
		t.Errorf("expected 0 validation errors, got %d: %+v", len(errs), errs)
	}
}

// TestValidateResources_DuplicateIdentifiers verifies that duplicate
// resource UIDs are detected during cross-resource validation.
func TestValidateResources_DuplicateIdentifiers(t *testing.T) {
	resources := []CaCResource{
		{
			ApiVersion: "complianceforge.io/v1",
			Kind:       "Policy",
			Metadata:   CaCResourceMetadata{Name: "same-policy", UID: "policy/same-policy"},
			Spec:       map[string]interface{}{"title": "Policy A"},
		},
		{
			ApiVersion: "complianceforge.io/v1",
			Kind:       "Policy",
			Metadata:   CaCResourceMetadata{Name: "same-policy", UID: "policy/same-policy"},
			Spec:       map[string]interface{}{"title": "Policy B"},
		},
	}

	engine := NewCaCEngine(nil)
	errs := engine.ValidateResources(resources)
	if len(errs) == 0 {
		t.Fatal("expected validation errors for duplicate UIDs, got none")
	}

	found := false
	for _, e := range errs {
		if e.Field == "metadata.uid" && e.Severity == "error" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a duplicate UID error, got: %+v", errs)
	}
}

// ============================================================
// DIFF PLAN TESTS
// ============================================================

// TestDiffPlan_NewResources verifies that resources not yet in
// the platform are identified as 'create' actions.
func TestDiffPlan_NewResources(t *testing.T) {
	// Build a minimal DiffPlan manually to test the logic without DB
	resources := []CaCResource{
		{
			ApiVersion: "complianceforge.io/v1",
			Kind:       "Policy",
			Metadata:   CaCResourceMetadata{Name: "new-policy", UID: "policy/new-policy"},
			Spec:       map[string]interface{}{"title": "New Policy"},
		},
		{
			ApiVersion: "complianceforge.io/v1",
			Kind:       "ControlImplementation",
			Metadata:   CaCResourceMetadata{Name: "new-control", UID: "controlimplementation/new-control"},
			Spec:       map[string]interface{}{"control_code": "A.5.1", "status": "planned"},
		},
	}

	// Simulate diffing with empty existing state
	plan := simulateDiff(resources, nil)

	if plan.Summary.ToCreate != 2 {
		t.Errorf("expected 2 creates, got %d", plan.Summary.ToCreate)
	}
	if plan.Summary.ToUpdate != 0 {
		t.Errorf("expected 0 updates, got %d", plan.Summary.ToUpdate)
	}
	if plan.Summary.ToDelete != 0 {
		t.Errorf("expected 0 deletes, got %d", plan.Summary.ToDelete)
	}

	for _, a := range plan.Actions {
		if a.Action != "create" {
			t.Errorf("expected action 'create', got %q for %s", a.Action, a.ResourceUID)
		}
	}
}

// TestDiffPlan_ModifiedResources verifies that changed resources
// are identified as 'update' actions.
func TestDiffPlan_ModifiedResources(t *testing.T) {
	resources := []CaCResource{
		{
			ApiVersion: "complianceforge.io/v1",
			Kind:       "Policy",
			Metadata:   CaCResourceMetadata{Name: "existing-policy", UID: "policy/existing-policy"},
			Spec:       map[string]interface{}{"title": "Updated Policy Title", "version": "2.0"},
		},
	}

	// Simulate existing with a different content hash
	existing := map[string]mockMapping{
		"policy/existing-policy": {
			Kind:        "Policy",
			ContentHash: "old-hash-that-differs",
			FilePath:    "default/policy/existing-policy.yaml",
		},
	}

	plan := simulateDiff(resources, existing)

	if plan.Summary.ToUpdate != 1 {
		t.Errorf("expected 1 update, got %d", plan.Summary.ToUpdate)
	}
	if plan.Summary.ToCreate != 0 {
		t.Errorf("expected 0 creates, got %d", plan.Summary.ToCreate)
	}

	if len(plan.Actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(plan.Actions))
	}
	if plan.Actions[0].Action != "update" {
		t.Errorf("expected action 'update', got %q", plan.Actions[0].Action)
	}
}

// TestDiffPlan_OrphanedResources verifies that resources in the
// platform but not in the repo are identified as 'delete' actions.
func TestDiffPlan_OrphanedResources(t *testing.T) {
	// Empty repo resources
	resources := []CaCResource{}

	// Existing platform resources that should be orphaned
	existing := map[string]mockMapping{
		"policy/old-policy": {
			Kind:        "Policy",
			ResourceName: "old-policy",
			ContentHash: "some-hash",
			FilePath:    "default/policy/old-policy.yaml",
		},
		"controlimplementation/old-control": {
			Kind:        "ControlImplementation",
			ResourceName: "old-control",
			ContentHash: "another-hash",
			FilePath:    "default/controlimplementation/old-control.yaml",
		},
	}

	plan := simulateDiff(resources, existing)

	if plan.Summary.ToDelete != 2 {
		t.Errorf("expected 2 deletes, got %d", plan.Summary.ToDelete)
	}
	if plan.Summary.ToCreate != 0 {
		t.Errorf("expected 0 creates, got %d", plan.Summary.ToCreate)
	}

	for _, a := range plan.Actions {
		if a.Action != "delete" {
			t.Errorf("expected action 'delete', got %q for %s", a.Action, a.ResourceUID)
		}
	}
}

// ============================================================
// WEBHOOK SIGNATURE TESTS
// ============================================================

// TestWebhookSignature_Valid verifies correct webhook signature validation.
func TestWebhookSignature_Valid(t *testing.T) {
	engine := NewCaCEngine(nil)
	secret := "my-webhook-secret"
	payload := []byte(`{"ref":"refs/heads/main","commits":[{"id":"abc123"}]}`)

	// Compute expected signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	if !engine.ValidateWebhookSignature(payload, secret, sig) {
		t.Error("expected valid signature to return true")
	}
}

// TestWebhookSignature_Invalid verifies that invalid signatures are rejected.
func TestWebhookSignature_Invalid(t *testing.T) {
	engine := NewCaCEngine(nil)
	secret := "my-webhook-secret"
	payload := []byte(`{"ref":"refs/heads/main","commits":[{"id":"abc123"}]}`)

	tests := []struct {
		name      string
		payload   []byte
		secret    string
		signature string
	}{
		{
			name:      "wrong signature",
			payload:   payload,
			secret:    secret,
			signature: "sha256=0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:      "empty secret",
			payload:   payload,
			secret:    "",
			signature: "sha256=abc123",
		},
		{
			name:      "empty signature",
			payload:   payload,
			secret:    secret,
			signature: "",
		},
		{
			name:      "tampered payload",
			payload:   []byte(`{"ref":"refs/heads/evil","commits":[]}`),
			secret:    secret,
			signature: computeTestSig(payload, secret),
		},
		{
			name:      "wrong secret",
			payload:   payload,
			secret:    "wrong-secret",
			signature: computeTestSig(payload, secret),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if engine.ValidateWebhookSignature(tt.payload, tt.secret, tt.signature) {
				t.Error("expected invalid signature to return false")
			}
		})
	}
}

// ============================================================
// MULTI-DOCUMENT YAML TESTS
// ============================================================

// TestParseMultiYAML verifies parsing of multiple YAML documents
// separated by --- dividers.
func TestParseMultiYAML(t *testing.T) {
	content := []byte(`
apiVersion: complianceforge.io/v1
kind: Policy
metadata:
  name: policy-one
spec:
  title: Policy One
---
apiVersion: complianceforge.io/v1
kind: ControlImplementation
metadata:
  name: control-one
spec:
  control_code: A.5.1
  status: implemented
---
apiVersion: complianceforge.io/v1
kind: EvidenceConfig
metadata:
  name: evidence-one
spec:
  evidence_type: screenshot
  collection_method: manual
`)

	engine := NewCaCEngine(nil)
	resources, errs := engine.ParseMultiYAML(content)

	if len(errs) != 0 {
		t.Errorf("expected no parse errors, got %d: %v", len(errs), errs)
	}
	if len(resources) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(resources))
	}

	if resources[0].Kind != "Policy" {
		t.Errorf("expected first resource kind 'Policy', got %q", resources[0].Kind)
	}
	if resources[1].Kind != "ControlImplementation" {
		t.Errorf("expected second resource kind 'ControlImplementation', got %q", resources[1].Kind)
	}
	if resources[2].Kind != "EvidenceConfig" {
		t.Errorf("expected third resource kind 'EvidenceConfig', got %q", resources[2].Kind)
	}
}

// TestParseMultiYAML_WithErrors verifies that valid resources are still
// returned even when some documents have errors.
func TestParseMultiYAML_WithErrors(t *testing.T) {
	content := []byte(`
apiVersion: complianceforge.io/v1
kind: Policy
metadata:
  name: good-policy
spec:
  title: Good Policy
---
apiVersion: complianceforge.io/v1
kind: InvalidKind
metadata:
  name: bad-resource
spec:
  field: value
`)

	engine := NewCaCEngine(nil)
	resources, errs := engine.ParseMultiYAML(content)

	if len(resources) != 1 {
		t.Errorf("expected 1 valid resource, got %d", len(resources))
	}
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d", len(errs))
	}
}

// ============================================================
// VALIDATE KIND-SPECIFIC SPEC TESTS
// ============================================================

// TestValidateResources_MissingSpecFields verifies that kind-specific
// required spec fields are validated.
func TestValidateResources_MissingSpecFields(t *testing.T) {
	resources := []CaCResource{
		{
			ApiVersion: "complianceforge.io/v1",
			Kind:       "ControlImplementation",
			Metadata:   CaCResourceMetadata{Name: "ctrl-missing-fields", UID: "controlimplementation/ctrl-missing"},
			Spec:       map[string]interface{}{}, // missing control_code and status
		},
		{
			ApiVersion: "complianceforge.io/v1",
			Kind:       "RiskAcceptance",
			Metadata:   CaCResourceMetadata{Name: "risk-missing", UID: "riskacceptance/risk-missing"},
			Spec:       map[string]interface{}{}, // missing risk_id and accepted_by
		},
	}

	engine := NewCaCEngine(nil)
	errs := engine.ValidateResources(resources)

	// ControlImplementation missing control_code + status = 2 errors
	// RiskAcceptance missing risk_id + accepted_by = 2 errors
	if len(errs) < 4 {
		t.Errorf("expected at least 4 validation errors, got %d: %+v", len(errs), errs)
	}
}

// ============================================================
// CONTENT HASH TESTS
// ============================================================

// TestComputeResourceHash verifies that the hash is deterministic
// and changes when content changes.
func TestComputeResourceHash(t *testing.T) {
	r1 := CaCResource{
		ApiVersion: "complianceforge.io/v1",
		Kind:       "Policy",
		Metadata:   CaCResourceMetadata{Name: "test-policy", UID: "policy/test"},
		Spec:       map[string]interface{}{"title": "Test"},
	}

	h1 := computeResourceHash(r1)
	h2 := computeResourceHash(r1)

	if h1 != h2 {
		t.Error("expected deterministic hash, got different values for same input")
	}
	if h1 == "" {
		t.Error("expected non-empty hash")
	}

	// Modify and verify hash changes
	r2 := r1
	r2.Spec = map[string]interface{}{"title": "Modified"}
	h3 := computeResourceHash(r2)
	if h1 == h3 {
		t.Error("expected different hash for modified resource")
	}
}

// TestBuildFilePath verifies default file path construction.
func TestBuildFilePath(t *testing.T) {
	r := CaCResource{
		Kind: "Policy",
		Metadata: CaCResourceMetadata{
			Name:      "My Test Policy",
			Namespace: "production",
		},
	}

	path := buildFilePath(r)
	expected := "production/policy/my-test-policy.yaml"
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}

	// Test default namespace
	r.Metadata.Namespace = ""
	path = buildFilePath(r)
	expected = "default/policy/my-test-policy.yaml"
	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

// ============================================================
// TEST HELPERS
// ============================================================

// mockMapping simulates an existing resource mapping for diff tests.
type mockMapping struct {
	Kind         string
	ResourceName string
	ContentHash  string
	FilePath     string
}

// simulateDiff performs an in-memory diff similar to DiffWithPlatform
// but without database access, for unit testing.
func simulateDiff(resources []CaCResource, existing map[string]mockMapping) *DiffPlan {
	plan := &DiffPlan{}
	seen := make(map[string]bool)

	for _, r := range resources {
		uid := r.Metadata.UID
		if uid == "" {
			uid = r.Kind + "/" + r.Metadata.Name
		}
		seen[uid] = true

		contentHash := computeResourceHash(r)

		if ex, exists := existing[uid]; exists {
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
					Resource:     &r,
				})
				plan.Summary.ToUpdate++
			}
		} else {
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
	return plan
}

// computeTestSig computes a test webhook signature.
func computeTestSig(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
