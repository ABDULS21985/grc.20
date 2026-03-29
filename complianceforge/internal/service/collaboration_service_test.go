package service

import (
	"testing"

	"github.com/google/uuid"
)

// ============================================================
// ParseMentions Tests — @user:UUID and @role:slug extraction
// ============================================================

func TestParseMentions_NoMentions(t *testing.T) {
	content := "This is a plain comment with no mentions."
	users, roles := ParseMentions(content)
	if len(users) != 0 {
		t.Errorf("Expected 0 user mentions, got %d", len(users))
	}
	if len(roles) != 0 {
		t.Errorf("Expected 0 role mentions, got %d", len(roles))
	}
}

func TestParseMentions_SingleUserMention(t *testing.T) {
	uid := uuid.New()
	content := "Hey @user:" + uid.String() + " please review this."
	users, roles := ParseMentions(content)
	if len(users) != 1 {
		t.Fatalf("Expected 1 user mention, got %d", len(users))
	}
	if users[0] != uid {
		t.Errorf("Expected user ID %s, got %s", uid, users[0])
	}
	if len(roles) != 0 {
		t.Errorf("Expected 0 role mentions, got %d", len(roles))
	}
}

func TestParseMentions_MultipleUserMentions(t *testing.T) {
	uid1 := uuid.New()
	uid2 := uuid.New()
	uid3 := uuid.New()
	content := "CC @user:" + uid1.String() + " and @user:" + uid2.String() + " and @user:" + uid3.String()
	users, roles := ParseMentions(content)
	if len(users) != 3 {
		t.Fatalf("Expected 3 user mentions, got %d", len(users))
	}
	expected := map[uuid.UUID]bool{uid1: true, uid2: true, uid3: true}
	for _, u := range users {
		if !expected[u] {
			t.Errorf("Unexpected user ID %s in mentions", u)
		}
	}
	if len(roles) != 0 {
		t.Errorf("Expected 0 role mentions, got %d", len(roles))
	}
}

func TestParseMentions_DuplicateUserMentions(t *testing.T) {
	uid := uuid.New()
	content := "@user:" + uid.String() + " already mentioned @user:" + uid.String()
	users, _ := ParseMentions(content)
	if len(users) != 1 {
		t.Errorf("Expected 1 deduplicated user mention, got %d", len(users))
	}
}

func TestParseMentions_SingleRoleMention(t *testing.T) {
	content := "Attention @role:compliance_officer please approve."
	users, roles := ParseMentions(content)
	if len(users) != 0 {
		t.Errorf("Expected 0 user mentions, got %d", len(users))
	}
	if len(roles) != 1 {
		t.Fatalf("Expected 1 role mention, got %d", len(roles))
	}
	if roles[0] != "compliance_officer" {
		t.Errorf("Expected role slug 'compliance_officer', got '%s'", roles[0])
	}
}

func TestParseMentions_MultipleRoleMentions(t *testing.T) {
	content := "@role:admin and @role:auditor please check."
	_, roles := ParseMentions(content)
	if len(roles) != 2 {
		t.Fatalf("Expected 2 role mentions, got %d", len(roles))
	}
	expected := map[string]bool{"admin": true, "auditor": true}
	for _, r := range roles {
		if !expected[r] {
			t.Errorf("Unexpected role slug '%s' in mentions", r)
		}
	}
}

func TestParseMentions_DuplicateRoleMentions(t *testing.T) {
	content := "@role:admin and again @role:admin"
	_, roles := ParseMentions(content)
	if len(roles) != 1 {
		t.Errorf("Expected 1 deduplicated role mention, got %d", len(roles))
	}
}

func TestParseMentions_MixedMentions(t *testing.T) {
	uid := uuid.New()
	content := "@user:" + uid.String() + " and @role:security_team please advise"
	users, roles := ParseMentions(content)
	if len(users) != 1 {
		t.Errorf("Expected 1 user mention, got %d", len(users))
	}
	if len(roles) != 1 {
		t.Errorf("Expected 1 role mention, got %d", len(roles))
	}
}

