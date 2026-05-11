# lib-response

[![CI](https://github.com/selfshop-dev/lib-response/actions/workflows/ci.yml/badge.svg)](https://github.com/selfshop-dev/lib-response/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/selfshop-dev/lib-response/branch/main/graph/badge.svg)](https://codecov.io/gh/selfshop-dev/lib-response)
[![Go Report Card](https://goreportcard.com/badge/github.com/selfshop-dev/lib-response)](https://goreportcard.com/report/github.com/selfshop-dev/lib-response)
[![Go version](https://img.shields.io/github/go-mod/go-version/selfshop-dev/lib-response)](go.mod)
[![License](https://img.shields.io/github/license/selfshop-dev/lib-response)](LICENSE)

RFC-9457 HTTP response serialization for Go services. The HTTP wire layer for the selfshop-dev stack — the single place where domain errors are turned into HTTP responses. A project by [selfshop-dev](https://github.com/selfshop-dev).

### Installation

```bash
go get -u github.com/selfshop-dev/lib-response
```

## Overview

All responses — successful and error — use a unified JSON envelope in RFC-9457 format. Successful responses carry `data`, error responses carry `detail` and `extensions`. Content-Type is `application/json` for 2xx, `application/problem+json` for 4xx/5xx.

**201 Created**
```json
{
    "type":     "about:blank",
    "status":   201,
    "title":    "Created",
    "data":     { "id": "uuid" },
    "instance": "/orders",
    "meta":     { "request_id": "abc123" }
}
```

**422 Unprocessable Entity**
```json
{
    "type":       "about:blank",
    "status":     422,
    "title":      "Unprocessable Entity",
    "detail":     "invalid order",
    "instance":   "/orders",
    "meta":       { "request_id": "abc123" },
    "extensions": {
        "fields": [
            { "field": "email", "code": "required", "message": "email is required" },
            { "field": "quantity", "code": "out_of_range", "message": "quantity must be between 1 and 100" }
        ]
    }
}
```

### Quick Start

```go
import response "github.com/selfshop-dev/lib-response"

var respond = response.NewWriter(func(r *http.Request) map[string]any {
    return map[string]any{"request_id": httpx.RequestIDFromContext(r.Context())}
})

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    order, err := h.svc.Create(r.Context(), cmd)
    if err != nil {
        respond.Error(w, r, err)
        return
    }
    respond.Created(w, r, order)
}
```

## Writer

`Writer` is the main entry point. It is created once at service startup via `NewWriter` with a `MetaExtractor` that builds the `meta` field from the incoming request. All methods are safe for concurrent use.

```go
var respond = response.NewWriter(func(r *http.Request) map[string]any {
    id := httpx.RequestIDFromContext(r.Context())
    if id == "" {
        return nil // meta field will be omitted
    }
    return map[string]any{"request_id": id}
})
```

Successful response methods accept an arbitrary `data` payload:

```go
respond.Ok(w, r, user)       // 200
respond.Created(w, r, order) // 201
respond.Accepted(w, r, job)  // 202
respond.NoContent(w, r)      // 204 — no body
```

Error response methods accept a `detail` string or an error:

```go
respond.BadRequest(w, r, "invalid json body")
respond.Unauthorized(w, r, "token expired")
respond.Forbidden(w, r, "admin role required")
respond.NotFound(w, r, "user not found")
respond.Conflict(w, r, "email already registered")
respond.InternalServerError(w, r) // detail is always omitted
```

## Error Mapping

`Writer.Error` inspects the error chain in priority order and selects the correct response automatically.

```go
respond.Error(w, r, err)
```

The handling priority is as follows. First, `*apperr.Error` is checked — the status is determined by `Kind`; for `KindInternal` and `KindUnknown` the `detail` is suppressed; if the `apperr` carries a `*validation.Error`, it is placed into `extensions.fields`. If no `*apperr.Error` is found, `*validation.Error` is checked — a 422 is returned with `extensions.fields`. For all other errors, a 500 is returned without `detail`.

| `apperr.Kind` | HTTP status |
|---|---|
| `KindNotFound` | 404 |
| `KindUnauthorized` | 401 |
| `KindForbidden` | 403 |
| `KindConflict` | 409 |
| `KindUnprocessable` | 422 |
| `KindUnavailable` | 503 |
| `KindTimeout` | 504 |
| `KindInternal`, `KindUnknown` | 500 (no detail) |

## Sentinels

The package exports ready-made `*Problem` values for common HTTP errors. Sentinels are safe for concurrent use across requests — `WithDetail` returns a copy and never mutates the receiver.

```go
respond.Write(w, r, response.ErrNotFound)
respond.Write(w, r, response.ErrNotFound.WithDetail("order not found"))
```

Available sentinels: `ErrBadRequest`, `ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, `ErrMethodNotAllowed`, `ErrConflict`, `ErrUnprocessable`, `ErrTooManyRequests`, `ErrInternalServerError`, `ErrNotImplemented`, `ErrServiceUnavailable`.

## License

[`MIT`](LICENSE) © 2026-present [`selfshop-dev`](https://github.com/selfshop-dev)