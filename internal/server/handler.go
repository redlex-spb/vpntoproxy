package server

import (
	"github.com/go-chi/render"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("home"))
	if err != nil {
		render.JSON(w, r, map[string]interface{}{"status": "fail", "error": err.Error(), "data": nil})
		return
	}
}
