package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperr "github.com/selfshop-dev/lib-apperr"

	response "github.com/selfshop-dev/lib-response"
)

func TestWriter_Ok(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodGet, "/users/1")

	w.Ok(rw, r, map[string]any{"id": "uuid-1"})

	assert.Equal(t, http.StatusOK, rw.Code)
	assert.Equal(t, response.MediaTypeJSON, rw.Header().Get("Content-Type"))

	var body envelope
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	assert.Equal(t, "about:blank", body.Type)
	assert.Equal(t, http.StatusOK, body.Status)
	assert.Equal(t, "OK", body.Title)
	assert.Equal(t, "/users/1", body.Instance)
	assert.NotNil(t, body.Data)
	assert.Empty(t, body.Detail)
	assert.Nil(t, body.Extensions)
}

func TestWriter_Created(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodPost, "/users")

	w.Created(rw, r, map[string]any{"id": "uuid-new"})

	assert.Equal(t, http.StatusCreated, rw.Code)
	assert.Equal(t, response.MediaTypeJSON, rw.Header().Get("Content-Type"))

	var body envelope
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	assert.Equal(t, http.StatusCreated, body.Status)
	assert.Equal(t, "Created", body.Title)
}

func TestWriter_Accepted(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodPost, "/jobs")

	w.Accepted(rw, r, map[string]any{"job_id": "job-1"})

	assert.Equal(t, http.StatusAccepted, rw.Code)
	assert.Equal(t, response.MediaTypeJSON, rw.Header().Get("Content-Type"))
}

func TestWriter_NoContent(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodDelete, "/users/1")

	w.NoContent(rw, r)

	assert.Equal(t, http.StatusNoContent, rw.Code)
	assert.Empty(t, rw.Body.Bytes())
}

func TestWriter_Write(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodGet, "/orders/99")

	w.Write(rw, r, response.ErrNotFound.WithDetail("order not found"))

	assert.Equal(t, http.StatusNotFound, rw.Code)
	assert.Equal(t, response.MediaTypeProblem, rw.Header().Get("Content-Type"))

	var body envelope
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	assert.Equal(t, "about:blank", body.Type)
	assert.Equal(t, http.StatusNotFound, body.Status)
	assert.Equal(t, "Not Found", body.Title)
	assert.Equal(t, "order not found", body.Detail)
	assert.Equal(t, "/orders/99", body.Instance)
	assert.Nil(t, body.Data)
}

func TestWriter_BadRequest(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodPost, "/users")

	w.BadRequest(rw, r, "invalid json")

	assert.Equal(t, http.StatusBadRequest, rw.Code)
	assert.Equal(t, response.MediaTypeProblem, rw.Header().Get("Content-Type"))

	var body envelope
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	assert.Equal(t, "invalid json", body.Detail)
}

func TestWriter_Unauthorized(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodGet, "/profile")

	w.Unauthorized(rw, r, "token expired")

	assert.Equal(t, http.StatusUnauthorized, rw.Code)
}

func TestWriter_Forbidden(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodDelete, "/admin")

	w.Forbidden(rw, r, "admin role required")

	assert.Equal(t, http.StatusForbidden, rw.Code)
}

func TestWriter_NotFound(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodGet, "/users/99")

	w.NotFound(rw, r, "user not found")

	assert.Equal(t, http.StatusNotFound, rw.Code)
}

func TestWriter_Conflict(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodPost, "/users")

	w.Conflict(rw, r, "email already registered")

	assert.Equal(t, http.StatusConflict, rw.Code)
}

func TestWriter_InternalServerError(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodGet, "/users")

	w.InternalServerError(rw, r)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)

	var body envelope
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	assert.Empty(t, body.Detail)
}

