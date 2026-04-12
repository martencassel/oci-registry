package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTags(t *testing.T) {
	// Setup
	basePath := "/tmp/oci-registry/tags"
	store := &TagStore{BasePath: basePath}

	// Test data
	repoKey := "test-repo"
	repoName := "test-image"
	tag := "latest"
	digest := "dummy-digest"

	// Test GetTag for non-existent tag
	_, err := store.GetTag(repoKey, repoName, tag)
	if err == nil {
		t.Errorf("GetTag did not return error for non-existent tag")
	}

	// Create tag file
	tagPath := store.tagPath(repoKey, repoName, tag)
	err = os.MkdirAll(filepath.Dir(tagPath), 0755)
	if err != nil {
		t.Fatalf("Failed to create directories for tag path: %v", err)
	}
	err = os.WriteFile(tagPath, []byte(digest), 0644)
	if err != nil {
		t.Fatalf("Failed to write tag file: %v", err)
	}

	// Test GetTag for existing tag
	gotDigest, err := store.GetTag(repoKey, repoName, tag)
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}
	if gotDigest != digest {
		t.Errorf("GetTag returned unexpected digest: got %s, want %s", gotDigest, digest)
	}

	// Test ListTags
	tags, err := store.ListTags(repoKey, repoName)
	if err != nil {
		t.Fatalf("ListTags failed: %v", err)
	}
	if len(tags) != 1 || tags[0] != tag {
		t.Errorf("ListTags returned unexpected tags: got %v, want [%s]", tags, tag)
	}

	// Test SetTag to update existing tag
	newDigest := "updated-digest"
	err = store.SetTag(repoKey, repoName, tag, newDigest)
	if err != nil {
		t.Fatalf("SetTag failed: %v", err)
	}
	gotDigest, err = store.GetTag(repoKey, repoName, tag)
	if err != nil {
		t.Fatalf("GetTag after SetTag failed: %v", err)
	}
	if gotDigest != newDigest {
		t.Errorf("GetTag after SetTag returned unexpected digest: got %s, want %s", gotDigest, newDigest)
	}
}
