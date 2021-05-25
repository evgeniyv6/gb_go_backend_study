package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

// вместо отдельного канала создаю структуру для хранения имени пользователя
type client struct {
	Messages chan<- string
	Nickname string
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

func Server(address string, port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server started.")
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
		case m := <-messages:
			for c := range clients {
				c.Messages <- m
			}
		case c := <-entering:
			clients[c] = true
			for cl := range clients {
				cl.Messages <- "Nick: " + c.Nickname
			}
		case c := <-leaving:
			delete(clients, c)
			close(c.Messages)
		}
	}
}

func handleConn(conn net.Conn) {
	timeout := time.NewTimer(100 * time.Second) // время простоя, если пользователь при подключении ничего не вводит - то закрываем соединение
	ch := make(chan string)                     // канал для отправки сообщений клиентам
	go clientWriter(conn, ch)

	enter := make(chan string) // входящие сообщения от клиентов
	go func() {
		input := bufio.NewScanner(conn)
		for input.Scan() {
			enter <- input.Text()
		}
	}()

	var who string
	ch <- "Enter your name: "
	who = <-enter // conn.RemoteAddr().String()

	cl := client{ch, who}
	messages <- "New user has arrived"
	entering <- cl
loop:
	for {
		select {
		case m := <-enter:
			messages <- who + " : " + m
		case <-timeout.C: // если пользователь при подключении ничего не вводит - то закрываем соединение
			err := conn.Close()
			if err != nil {
				log.Fatal(err)
			}
			break loop
		}

	}

	leaving <- cl
	messages <- who + " has left"
	err := conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		_, err := fmt.Fprintln(conn, msg)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	Server("localhost", 8000)
}
