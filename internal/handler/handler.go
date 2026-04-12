package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"

	uuid "github.com/google/uuid"
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
	meta := oci.ClassifyRequest(r.Method, r.URL.Path)

	switch meta.Kind {
	case oci.KindBlobUploadPost:
		h.handleBlobUploadPost(w, r, meta, cfg)
	case oci.KindBlobUploadPatch:
		h.handleBlobUploadPatch(w, r, meta, cfg)
	case oci.KindBlobUploadPut:
		h.handleBlobUploadPut(w, r, meta, cfg)
	case oci.KindBlobHead:
		h.handleBlobHead(w, r, meta, cfg)
	case oci.KindManifestHead:
		h.handleManifestHead(w, r, meta, cfg)
	default:
		http.Error(w, fmt.Sprintf("unsupported request kind: %s", meta.Kind), http.StatusNotImplemented)
	}

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

func (h *OCIRegistryHandler) handleManifestHead(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta, cfg *config.RepoConfig) {
	var err oci.OCIError
	err = oci.ErrManifestNotFound
	ociErr := oci.ToOCI(err)
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
	oci.WriteError(w, ociErr)
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

// POST /v2/oci-local/alpine/blobs/uploads/
// or
// POST /v2/oci-local/alpine/blobs/uploads/{uuid}
func (h *OCIRegistryHandler) handleCreateBlobUpload(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta, cfg *config.RepoConfig) {
	uploadUUID := uuid.New().String()
	err := os.MkdirAll("/tmp/oci-registry", 0755)
	if err != nil {
		log.Errorf("Failed to create temp directory: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	uploadPath := fmt.Sprintf("/tmp/oci-registry/%s", uploadUUID)
	// Create an empty file to represent the upload session
	f, err := os.Create(uploadPath)
	if err != nil {
		log.Errorf("Failed to create upload file: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	f.Close()
	log.Infof("Created blob upload session for repoKey: %s, uploadUUID: %s", meta.RepoKey, uploadUUID)
	// Accepted
	w.WriteHeader(http.StatusAccepted)
	// Location: /v2/oci-local/alpine/blobs/uploads/{uuid}
	w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/uploads/%s", meta.RepoKey, uploadUUID))
	// Docker-Upload-UUID: {uuid}
	w.Header().Set("Docker-Upload-UUID", uploadUUID)
	// Docker-Distribution-API-Version: registry/2.0
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")

}

func (h *OCIRegistryHandler) handleBlobUploadPatch(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta, cfg *config.RepoConfig) {
	id := meta.UploadUUID
	log.Infof("Received blob upload PATCH for repoKey: %s, uploadUUID: %s", meta.RepoKey, id)
	// For now, just return 202 Accepted with Location header pointing to the upload URL
	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/uploads/%s", meta.RepoKey, id))
	w.Header().Set("Docker-Upload-UUID", id)
	w.Header().Set("Docker-Distribution-API-Version", "registry/2.0")
}

func (h *OCIRegistryHandler) handleBlobUploadPut(w http.ResponseWriter, r *http.Request, meta oci.RequestMeta, cfg *config.RepoConfig) {
	id := meta.UploadUUID
	digest := meta.Query.Get("digest")
	log.Infof("Received blob upload PUT for repoKey: %s, uploadUUID: %s, digest: %s", meta.RepoKey, id, digest)
	// For now, just return 201 Created with Location header pointing to the blob URL
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/%s", meta.RepoKey, digest))
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
