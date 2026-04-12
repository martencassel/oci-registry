package oci

import (
	"errors"
	"strings"
)

var (
	ErrInvalidPath = errors.New("invalid path")
)

var verbSet = map[string]struct{}{
	"blobs":     {},
	"manifests": {},
	"tags":      {},
	"referrers": {},
	"uploads":   {},
}

type Token struct {
	Value     string
	IsVerb    bool
	IsSpecial bool
}

func lex(path string) ([]Token, error) {
	// 1. Must start with /v2/
	if !strings.HasPrefix(path, "/v2/") {
		return nil, ErrInvalidPath
	}

	// Could be /v2/ ping → return empty tokens
	if path == "/v2/" {
		return []Token{}, nil
	}

	// 2. Strip prefix
	rest := strings.TrimPrefix(path, "/v2/")

	// 3. Split off query string
	if i := strings.Index(rest, "?"); i != -1 {
		rest = rest[:i]
	}

	// 4. Split into segments
	segments := strings.Split(rest, "/")

	// 5. Validate empty segments (only allow trailing slash)
	for i, seg := range segments {
		if seg == "" {
			if i == len(segments)-1 {
				continue
			}
			return nil, ErrInvalidPath
		}
	}

	// 6. Classify tokens
	tokens := make([]Token, 0, len(segments))
	for _, seg := range segments {
		if seg == "" {
			// trailing slash → ignore
			continue
		}

		_, isVerb := verbSet[seg]
		isSpecial := strings.HasPrefix(seg, "_")

		tokens = append(tokens, Token{
			Value:     seg,
			IsVerb:    isVerb,
			IsSpecial: isSpecial,
		})
	}

	if len(tokens) == 0 {
		return nil, ErrInvalidPath
	}

	return tokens, nil
}
