// Package storage provides a file storage abstraction for ComplianceForge.
// Supports local filesystem and S3-compatible backends.
// Used for storing control evidence, policy documents, audit reports, and attachments.
package storage

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/complianceforge/platform/internal/config"
)

// StoredFile contains metadata about a stored file.
type StoredFile struct {
	Path      string `json:"path"`
	FileName  string `json:"file_name"`
	Size      int64  `json:"size_bytes"`
	MimeType  string `json:"mime_type"`
	SHA256    string `json:"sha256"`
	StoredAt  time.Time `json:"stored_at"`
}

// Storage defines the file storage interface.
type Storage interface {
	Store(orgID string, category string, fileName string, reader io.Reader) (*StoredFile, error)
	Retrieve(path string) (io.ReadCloser, error)
	Delete(path string) error
}

// LocalStorage implements Storage using the local filesystem.
type LocalStorage struct {
	basePath string
}

// NewStorage creates a storage backend based on configuration.
func NewStorage(cfg config.StorageConfig) (Storage, error) {
	switch cfg.Driver {
	case "local":
		return NewLocalStorage(cfg.LocalPath)
	case "s3":
		// S3 implementation would go here
		return nil, fmt.Errorf("S3 storage not yet implemented — use local")
	default:
		return NewLocalStorage(cfg.LocalPath)
	}
}

// NewLocalStorage creates a local filesystem storage backend.
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	if err := os.MkdirAll(basePath, 0750); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	log.Info().Str("path", basePath).Msg("Local file storage initialised")
	return &LocalStorage{basePath: basePath}, nil
}

// Store saves a file to local storage, organized by org and category.
// Category examples: "evidence", "policies", "audit-reports", "attachments"
func (s *LocalStorage) Store(orgID, category, fileName string, reader io.Reader) (*StoredFile, error) {
	// Create directory structure: basePath/orgID/category/YYYY-MM/
	yearMonth := time.Now().Format("2006-01")
	dir := filepath.Join(s.basePath, orgID, category, yearMonth)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate unique filename to prevent collisions
	ext := filepath.Ext(fileName)
	uniqueName := fmt.Sprintf("%s_%s%s", uuid.New().String()[:8], sanitizeFileName(fileName), ext)
	filePath := filepath.Join(dir, uniqueName)

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write file and compute SHA-256 hash simultaneously
	hasher := sha256.New()
	tee := io.TeeReader(reader, hasher)

	size, err := io.Copy(file, tee)
	if err != nil {
		os.Remove(filePath) // Clean up on failure
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	// Return relative path from basePath
	relativePath := filepath.Join(orgID, category, yearMonth, uniqueName)

	return &StoredFile{
		Path:     relativePath,
		FileName: fileName,
		Size:     size,
		SHA256:   hash,
		StoredAt: time.Now(),
	}, nil
}

// Retrieve opens a stored file for reading.
func (s *LocalStorage) Retrieve(path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	return file, nil
}

// Delete removes a stored file.
func (s *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(s.basePath, path)
	return os.Remove(fullPath)
}

// sanitizeFileName removes potentially dangerous characters from filenames.
func sanitizeFileName(name string) string {
	// Remove extension first
	name = name[:len(name)-len(filepath.Ext(name))]
	// Keep only alphanumeric, hyphens, underscores
	safe := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			safe = append(safe, c)
		}
	}
	if len(safe) == 0 {
		return "file"
	}
	if len(safe) > 50 {
		safe = safe[:50]
	}
	return string(safe)
}
