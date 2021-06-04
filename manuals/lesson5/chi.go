package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

func main() {
	router := chi.NewRouter()

	router.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		fmt.Fprintf(w, "Chi GET HANDLER, id: %s", id)
		return
	})

	router.Post("/{id}/name/{name}", func(w http.ResponseWriter, r *http.Request) {
		id, name := chi.URLParam(r, "id"), chi.URLParam(r, "name")
		fmt.Fprintf(w, "Chi POST HANDLER. id: %s, name: %s", id, name)
	})

	router.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "got panic here")
			}
		}()
		panic("panic!!")
		fmt.Fprintln(w, "GET CHI PANIC")
	})

	log.Fatal(http.ListenAndServe(":6097", router))
}
