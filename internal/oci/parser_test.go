package oci

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestParser(t *testing.T) {
	// Uploads
	tests := []struct {
		name    string
		path    string
		method  string
		want    *ParseResult
		wantErr error
	}{
		{
			name:   "simple upload",
			method: "POST",
			path:   "/v2/oci-local/ubuntu/blobs/uploads",
			want: &ParseResult{
				RepoKey:    "oci-local",
				Repository: "ubuntu",
				Verb:       VerbBlobs,
				SubVerb:    "uploads",
				IsUpload:   true,
				UploadUUID: "",
				RawPath:    "/v2/oci-local/ubuntu/blobs/uploads",
			},
		},
		{
			name:   "upload with UUID",
			method: "POST",
			path:   "/v2/oci-local/ubuntu/blobs/uploads/123e4567-e89b-12d3-a456-426614174000",
			want: &ParseResult{
				RepoKey:    "oci-local",
				Repository: "ubuntu",
				Verb:       VerbBlobs,
				SubVerb:    "uploads",
				IsUpload:   true,
				UploadUUID: "123e4567-e89b-12d3-a456-426614174000",
				RawPath:    "/v2/oci-local/ubuntu/blobs/uploads/123e4567-e89b-12d3-a456-426614174000",
			},
		},
		{
			name:    "invalid upload path",
			method:  "POST",
			path:    "/v2/oci-local/ubuntu/blobs/uploads/123e4567-e89b-12d3-a456-426614174000/extra",
			wantErr: ErrInvalidPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parser(tt.method, tt.path)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("%s expected error %v, got %v", tt.name, tt.wantErr, err)
			}
			if err != nil {
				return
			}
			assert.Equal(t, tt.want, &got)
		})
	}
}
