package response

import "net/http"

const (
	// MediaTypeJSON is the Content-Type for successful responses.
	MediaTypeJSON = "application/json"

	// MediaTypeProblem is the RFC-9457 Content-Type for error responses.
	MediaTypeProblem = "application/problem+json"

	typeBlank = "about:blank"
)

// envelope is the unified JSON shape for every response — success and error alike.
// Implements the RFC-9457 problem detail format extended with data and meta fields.
//
// Field ordering is chosen for optimal struct alignment: pointer/interface fields
// first, then map, then strings, then int.
type envelope struct {
	Data       any            `json:"data,omitempty"`       // success payload; omitted on errors
	Extensions any            `json:"extensions,omitempty"` // validation field errors; omitted on success
	Meta       map[string]any `json:"meta,omitempty"`       // per-request metadata (e.g. request_id); omitted if extractor returns nil

	// Type is the RFC-9457 problem type URI. Always "about:blank" —
	// signals that the status code is the primary classification.
	Type string `json:"type"`

	Title    string `json:"title"`              // HTTP status text, e.g. "Not Found"
	Detail   string `json:"detail,omitempty"`   // human-readable error detail; omitted on success and KindInternal
	Instance string `json:"instance,omitempty"` // request path, e.g. "/users/42"
	Status   int    `json:"status"`             // HTTP status code, e.g. 404
}

func newEnvelope(status int) envelope {
	return envelope{
		Type:   typeBlank,
		Status: status,
		Title:  http.StatusText(status),
	}
}
