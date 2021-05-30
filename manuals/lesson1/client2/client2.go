package client2

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func Client2(address string, port int) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", address, port))
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	go func() {
		io.Copy(os.Stdout, conn)
	}()

	io.Copy(conn, os.Stdin)
	fmt.Printf("%s: exit", conn.LocalAddr())
}
