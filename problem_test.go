package response_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	apperr "github.com/selfshop-dev/lib-apperr"
	validation "github.com/selfshop-dev/lib-validation"

	response "github.com/selfshop-dev/lib-response"
)

// newRecorder returns a fresh ResponseRecorder for each test.
func newRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// newRequest returns a new incoming request for the given method and path.
func newRequest(method, path string) *http.Request {
	return httptest.NewRequestWithContext(context.Background(), method, path, nil)
}

// noMeta is a MetaExtractor that returns nil — omits the meta field entirely.
func noMeta(_ *http.Request) map[string]any { return nil }

// envelope is the base JSON shape shared by all responses.
type envelope struct {
	Data       any            `json:"data"`
	Meta       map[string]any `json:"meta"`
	Extensions any            `json:"extensions"`
	Type       string         `json:"type"`
	Title      string         `json:"title"`
	Detail     string         `json:"detail"`
	Instance   string         `json:"instance"`
	Status     int            `json:"status"`
}

// envelopeWithExtensions is used when asserting on validation field errors.
type envelopeWithExtensions struct {
	Extensions *fieldsExtBody `json:"extensions"`
	Meta       map[string]any `json:"meta"`
	Type       string         `json:"type"`
	Title      string         `json:"title"`
	Detail     string         `json:"detail"`
	Instance   string         `json:"instance"`
	Status     int            `json:"status"`
}

type fieldsExtBody struct {
	Fields []fieldBody `json:"fields"`
}

type fieldBody struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// makeAppErr returns a minimal [*apperr.Error] with the given Kind.
func makeAppErr(k apperr.Kind) error {
	return apperr.Wrap(errors.New(""), k, "")
}

// makeAppErrWithMessage returns a [*apperr.Error] with Kind and Message set.
func makeAppErrWithMessage(k apperr.Kind, msg string) error {
	return apperr.Wrap(errors.New(msg), k, msg)
}

// makeAppErrWithValidation returns a KindConflict error with a [validation.Error] attached.
func makeAppErrWithValidation() error {
	return apperr.ConflictField("email", "already registered")
}

// makeValidationErr returns a [*validation.Error] with two field errors.
func makeValidationErr() error {
	c := validation.NewCollector("invalid user")
	return c.Add(
		validation.Required("name"),
		validation.Invalid("email", "must be a valid address"),
	).Err()
}

func TestProblem_WithDetail(t *testing.T) {
	t.Parallel()

	t.Run("returns copy with detail set", func(t *testing.T) {
		t.Parallel()
		p := response.ErrNotFound.WithDetail("order not found")
		assert.Equal(t, "order not found", p.Detail())
	})

	t.Run("does not mutate original sentinel", func(t *testing.T) {
		t.Parallel()
		_ = response.ErrNotFound.WithDetail("mutated")
		// ErrNotFound должен остаться без detail
		assert.Empty(t, response.ErrNotFound.Detail())
	})

	t.Run("chaining produces independent copies", func(t *testing.T) {
		t.Parallel()
		p1 := response.ErrNotFound.WithDetail("first")
		p2 := response.ErrNotFound.WithDetail("second")
		assert.Equal(t, "first", p1.Detail())
		assert.Equal(t, "second", p2.Detail())
	})
}

func TestKindToStatus(t *testing.T) {
	t.Parallel()

	tests := [...]struct {
		name       string
		wantStatus int
		err        error
	}{
		{"not found", http.StatusNotFound, makeAppErr(apperr.KindNotFound)},
		{"unauthorized", http.StatusUnauthorized, makeAppErr(apperr.KindUnauthorized)},
		{"forbidden", http.StatusForbidden, makeAppErr(apperr.KindForbidden)},
		{"conflict", http.StatusConflict, makeAppErr(apperr.KindConflict)},
		{"unprocessable", http.StatusUnprocessableEntity, makeAppErr(apperr.KindUnprocessable)},
		{"unavailable", http.StatusServiceUnavailable, makeAppErr(apperr.KindUnavailable)},
		{"timeout", http.StatusGatewayTimeout, makeAppErr(apperr.KindTimeout)},
		{"internal", http.StatusInternalServerError, makeAppErr(apperr.KindInternal)},
		{"unknown", http.StatusInternalServerError, makeAppErr(apperr.KindUnknown)},
		{"default fallback", http.StatusInternalServerError, &apperr.Error{Kind: apperr.Kind(255)}},
	}

	w := response.NewWriter(noMeta)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rw, r := newRecorder(), newRequest(http.MethodGet, "/test")
			w.Error(rw, r, tc.err)
			assert.Equal(t, tc.wantStatus, rw.Code)
		})
	}
}

func TestSentinels_StatusCodes(t *testing.T) {
	t.Parallel()

	// Sentinel status codes are part of the public API contract.
	tests := [...]struct {
		name       string
		sentinel   *response.Problem
		wantStatus int
	}{
		{"ErrBadRequest", response.ErrBadRequest, http.StatusBadRequest},
		{"ErrUnauthorized", response.ErrUnauthorized, http.StatusUnauthorized},
		{"ErrForbidden", response.ErrForbidden, http.StatusForbidden},
		{"ErrNotFound", response.ErrNotFound, http.StatusNotFound},
		{"ErrMethodNotAllowed", response.ErrMethodNotAllowed, http.StatusMethodNotAllowed},
		{"ErrConflict", response.ErrConflict, http.StatusConflict},
		{"ErrUnprocessable", response.ErrUnprocessable, http.StatusUnprocessableEntity},
		{"ErrTooManyRequests", response.ErrTooManyRequests, http.StatusTooManyRequests},
		{"ErrInternalServerError", response.ErrInternalServerError, http.StatusInternalServerError},
		{"ErrNotImplemented", response.ErrNotImplemented, http.StatusNotImplemented},
		{"ErrServiceUnavailable", response.ErrServiceUnavailable, http.StatusServiceUnavailable},
	}

	w := response.NewWriter(noMeta)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rw, r := newRecorder(), newRequest(http.MethodGet, "/")
			w.Write(rw, r, tc.sentinel)
			assert.Equal(t, tc.wantStatus, rw.Code)
		})
	}
}
