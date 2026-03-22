# lib-response

[![CI](https://github.com/selfshop-dev/lib-response/actions/workflows/ci.yml/badge.svg)](https://github.com/selfshop-dev/lib-response/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/selfshop-dev/lib-response/branch/main/graph/badge.svg)](https://codecov.io/gh/selfshop-dev/lib-response)
[![Go Report Card](https://goreportcard.com/badge/github.com/selfshop-dev/lib-response)](https://goreportcard.com/report/github.com/selfshop-dev/lib-response)
[![Go version](https://img.shields.io/github/go-mod/go-version/selfshop-dev/lib-response)](go.mod)
[![License](https://img.shields.io/github/license/selfshop-dev/lib-response)](LICENSE)

RFC-9457 HTTP-ответы для Go-сервисов — единый конверт для success и error, без дублирования логики маппинга по хендлерам.

## Overview

Каждый HTTP-сервис решает одни и те же задачи: сериализовать успешный ответ, замаппить domain-ошибку в статус-код, не протечь внутренние детали наружу. `lib-response` централизует это в одном месте.

Пакет знает про [`lib-apperr`](https://github.com/selfshop-dev/lib-apperr) и [`lib-validation`](https://github.com/selfshop-dev/lib-validation) и является единственным местом в стеке, где domain-ошибки превращаются в HTTP-ответы — все остальные пакеты остаются transport-agnostic.

```
lib-validation — что именно не так (поля, коды)
lib-apperr     — почему не так (kind, op, cause, stack)
lib-response   — как сообщить клиенту (HTTP, RFC-9457)
```

### Installation

```bash
go get -u github.com/selfshop-dev/lib-response
```

### Быстрый старт

```go
import "github.com/selfshop-dev/lib-response"

var respond = response.NewWriter(func(r *http.Request) map[string]any {
    return map[string]any{
        "request_id": httpx.RequestIDFromContext(r.Context()),
    }
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

## Конверт

Каждый ответ — success и error alike — использует один JSON-конверт, построенный поверх [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457):

```jsonc
// 201 Created
{
    "type":     "about:blank",
    "status":   201,
    "title":    "Created",
    "data":     { "id": "uuid" },
    "instance": "/orders",
    "meta":     { "request_id": "abc123" }
}

// 422 Unprocessable Entity
{
    "type":       "about:blank",
    "status":     422,
    "title":      "Unprocessable Entity",
    "detail":     "invalid order",
    "instance":   "/orders",
    "meta":       { "request_id": "abc123" },
    "extensions": { "fields": [ ... ] }
}
```

`Content-Type` — `application/json` для 2xx и `application/problem+json` для 4xx/5xx. Поле `data` опускается в ошибках; `detail` и `extensions` — в успешных ответах.

## Writer

[`Writer`](writer.go) — основная точка входа. Конструируй один раз при старте через `NewWriter` и переиспользуй в каждом хендлере.

### Успешные ответы

| Метод | Статус | Content-Type |
|---|---|---|
| `OK(w, r, data)` | 200 | `application/json` |
| `Created(w, r, data)` | 201 | `application/json` |
| `Accepted(w, r, data)` | 202 | `application/json` |
| `NoContent(w, r)` | 204 | — |

### Ошибки

`Error` инспектирует цепочку ошибок в порядке приоритета:

1. `*apperr.Error` → статус из `Kind`; `KindInternal` и `KindUnknown` никогда не раскрывают `Message` клиенту. Если `apperr` несёт `*validation.Error` — включается как `extensions.fields`.
2. `*validation.Error` → 422 с `extensions.fields`.
3. Всё остальное → 500 без `detail`.

Convenience-методы для типичных ошибок хендлера — до вызова сервиса:

| Метод | Статус |
|---|---|
| `BadRequest(w, r, detail)` | 400 |
| `Unauthorized(w, r, detail)` | 401 |
| `Forbidden(w, r, detail)` | 403 |
| `NotFound(w, r, detail)` | 404 |
| `Conflict(w, r, detail)` | 409 |
| `InternalServerError(w, r)` | 500 |

## Sentinels

Пакетные sentinel-проблемы безопасны для шаринга между запросами — [`WithDetail`](problem.go) возвращает копию и никогда не мутирует receiver:

```go
respond.Write(w, r, response.ErrNotFound)
respond.Write(w, r, response.ErrNotFound.WithDetail("order not found"))
```

Доступные сентинели: `ErrBadRequest`, `ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, `ErrMethodNotAllowed`, `ErrConflict`, `ErrUnprocessable`, `ErrTooManyRequests`, `ErrInternalServerError`, `ErrNotImplemented`, `ErrServiceUnavailable`.

## Meta

`MetaExtractor` вызывается при каждом ответе. Верни `nil` чтобы поле `meta` не появлялось в ответе:

```go
var respond = response.NewWriter(func(r *http.Request) map[string]any {
    id := httpx.RequestIDFromContext(r.Context())
    if id == "" {
        return nil
    }
    return map[string]any{"request_id": id}
})
```

## Производительность

Результаты (`go test -bench=. -benchmem -benchtime=3s ./...`):

```
BenchmarkProblem_WithDetail-12                  1000000000          2.5 ns/op          0 B/op      0 allocs/op
BenchmarkWriter_NoContent-12                      38001510         87.2 ns/op         48 B/op      1 allocs/op
BenchmarkWriter_Write_Sentinel-12                  2153740       1695   ns/op       1280 B/op     11 allocs/op
BenchmarkWriter_OK-12                              1511598       2382   ns/op       1424 B/op     16 allocs/op
BenchmarkWriter_OK_WithMeta-12                     1277067       2817   ns/op       1841 B/op     19 allocs/op
BenchmarkWriter_Error_AppErr_NotFound-12           2223900       1662   ns/op       1328 B/op     12 allocs/op
BenchmarkWriter_Error_AppErr_Internal-12           2236429       1620   ns/op       1296 B/op     12 allocs/op
BenchmarkWriter_Error_AppErr_WithValidation-12     1527910       2350   ns/op       1593 B/op     14 allocs/op
BenchmarkWriter_Error_ValidationErr-12              978885       3440   ns/op       1929 B/op     15 allocs/op
```

Несколько ориентиров:

- **`WithDetail` (~2.5 ns, 0 allocs)** — struct copy на стеке; sentinel-паттерн не стоит ничего на горячем пути.
- **`NoContent` (~87 ns, 1 alloc)** — минимальный путь, только `WriteHeader`. Аллокация — сам `httptest.ResponseRecorder` в бенчмарке.
- **`OK` / error-методы (~1.6–2.4 µs)** — стоимость определяется `json.Marshal` конверта и аллокациями `http.ResponseWriter`. `WithMeta` добавляет ~400 ns за map-аллокацию в экстракторе.
- **`ValidationErr` (~3.4 µs)** — самый тяжёлый путь из-за сериализации `extensions.fields`; линейно растёт с числом field-ошибок.

## Makefile

Основные возможности:

| Цель | Описание |
|---|---|
| `make code-gen` | Запустить `go generate ./...` |
| `make lint` | Запустить golangci-lint |
| `make test` | Генерация кода + тесты с coverage |
| `make prof` | Собрать профили (cpu, mem, block, mutex) |
| `make prof-view` | Открыть профиль в браузере (`FILE=cpu.out` по умолчанию) |

## Лицензия

[`MIT`](LICENSE) © 2026-present [`selfshop-dev`](https://github.com/selfshop-dev)