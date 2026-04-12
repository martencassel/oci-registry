package oci

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestClassifyStreamedPush(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   Kind
	}{
		{method: "HEAD", path: "/v2/oci-local/alpine/blobs/sha256:989e799e634906e94dc9a5ee2ee26fc92ad260522990f26e707861a5f52bf64e", want: KindBlobHead},
		{method: "HEAD", path: "/v2/oci-local/alpine/blobs/sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", want: KindBlobHead},
		{method: "HEAD", path: "/v2/oci-local/hello/blobs/sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", want: KindBlobHead},
		{method: "POST", path: "/v2/oci-local/alpine/blobs/uploads/", want: KindBlobUploadPost},
		{method: "PATCH", path: "/v2/oci-local/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000", want: KindBlobUploadPatch},
		{method: "PUT", path: "/v2/oci-local/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000?digest=ha256%3A60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", want: KindBlobUploadPut},
	}

	for _, tt := range tests {
		meta := ClassifyRequest(tt.method, tt.path)
		assert.Equal(t, tt.want, meta.Kind)
		assert.Equal(t, "oci-local", meta.RepoKey)
	}

}

func TestParseBlobUpload(t *testing.T) {
	meta, err := ParseV1("POST", "/v2/oci-local/alpine/blobs/uploads")
	assert.NoError(t, err)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "alpine", meta.Repository)
	assert.Equal(t, "uploads", meta.SubVerb)
}

func TestParseBlobUploadWithUUID(t *testing.T) {
	meta, err := ParseV1("PATCH", "/v2/oci-local/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000")
	assert.NoError(t, err)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "alpine", meta.Repository)
	assert.Equal(t, "uploads", meta.SubVerb)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", meta.UploadUUID)
}

func TestParseBlobUploadPutWithDigest(t *testing.T) {
	meta, err := ParseV1("PUT", "/v2/oci-local/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000?digest=sha256%3A60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb")
	assert.NoError(t, err)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "alpine", meta.Repository)
	assert.Equal(t, "uploads", meta.SubVerb)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", meta.UploadUUID)
	assert.Equal(t, "sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", meta.Query.Get("digest"))
}
