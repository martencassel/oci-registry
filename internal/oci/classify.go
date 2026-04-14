package oci

import (
	"net/http"
	"net/url"
	"strings"
)

// ErrNotOCI
var ErrNotOCI = http.ErrNotSupported

type RequestMeta struct {
	Kind        Kind
	RepoKey     string
	Verb        VerbType
	Repository  string
	SubVerb     string
	Repo        string
	Digest      string
	Reference   string
	IsDigestRef bool
	UploadUUID  string
	Query       url.Values
	RawPath     string
}

type ResponseMeta struct {
	Kind          Kind
	ContentType   string
	ContentLength int64
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}

func ExtractRepoKey(path string) (repoKey string, remainder string, ok bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 || parts[0] != "v2" {
		return "", "", false
	}

	repoKey = parts[1]
	remainder = "/" + strings.Join(parts[2:], "/")
	return repoKey, remainder, true
}

// VerbType
type VerbType int

const (
	VerbUnknown VerbType = iota
	VerbBlobs
	VerbManifests
	VerbTags
	VerbReferrers
)

type ParseResult struct {
	IsPing   bool
	IsUpload bool

	RepoKey    string
	Repository string

	// Operation classification
	Verb    VerbType
	SubVerb string // e.g. "uploads" for blobs/uploads or "list" for tags/list

	// Identifiers
	Digest     string
	Reference  string
	UploadUUID string

	// Query parameters
	Query url.Values

	// Raw
	RawPath string
}

type Kind int

func (k Kind) String() string {
	switch k {
	case KindUnknown:
		return "Unknown"
	case KindPing:
		return "Ping"
	case KindDownloadBlob:
		return "DownloadBlob"
	case KindCheckBlobExists:
		return "CheckBlobExists"
	case KindGetManifest:
		return "GetManifest"
	case KindCheckManifestExists:
		return "CheckManifestExists"
	case KindStartBlobUpload:
		return "StartBlobUpload"
	case KindMonolithicBlobUpload:
		return "MonolithicBlobUpload"
	case KindUploadBlobChunk:
		return "UploadBlobChunk"
	case KindCompleteBlobUpload:
		return "CompleteBlobUpload"
	case KindCancelBlobUpload:
		return "CancelBlobUpload"
	case KindGetUploadStatus:
		return "GetUploadStatus"
	case KindMountBlob:
		return "MountBlob"
	case KindGetBlobUpload:
		return "GetBlobUpload"
	case KindUploadManifest:
		return "UploadManifest"
	case KindDeleteBlobUpload:
		return "DeleteBlobUpload"
	case KindListTags:
		return "ListTags"
	case KindListTagsPaginated:
		return "ListTagsPaginated"
	case KindListReferrers:
		return "ListReferrers"
	default:
		return "Unknown"
	}
}

const (
	KindUnknown Kind = iota

	// Meta
	KindPing

	// Blob Pull
	KindDownloadBlob
	KindCheckBlobExists

	// Manifest Pull
	KindGetManifest
	KindCheckManifestExists

	// Blob Upload
	KindStartBlobUpload
	KindMonolithicBlobUpload
	KindUploadBlobChunk
	KindCompleteBlobUpload
	KindCancelBlobUpload
	KindGetUploadStatus
	KindMountBlob
	KindGetBlobUpload

	// Manifest Push
	KindUploadManifest
	KindDeleteBlobUpload
	KindDeleteManifest

	// Tag Discovery
	KindListTags
	KindListTagsPaginated

	// Referrers API
	KindListReferrers
)

func ClassifyRequest(method, path string, req *http.Request) RequestMeta {
	if method == "GET" && path == "/v2/" {
		return RequestMeta{Kind: KindPing}
	}
	if meta, ok := checkBlob(method, path); ok {
		return meta
	}
	if meta, ok := checkManifest(method, path); ok {
		return meta
	}
	if meta, ok := checkBlobUpload(method, req); ok {
		return meta
	}
	if meta, ok := checkTagsList(method, req); ok {
		return meta
	}
	if meta, ok := checkReferrers(method, path); ok {
		return meta
	}
	return RequestMeta{Kind: KindUnknown}
}
