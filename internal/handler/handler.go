package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	config "github.com/martencassel/oci-registry/internal/config"
	"github.com/martencassel/oci-registry/internal/oci"
	"github.com/martencassel/oci-registry/internal/store"
	log "github.com/sirupsen/logrus"
)

type OCIRegistryHandler struct {
	Uploads            *store.Uploads
	ManifestStore      *store.ManifestStore
	RepoConfigProvider config.RepoConfigProvider
}

func NewOCIRegistryHandler(repoConfigProvider config.RepoConfigProvider, uploads *store.Uploads, manifestStore *store.ManifestStore) *OCIRegistryHandler {
	return &OCIRegistryHandler{
		RepoConfigProvider: repoConfigProvider,
		Uploads:            uploads,
		ManifestStore:      manifestStore,
	}
}

func (h *OCIRegistryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle /v2/ ping
	if r.URL.Path == "/v2/" {
		h.handlePing(w, r, oci.RequestMeta{Kind: oci.KindPing})
		return
	}
	// 1. Extract repo key and remainder
	repoKey, remainder, ok := oci.ExtractRepoKey(r.URL.Path)
	if !ok {
		log.Warnf("Failed to extract repo key from path: %s", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	if !ok {
		http.NotFound(w, r)
		return
	}

	cfg, err := h.RepoConfigProvider.GetRepoConfig(repoKey)
	if err != nil {
		log.Infof("Repository config not found for key: %s", repoKey)
		http.Error(w, fmt.Sprintf("repository config not found for key: %s", repoKey), http.StatusNotFound)
		return
	}
	log.Infof("Received request for repoKey: %s, remainder: %s", repoKey, remainder)

	// 2. Classify request BEFORE rewriting
	log.Infof("Classifying request for method: %s, path: %s", r.Method, r.URL.Path)
	meta := oci.ClassifyRequest(r.Method, r.URL.Path, r)
	log.Infof("Classified request as kind: %s", meta.Kind.String())

	switch meta.Kind {
	case oci.KindPing:
		h.handlePing(w, r, meta)
		return
	//
	// GET / HEAD /v2/<repoKey>/<name>/blobs/<digest>
	//
	case oci.KindDownloadBlob:
		h.handleDownloadBlob(w, r, meta, cfg)
		return
	case oci.KindCheckBlobExists:
		h.handleBlobHead(w, r, meta, cfg)
		return
	//
	// GET / HEAD /v2/<repoKey>/<name>/manifests/<reference>
	//
	case oci.KindGetManifest:
		h.handleGetManifest(w, r, meta, cfg)
		return
	case oci.KindCheckManifestExists:
		h.handleManifestHead(w, r, meta, cfg)
		return

	//
	// POST /v2/<repoKey>/<name>/blobs/uploads/
	// POST /v2/<repoKey>/<name>/blobs/uploads/<uploadUUID>
	// POST /v2/<repoKey>/<name>/blobs/uploads/<uploadUUID>?digest=<digest>
	//
	case oci.KindStartBlobUpload:
		h.handleBlobUploadPost(w, r, meta, cfg)
		return
	//
	// PATCH /v2/<repoKey>/<name>/blobs/uploads/<uploadUUID>
	//
	case oci.KindUploadBlobChunk:
		h.handleBlobUploadPatch(w, r, meta, cfg)
		return
	//
	// PUT /v2/<repoKey>/<name>/blobs/uploads/<uploadUUID>?digest=<digest>
	// PUT /v2/<repoKey>/<name>/blobs/uploads/<uploadUUID>
	//
	case oci.KindCompleteBlobUpload:
		h.handleBlobUploadPut(w, r, meta, cfg)
		return
	//
	// PUT /v2/<repoKey>/<name>/manifests/<reference>
	//
	case oci.KindUploadManifest:
		h.handleManifestPut(w, r, meta, cfg)
		return
	default:
		http.Error(w, fmt.Sprintf("unsupported request kind: %s for %s", meta.Kind, r.URL.Path), http.StatusNotImplemented)
	}

}

// handleDownloadBlob
func (h *OCIRegistryHandler) handleDownloadBlob(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta, cfg *config.RepoConfig) {
	var err oci.OCIError
	err = oci.ErrBlobNotFound
	ociErr := oci.ToOCI(err)
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
	oci.WriteError(w, ociErr)
}

func (h *OCIRegistryHandler) handlePing(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta) {
	// Just return 200 OK for /v2/ ping requests
	w.WriteHeader(http.StatusOK)
	// Header
	// distribution-spec-version: 2.0
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
}

func (h *OCIRegistryHandler) handleBlobHead(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta, cfg *config.RepoConfig) {
	var err oci.OCIError
	err = oci.ErrBlobNotFound
	ociErr := oci.ToOCI(err)
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
	oci.WriteError(w, ociErr)
}

func (h *OCIRegistryHandler) handleGetManifest(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta, cfg *config.RepoConfig) {
	var err oci.OCIError
	err = oci.ErrManifestNotFound
	ociErr := oci.ToOCI(err)
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
	oci.WriteError(w, ociErr)
}

func (h *OCIRegistryHandler) handleManifestHead(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta, cfg *config.RepoConfig) {
	var err oci.OCIError
	err = oci.ErrManifestNotFound
	ociErr := oci.ToOCI(err)
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
	oci.WriteError(w, ociErr)
}

// POST /v2/oci-local/alpine/blobs/uploads/
// or
// POST /v2/oci-local/alpine/blobs/uploads/{uuid}
func (h *OCIRegistryHandler) handleCreateBlobUpload(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta, cfg *config.RepoConfig) {
	id, err := h.Uploads.CreateUploadSession(meta.RepoKey)
	if err != nil {
		log.Errorf("Failed to create upload session: %v", err)
		http.Error(w, "failed to create upload session", http.StatusInternalServerError)
		return
	}
	log.Infof("Created blob upload session with ID: %s", id)
	log.Infof("Created blob upload session for repoKey: %s, uploadUUID: %s", meta.RepoKey, id)
	// Accepted
	w.WriteHeader(http.StatusAccepted)
	// Location: /v2/oci-local/alpine/blobs/uploads/{uuid}
	w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/uploads/%s", meta.RepoKey, id))
	// Docker-Upload-UUID: {uuid}
	w.Header().Set("Docker-Upload-UUID", id)
	// Docker-Distribution-API-Version: registry/2.0
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")

}

// POST http://localhost:8080/v2/oci-local/alpine/blobs/uploads/
func (h *OCIRegistryHandler) handleBlobUploadPost(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta, cfg *config.RepoConfig) {
	if meta.UploadUUID == "" {
		h.handleCreateBlobUpload(w, r, meta, cfg)
		return
	}
	id := meta.UploadUUID
	log.Infof("Received blob upload POST for repoKey: %s, uploadUUID: %s", meta.RepoKey, id)

	// For now, just return 202 Accepted with Location header pointing to the upload URL
	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/uploads/%s", meta.RepoKey, id))
	w.Header().Set("Docker-Upload-UUID", id)
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
}

func SummarizeBinary(body []byte) string {
	h := sha256.Sum256(body)
	return fmt.Sprintf("<binary %d bytes, sha256=%s>", len(body), hex.EncodeToString(h[:]))
}

func (h *OCIRegistryHandler) handleBlobUploadPatch(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta, cfg *config.RepoConfig) {
	id := meta.UploadUUID
	log.Infof("Received blob upload PATCH for repoKey: %s, uploadUUID: %s", meta.RepoKey, id)

	err := h.Uploads.WriteChunk(id, r.Body)
	if err != nil {
		log.Errorf("Failed to write chunk to upload session: %v", err)
		http.Error(w, "failed to write chunk to upload session", http.StatusInternalServerError)
		return
	}

	// For now, just return 202 Accepted with Location header pointing to the upload URL
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/uploads/%s", meta.RepoKey, id))
	w.Header().Set("Docker-Upload-UUID", id)
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
}

func (h *OCIRegistryHandler) handleBlobUploadPut(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta, cfg *config.RepoConfig) {
	id := meta.UploadUUID
	digestQuery := r.URL.Query().Get("digest")
	log.Infof("Received blob upload PUT for repoKey: %s, uploadUUID: %s, digest: %s", meta.RepoKey, id, digestQuery)
	err := h.Uploads.FinalizeUploadSession(id, digestQuery)
	if err != nil {
		log.Errorf("Failed to finalize upload session: %v", err)
		http.Error(w, "failed to finalize upload session", http.StatusInternalServerError)
		return
	}

	// For now, just return 201 Created with Location header pointing to the blob URL
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/%s", meta.RepoKey, digestQuery))
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
}

func (h *OCIRegistryHandler) handleManifestPut(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta, cfg *config.RepoConfig) {
	log.Infof("Received manifest PUT for repoKey: %s, repoName: %s, reference: %s", meta.RepoKey, meta.Repository, meta.Reference)

	if meta.Reference == "" {
		http.Error(w, "reference is required", http.StatusBadRequest)
		return
	}

	// Read manifest content
	manifestData, err := io.ReadAll(r.Body)
	if err != nil {
		log.Errorf("Failed to read manifest data: %v", err)
		http.Error(w, "failed to read manifest data", http.StatusInternalServerError)
		return
	}
	// Write manifest to store
	err = h.ManifestStore.PutManifest(meta.RepoKey, meta.Repository, meta.Reference, manifestData)
	if err != nil {
		log.Errorf("Failed to store manifest: %v", err)
		http.Error(w, "failed to store manifest", http.StatusInternalServerError)
		return
	}
	// Return 201 Created with Location header pointing to the manifest URL
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Location", fmt.Sprintf("/v2/%s/%s/manifests/%s", meta.RepoKey, meta.Repository, meta.Reference))
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
}
