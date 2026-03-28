package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/complianceforge/platform/internal/database"
	"github.com/complianceforge/platform/internal/middleware"
	"github.com/complianceforge/platform/internal/models"
	"github.com/complianceforge/platform/internal/pkg/storage"
)

type EvidenceHandler struct {
	db      *database.DB
	storage storage.Storage
}

func NewEvidenceHandler(db *database.DB, store storage.Storage) *EvidenceHandler {
	return &EvidenceHandler{db: db, storage: store}
}

// UploadEvidence uploads a file as evidence for a control implementation.
// POST /api/v1/controls/{id}/evidence
func (h *EvidenceHandler) UploadEvidence(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	implID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid implementation ID")
		return
	}

	// Parse multipart form (max 50MB)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "FILE_TOO_LARGE", "File size exceeds 50MB limit")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "MISSING_FILE", "No file uploaded")
		return
	}
	defer file.Close()

	title := r.FormValue("title")
	if title == "" {
		title = header.Filename
	}
	description := r.FormValue("description")
	evidenceType := r.FormValue("evidence_type")
	if evidenceType == "" {
		evidenceType = "document"
	}

	// Store the file
	stored, err := h.storage.Store(orgID.String(), "evidence", header.Filename, file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "STORAGE_ERROR", "Failed to store file")
		return
	}

	// Create evidence record
	var evidenceID uuid.UUID
	err = h.db.Pool.QueryRow(r.Context(), `
		INSERT INTO control_evidence (
			organization_id, control_implementation_id, title, description,
			evidence_type, file_path, file_name, file_size_bytes, mime_type,
			file_hash, collection_method, collected_by, is_current, review_status, metadata
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'manual_upload',$11,true,'pending','{}')
		RETURNING id`,
		orgID, implID, title, description, evidenceType,
		stored.Path, stored.FileName, stored.Size, header.Header.Get("Content-Type"),
		stored.SHA256, userID,
	).Scan(&evidenceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create evidence record")
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"evidence_id": evidenceID,
			"file_name":   stored.FileName,
			"file_size":   stored.Size,
			"sha256":      stored.SHA256,
			"message":     "Evidence uploaded successfully",
		},
	})
}

// ListEvidence returns all evidence for a control implementation.
// GET /api/v1/controls/{id}/evidence
func (h *EvidenceHandler) ListEvidence(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	implID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid implementation ID")
		return
	}

	rows, err := h.db.Pool.Query(r.Context(), `
		SELECT id, title, description, evidence_type, file_name, file_size_bytes,
			mime_type, file_hash, collection_method, collected_at,
			valid_from, valid_until, is_current, review_status,
			reviewed_by, reviewed_at, review_notes
		FROM control_evidence
		WHERE control_implementation_id = $1 AND organization_id = $2 AND deleted_at IS NULL
		ORDER BY collected_at DESC`,
		implID, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list evidence")
		return
	}
	defer rows.Close()

	var evidence []map[string]interface{}
	for rows.Next() {
		var e models.ControlEvidence
		if err := rows.Scan(
			&e.ID, &e.Title, &e.Description, &e.EvidenceType, &e.FileName,
			&e.FileSizeBytes, &e.MimeType, &e.FileHash, &e.CollectionMethod,
			&e.CollectedAt, &e.ValidFrom, &e.ValidUntil, &e.IsCurrent,
			&e.ReviewStatus, &e.ReviewedBy, &e.ReviewedAt, &e.ReviewNotes,
		); err != nil {
			continue
		}
		evidence = append(evidence, map[string]interface{}{
			"id":                e.ID,
			"title":             e.Title,
			"evidence_type":     e.EvidenceType,
			"file_name":         e.FileName,
			"file_size_bytes":   e.FileSizeBytes,
			"file_size_display": formatFileSize(e.FileSizeBytes),
			"collection_method": e.CollectionMethod,
			"collected_at":      e.CollectedAt,
			"is_current":        e.IsCurrent,
			"review_status":     e.ReviewStatus,
		})
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: evidence})
}

// DownloadEvidence streams an evidence file to the client.
// GET /api/v1/evidence/{id}/download
func (h *EvidenceHandler) DownloadEvidence(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	evidenceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid evidence ID")
		return
	}

	var filePath, fileName, mimeType string
	err = h.db.Pool.QueryRow(r.Context(), `
		SELECT file_path, file_name, mime_type
		FROM control_evidence
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`,
		evidenceID, orgID,
	).Scan(&filePath, &fileName, &mimeType)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Evidence not found")
		return
	}

	reader, err := h.storage.Retrieve(filePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "STORAGE_ERROR", "Failed to retrieve file")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	io.Copy(w, reader)
}

// ReviewEvidence accepts or rejects evidence submitted for a control.
// PUT /api/v1/evidence/{id}/review
func (h *EvidenceHandler) ReviewEvidence(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	evidenceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid evidence ID")
		return
	}

	var req struct {
		Status string `json:"status"` // accepted, rejected
		Notes  string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}

	if req.Status != "accepted" && req.Status != "rejected" {
		writeError(w, http.StatusBadRequest, "INVALID_STATUS", "Status must be 'accepted' or 'rejected'")
		return
	}

	_, err = h.db.Pool.Exec(r.Context(), `
		UPDATE control_evidence SET
			review_status = $1, reviewed_by = $2, reviewed_at = NOW(), review_notes = $3
		WHERE id = $4 AND organization_id = $5 AND deleted_at IS NULL`,
		req.Status, userID, req.Notes, evidenceID, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to review evidence")
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"evidence_id": evidenceID,
			"status":      req.Status,
			"message":     "Evidence reviewed",
		},
	})
}

func formatFileSize(bytes int64) string {
	if bytes < 1024 {
		return strconv.FormatInt(bytes, 10) + " B"
	}
	if bytes < 1024*1024 {
		return strconv.FormatFloat(float64(bytes)/1024, 'f', 1, 64) + " KB"
	}
	return strconv.FormatFloat(float64(bytes)/(1024*1024), 'f', 1, 64) + " MB"
}
