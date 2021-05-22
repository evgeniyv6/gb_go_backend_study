package main

import (
	"bufio"
	"fmt"
	config2 "gb_go_backend_study/homework/lesson2/exercise1/config"
	"go.uber.org/zap"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var (
	messages = make(chan string)
	sugar    *zap.SugaredLogger
	ticker   *time.Ticker
)

type SrvParams struct {
	address, port string
	sugar         *zap.SugaredLogger
	ticker        *time.Ticker
}

func NewSrvParams(addr, port string, dur time.Duration) *SrvParams {
	logger, err := zap.NewProduction()
	defer logger.Sync()
	if err != nil {
		log.Println(err)
	}
	return &SrvParams{
		addr,
		port,
		logger.Sugar(),
		time.NewTicker(dur * time.Second),
	}
}

func spammer(clients *[]net.Conn) {
	for {
		select {
		case tickerTime := <-ticker.C:
			for _, conn := range *clients {
				_, err := io.WriteString(conn, fmt.Sprintf("%s\n", tickerTime.Format(time.RFC850)))
				if err != nil {
					sugar.Error(err)
					return
				}
			}
		case msg := <-messages:
			for _, conn := range *clients {
				_, err := io.WriteString(conn, fmt.Sprintf("%s\n", msg))
				if err != nil {
					sugar.Error(err)
					return
				}
			}

		}
	}
}

func handleConn(clients *[]net.Conn) {
	go spammer(clients)

	input := bufio.NewScanner(os.Stdin)
	fmt.Print("Type here: ")
	for input.Scan() {
		fmt.Print("Type here: ")
		messages <- input.Text()
	}
	if err := input.Err(); err != nil {
		sugar.Errorf("Bufio scanner err: %s", err)
	}
	for _, conn := range *clients {
		if err := conn.Close(); err != nil {
			sugar.Error(err)
		}
	}
}

func Server(params *SrvParams) {
	clients := []net.Conn{}
	sugar = params.sugar
	ticker = params.ticker

	l, err := net.Listen("tcp", net.JoinHostPort(params.address, params.port))
	sugar.Infof("Server started. Listening to address: %s on port: %s", params.address, params.port)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			sugar.Error(err)
			continue
		}
		clients = append(clients, conn)
		sugar.Infof("Get connection from client #%d", len(clients))
		go handleConn(&clients)
	}
}

func main() {
	config, err := config2.ReadConfig("./config.json")
	if err != nil {
		log.Fatalf("Couldnot read configuration file. Err: %s", err)
	}
	params := NewSrvParams(config.Server.Address, config.Server.Port, config.Ticker.Delay)
	Server(params)
}
