package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type MetadataFile struct {
	// Base directory for all metadata records
	BasePath string
}

type MetadataRecord struct {
	RepoKey   string
	RepoName  string
	Reference string
	SHA256    string
}

type ManifestStore struct {
	// Base directory for all manifests
	BasePath      string
	ReferrerStore *ReferrerStore
}

// --------------------
// Helpers
// --------------------

func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	return nil
}

func ensureDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// --------------------
// MetadataFile
// --------------------

// pathForRecord returns the full file path for a metadata record.
func (m *MetadataFile) PathForRecord(record MetadataRecord) string {
	return filepath.Join(
		m.BasePath,
		record.RepoKey,
		record.RepoName,
		"metadata",
		record.Reference+".sha256",
	)
}

func (m *MetadataFile) WriteRecord(record MetadataRecord) error {
	path := m.PathForRecord(record)
	dir := filepath.Dir(path)

	if err := ensureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory for metadata record: %w", err)
	}

	if err := atomicWrite(path, []byte(record.SHA256)); err != nil {
		return fmt.Errorf("failed to write metadata record: %w", err)
	}

	return nil
}

func (m *MetadataFile) ReadRecord(repoKey, repoName, reference string) (MetadataRecord, error) {
	path := m.PathForRecord(MetadataRecord{
		RepoKey:   repoKey,
		RepoName:  repoName,
		Reference: reference,
	})

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return MetadataRecord{}, fmt.Errorf("metadata record not found")
		}
		return MetadataRecord{}, fmt.Errorf("failed to read metadata record: %w", err)
	}

	return MetadataRecord{
		RepoKey:   repoKey,
		RepoName:  repoName,
		Reference: reference,
		SHA256:    string(data),
	}, nil
}

// Raw file access if you still want it

func (m *MetadataFile) Write(data []byte) error {
	// Treat BasePath as a file path here (legacy behavior)
	dir := filepath.Dir(m.BasePath)
	if err := ensureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory for metadata file: %w", err)
	}
	if err := atomicWrite(m.BasePath, data); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}
	return nil
}

func (m *MetadataFile) Read() ([]byte, error) {
	data, err := os.ReadFile(m.BasePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("metadata file not found")
		}
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}
	return data, nil
}

// --------------------
// ManifestStore
// --------------------

func NewManifestStore(basePath string, referrerStore *ReferrerStore) *ManifestStore {
	return &ManifestStore{
		BasePath:      basePath,
		ReferrerStore: referrerStore,
	}
}

func (s *ManifestStore) manifestPath(repoKey, repoName, reference string) string {
	return filepath.Join(
		s.BasePath,
		repoKey,
		repoName,
		"manifests",
		reference,
	)
}

func (s *ManifestStore) PathForRecord(record MetadataRecord) string {
	return s.manifestPath(record.RepoKey, record.RepoName, record.Reference)
}

func (s *ManifestStore) GetManifest(repoKey, repoName, reference string) ([]byte, error) {
	manifestPath := s.manifestPath(repoKey, repoName, reference)

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("manifest not found")
		}
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	return data, nil
}

func (s *ManifestStore) PutManifest(repoKey, repoName, reference string, data []byte) error {
	manifestPath := s.manifestPath(repoKey, repoName, reference)
	dir := filepath.Dir(manifestPath)

	if err := ensureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory for manifest: %w", err)
	}

	if err := atomicWrite(manifestPath, data); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

type OCIManifest struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType"`
	ArtifactType  string            `json:"artifactType,omitempty"`
	Subject       *Descriptor       `json:"subject,omitempty"`
	Descriptors   []Descriptor      `json:"descriptors,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

type Descriptor struct {
	MediaType string `json:"mediaType"`
	Size      int64  `json:"size"`
	Digest    string `json:"digest"`
}

func (s *ManifestStore) PutManifest2(repoKey, repoName, reference string, data []byte) error {
	// 1. Write the manifest itself
	manifestPath := filepath.Join(s.BasePath, repoKey, repoName, "manifests", reference)
	if err := ensureDir(filepath.Dir(manifestPath)); err != nil {
		return err
	}
	if err := atomicWrite(manifestPath, data); err != nil {
		return err
	}

	// 2. Parse manifest
	var m OCIManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}

	// 3. If no subject → nothing to do
	if m.Subject == nil {
		return nil
	}

	// 4. Compute digest of this manifest
	digest := computeDigest(data)

	// 5. Store referrer descriptor
	desc := ReferrerDescriptor{
		Digest:       digest,
		ArtifactType: m.ArtifactType,
		MediaType:    m.MediaType,
		Size:         int64(len(data)),
		Annotations:  m.Annotations,
	}

	return s.ReferrerStore.PutReferrer(repoKey, repoName, m.Subject.Digest, desc)
}

func computeDigest(data []byte) string {
	// Placeholder for actual digest computation logic
	return "sha256:dummy-digest"
}
