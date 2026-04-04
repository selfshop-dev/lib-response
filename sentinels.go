package response

import "net/http"

// Package-level sentinel problems. Safe to share across requests —
// [Problem.WithDetail] returns a copy and never mutates the receiver.
//
//	respond.Write(w, r, ErrNotFound)
//	respond.Write(w, r, ErrNotFound.WithDetail("order not found"))
var (
	ErrBadRequest          = newProblem(http.StatusBadRequest)
	ErrUnauthorized        = newProblem(http.StatusUnauthorized)
	ErrForbidden           = newProblem(http.StatusForbidden)
	ErrNotFound            = newProblem(http.StatusNotFound)
	ErrMethodNotAllowed    = newProblem(http.StatusMethodNotAllowed)
	ErrConflict            = newProblem(http.StatusConflict)
	ErrUnprocessable       = newProblem(http.StatusUnprocessableEntity)
	ErrTooManyRequests     = newProblem(http.StatusTooManyRequests)
	ErrInternalServerError = newProblem(http.StatusInternalServerError)
	ErrNotImplemented      = newProblem(http.StatusNotImplemented)
	ErrServiceUnavailable  = newProblem(http.StatusServiceUnavailable)
)
