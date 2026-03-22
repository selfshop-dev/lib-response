package response

import (
	"encoding/json"
	"net/http"

	apperr "github.com/selfshop-dev/lib-apperr"
	validation "github.com/selfshop-dev/lib-validation"
)

// MetaExtractor builds the meta payload from the incoming request.
// Called once per response — must be safe for concurrent use.
// Return nil to omit the meta field from the response entirely.
type MetaExtractor func(r *http.Request) map[string]any

// Writer serialises responses using the RFC-9457 envelope.
// Construct once at startup via [NewWriter]; safe for concurrent use.
type Writer struct {
	extractMeta MetaExtractor
}

// NewWriter creates a [Writer] with the given meta extractor.
// The extractor is called on every response to populate the meta field.
//
//	var respond = response.NewWriter(func(r *http.Request) map[string]any {
//	    id := httpx.RequestIDFromContext(r.Context())
//	    if id == "" {
//	        return nil
//	    }
//	    return map[string]any{"request_id": id}
//	})
func NewWriter(extractor MetaExtractor) *Writer {
	return &Writer{extractMeta: extractor}
}

// OK writes a 200 OK response with data as application/json.
func (w *Writer) OK(rw http.ResponseWriter, r *http.Request, data any) {
	w.writeSuccess(rw, r, http.StatusOK, data)
}

// Created writes a 201 Created response with data as application/json.
func (w *Writer) Created(rw http.ResponseWriter, r *http.Request, data any) {
	w.writeSuccess(rw, r, http.StatusCreated, data)
}

// Accepted writes a 202 Accepted response with data as application/json.
func (w *Writer) Accepted(rw http.ResponseWriter, r *http.Request, data any) {
	w.writeSuccess(rw, r, http.StatusAccepted, data)
}

// NoContent writes a 204 No Content response with no body.
func (w *Writer) NoContent(rw http.ResponseWriter, _ *http.Request) {
	rw.WriteHeader(http.StatusNoContent)
}

// Write writes p as an RFC-9457 problem response as application/problem+json.
// Use for sending sentinel errors directly:
//
//	respond.Write(w, r, response.ErrNotFound.WithDetail("order not found"))
func (w *Writer) Write(rw http.ResponseWriter, r *http.Request, p *Problem) {
	w.writeProblem(rw, r, p)
}

// Error maps err to the appropriate RFC-9457 problem response.
// Inspects the error chain in priority order:
//
//  1. [*apperr.Error] → status from Kind; KindInternal detail suppressed.
//     If apperr carries a [*validation.Error] it is included as extensions.fields.
//  2. [*validation.Error] → 422 with extensions.fields.
//  3. anything else → 500 with no detail.
func (w *Writer) Error(rw http.ResponseWriter, r *http.Request, err error) {
	if ae, ok := apperr.As(err); ok {
		w.writeProblem(rw, r, problemFromAppErr(ae))
		return
	}
	if ve, ok := validation.As(err); ok {
		w.writeProblem(rw, r, problemFromValidation(ve))
		return
	}
	w.writeProblem(rw, r, ErrInternalServerError)
}

// BadRequest writes a 400 Bad Request problem with the given detail.
// Use for handler-level input errors before any service call.
func (w *Writer) BadRequest(rw http.ResponseWriter, r *http.Request, detail string) {
	w.writeProblem(rw, r, ErrBadRequest.WithDetail(detail))
}

// Unauthorized writes a 401 Unauthorized problem with the given detail.
func (w *Writer) Unauthorized(rw http.ResponseWriter, r *http.Request, detail string) {
	w.writeProblem(rw, r, ErrUnauthorized.WithDetail(detail))
}

// Forbidden writes a 403 Forbidden problem with the given detail.
func (w *Writer) Forbidden(rw http.ResponseWriter, r *http.Request, detail string) {
	w.writeProblem(rw, r, ErrForbidden.WithDetail(detail))
}

// NotFound writes a 404 Not Found problem with the given detail.
func (w *Writer) NotFound(rw http.ResponseWriter, r *http.Request, detail string) {
	w.writeProblem(rw, r, ErrNotFound.WithDetail(detail))
}

// Conflict writes a 409 Conflict problem with the given detail.
func (w *Writer) Conflict(rw http.ResponseWriter, r *http.Request, detail string) {
	w.writeProblem(rw, r, ErrConflict.WithDetail(detail))
}

// InternalServerError writes a 500 Internal Server Error problem with no detail.
// Detail is intentionally omitted to prevent leaking internal state to clients.
func (w *Writer) InternalServerError(rw http.ResponseWriter, r *http.Request) {
	w.writeProblem(rw, r, ErrInternalServerError)
}

func (w *Writer) writeSuccess(rw http.ResponseWriter, r *http.Request, status int, data any) {
	env := newEnvelope(status)
	env.Data = data
	env.Instance = r.URL.Path
	env.Meta = w.extractMeta(r)
	writeJSON(rw, status, env, MediaTypeJSON)
}

func (w *Writer) writeProblem(rw http.ResponseWriter, r *http.Request, p *Problem) {
	env := newEnvelope(p.status)
	env.Detail = p.detail
	env.Extensions = p.extensions
	env.Instance = r.URL.Path
	env.Meta = w.extractMeta(r)
	writeJSON(rw, p.status, env, MediaTypeProblem)
}

// writeJSON serialises v and writes it to rw.
// On marshal failure it falls back to a plain 500 text response —
// headers have not been written yet at that point.
func writeJSON(rw http.ResponseWriter, status int, v any, contentType string) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(rw, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", contentType)
	rw.WriteHeader(status)
	_, _ = rw.Write(b) //nolint:errcheck // write error after WriteHeader is unrecoverable
}
