package server

import (
	"github.com/go-chi/chi"
	"net/http"
	"vpntoproxy/internal/server/vpn"
)

func route() http.Handler {
	// create `ServerMux`
	mux := chi.NewRouter()

	mux.Get("/", home)
	mux.Mount("/api", apiRoute())

	return mux
}

func apiRoute() http.Handler {
	r := chi.NewRouter()

	r.Mount("/vpn", vpn.Router())

	return r
}
