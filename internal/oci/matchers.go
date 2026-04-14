package oci

import (
	"net/http"
	"regexp"
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
