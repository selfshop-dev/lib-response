# lib-response

[![CI](https://github.com/selfshop-dev/lib-response/actions/workflows/ci.yml/badge.svg)](https://github.com/selfshop-dev/lib-response/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/selfshop-dev/lib-response/branch/main/graph/badge.svg)](https://codecov.io/gh/selfshop-dev/lib-response)
[![Go Report Card](https://goreportcard.com/badge/github.com/selfshop-dev/lib-response)](https://goreportcard.com/report/github.com/selfshop-dev/lib-response)
[![Go version](https://img.shields.io/github/go-mod/go-version/selfshop-dev/lib-response)](go.mod)
[![License](https://img.shields.io/github/license/selfshop-dev/lib-response)](LICENSE)

RFC-9457 сериализация HTTP-ответов для Go-сервисов. HTTP wire layer для стека selfshop-dev — единственное место, где доменные ошибки превращаются в HTTP-ответы. Проект организации [selfshop-dev](https://github.com/selfshop-dev).

### Installation

```bash
go get -u github.com/selfshop-dev/lib-response
```

## Overview

Все ответы — успешные и ошибочные — используют единый JSON-конверт в формате RFC-9457. Успешные ответы несут `data`, ошибочные — `detail` и `extensions`. Content-Type `application/json` для 2xx, `application/problem+json` для 4xx/5xx.

**201 Created**
```
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

### Быстрый старт

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

`Writer` — основная точка входа. Создаётся один раз при старте сервиса через `NewWriter` с `MetaExtractor`, который строит `meta`-поле из входящего запроса. Все методы безопасны для конкурентного использования.

```go
var respond = response.NewWriter(func(r *http.Request) map[string]any {
    id := httpx.RequestIDFromContext(r.Context())
    if id == "" {
        return nil // meta поле будет опущено
    }
    return map[string]any{"request_id": id}
})
```

Методы успешных ответов принимают произвольный `data`-payload:

```go
respond.Ok(w, r, user)       // 200
respond.Created(w, r, order) // 201
respond.Accepted(w, r, job)  // 202
respond.NoContent(w, r)      // 204 — без тела
```

Методы ошибочных ответов принимают `detail`-строку или ошибку:

```go
respond.BadRequest(w, r, "invalid json body")
respond.Unauthorized(w, r, "token expired")
respond.Forbidden(w, r, "admin role required")
respond.NotFound(w, r, "user not found")
respond.Conflict(w, r, "email already registered")
respond.InternalServerError(w, r) // detail всегда опускается
```

## Маппинг ошибок

`Writer.Error` инспектирует цепочку ошибок в порядке приоритета и выбирает правильный ответ автоматически.

```go
respond.Error(w, r, err)
```

Приоритет обработки следующий. Первым проверяется `*apperr.Error` — статус определяется по `Kind`, для `KindInternal` и `KindUnknown` `detail` подавляется; если `apperr` несёт `*validation.Error`, она попадает в `extensions.fields`. Если `*apperr.Error` не найден, проверяется `*validation.Error` — возвращается 422 с `extensions.fields`. Для всех прочих ошибок возвращается 500 без `detail`.

| `apperr.Kind` | HTTP статус |
|---|---|
| `KindNotFound` | 404 |
| `KindUnauthorized` | 401 |
| `KindForbidden` | 403 |
| `KindConflict` | 409 |
| `KindUnprocessable` | 422 |
| `KindUnavailable` | 503 |
| `KindTimeout` | 504 |
| `KindInternal`, `KindUnknown` | 500 (без detail) |

## Сентинелы

Пакет экспортирует готовые `*Problem` для типичных HTTP-ошибок. Сентинелы безопасны для совместного использования между запросами — `WithDetail` возвращает копию и никогда не мутирует получателя.

```go
respond.Write(w, r, response.ErrNotFound)
respond.Write(w, r, response.ErrNotFound.WithDetail("order not found"))
```

Доступные сентинелы: `ErrBadRequest`, `ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, `ErrMethodNotAllowed`, `ErrConflict`, `ErrUnprocessable`, `ErrTooManyRequests`, `ErrInternalServerError`, `ErrNotImplemented`, `ErrServiceUnavailable`.

## Лицензия

[`MIT`](LICENSE) © 2026-present [`selfshop-dev`](https://github.com/selfshop-dev)