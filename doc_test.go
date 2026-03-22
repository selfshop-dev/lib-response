package response_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	apperr "github.com/selfshop-dev/lib-apperr"
	validation "github.com/selfshop-dev/lib-validation"

	response "github.com/selfshop-dev/lib-response"
)

// testWriter is a Writer with no meta for use in examples.
var testWriter = response.NewWriter(func(_ *http.Request) map[string]any { return nil })

// ExampleWriter_OK demonstrates a successful 200 response.
func ExampleWriter_OK() {
	rw := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/users/1", nil)

	testWriter.OK(rw, r, map[string]any{"id": "uuid-1", "name": "Alice"})

	fmt.Println(rw.Code)
	fmt.Println(rw.Header().Get("Content-Type"))

	// Output:
	// 200
	// application/json
}

// ExampleWriter_Created demonstrates a 201 Created response.
func ExampleWriter_Created() {
	rw := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/users", nil)

	testWriter.Created(rw, r, map[string]any{"id": "uuid-new"})

	fmt.Println(rw.Code)
	fmt.Println(rw.Header().Get("Content-Type"))

	// Output:
	// 201
	// application/json
}

// ExampleWriter_NoContent demonstrates a 204 No Content response.
func ExampleWriter_NoContent() {
	rw := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), http.MethodDelete, "/users/1", nil)

	testWriter.NoContent(rw, r)

	fmt.Println(rw.Code)
	fmt.Println(rw.Body.Len())

	// Output:
	// 204
	// 0
}

// ExampleWriter_Error_appErr demonstrates mapping a *apperr.Error to an HTTP response.
func ExampleWriter_Error_appErr() {
	rw := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/users/42", nil)

	err := apperr.NotFound("user", 42)
	testWriter.Error(rw, r, err)

	fmt.Println(rw.Code)
	fmt.Println(rw.Header().Get("Content-Type"))

	// Output:
	// 404
	// application/problem+json
}

// ExampleWriter_Error_validationErr demonstrates mapping a *validation.Error to a 422 response.
func ExampleWriter_Error_validationErr() {
	rw := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/users", nil)

	c := validation.NewCollector("invalid user")
	c.Add(validation.Required("name"))
	c.Add(validation.Invalid("email", "must be a valid address"))

	testWriter.Error(rw, r, c.Err())

	fmt.Println(rw.Code)
	fmt.Println(rw.Header().Get("Content-Type"))

	// Output:
	// 422
	// application/problem+json
}

// ExampleWriter_Error_internalErr demonstrates that KindInternal never leaks detail.
func ExampleWriter_Error_internalErr() {
	rw := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)

	err := apperr.Internal("secret db password: hunter2")
	testWriter.Error(rw, r, err)

	fmt.Println(rw.Code)
	// Detail is suppressed — 500 responses never expose internal messages.

	// Output:
	// 500
}

// ExampleWriter_Write demonstrates sending a sentinel problem directly.
func ExampleWriter_Write() {
	rw := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/orders/99", nil)

	testWriter.Write(rw, r, response.ErrNotFound.WithDetail("order not found"))

	fmt.Println(rw.Code)

	// Output:
	// 404
}

// ExampleProblem_WithDetail demonstrates the copy-on-write sentinel pattern.
func ExampleProblem_WithDetail() {
	p := response.ErrNotFound.WithDetail("user not found")

	fmt.Println(p.Status())
	fmt.Println(p.Detail())
	// Original sentinel is unchanged:
	fmt.Println(response.ErrNotFound.Detail())

	// Output:
	// 404
	// user not found
	//
}

// ExampleNewWriter demonstrates constructing a Writer with a meta extractor.
func ExampleNewWriter() {
	w := response.NewWriter(func(_ *http.Request) map[string]any {
		// In production: httpx.RequestIDFromContext(r.Context())
		return map[string]any{"request_id": "req-abc123"}
	})

	rw := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/users", nil)

	w.OK(rw, r, nil)

	fmt.Println(rw.Code)

	// Output:
	// 200
}
