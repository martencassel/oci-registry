package main

import (
	"github.com/gin-gonic/gin"
	config "github.com/martencassel/oci-registry/internal/config"
	"github.com/martencassel/oci-registry/internal/handler"
	"github.com/martencassel/oci-registry/internal/store"
	"github.com/martencassel/oci-registry/internal/transport"
)

func main() {

	repoConfigProvider := config.NewRepoConfigInMemory()
	repoConfigProvider.AddRepoConfig(&config.RepoConfig{
		RepoKey: "oci-local",
	})
	blobs := store.NewFSBlobStore("/tmp/oci-registry")
	uploads := store.NewUploads("/tmp/oci-registry/uploads", blobs)
	referers := store.NewReferrerStore("/tmp/oci-registry/referrers")
	manifestStore := store.NewManifestStore("/tmp/oci-registry/manifests", referers)
	handler := handler.NewOCIRegistryHandler(repoConfigProvider, uploads, manifestStore)
	r := gin.Default()
	r.Use(transport.TraceHandler())
	r.Use(func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
		c.Abort() // Prevent further handlers from running
	})
	r.Run("172.21.22.152:8080")
}
