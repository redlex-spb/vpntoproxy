package vpn

import (
	"github.com/go-chi/chi"
	"net/http"
)

func Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/", list)
	r.Get("/{ID}", detail)
	r.Post("/", create)
	r.Delete("/{ID}", del)

	r.Get("/checkVpn", checkVpn)
	r.Get("/checkProxy", checkProxy)

	return r
}
