package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

// клиент особо не модифицировал, добавил обработку ошибок
func Client(address string, port int) {
	conn, err := net.Dial("tcp", net.JoinHostPort(address, strconv.Itoa(port)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Connected to the server %s:%d\n", address, port)
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil {
			log.Fatal(err)
		}
	}()

	_, err = io.Copy(conn, os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s: exit", conn.LocalAddr())
}

func main() {
	Client("localhost", 8000)
}
