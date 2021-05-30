package server2

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client chan<- string

var (
	entering = make(chan client)
	leaving  = make(chan client)
	msg      = make(chan string)
)

func Server2(address string, port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConn(conn)
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case m := <-msg:
			for c := range clients {
				c <- m
			}
		case c := <-entering:
			clients[c] = true
		case c := <-leaving:
			delete(clients, c)
			close(c)
		}
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

	who := conn.RemoteAddr().String()
	ch <- "You are" + who
	msg <- who + "has arrived"
	entering <- ch

	input := bufio.NewScanner(conn)

	for input.Scan() {
		msg <- who + ": " + input.Text()
	}

	leaving <- ch
	msg <- who + "has left"
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for l := range ch {
		fmt.Fprintln(conn, l)
	}
}
