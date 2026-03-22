package response_test

import (
	"net/http"
	"runtime"
	"testing"

	apperr "github.com/selfshop-dev/lib-apperr"
	response "github.com/selfshop-dev/lib-response"
)

// BenchmarkProblem_WithDetail measures the copy-on-write sentinel decoration —
// called once per request when adding per-request detail to a sentinel.
func BenchmarkProblem_WithDetail(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		p := response.ErrNotFound.WithDetail("user 42 not found")
		runtime.KeepAlive(p)
	}
}

// BenchmarkKindToStatus measures kindToStatus via Writer.Error —
// the switch over apperr.Kind that runs on every error response.
func BenchmarkKindToStatus(b *testing.B) {
	w := response.NewWriter(func(_ *http.Request) map[string]any { return nil })
	r := newBenchRequest()
	err := makeAppErr(apperr.KindNotFound)
	b.ReportAllocs()
	for b.Loop() {
		rw := newRecorder()
		w.Error(rw, r, err)
		runtime.KeepAlive(rw)
	}
}
