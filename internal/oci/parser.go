package oci

/*
Request := RepoKey RepoPath? Special? Verb Tail?
RepoPath := Segment*
Special := "_" Segment
Verb := "blobs" | "manifests" | "tags" | "referrers" | "uploads"
Tail := depends on Verb
*/

type state int

const (
	stRepoKey state = iota
	stRepoPath
	stSpecial
	stVerb
	stTail
)

var verbHandlers = map[string]func(*ParseResult, []Token) error{
	"blobs":     parseBlobs,
	"manifests": parseManifests,
	"tags":      parseTags,
	"referrers": parseReferrers,
	"uploads":   parseUploads,
}

func parseUploads(meta *ParseResult, tail []Token) error {
	meta.IsUpload = true
	meta.SubVerb = "uploads"
	if len(tail) == 0 {
		return nil
	}
	if len(tail) == 1 {
		meta.UploadUUID = tail[0].Value
		return nil
	}
	return ErrInvalidPath
}

func parseBlobs(meta *ParseResult, tail []Token) error {
	if len(tail) == 0 {
		return nil
	}
	meta.Digest = tail[0].Value
	return nil
}

func parseManifests(meta *ParseResult, tail []Token) error {
	if len(tail) == 0 {
		return nil
	}
	meta.Reference = tail[0].Value
	return nil
}

func parseTags(meta *ParseResult, tail []Token) error {
	if len(tail) == 0 {
		return nil
	}
	if tail[0].Value == "list" {
		meta.SubVerb = "list"
	}
	return nil
}

func parseReferrers(meta *ParseResult, tail []Token) error {
	if len(tail) == 0 {
		return nil
	}
	meta.Reference = tail[0].Value
	return nil
}

var verbMap = map[string]VerbType{
	"blobs":     VerbBlobs,
	"manifests": VerbManifests,
	"tags":      VerbTags,
	"referrers": VerbReferrers,
	"uploads":   VerbBlobs,
}

func Parser(method, rawPath string) (ParseResult, error) {
	tokens, err := lex(rawPath)
	if err != nil {
		return ParseResult{}, err
	}

	p := ParseResult{RawPath: rawPath}

	sm := stRepoKey
	i := 0

	for i < len(tokens) {
		t := tokens[i]

		switch sm {
		case stRepoKey:
			p.RepoKey = t.Value
			sm = stRepoPath
		case stRepoPath:
			if t.IsVerb {
				sm = stVerb
				continue
			}
			if t.IsSpecial {
				p.SubVerb = t.Value
				sm = stSpecial
				continue
			}
			p.Repository = appendPath(p.Repository, t.Value)
		case stSpecial:
			sm = stVerb
		case stVerb:
			p.Verb = verbMap[t.Value]
			sm = stTail
		case stTail:
			handler := verbHandlers[t.Value]
			if handler == nil {
				return ParseResult{}, ErrInvalidPath
			}
			if err := handler(&p, tokens[i+1:]); err != nil {
				return ParseResult{}, err
			}
			return p, nil
		}
		i++
	}

	return p, nil
}

func appendPath(base, segment string) string {
	if base == "" {
		return segment
	}
	return base + "/" + segment
}
