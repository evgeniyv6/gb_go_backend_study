package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fmt.Fprintln(w, "GET HANDLER", vars["id"])
		return
	}).Methods(http.MethodGet)

	router.HandleFunc("/{id}/name/{name}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fmt.Fprintf(w, "GET HANDLET id: %s, name: name: %s", vars["id"], vars["name"])
	}).Methods(http.MethodGet)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "POST HANDLER")
	}).Methods(http.MethodPost)

	log.Fatal(http.ListenAndServe(":6096", router))
}
