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
	addr      string
	suggester *suggester
}

func newServer(addr string, sug *suggester) *server {
	return &server{
		addr:      addr,
		suggester: sug,
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

	q := r.URL.Query().Get("q")
	lat := r.URL.Query().Get("latitude")
	lon := r.URL.Query().Get("longitude")
	log.Printf("serving request with params q=%#v lat=%#v lon=%#v", q, lat, lon)

	ms, err := s.suggester.Match(q, lat, lon)
	if err != nil {
		log.Printf("failed to find suggestions err=%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentTypeJSON)

	sr := &SuggestionsResponse{Suggestions: ms}
	err = json.NewEncoder(w).Encode(sr)
	if err != nil {
		log.Printf("failed to encode response err=%v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
