package util

import (
	"net/http"
)

// Mux is a standard mux interface for HTTP
type Mux interface {
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}
