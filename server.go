package main

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	pathSuggestions = "/suggestions"

	contentTypeJSON = "application/json"
)

type server struct {
	addr string
}

func newServer(addr string) *server {
	return &server{
		addr: addr,
	}
}

func (s *server) start() error {
	hs := &http.Server{Addr: s.addr, Handler: s}
	return hs.ListenAndServe()
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Printf("unsupport method=%#v", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if r.URL.Path != pathSuggestions {
		log.Printf("unsupported path=%#v", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	sr := newSuggestionsResponse()
	err := json.NewEncoder(w).Encode(sr)
	if err != nil {
		log.Printf("failed to encode response err=%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
