package client

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func Client(address string, port int) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 256)

	for {
		_, err = conn.Read(buf)
		if err == io.EOF {
			break
		}
		io.WriteString(os.Stdout, fmt.Sprintf("output: %s\n", string(buf)))
	}
}