func TestParseMentions_InvalidUUID(t *testing.T) {
	content := "@user:not-a-valid-uuid should be ignored"
	users, _ := ParseMentions(content)
	if len(users) != 0 {
		t.Errorf("Expected 0 user mentions for invalid UUID, got %d", len(users))
	}
}

func TestParseMentions_MentionInMiddleOfText(t *testing.T) {
	uid := uuid.New()
	content := "I think we should ask @user:" + uid.String() + " about the compliance gap on control A.1."
	users, _ := ParseMentions(content)
	if len(users) != 1 {
		t.Fatalf("Expected 1 user mention in middle of text, got %d", len(users))
	}
	if users[0] != uid {
		t.Errorf("Expected user ID %s, got %s", uid, users[0])
	}
}

func TestParseMentions_EmptyString(t *testing.T) {
	users, roles := ParseMentions("")
	if len(users) != 0 {
		t.Errorf("Expected 0 user mentions for empty string, got %d", len(users))
	}
	if len(roles) != 0 {
		t.Errorf("Expected 0 role mentions for empty string, got %d", len(roles))
	}
}

// ============================================================
// SanitizeContentHTML Tests — XSS prevention
// ============================================================

func TestSanitizeContentHTML_PlainText(t *testing.T) {
	input := "Hello, this is a comment."
	result := SanitizeContentHTML(input)
	expected := "Hello, this is a comment."
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestSanitizeContentHTML_HTMLEscaping(t *testing.T) {
	input := "<script>alert('xss')</script>"
	result := SanitizeContentHTML(input)
	if result == input {
		t.Error("XSS payload was not sanitized")
	}
	// Should contain escaped entities
	if !containsStr(result, "&lt;script&gt;") {
		t.Errorf("Expected HTML-escaped script tag, got '%s'", result)
	}
}

func TestSanitizeContentHTML_EventHandler(t *testing.T) {
	input := `<img src=x onerror="alert('xss')">`
	result := SanitizeContentHTML(input)
	if result == input {
		t.Error("Event handler XSS payload was not sanitized")
	}
	if !containsStr(result, "&lt;img") {
		t.Errorf("Expected HTML-escaped img tag, got '%s'", result)
	}
}

func TestSanitizeContentHTML_LineBreaks(t *testing.T) {
	input := "Line 1\nLine 2\nLine 3"
	result := SanitizeContentHTML(input)
	if !containsStr(result, "<br>") {
		t.Errorf("Expected <br> for line breaks, got '%s'", result)
	}
}

func TestSanitizeContentHTML_Ampersand(t *testing.T) {
	input := "A & B"
	result := SanitizeContentHTML(input)
	if !containsStr(result, "&amp;") {
		t.Errorf("Expected escaped ampersand, got '%s'", result)
	}
}

func TestSanitizeContentHTML_QuotesEscaped(t *testing.T) {
	input := `He said "hello" and she said 'hi'`
	result := SanitizeContentHTML(input)
	if !containsStr(result, "&#34;") {
		t.Errorf("Expected escaped double quotes, got '%s'", result)
	}
}

// ============================================================
// Comment Threading Tests — buildCommentTree logic
// ============================================================

func TestBuildCommentTree_EmptyList(t *testing.T) {
	result := buildCommentTree([]Comment{})
	if len(result) != 0 {
		t.Errorf("Expected empty tree for empty input, got %d roots", len(result))
	}
}

func TestBuildCommentTree_FlatComments(t *testing.T) {
	c1 := Comment{ID: uuid.New(), ParentCommentID: nil, Replies: []Comment{}}
	c2 := Comment{ID: uuid.New(), ParentCommentID: nil, Replies: []Comment{}}
	c3 := Comment{ID: uuid.New(), ParentCommentID: nil, Replies: []Comment{}}

	result := buildCommentTree([]Comment{c1, c2, c3})
	if len(result) != 3 {
		t.Fatalf("Expected 3 root comments, got %d", len(result))
	}
	for _, r := range result {
		if len(r.Replies) != 0 {
			t.Errorf("Expected 0 replies on root, got %d", len(r.Replies))
		}
	}
}

func TestBuildCommentTree_SingleThread(t *testing.T) {
	rootID := uuid.New()
	childID := uuid.New()
	grandchildID := uuid.New()

	root := Comment{ID: rootID, ParentCommentID: nil, ThreadDepth: 0}
	child := Comment{ID: childID, ParentCommentID: &rootID, ThreadDepth: 1}
	grandchild := Comment{ID: grandchildID, ParentCommentID: &childID, ThreadDepth: 2}

	result := buildCommentTree([]Comment{root, child, grandchild})
	if len(result) != 1 {
		t.Fatalf("Expected 1 root, got %d", len(result))
	}
	if len(result[0].Replies) != 1 {
		t.Fatalf("Expected 1 reply on root, got %d", len(result[0].Replies))
	}
	if len(result[0].Replies[0].Replies) != 1 {
		t.Fatalf("Expected 1 nested reply, got %d", len(result[0].Replies[0].Replies))
	}
	if result[0].Replies[0].Replies[0].ID != grandchildID {
		t.Error("Grandchild ID mismatch")
	}
}

func TestBuildCommentTree_MultipleThreads(t *testing.T) {
	root1ID := uuid.New()
	root2ID := uuid.New()
	child1ID := uuid.New()
	child2ID := uuid.New()

	root1 := Comment{ID: root1ID, ParentCommentID: nil}
	root2 := Comment{ID: root2ID, ParentCommentID: nil}
	child1 := Comment{ID: child1ID, ParentCommentID: &root1ID}
	child2 := Comment{ID: child2ID, ParentCommentID: &root2ID}

	result := buildCommentTree([]Comment{root1, root2, child1, child2})
	if len(result) != 2 {
		t.Fatalf("Expected 2 roots, got %d", len(result))
	}

	// Find root1 in result (order may vary since they're both roots)
	var r1, r2 *Comment
	for i := range result {
		if result[i].ID == root1ID {
			r1 = &result[i]
		}
		if result[i].ID == root2ID {
			r2 = &result[i]
		}
	}

	if r1 == nil || r2 == nil {
		t.Fatal("Could not find both roots in result")
	}
	if len(r1.Replies) != 1 {
		t.Errorf("Root1 expected 1 reply, got %d", len(r1.Replies))
	}
	if len(r2.Replies) != 1 {
		t.Errorf("Root2 expected 1 reply, got %d", len(r2.Replies))
	}
}

func TestBuildCommentTree_OrphanedReply(t *testing.T) {
	// Reply whose parent doesn't exist should be treated as a root.
	missingParentID := uuid.New()
	orphan := Comment{
		ID:              uuid.New(),
		ParentCommentID: &missingParentID,
	}

	result := buildCommentTree([]Comment{orphan})
	if len(result) != 1 {
		t.Fatalf("Expected orphaned reply to become root, got %d roots", len(result))
	}
}

// ============================================================
// Auto-follow Rule Validation Tests
// ============================================================

func TestAllowedReactions(t *testing.T) {
	validReactions := []string{"thumbs_up", "thumbs_down", "check", "eyes", "rocket", "warning"}
	for _, r := range validReactions {
		if !allowedReactions[r] {
			t.Errorf("Expected '%s' to be an allowed reaction", r)
		}
	}

	invalidReactions := []string{"heart", "fire", "poop", "", "thumbs_up ", " thumbs_up"}
	for _, r := range invalidReactions {
		if allowedReactions[r] {
			t.Errorf("Expected '%s' to NOT be an allowed reaction", r)
		}
	}
}

func TestMaxThreadDepthConstant(t *testing.T) {
	if MaxThreadDepth != 3 {
		t.Errorf("Expected MaxThreadDepth = 3, got %d", MaxThreadDepth)
	}
}

func TestMaxCommentEditWindowHours(t *testing.T) {
	if MaxCommentEditWindowHours != 24 {
		t.Errorf("Expected MaxCommentEditWindowHours = 24, got %d", MaxCommentEditWindowHours)
	}
}

// ============================================================
// Helpers
// ============================================================

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
