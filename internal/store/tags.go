package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type TagStore struct {
	BasePath string
}

func NewTagStore(basePath string) *TagStore {
	return &TagStore{BasePath: basePath}
}

func (s *TagStore) tagPath(repoKey, repoName, tag string) string {
	return filepath.Join(s.BasePath, repoKey, repoName, "tags", tag)
}

// -----------------------------
// List all tags
// -----------------------------
func (s *TagStore) ListTags(repoKey, repoName string) ([]string, error) {
	dir := filepath.Join(s.BasePath, repoKey, repoName, "tags")

	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}

	tags := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			tags = append(tags, e.Name())
		}
	}
	return tags, nil
}

// -----------------------------
// Get digest for a tag
// -----------------------------
func (s *TagStore) GetTag(repoKey, repoName, tag string) (string, error) {
	path := s.tagPath(repoKey, repoName, tag)

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("tag not found")
		}
		return "", err
	}

	return string(data), nil
}

// -----------------------------
// Create or update a tag
// -----------------------------
func (s *TagStore) SetTag(repoKey, repoName, tag, digest string) error {
	path := s.tagPath(repoKey, repoName, tag)
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	return atomicWrite(path, []byte(digest))
}

// -----------------------------
// Delete a tag
// -----------------------------
func (s *TagStore) DeleteTag(repoKey, repoName, tag string) error {
	path := s.tagPath(repoKey, repoName, tag)
	if err := os.Remove(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return nil
}
