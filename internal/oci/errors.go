package oci

import (
	"errors"
	"fmt"
	"net/http"
)

// Errors
const (
	// BLOB_UNKNOWN
	ErrCodeBlobNotFound = "BLOB_UNKNOWN"
	// BLOB_UPLOAD_INVALID
	ErrCodeBlobUploadInvalid = "BLOB_UPLOAD_INVALID"
	// BLOB_UPLOAD_UNKNOWN
	ErrCodeBlobUploadUnknown = "BLOB_UPLOAD_UNKNOWN"
	// DIGEST_INVALID
	ErrCodeDigestInvalid = "DIGEST_INVALID"
	// MANIFEST_BLOB_UNKNOWN
	ErrCodeManifestBlobUnknown = "MANIFEST_BLOB_UNKNOWN"
	// MANIFEST_INVALID
	ErrCodeManifestInvalid = "MANIFEST_INVALID"
	// MANIFEST_UNKNOWN
	ErrCodeManifestNotFound = "MANIFEST_UNKNOWN"
	// NAME_INVALID
	ErrCodeNameInvalid = "NAME_INVALID"
	// NAME_UNKNOWN
	ErrCodeNameUnknown = "NAME_UNKNOWN"
	// SIZE_INVALID
	ErrCodeSizeInvalid = "SIZE_INVALID"
	// UNAUTHORIZED
	ErrCodeUnauthorized = "UNAUTHORIZED"
	// DENIED
	ErrCodeDenied = "DENIED"
	// UNSUPPORTED
	ErrCodeUnsupported = "UNSUPPORTED"
	// TOO_MANY_REQUESTS
	ErrCodeTooManyRequests = "TOO_MANY_REQUESTS"
)

var (
	ErrBlobNotFound        = OCIError{Code: ErrCodeBlobNotFound, Message: "blob unknown to registry", HTTPStatus: http.StatusNotFound}
	ErrBlobUploadInvalid   = OCIError{Code: ErrCodeBlobUploadInvalid, Message: "blob upload invalid", HTTPStatus: http.StatusBadRequest}
	ErrBlobUploadUnknown   = OCIError{Code: ErrCodeBlobUploadUnknown, Message: "blob upload unknown to registry", HTTPStatus: http.StatusNotFound}
	ErrDigestInvalid       = OCIError{Code: ErrCodeDigestInvalid, Message: "provided digest did not match uploaded content", HTTPStatus: http.StatusBadRequest}
	ErrManifestBlobUnknown = OCIError{Code: ErrCodeManifestBlobUnknown, Message: "manifest references a manifest or blob unknown to registry", HTTPStatus: http.StatusBadRequest}
	ErrManifestInvalid     = OCIError{Code: ErrCodeManifestInvalid, Message: "manifest invalid", HTTPStatus: http.StatusBadRequest}
	ErrManifestNotFound    = OCIError{Code: ErrCodeManifestNotFound, Message: "manifest unknown to registry", HTTPStatus: http.StatusNotFound}
	ErrNameInvalid         = OCIError{Code: ErrCodeNameInvalid, Message: "invalid repository name", HTTPStatus: http.StatusBadRequest}
	ErrNameUnknown         = OCIError{Code: ErrCodeNameUnknown, Message: "repository name not known to registry", HTTPStatus: http.StatusNotFound}
	ErrSizeInvalid         = OCIError{Code: ErrCodeSizeInvalid, Message: "provided length did not match content length", HTTPStatus: http.StatusBadRequest}
	ErrUnauthorized        = OCIError{Code: ErrCodeUnauthorized, Message: "authentication required", HTTPStatus: http.StatusUnauthorized}
	ErrDenied              = OCIError{Code: ErrCodeDenied, Message: "denied", HTTPStatus: http.StatusForbidden}
	ErrUnsupported         = OCIError{Code: ErrCodeUnsupported, Message: "unsupported", HTTPStatus: http.StatusNotImplemented}
	ErrTooManyRequests     = OCIError{Code: ErrCodeTooManyRequests, Message: "too many requests", HTTPStatus: http.StatusTooManyRequests}
)

type OCIError struct {
	Code       string      `json:"code"`
	Message    string      `json:"message"`
	Detail     interface{} `json:"detail,omitempty"`
	HTTPStatus int         `json:"-"`
}

func (e OCIError) Error() string {
	return e.Message
}

func (e OCIError) JSON() string {
	return fmt.Sprintf(`{"errors":[{"code":"%s","message":"%s","detail":"%v"}]}`, e.Code, e.Message, e.Detail)
}

func WriteError(w http.ResponseWriter, err OCIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.HTTPStatus)
	w.Write([]byte(err.JSON()))
}

func ToOCI(err error) OCIError {
	switch {
	case errors.Is(err, ErrBlobNotFound):
		e := ErrBlobNotFound
		e.Detail = err.Error()
		return e

	case errors.Is(err, ErrUnauthorized):
		e := ErrUnauthorized
		e.Detail = err.Error()
		return e

	// Example: digest mismatch from CAS layer
	case errors.Is(err, ErrDigestInvalid):
		e := ErrDigestInvalid
		e.Detail = err.Error()
		return e
	case errors.Is(err, ErrManifestBlobUnknown):
		e := ErrManifestBlobUnknown
		e.Detail = err.Error()
		return e
	case errors.Is(err, ErrManifestInvalid):
		e := ErrManifestInvalid
		e.Detail = err.Error()
		return e
	case errors.Is(err, ErrManifestNotFound):
		e := ErrManifestNotFound
		e.Detail = err.Error()
		return e
	case errors.Is(err, ErrNameInvalid):
		e := ErrNameInvalid
		e.Detail = err.Error()
		return e
	case errors.Is(err, ErrNameUnknown):
		e := ErrNameUnknown
		e.Detail = err.Error()
		return e
	default:
		// Fallback: preserve detail for operator logs
		e := ErrUnsupported
		e.Detail = err.Error()
		return e
	}
}
