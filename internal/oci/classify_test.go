package oci

import (
	"net/http"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestOCIRequestClassification(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		wantKind       Kind
		wantRepo       string
		wantRef        string
		wantDigest     string
		wantUploadUUID string
	}{
		{
			// end-1	GET	/v2/
			name:     "Ping",
			method:   "GET",
			path:     "/v2/",
			wantKind: KindPing,
		},
		{
			// end-2	GET / HEAD	/v2/<name>/blobs/<digest>
			name:       "Download Blob",
			method:     "GET",
			path:       "/v2/oci-local/alpine/blobs/sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantKind:   KindDownloadBlob,
			wantRepo:   "oci-local",
			wantDigest: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			// end-3	GET / HEAD	/v2/<name>/manifests/<reference>
			name:       "Check Blob Exists",
			method:     "HEAD",
			path:       "/v2/oci-local/alpine/blobs/sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantKind:   KindCheckBlobExists,
			wantRepo:   "oci-local",
			wantDigest: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			// end-4a	POST	/v2/<name>/blobs/uploads/
			name:     "Start Blob Upload",
			method:   "POST",
			path:     "/v2/oci-local/alpine/blobs/uploads/",
			wantKind: KindStartBlobUpload,
			wantRepo: "oci-local",
		},
		{
			// end-4b	POST	/v2/<name>/blobs/uploads/?digest=<digest>
			name:       "Start Blob Upload with digest",
			method:     "POST",
			path:       "/v2/oci-local/alpine/blobs/uploads/?digest=sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantKind:   KindMonolithicBlobUpload,
			wantRepo:   "oci-local",
			wantDigest: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			// end-5	PATCH	/v2/<name>/blobs/uploads/<reference>
			name:           "Patch Blob Upload Chunk",
			method:         "PATCH",
			path:           "/v2/oci-local/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000",
			wantKind:       KindUploadBlobChunk,
			wantRepo:       "oci-local",
			wantUploadUUID: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			// end-6	PUT	/v2/<name>/blobs/uploads/<reference>?digest=<digest>
			name:           "Complete Blob Upload",
			method:         "PUT",
			path:           "/v2/oci-local/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000?digest=sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantKind:       KindCompleteBlobUpload,
			wantRepo:       "oci-local",
			wantUploadUUID: "123e4567-e89b-12d3-a456-426614174000",
			wantDigest:     "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			// end-7	PUT	/v2/<name>/manifests/<reference>
			name:     "Upload Manifest",
			method:   "PUT",
			path:     "/v2/oci-local/alpine/manifests/v1.0.0",
			wantKind: KindUploadManifest,
			wantRepo: "oci-local",
			wantRef:  "v1.0.0",
		},
		{
			// end-8a	GET	/v2/<name>/tags/list
			name:     "List Tags",
			method:   "GET",
			path:     "/v2/oci-local/alpine/tags/list",
			wantKind: KindListTags,
			wantRepo: "oci-local",
		},
		{
			// end-8b	GET	/v2/<name>/tags/list?n=<integer>&last=<tagname>
			name:     "List Tags with pagination",
			method:   "GET",
			path:     "/v2/oci-local/alpine/tags/list?n=10&last=v1.0.0",
			wantKind: KindListTags,
			wantRepo: "oci-local",
		},
		{
			// end-9	DELETE	/v2/<name>/manifests/<reference>
			name:     "Delete Manifest",
			method:   "DELETE",
			path:     "/v2/oci-local/alpine/manifests/v1.0.0",
			wantKind: KindDeleteManifest,
			wantRepo: "oci-local",
			wantRef:  "v1.0.0",
		},
		{
			// end-10	DELETE	/v2/<name>/blobs/<digest>
			name:       "Delete Blob",
			method:     "DELETE",
			path:       "/v2/oci-local/alpine/blobs/sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantKind:   KindDeleteBlobUpload,
			wantRepo:   "oci-local",
			wantDigest: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			// end-11	POST	/v2/<name>/blobs/uploads/?mount=<digest>&from=<other_name>
			name:       "Mount Blob",
			method:     "POST",
			path:       "/v2/oci-local/alpine/blobs/uploads/?mount=sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855&from=other-repo",
			wantKind:   KindMountBlob,
			wantRepo:   "oci-local",
			wantDigest: "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			// end-12a	GET	/v2/<name>/referrers/<digest>
			name:     "List Referrers without artifactType",
			method:   "GET",
			path:     "/v2/oci-local/alpine/referrers/sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantKind: KindListReferrers,
			wantRepo: "oci-local",
			wantRef:  "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			// end-12b	GET	/v2/<name>/referrers/<digest>?artifactType=<artifactType>
			name:       "List Referrers",
			method:     "GET",
			path:       "/v2/oci-local/alpine/referrers/sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855?artifactType=application/vnd.cncf.notary.signature",
			wantKind:   KindListReferrers,
			wantRepo:   "oci-local",
			wantRef:    "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			wantDigest: "application/vnd.cncf.notary.signature",
		},
		{
			// end-13	GET	/v2/<name>/blobs/uploads/<reference>
			name:           "Get Blob Upload Status",
			method:         "GET",
			path:           "/v2/oci-local/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000",
			wantKind:       KindGetUploadStatus,
			wantRepo:       "oci-local",
			wantUploadUUID: "123e4567-e89b-12d3-a456-426614174000",
		},
		{
			// end-14	DELETE	/v2/<name>/blobs/uploads/<reference>
			name:           "Cancel Blob Upload",
			method:         "DELETE",
			path:           "/v2/oci-local/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000",
			wantKind:       KindCancelBlobUpload,
			wantRepo:       "oci-local",
			wantUploadUUID: "123e4567-e89b-12d3-a456-426614174000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			assert.NoError(t, err)
			meta := ClassifyRequest(req.Method, req.URL.Path, req)
			assert.Equal(t, tt.wantKind, meta.Kind)
			assert.Equal(t, tt.wantRepo, meta.RepoKey)
			assert.Equal(t, tt.wantRef, meta.Reference)
			assert.Equal(t, tt.wantDigest, meta.Digest)
			assert.Equal(t, tt.wantUploadUUID, meta.UploadUUID)
		})
	}
}
