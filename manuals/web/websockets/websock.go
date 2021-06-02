package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)

		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			log.Printf("%s sent: %s\n", conn.RemoteAddr(), string(msg))

			if err = conn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	})

	http.HandleFunc("/main", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "websock.html")
	})

	http.ListenAndServe(":6094", nil)
}
