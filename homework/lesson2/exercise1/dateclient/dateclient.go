package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

func Client(address string, port int) {
	conn, err := net.Dial("tcp", net.JoinHostPort(address, strconv.Itoa(port)))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Connected to the server %s:%d\n", address, port)
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	buf := make([]byte, 256)

	for {
		_, err = conn.Read(buf)
		if err == io.EOF {
			break
		}
		_, err = io.WriteString(os.Stdout, string(buf))
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func main() {
	Client("localhost", 8000)
}
