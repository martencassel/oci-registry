package oci

import (
	"errors"
)

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
