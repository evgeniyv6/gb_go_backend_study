package main

import (
	"fmt"
	"github.com/gorilla/sessions"
	"net/http"
)

var (
	key   = []byte("secret-key")
	store = sessions.NewCookieStore(key)
)

func secret(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, "cookie-name")

	if auth, ok := sess.Values["authenticated"].(bool); !ok || !auth {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	fmt.Fprintf(w, "The secret sentence")
}

func login(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, "cookie-name")
	sess.Values["authenticated"] = true
	sess.Save(r, w)
}

func logout(w http.ResponseWriter, r *http.Request) {
	sess, _ := store.Get(r, "cookie-name")
	sess.Values["authenticated"] = false
	sess.Save(r, w)
}

func main() {
	http.HandleFunc("/secret", secret)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)

	http.ListenAndServe(":6093", nil)

}
