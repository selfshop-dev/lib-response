// Package response provides RFC-9457 problem detail serialisation for HTTP handlers.
//
// It is the HTTP wire layer for the selfshop-dev service stack. It knows about
// [apperr] and [validation] and is the single place that maps domain errors to
// HTTP responses — all other packages stay transport-agnostic.
//
// # Envelope
//
// Every response — success and error alike — uses the same JSON envelope:
//
//	// 201 Created
//	{
//	    "type":     "about:blank",
//	    "status":   201,
//	    "title":    "Created",
//	    "data":     { "id": "uuid" },
//	    "instance": "/orders",
//	    "meta":     { "request_id": "abc123" }
//	}
//
//	// 422 Unprocessable Entity
//	{
//	    "type":       "about:blank",
//	    "status":     422,
//	    "title":      "Unprocessable Entity",
//	    "detail":     "invalid order",
//	    "instance":   "/orders",
//	    "meta":       { "request_id": "abc123" },
//	    "extensions": { "fields": [ ... ] }
//	}
//
// Content-Type is application/json for 2xx and application/problem+json for
// 4xx/5xx. The data field is omitted on errors; detail and extensions are
// omitted on success.
//
// # Writer
//
// [Writer] is the main entry point. Construct one at startup with a
// [MetaExtractor] that builds your service's meta map from a request, then
// use it in every handler:
//
//	var respond = response.NewWriter(func(r *http.Request) map[string]any {
//	    return map[string]any{
//	        "request_id": httpx.RequestIDFromContext(r.Context()),
//	    }
//	})
//
//	func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
//	    order, err := h.svc.Create(r.Context(), cmd)
//	    if err != nil {
//	        respond.Error(w, r, err)
//	        return
//	    }
//	    respond.Created(w, r, order)
//	}
//
// # Error mapping
//
// [Writer.Error] inspects the error chain in priority order:
//
//  1. *[apperr.Error] → status from Kind; KindInternal detail suppressed.
//     If apperr carries a *[validation.Error] it is included as extensions.fields.
//  2. *[validation.Error] → 422 with extensions.fields.
//  3. anything else → 500 with no detail.
//
// # Sentinels
//
// Package-level sentinel problems are safe to share across requests — [Problem.WithDetail]
// returns a copy and never mutates the receiver:
//
//	respond.Write(w, r, response.ErrNotFound)
//	respond.Write(w, r, response.ErrNotFound.WithDetail("order not found"))
//
// # Meta
//
// The [MetaExtractor] is called on every response. Return nil to omit the meta
// field entirely:
//
//	var respond = response.NewWriter(func(r *http.Request) map[string]any {
//	    id := httpx.RequestIDFromContext(r.Context())
//	    if id == "" {
//	        return nil
//	    }
//	    return map[string]any{"request_id": id}
//	})
//
// # Concurrency
//
// [Writer] is safe for concurrent use. The [MetaExtractor] is called once per
// response and must itself be safe for concurrent use.
package response
