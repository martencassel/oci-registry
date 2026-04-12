package oci

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestLexer1(t *testing.T) {
	path := "/v2/oci-local/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000?digest=sha256%3A60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb"
	tokens, err := lex(path)
	assert.NoError(t, err)
	expected := []Token{
		{Value: "oci-local", IsVerb: false, IsSpecial: false},
		{Value: "alpine", IsVerb: false, IsSpecial: false},
		{Value: "blobs", IsVerb: true, IsSpecial: false},
		{Value: "uploads", IsVerb: true, IsSpecial: false},
		{Value: "123e4567-e89b-12d3-a456-426614174000", IsVerb: false, IsSpecial: false},
	}
	assert.Equal(t, expected, tokens)
}

func TestLexerInvalidPath(t *testing.T) {
	path := "/invalid/path"
	_, err := lex(path)
	assert.ErrorIs(t, err, ErrInvalidPath)
}

func TestLexerEmptySegment(t *testing.T) {
	path := "/v2/oci-local//blobs/uploads/"
	_, err := lex(path)
	assert.ErrorIs(t, err, ErrInvalidPath)
}

func TestLexerTrailingSlash(t *testing.T) {
	path := "/v2/oci-local/alpine/blobs/uploads/"
	tokens, err := lex(path)
	assert.NoError(t, err)
	expected := []Token{
		{Value: "oci-local", IsVerb: false, IsSpecial: false},
		{Value: "alpine", IsVerb: false, IsSpecial: false},
		{Value: "blobs", IsVerb: true, IsSpecial: false},
		{Value: "uploads", IsVerb: true, IsSpecial: false},
	}
	assert.Equal(t, expected, tokens)
}

func TestLexerNoV2Prefix(t *testing.T) {
	path := "/oci-local/alpine/blobs/uploads/"
	_, err := lex(path)
	assert.ErrorIs(t, err, ErrInvalidPath)
}

func TestLexerOnlyV2(t *testing.T) {
	path := "/v2/"
	tokens, err := lex(path)
	assert.NoError(t, err)
	expected := []Token{}
	assert.Equal(t, expected, tokens)
}

// /v2/<name>/blobs/uploads/<reference>?digest=<digest>
func TestParseBlobUploadPutWithDigest2(t *testing.T) {
	meta, err := ParseV1("PUT", "/v2/oci-local/alpine/blobs/uploads/123e4567-e89b-12d3-a456-426614174000?digest=sha256%3A60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb")
	assert.NoError(t, err)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "alpine", meta.Repository)
	assert.Equal(t, "uploads", meta.SubVerb)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", meta.UploadUUID)
	assert.Equal(t, "sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", meta.Query.Get("digest"))
}

// POST	/v2/<name>/blobs/uploads/?digest=<digest>
func TestParseBlobUploadPostWithDigest(t *testing.T) {
	meta, err := ParseV1("POST", "/v2/oci-local/alpine/blobs/uploads/?digest=sha256%3A60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb")
	assert.NoError(t, err)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "alpine", meta.Repository)
	assert.Equal(t, "uploads", meta.SubVerb)
	assert.Equal(t, "sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", meta.Query.Get("digest"))
}

// /v2/<name>/blobs/<digest>
func TestParseBlobHeadWithDigest(t *testing.T) {
	meta, err := ParseV1("HEAD", "/v2/oci-local/alpine/blobs/sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb")
	assert.NoError(t, err)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "alpine", meta.Repository)
	assert.Equal(t, "sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", meta.Digest)
}

// /v2/<name>/manifests/<reference>
func TestParseManifestHeadWithReference(t *testing.T) {
	meta, err := ParseV1("HEAD", "/v2/oci-local/alpine/manifests/latest")
	assert.NoError(t, err)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "alpine", meta.Repository)
	assert.Equal(t, "latest", meta.Reference)
}

// /v2/<name>/tags/list
func TestParseTagsList(t *testing.T) {
	meta, err := ParseV1("GET", "/v2/oci-local/alpine/tags/list")
	assert.NoError(t, err)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "alpine", meta.Repository)
	assert.Equal(t, "list", meta.SubVerb)
}

// /v2/<name>/tags/list?n=<integer>&last=<tagname>
func TestParseTagsListWithQuery(t *testing.T) {
	meta, err := ParseV1("GET", "/v2/oci-local/alpine/tags/list?n=10&last=beta")
	assert.NoError(t, err)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "alpine", meta.Repository)
	assert.Equal(t, "list", meta.SubVerb)
	assert.Equal(t, "10", meta.Query.Get("n"))
	assert.Equal(t, "beta", meta.Query.Get("last"))
}

// POST mount /v2/<name>/tags/list?n=<integer>&last=<tagname>
func TestParseBlobMountWithQuery(t *testing.T) {
	meta, err := ParseV1("POST", "/v2/oci-local/alpine/blobs/uploads/?digest=sha256%3A60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb")
	assert.NoError(t, err)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "alpine", meta.Repository)
	assert.Equal(t, "uploads", meta.SubVerb)
	assert.Equal(t, "sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", meta.Query.Get("digest"))
}

// /v2/<name>/referrers/<digest>
func TestParseReferrersWithDigest(t *testing.T) {
	meta, err := ParseV1("GET", "/v2/oci-local/alpine/referrers/sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb")
	assert.NoError(t, err)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "alpine", meta.Repository)
	assert.Equal(t, "sha256:60bcbda02295d1a14b019504e031b5fac8410b19b08460af05b2218069a44efb", meta.Reference)
}

// /v2/<name>/referrers/<digest>
func TestParseReferrersWithReference(t *testing.T) {
	meta, err := ParseV1("GET", "/v2/oci-local/alpine/referrers/latest")
	assert.NoError(t, err)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "alpine", meta.Repository)
	assert.Equal(t, "latest", meta.Reference)
}

// /v2/<name>/blobs/uploads/<reference>
func TestParseBlobUploadWithReference(t *testing.T) {
	meta, err := ParseV1("POST", "/v2/oci-local/alpine/blobs/uploads/latest")
	assert.NoError(t, err)
	assert.Equal(t, "oci-local", meta.RepoKey)
	assert.Equal(t, "alpine", meta.Repository)
	assert.Equal(t, "uploads", meta.SubVerb)
	assert.Equal(t, "latest", meta.UploadUUID)
}
