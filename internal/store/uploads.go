package store

import (
	"fmt"
	"os"
	"path/filepath"

	// google uuid
	uuid "github.com/google/uuid"
	digest "github.com/opencontainers/go-digest"
)

type Uploads struct {
	Blobs    BlobStore
	BasePath string
}

func NewUploads(basePath string, blobs BlobStore) *Uploads {
	return &Uploads{
		BasePath: basePath,
		Blobs:    blobs,
	}
}

func (u *Uploads) CreateUploadSession(repoKey string) (string, error) {
	// Generate a unique upload ID (e.g., using UUID)
	uploadID := uuid.New().String()
	// Create a file to represent the upload session
	uploadPath := fmt.Sprintf("%s/%s", u.BasePath, uploadID)
	if err := os.MkdirAll(u.BasePath, 0755); err != nil {
		return "", err
	}

	f, err := os.Create(uploadPath)
	if err != nil {
		return "", err
	}
	f.Close()
	return uploadID, nil
}

// WriteChunk writes a chunk of data to the upload session file
func (u *Uploads) WriteChunk(uploadID string, data []byte) error {
	uploadPath := fmt.Sprintf("%s/%s", u.BasePath, uploadID)
	f, err := os.OpenFile(uploadPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

func (u *Uploads) GetUploadSession(uploadID string) (string, error) {
	uploadPath := fmt.Sprintf("%s/%s", u.BasePath, uploadID)
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		return "", fmt.Errorf("upload session not found")
	}
	return uploadPath, nil
}

func (u *Uploads) DeleteUploadSession(uploadID string) error {
	uploadPath := fmt.Sprintf("%s/%s", u.BasePath, uploadID)
	return os.Remove(uploadPath)
}

func (u *Uploads) FinalizeUploadSession(uploadID string, dgst string) error {
	uploadPath := fmt.Sprintf("%s/%s", u.BasePath, uploadID)

	// Check if the upload session exists
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		return fmt.Errorf("upload session not found")
	}

	// Move the uploaded file to the blob store with the correct digest
	blobPath := u.Blobs.BlobPath(digest.Digest(dgst))
	err := os.MkdirAll(filepath.Dir(blobPath), 0755)
	if err != nil {
		return err
	}
	if err := os.Rename(uploadPath, blobPath); err != nil {
		return err
	}

	return nil
}
