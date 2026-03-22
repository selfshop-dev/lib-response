package response

import "net/http"

const (
	// MediaTypeJSON is the Content-Type for successful responses.
	MediaTypeJSON = "application/json"

	// MediaTypeProblem is the RFC-9457 Content-Type for error responses.
	MediaTypeProblem = "application/problem+json"

	typeBlank = "about:blank"
)

// envelope is the unified JSON shape for every response — success and error.
//
// Fields are grouped to minimise struct padding:
// pointer/slice/map fields first, then string fields, then int, then bool.
type envelope struct {
	Data       any            `json:"data,omitempty"`
	Extensions any            `json:"extensions,omitempty"`
	Meta       map[string]any `json:"meta,omitempty"`
	Type       string         `json:"type"`
	Title      string         `json:"title"`
	Detail     string         `json:"detail,omitempty"`
	Instance   string         `json:"instance,omitempty"`
	Status     int            `json:"status"`
}

func newEnvelope(status int) envelope {
	return envelope{
		Type:   typeBlank,
		Status: status,
		Title:  http.StatusText(status),
	}
}
