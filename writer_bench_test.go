package response_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	response "github.com/selfshop-dev/lib-response"
)

func BenchmarkWriter_Ok(b *testing.B) {
	meta := map[string]any{"request_id": "req-abc"}
	w := response.NewWriter(func(_ *http.Request) map[string]any { return meta })
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/users/1", nil)
	d := map[string]any{"id": "uuid-1", "name": "Alice"}
	rw := httptest.NewRecorder()
	for b.Loop() {
		rw.Body.Reset()
		w.Ok(rw, r, d)
	}
}
