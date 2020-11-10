package couchdb

import (
	"net/http"

	"github.com/go-kivik/kivik/v4"
)

type Error = kivik.Error

// ErrorStatus checks whether the given error's status code matches statusCode.
func ErrorStatus(err error, statusCode int) bool {
	return kivik.StatusCode(err) == statusCode
}

// NotFound checks whether the given error has StatusCode == 404.
func NotFound(err error) bool {
	return ErrorStatus(err, http.StatusNotFound)
}

// Conflict checks whether the given error has StatusCode == 409.
func Conflict(err error) bool {
	return ErrorStatus(err, http.StatusConflict)
}

// Unauthorized checks whether the given error has StatusCode == 401.
func Unauthorized(err error) bool {
	return ErrorStatus(err, http.StatusUnauthorized)
}
