// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package adminserver

import (
	"context"
	"net/http"
)

type HTTPServer struct {
	server *http.Server
	mux    *http.ServeMux
}

// New initializes a new HttpServer with a specified configuration
func New(addr string) *HTTPServer {
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    addr,
	}
	return &HTTPServer{
		server: server,
		mux:    mux,
	}
}

// RegisterHandler adds a new route and its associated handler to the server
func (hs *HTTPServer) RegisterHandler(route string, handler http.Handler) {
	hs.mux.Handle(route, handler)
}

// Start begins running the HTTP server
func (hs *HTTPServer) Start() error {
	return hs.server.ListenAndServe()
}

func (hs *HTTPServer) Shutdown(ctx context.Context) error {
	return hs.server.Shutdown(ctx)
}
