package oci

import (
	"net/http"
	"net/url"
	"regexp"
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

func checkBlob(method string, path string) (RequestMeta, bool) {
	if repoKey, name, digest, ok := parseBlobPath(path); ok {
		switch method {
		case "GET":
			return RequestMeta{Kind: KindDownloadBlob, RepoKey: repoKey, Repository: name, Digest: digest}, true
		case "HEAD":
			return RequestMeta{Kind: KindCheckBlobExists, RepoKey: repoKey, Repository: name, Digest: digest}, true
		}
	}
	return RequestMeta{}, false
}

func checkManifest(method string, path string) (RequestMeta, bool) {
	if repoKey, name, reference, ok := parseManifestPath(path); ok {
		switch method {
		case "GET":
			return RequestMeta{Kind: KindGetManifest, RepoKey: repoKey, Repository: name, Reference: reference}, true
		case "HEAD":
			return RequestMeta{Kind: KindCheckManifestExists, RepoKey: repoKey, Repository: name, Reference: reference}, true
		case "PUT":
			return RequestMeta{Kind: KindUploadManifest, RepoKey: repoKey, Repository: name, Reference: reference}, true
		}
	}
	return RequestMeta{}, false
}

func checkBlobUpload(method string, req *http.Request) (RequestMeta, bool) {
	if repoKey, name, uploadUUID, digest, mount, from, ok := parseBlobUploadPath(req); ok {
		switch method {
		case "POST":
			if mount != "" && from != "" {
				return RequestMeta{Kind: KindMountBlob, RepoKey: repoKey, Repository: name, UploadUUID: uploadUUID, Digest: digest}, true
			}
			if digest != "" && uploadUUID == "" {
				return RequestMeta{Kind: KindMonolithicBlobUpload, RepoKey: repoKey, Repository: name, Digest: digest}, true
			}
			return RequestMeta{Kind: KindStartBlobUpload, RepoKey: repoKey, Repository: name, UploadUUID: uploadUUID, Digest: digest}, true

		case "PATCH":
			if uploadUUID != "" {
				return RequestMeta{Kind: KindUploadBlobChunk, RepoKey: repoKey, Repository: name, UploadUUID: uploadUUID, Digest: digest}, true
			}

		case "PUT":
			if uploadUUID != "" && digest != "" {
				return RequestMeta{Kind: KindCompleteBlobUpload, RepoKey: repoKey, Repository: name, UploadUUID: uploadUUID, Digest: digest}, true
			}

		case "GET":
			if uploadUUID != "" {
				return RequestMeta{Kind: KindGetUploadStatus, RepoKey: repoKey, Repository: name, UploadUUID: uploadUUID, Digest: digest}, true
			}

		case "DELETE":
			if uploadUUID != "" {
				return RequestMeta{Kind: KindCancelBlobUpload, RepoKey: repoKey, Repository: name, UploadUUID: uploadUUID, Digest: digest}, true
			}
		}
	}
	return RequestMeta{}, false
}

func checkTagsList(method string, req *http.Request) (RequestMeta, bool) {
	if repoKey, name, ok := parseTagsListPath(req); ok {
		switch method {
		case "GET":
			if req.URL.Query().Has("n") || req.URL.Query().Has("last") {
				return RequestMeta{Kind: KindListTagsPaginated, RepoKey: repoKey, Repository: name}, true
			}
			return RequestMeta{Kind: KindListTags, RepoKey: repoKey, Repository: name}, true
		}
	}
	return RequestMeta{}, false
}

func checkReferrers(method string, path string) (RequestMeta, bool) {
	if repoKey, name, reference, ok := parseReferrersPath(path); ok {
		switch method {
		case "GET":
			return RequestMeta{Kind: KindListReferrers, RepoKey: repoKey, Repository: name, Reference: reference}, true
		}
	}
	return RequestMeta{}, false
}

func ClassifyRequest(method, path string, req *http.Request) RequestMeta {
	// 1. Ping
	if method == "GET" && path == "/v2/" {
		return RequestMeta{Kind: KindPing}
	}
	// 2. Blob existence / download
	if meta, ok := checkBlob(method, path); ok {
		return meta
	}
	// 3. Manifest existence / download
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

var blobPathRE = regexp.MustCompile(`^/v2/([^/]+)/(.+)/blobs/([^/]+)$`)

func parseBlobPath(path string) (repoKey string, name string, digest string, ok bool) {
	matches := blobPathRE.FindStringSubmatch(path)
	if len(matches) != 4 {
		return "", "", "", false
	}
	repoKey = matches[1]
	name = matches[2]
	digest = matches[3]
	ok = true
	return repoKey, name, digest, ok
}

var manifestPathRE = regexp.MustCompile(`^/v2/([^/]+)/(.+?)/manifests/([^/]+)$`)

func parseManifestPath(path string) (repoKey string, name string, reference string, ok bool) {
	matches := manifestPathRE.FindStringSubmatch(path)
	if len(matches) != 4 {
		return "", "", "", false
	}
	repoKey = matches[1]
	name = matches[2]
	reference = matches[3]
	ok = true
	return repoKey, name, reference, ok
}

// /v2/<repoKey>/<name>/blobs/uploads
// /v2/<repoKey>/<name>/blobs/uploads/
// /v2/<repoKey>/<name>/blobs/uploads/<uploadUUID>
// /v2/<repoKey>/<name>/blobs/uploads/<uploadUUID>?digest=<digest>
var blobUploadPathRE = regexp.MustCompile(
	`^/v2/([^/]+)(?:/(.+?))?/blobs/uploads` +
		`(?:/([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}))?` +
		`/?$`,
)

func parseBlobUploadPath(req *http.Request) (repoKey, name, uploadUUID, digest, mount, from string, ok bool) {
	matches := blobUploadPathRE.FindStringSubmatch(req.URL.Path)
	if len(matches) != 4 {
		return "", "", "", "", "", "", false
	}

	repoKey = matches[1]
	name = matches[2]
	uploadUUID = matches[3]

	q := req.URL.Query()

	digest = q.Get("digest")
	mount = q.Get("mount")
	from = q.Get("from")

	return repoKey, name, uploadUUID, digest, mount, from, true
}

var tagsListPathRE = regexp.MustCompile(
	`^/v2/([^/]+)/(.+?)/tags/list(?:\?.*)?$`,
)

func parseTagsListPath(req *http.Request) (repoKey, name string, ok bool) {
	matches := tagsListPathRE.FindStringSubmatch(req.URL.Path)
	if len(matches) != 3 {
		return "", "", false
	}
	return matches[1], matches[2], true
}

var referrersPathRE = regexp.MustCompile(`^/v2/([^/]+)/(.+?)/referrers/([^/]+)$`)

func parseReferrersPath(path string) (repoKey string, name string, reference string, ok bool) {
	matches := referrersPathRE.FindStringSubmatch(path)
	if len(matches) != 4 {
		return "", "", "", false
	}
	repoKey = matches[1]
	name = matches[2]
	reference = matches[3]
	ok = true
	return repoKey, name, reference, ok
}
