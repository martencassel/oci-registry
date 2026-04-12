package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type ReferrerStore struct {
	BasePath string
}

func NewReferrerStore(basePath string) *ReferrerStore {
	return &ReferrerStore{BasePath: basePath}
}

func (s *ReferrerStore) referrerDir(repoKey, repoName, subjectDigest string) string {
	return filepath.Join(s.BasePath, repoKey, repoName, "referrers", subjectDigest)
}

func (s *ReferrerStore) referrerPath(repoKey, repoName, subjectDigest, referrerDigest string) string {
	return filepath.Join(s.referrerDir(repoKey, repoName, subjectDigest), referrerDigest+".json")
}

type ReferrerDescriptor struct {
	Digest       string            `json:"digest"`
	ArtifactType string            `json:"artifactType"`
	MediaType    string            `json:"mediaType"`
	Size         int64             `json:"size"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

// -----------------------------
// Add or update a referrer
// -----------------------------
func (s *ReferrerStore) PutReferrer(repoKey, repoName, subjectDigest string, desc ReferrerDescriptor) error {
	dir := s.referrerDir(repoKey, repoName, subjectDigest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create referrer directory: %w", err)
	}

	path := s.referrerPath(repoKey, repoName, subjectDigest, desc.Digest)

	data, err := json.Marshal(desc)
	if err != nil {
		return fmt.Errorf("failed to marshal referrer descriptor: %w", err)
	}

	return atomicWrite(path, data)
}

// -----------------------------
// List referrers for a digest
// Optional filtering by artifactType
// -----------------------------
func (s *ReferrerStore) GetReferrers(repoKey, repoName, subjectDigest, artifactType string) ([]ReferrerDescriptor, error) {
	dir := s.referrerDir(repoKey, repoName, subjectDigest)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []ReferrerDescriptor{}, nil
		}
		return nil, err
	}

	var out []ReferrerDescriptor

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}

		var desc ReferrerDescriptor
		if err := json.Unmarshal(data, &desc); err != nil {
			return nil, err
		}

		if artifactType != "" && desc.ArtifactType != artifactType {
			continue
		}

		out = append(out, desc)
	}

	return out, nil
}
