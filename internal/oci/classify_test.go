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

func TestOCIPush(t *testing.T) {
	// end-1
	//
	// GET /v2/
	req, err := http.NewRequest("GET", "/v2/", nil)
	assert.NoError(t, err)
	meta := ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindPing, meta.Kind)

	// end-2
	//
	// GET / HEAD, /v2/<repoKey>/<name>/blobs/<digest>
	req, err = http.NewRequest("GET", "/v2/oci-local/one/two/three/alpine/blobs/sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindDownloadBlob, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", meta.Digest)

	req, err = http.NewRequest("HEAD", "/v2/oci-local/one/two/three/alpine/blobs/sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindCheckBlobExists, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", meta.Digest)

	// end-3
	// GET / HEAD, /v2/<repoKey>/<name>/manifests/<reference>
	//
	req, err = http.NewRequest("GET", "/v2/oci-local/one/two/three/alpine/manifests/v1.0.0", nil)
	assert.NoError(t, err)
	repoKey, name, reference, ok := parseManifestPath(req.URL.Path)
	assert.True(t, ok)
	assert.Equal(t, "oci-local", repoKey)
	assert.Equal(t, "one/two/three/alpine", name)
	assert.Equal(t, "v1.0.0", reference)

	req, err = http.NewRequest("HEAD", "/v2/oci-local/one/two/three/alpine/manifests/sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindCheckManifestExists, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", meta.Reference)

	req, err = http.NewRequest("HEAD", "/v2/oci-local/one/two/three/alpine/manifests/v1.0.0", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindCheckManifestExists, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "v1.0.0", meta.Reference)

	req, err = http.NewRequest("HEAD", "/v2/oci-local/one/two/three/alpine/manifests/v1.0.0", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindCheckManifestExists, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "v1.0.0", meta.Reference)

	// end-4a, POST	/v2/<name>/blobs/uploads/
	req, err = http.NewRequest("POST", "/v2/oci-local/one/two/three/alpine/blobs/uploads/", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindStartBlobUpload, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "", meta.UploadUUID)
	assert.Equal(t, "", meta.Digest)

	// end-4b	POST	/v2/<name>/blobs/uploads/?digest=<digest>
	req, err = http.NewRequest("POST", "/v2/oci-local/one/two/three/alpine/blobs/uploads/?digest=sha256%3A60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindMonolithicBlobUpload, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "", meta.UploadUUID)
	assert.Equal(t, "sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", meta.Digest)

	// end-5	PATCH	/v2/<name>/blobs/uploads/<reference>
	req, err = http.NewRequest("PATCH", "/v2/oci-local/one/two/three/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindUploadBlobChunk, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", meta.UploadUUID)
	assert.Equal(t, "", meta.Digest)

	// end-6	PUT	/v2/<name>/blobs/uploads/<reference>?digest=<digest>
	req, err = http.NewRequest("PUT", "/v2/oci-local/one/two/three/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000?digest=sha256%3A60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindCompleteBlobUpload, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", meta.UploadUUID)
	assert.Equal(t, "sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", meta.Digest)

	// end-7	PUT	/v2/<name>/manifests/<reference>
	req, err = http.NewRequest("PUT", "/v2/oci-local/one/two/three/alpine/manifests/v1.0.0", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindUploadManifest, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "v1.0.0", meta.Reference)

	// end-8a	GET	/v2/<name>/tags/list
	req, err = http.NewRequest("GET", "/v2/oci-local/one/two/three/alpine/tags/list", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindListTags, meta.Kind)
	assert.Equal(t, "oci-local", repoKey)
	assert.Equal(t, "one/two/three/alpine", name)

	// end-8b	GET	/v2/<name>/tags/list?n=<num_tags>&last=<last_tag>
	req, err = http.NewRequest("GET", "/v2/oci-local/one/two/three/alpine/tags/list?n=10&last=v1.0.0", nil)
	assert.NoError(t, err)
	repoKey, name, ok = parseTagsListPath(req)
	assert.True(t, ok)
	assert.Equal(t, "oci-local", repoKey)
	assert.Equal(t, "one/two/three/alpine", name)
	n := req.URL.Query().Get("n")
	last := req.URL.Query().Get("last")
	assert.Equal(t, "10", n)
	assert.Equal(t, "v1.0.0", last)

	// end-11, POST	/v2/<name>/blobs/uploads/?mount=<digest>&from=<other_name>
	req, err = http.NewRequest("POST", "/v2/oci-local/one/two/three/alpine/blobs/uploads/?mount=sha256%3A60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb&from=other/repo", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindMountBlob, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "", meta.UploadUUID)
	mount := req.URL.Query().Get("mount")
	from := req.URL.Query().Get("from")
	assert.Equal(t, "sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", mount)
	assert.Equal(t, "other/repo", from)

	// end-12a	GET	/v2/<name>/referrers/<digest>
	req, err = http.NewRequest("GET", "/v2/oci-local/one/two/three/alpine/referrers/sha256%3A60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindListReferrers, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", meta.Reference)

	// end-12b	GET	/v2/<name>/referrers/<digest>?artifactType=<artifactType>
	req, err = http.NewRequest("GET", "/v2/oci-local/one/two/three/alpine/referrers/sha256%3A60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb?artifactType=application%2Fvnd.cncf.helm.chart", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindListReferrers, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", meta.Reference)
	artifactType := req.URL.Query().Get("artifactType")
	assert.Equal(t, "application/vnd.cncf.helm.chart", artifactType)

	// end-13	GET	/v2/<name>/blobs/uploads/<reference>
	req, err = http.NewRequest("GET", "/v2/oci-local/one/two/three/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindGetUploadStatus, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", meta.UploadUUID)
	assert.Equal(t, "", meta.Digest)

	// end-14	DELETE	/v2/<name>/blobs/uploads/<reference>
	req, err = http.NewRequest("DELETE", "/v2/oci-local/one/two/three/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000", nil)
	assert.NoError(t, err)
	meta = ClassifyRequest(req.Method, req.URL.Path, req)
	assert.Equal(t, KindCancelBlobUpload, meta.Kind)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "one/two/three/alpine", meta.Repository)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", meta.UploadUUID)
	assert.Equal(t, "", meta.Digest)
}
