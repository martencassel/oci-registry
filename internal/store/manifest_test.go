package store

import "testing"

func TestManifestStore(t *testing.T) {
	// Setup
	basePath := "/tmp/oci-registry/manifests"
	store := &ManifestStore{BasePath: basePath}

	// Test data
	record := MetadataRecord{
		RepoKey:   "test-repo",
		RepoName:  "test-image",
		Reference: "latest",
		SHA256:    "dummy-sha256",
	}

	// Test path generation
	expectedPath := "/tmp/oci-registry/manifests/test-repo/test-image/manifests/latest"
	if store.PathForRecord(record) != expectedPath {
		t.Errorf("pathForRecord returned unexpected path: got %s, want %s", store.PathForRecord(record), expectedPath)
	}

	// Test WriteRecord
	err := store.PutManifest(record.RepoKey, record.RepoName, record.Reference, []byte(record.SHA256))
	if err != nil {
		t.Fatalf("PutManifest failed: %v", err)
	}

	// Test GetManifest
	data, err := store.GetManifest(record.RepoKey, record.RepoName, record.Reference)
	if err != nil {
		t.Fatalf("GetManifest failed: %v", err)
	}
	if string(data) != record.SHA256 {
		t.Errorf("GetManifest returned unexpected data: got %s, want %s", string(data), record.SHA256)
	}

	// Test GetManifest for non-existent record
	_, err = store.GetManifest("nonexistent-repo", "nonexistent-image", "nonexistent-tag")
	if err == nil {
		t.Errorf("GetManifest did not return error for non-existent record")
	}

}