func TestWriter_Error_AppErr(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)

	t.Run("maps KindNotFound to 404", func(t *testing.T) {
		t.Parallel()
		rw, r := newRecorder(), newRequest(http.MethodGet, "/users/1")
		w.Error(rw, r, makeAppErr(apperr.KindNotFound))
		assert.Equal(t, http.StatusNotFound, rw.Code)
	})

	t.Run("includes detail from message", func(t *testing.T) {
		t.Parallel()
		rw, r := newRecorder(), newRequest(http.MethodGet, "/users/1")
		w.Error(rw, r, makeAppErrWithMessage(apperr.KindNotFound, "user 42 not found"))

		var body envelope
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
		assert.Equal(t, "user 42 not found", body.Detail)
	})

	t.Run("suppresses detail for KindInternal", func(t *testing.T) {
		t.Parallel()
		rw, r := newRecorder(), newRequest(http.MethodGet, "/users/1")
		w.Error(rw, r, makeAppErrWithMessage(apperr.KindInternal, "secret internal detail"))

		var body envelope
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
		assert.Empty(t, body.Detail)
	})

	t.Run("suppresses detail for KindUnknown", func(t *testing.T) {
		t.Parallel()
		rw, r := newRecorder(), newRequest(http.MethodGet, "/")
		w.Error(rw, r, makeAppErrWithMessage(apperr.KindUnknown, "secret"))

		var body envelope
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
		assert.Empty(t, body.Detail)
	})

	t.Run("includes validation fields in extensions", func(t *testing.T) {
		t.Parallel()
		rw, r := newRecorder(), newRequest(http.MethodPost, "/users")
		w.Error(rw, r, makeAppErrWithValidation())

		assert.Equal(t, http.StatusConflict, rw.Code)
		var body envelopeWithExtensions
		require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
		require.NotNil(t, body.Extensions)
		assert.NotEmpty(t, body.Extensions.Fields)
		assert.Equal(t, "email", body.Extensions.Fields[0].Field)
	})
}

func TestWriter_Error_ValidationErr(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodPost, "/users")

	w.Error(rw, r, makeValidationErr())

	assert.Equal(t, http.StatusUnprocessableEntity, rw.Code)
	assert.Equal(t, response.MediaTypeProblem, rw.Header().Get("Content-Type"))

	var body envelopeWithExtensions
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	assert.Equal(t, "invalid user", body.Detail)
	require.NotNil(t, body.Extensions)
	assert.Len(t, body.Extensions.Fields, 2)
	assert.Equal(t, "name", body.Extensions.Fields[0].Field)
	assert.Equal(t, "email", body.Extensions.Fields[1].Field)
}

func TestWriter_Error_UnknownErr(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodGet, "/")

	w.Error(rw, r, errors.New("unexpected plain error"))

	assert.Equal(t, http.StatusInternalServerError, rw.Code)

	var body envelope
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	assert.Empty(t, body.Detail)
}

func TestWriter_Meta_PopulatedFromExtractor(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(func(_ *http.Request) map[string]any {
		return map[string]any{"request_id": "req-abc"}
	})
	rw, r := newRecorder(), newRequest(http.MethodGet, "/users")

	w.Ok(rw, r, nil)

	var body envelope
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	require.NotNil(t, body.Meta)
	assert.Equal(t, "req-abc", body.Meta["request_id"])
}

func TestWriter_Meta_OmittedWhenNil(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodGet, "/users")

	w.Ok(rw, r, nil)

	var body envelope
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	assert.Nil(t, body.Meta)
}

func TestWriter_Meta_InErrorResponse(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(func(_ *http.Request) map[string]any {
		return map[string]any{"request_id": "req-xyz"}
	})
	rw, r := newRecorder(), newRequest(http.MethodGet, "/")

	w.Error(rw, r, errors.New("boom"))

	var body envelope
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	require.NotNil(t, body.Meta)
	assert.Equal(t, "req-xyz", body.Meta["request_id"])
}

func TestWriter_Instance_SetFromRequestPath(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)

	paths := []string{"/users/42", "/orders", "/lists/1/todos/2"}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			rw, r := newRecorder(), newRequest(http.MethodGet, path)
			w.Ok(rw, r, nil)

			var body envelope
			require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
			assert.Equal(t, path, body.Instance)
		})
	}
}

func TestWriter_ContentType(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)

	t.Run("success uses application/json", func(t *testing.T) {
		t.Parallel()
		rw, r := newRecorder(), newRequest(http.MethodGet, "/")
		w.Ok(rw, r, nil)
		assert.Equal(t, response.MediaTypeJSON, rw.Header().Get("Content-Type"))
	})

	t.Run("error uses application/problem+json", func(t *testing.T) {
		t.Parallel()
		rw, r := newRecorder(), newRequest(http.MethodGet, "/")
		w.Error(rw, r, errors.New("boom"))
		assert.Equal(t, response.MediaTypeProblem, rw.Header().Get("Content-Type"))
	})

	t.Run("Write uses application/problem+json", func(t *testing.T) {
		t.Parallel()
		rw, r := newRecorder(), newRequest(http.MethodGet, "/")
		w.Write(rw, r, response.ErrNotFound)
		assert.Equal(t, response.MediaTypeProblem, rw.Header().Get("Content-Type"))
	})
}

func TestWriteJSON_MarshalFailure(t *testing.T) {
	t.Parallel()

	w := response.NewWriter(noMeta)
	rw, r := newRecorder(), newRequest(http.MethodGet, "/")

	// channel is not JSON-marshalable — triggers the marshal error branch
	w.Ok(rw, r, make(chan struct{}))

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
}
