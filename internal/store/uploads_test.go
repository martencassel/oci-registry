package store

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	digest "github.com/opencontainers/go-digest"
)

func TestUploads(t *testing.T) {
	// Setup
	basePath := "/tmp/oci-registry/uploads"
	blobs := NewFSBlobStore("/tmp/oci-registry/blobs")
	uploads := NewUploads(basePath, blobs)

	// Test CreateUploadSession
	repoKey := "test-repo"
	uploadID, err := uploads.CreateUploadSession(repoKey)
	if err != nil {
		t.Fatalf("CreateUploadSession failed: %v", err)
	}
	// Check that the upload session file was created
	uploadPath := fmt.Sprintf("%s/%s", basePath, uploadID)
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		t.Fatalf("Upload session file was not created at expected path: %s", uploadPath)
	}

	// Test GetUploadSession
	sessionPath, err := uploads.GetUploadSession(uploadID)
	if err != nil {
		t.Fatalf("GetUploadSession failed: %v", err)
	}
	expectedPath := fmt.Sprintf("%s/%s", basePath, uploadID)
	if sessionPath != expectedPath {
		t.Errorf("GetUploadSession returned unexpected path: got %s, want %s", sessionPath, expectedPath)
	}

	// Write chunk
	chunkData := []byte("test chunk data")
	br := io.NopCloser(bytes.NewReader(chunkData))
	err = uploads.WriteChunk(uploadID, br)
	if err != nil {
		t.Fatalf("WriteChunk failed: %v", err)
	}

	// Commit it
	chunkDataDgst := digest.FromBytes(chunkData)
	err = uploads.FinalizeUploadSession(uploadID, chunkDataDgst.String())
	if err != nil {
		t.Fatalf("FinalizeUploadSession failed: %v", err)
	}

	// Verify blob is in blob store
	has, err := blobs.Has(nil, chunkDataDgst)
	if err != nil {
		t.Fatalf("BlobStore Has failed: %v", err)
	}
	if !has {
		t.Errorf("BlobStore does not have expected blob with digest: %s", chunkDataDgst)
	}

}
