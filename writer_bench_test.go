package response_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	apperr "github.com/selfshop-dev/lib-apperr"
	response "github.com/selfshop-dev/lib-response"
	validation "github.com/selfshop-dev/lib-validation"
)

// newBenchRequest returns a pre-built request for use in benchmarks.
// Constructed once per benchmark to avoid measuring http.NewRequest overhead.
func newBenchRequest() *http.Request {
	return httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/users/42", nil)
}

// BenchmarkWriter_OK measures the full success response path:
// meta extraction + envelope build + JSON marshal + write.
func BenchmarkWriter_OK(b *testing.B) {
	w := response.NewWriter(noMeta)
	r := newBenchRequest()
	data := map[string]any{"id": "uuid-1", "name": "Alice"}
	b.ReportAllocs()
	for b.Loop() {
		rw := httptest.NewRecorder()
		w.OK(rw, r, data)
		runtime.KeepAlive(rw)
	}
}

// BenchmarkWriter_OK_WithMeta measures the overhead of meta extraction and
// map allocation on every response — the production case with request_id.
func BenchmarkWriter_OK_WithMeta(b *testing.B) {
	w := response.NewWriter(func(_ *http.Request) map[string]any {
		return map[string]any{"request_id": "req-abc123"}
	})
	r := newBenchRequest()
	data := map[string]any{"id": "uuid-1"}
	b.ReportAllocs()
	for b.Loop() {
		rw := httptest.NewRecorder()
		w.OK(rw, r, data)
		runtime.KeepAlive(rw)
	}
}

// BenchmarkWriter_NoContent measures the minimal response path — no body,
// no JSON marshal, only WriteHeader.
func BenchmarkWriter_NoContent(b *testing.B) {
	w := response.NewWriter(noMeta)
	r := newBenchRequest()
	b.ReportAllocs()
	for b.Loop() {
		rw := httptest.NewRecorder()
		w.NoContent(rw, r)
		runtime.KeepAlive(rw)
	}
}

// BenchmarkWriter_Error_AppErr_NotFound measures the common error path:
// apperr.As + problem build + JSON marshal.
func BenchmarkWriter_Error_AppErr_NotFound(b *testing.B) {
	w := response.NewWriter(noMeta)
	r := newBenchRequest()
	err := apperr.NotFound("user", int64(42))
	b.ReportAllocs()
	for b.Loop() {
		rw := httptest.NewRecorder()
		w.Error(rw, r, err)
		runtime.KeepAlive(rw)
	}
}

// BenchmarkWriter_Error_AppErr_Internal measures the internal error path —
// detail suppressed, minimal envelope.
func BenchmarkWriter_Error_AppErr_Internal(b *testing.B) {
	w := response.NewWriter(noMeta)
	r := newBenchRequest()
	err := apperr.Internal("unexpected failure")
	b.ReportAllocs()
	for b.Loop() {
		rw := httptest.NewRecorder()
		w.Error(rw, r, err)
		runtime.KeepAlive(rw)
	}
}

// BenchmarkWriter_Error_ValidationErr measures the 422 path with field errors
// — the heaviest error response due to extensions serialisation.
func BenchmarkWriter_Error_ValidationErr(b *testing.B) {
	w := response.NewWriter(noMeta)
	r := newBenchRequest()

	c := validation.NewCollector("invalid user")
	c.Add(
		validation.Required("name"),
		validation.Invalid("email", "must be a valid address"),
		validation.TooLong("bio", 500),
	)
	err := c.Err()

	b.ReportAllocs()
	for b.Loop() {
		rw := httptest.NewRecorder()
		w.Error(rw, r, err)
		runtime.KeepAlive(rw)
	}
}

// BenchmarkWriter_Error_AppErr_WithValidation measures a conflict error that
// carries a validation.Error — exercises both apperr.As and fields extension.
func BenchmarkWriter_Error_AppErr_WithValidation(b *testing.B) {
	w := response.NewWriter(noMeta)
	r := newBenchRequest()
	err := apperr.ConflictField("email", "already registered")
	b.ReportAllocs()
	for b.Loop() {
		rw := httptest.NewRecorder()
		w.Error(rw, r, err)
		runtime.KeepAlive(rw)
	}
}

// BenchmarkWriter_Write_Sentinel measures sending a pre-built sentinel problem
// — the fastest error path, no error chain traversal.
func BenchmarkWriter_Write_Sentinel(b *testing.B) {
	w := response.NewWriter(noMeta)
	r := newBenchRequest()
	p := response.ErrNotFound.WithDetail("user not found")
	b.ReportAllocs()
	for b.Loop() {
		rw := httptest.NewRecorder()
		w.Write(rw, r, p)
		runtime.KeepAlive(rw)
	}
}
