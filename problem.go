package response

import (
	"net/http"

	apperr "github.com/selfshop-dev/lib-apperr"
	validation "github.com/selfshop-dev/lib-validation"
)

// Problem is an RFC-9457 problem detail.
//
// Immutable by convention: [Problem.WithDetail] returns a shallow copy so
// package-level sentinels can be safely decorated per request without races.
type Problem struct {
	extensions any
	detail     string
	status     int
}

func newProblem(status int) *Problem {
	return &Problem{status: status}
}

// Detail returns the detail message of the problem.
func (p *Problem) Detail() string { return p.detail }

// Status returns the HTTP status code of the problem.
func (p *Problem) Status() int { return p.status }

// WithDetail returns a copy of p with Detail set.
// Safe to call on package-level sentinels — never mutates the receiver.
func (p *Problem) WithDetail(detail string) *Problem {
	cp := *p
	cp.detail = detail
	return &cp
}

// withExtensions returns a copy with Extensions set.
func (p *Problem) withExtensions(ext any) *Problem {
	cp := *p
	cp.extensions = ext
	return &cp
}

// fieldsExt is the RFC-9457 extension payload for validation errors.
// Nested under "extensions" so the per-field list never collides with
// standard RFC-9457 members (type, title, status, detail, instance).
type fieldsExt struct {
	Fields []validation.FieldError `json:"fields"`
}

// problemFromAppErr maps a *apperr.Error to a Problem.
// KindInternal and KindUnknown never expose their message to the client.
func problemFromAppErr(ae *apperr.Error) *Problem {
	status := kindToStatus(ae.Kind)

	if ae.Kind == apperr.KindInternal || ae.Kind == apperr.KindUnknown {
		return newProblem(status)
	}

	p := newProblem(status).WithDetail(ae.Message)

	if ve := ae.Validation(); ve != nil && ve.HasErrors() {
		return p.withExtensions(fieldsExt{Fields: ve.Fields})
	}
	return p
}

// problemFromValidation maps a *validation.Error to a 422 Problem.
func problemFromValidation(ve *validation.Error) *Problem {
	return newProblem(http.StatusUnprocessableEntity).
		WithDetail(ve.Summary).
		withExtensions(fieldsExt{Fields: ve.Fields})
}

// kindToStatus maps [apperr.Kind] to an HTTP status code.
// Lives here — not in apperr — so apperr stays transport-agnostic.
// The default branch prevents a new Kind from silently producing 200.
func kindToStatus(k apperr.Kind) int {
	switch k {
	case apperr.KindNotFound:
		return http.StatusNotFound
	case apperr.KindUnauthorized:
		return http.StatusUnauthorized
	case apperr.KindForbidden:
		return http.StatusForbidden
	case apperr.KindConflict:
		return http.StatusConflict
	case apperr.KindUnprocessable:
		return http.StatusUnprocessableEntity
	case apperr.KindUnavailable:
		return http.StatusServiceUnavailable
	case apperr.KindTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}
