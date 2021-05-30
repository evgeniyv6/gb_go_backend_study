package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

var msg chan string

func Server(address string, port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(c net.Conn) {
	defer c.Close()

	for {
		_, err := io.WriteString(c, time.Now().Format(time.RFC850))
		if err != nil {
			return
		}
		time.Sleep(1 * time.Second)
	}
}
