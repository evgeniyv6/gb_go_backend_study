package main

import (
	"bufio"
	"context"
	"fmt"
	config2 "gb_go_backend_study/homework/lesson2/exercise1/config"
	"go.uber.org/zap"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	messages = make(chan string) // канал для бродкастинга сообщений из консоли сервера всем подключенным клиентам
	sugar    *zap.SugaredLogger  // https://github.com/uber-go/zap логгер,
	ticker   *time.Ticker        // тиккер для рассылки клиентам datetime
	mu       sync.Mutex          // мьютекс для предотвращения гонки при доступе к списку клиентских соединений
)

// структура с параметрами для сервера
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
		address: addr,
		port:    port,
		sugar:   logger.Sugar(),
		ticker:  time.NewTicker(dur * time.Second),
	}
}

// рассылает либо datetime тиккера, либо сообщения из консоли сервера, либо завершается при ctx.Done
func spammer(clients *[]net.Conn, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			sugar.Info("Stop spammer")
			return
		case tickerTime := <-ticker.C:
			mu.Lock()
			for _, conn := range *clients {
				_, err := io.WriteString(conn, fmt.Sprintf("%s\n", tickerTime.Format(time.RFC850)))
				if err != nil {
					sugar.Error(err)
					return
				}
			}
			mu.Unlock()
		case msg := <-messages:
			mu.Lock()
			for _, conn := range *clients {
				_, err := io.WriteString(conn, fmt.Sprintf("%s\n", msg))
				if err != nil {
					sugar.Error(err)
					return
				}
			}
			mu.Unlock()
		}
	}
}

// запускает в горутине spammer
// ожидает ввода с консоли сервера и передает сообщение в канал messages
func (s *SrvParams) handleConn(clients *[]net.Conn, ctx context.Context) {
	go spammer(clients, ctx)

	input := bufio.NewScanner(os.Stdin)
	fmt.Print("Type here (ctrl+C for exit) > \n")
	for input.Scan() {
		fmt.Print("Type here (ctrl+C for exit) > \n")
		messages <- input.Text()
	}
	if err := input.Err(); err != nil {
		sugar.Errorf("Bufio scanner err: %s", err)
	}

	select {
	case <-ctx.Done():
		mu.Lock()
		for _, conn := range *clients {
			if err := conn.Close(); err != nil {
				sugar.Error(err)
			}

		}
		mu.Unlock()
		return
	}
}

func (s *SrvParams) Server() {
	var (
		clients      = []net.Conn{}
		ctx, cancel  = context.WithCancel(context.Background())
		cancelSignal = make(chan os.Signal, 1)
		done         = make(chan bool, 1) // канал, при получении сообщения в который - return

		// обработка сигнала прерывания
		catchSignal = func(cancelFunc context.CancelFunc, l net.Listener) {
			sig := <-cancelSignal
			sugar.Warnf("Received stop signal - %v", sig)
			done <- true
			cancelFunc()
			err := l.Close()
			if err != nil {
				sugar.Error(err)
			}
		}
	)
	signal.Notify(cancelSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	sugar = s.sugar
	ticker = s.ticker

	l, err := net.Listen("tcp", net.JoinHostPort(s.address, s.port))
	sugar.Infof("Server started. Listening to address: %s on port: %s", s.address, s.port)
	if err != nil {
		log.Fatal(err)
	}
	go catchSignal(cancel, l)
	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case d := <-done:
				sugar.Warn("Server stopted ", d)
				return
			default:
				sugar.Error("Caught error: ", err)
			}
		} else {
			mu.Lock()
			clients = append(clients, conn)
			mu.Unlock()
			sugar.Infof("Get connection from client #%d", len(clients))
			go s.handleConn(&clients, ctx)
		}
	}
}

func main() {
	config, err := config2.ReadConfig("./config.json") // параметры для сервера считываем из файла config.json см. package config
	if err != nil {
		log.Fatalf("Couldnot read configuration file. Err: %s", err)
	}
	srv := NewSrvParams(config.Server.Address, config.Server.Port, config.Ticker.Delay)
	srv.Server()
}
