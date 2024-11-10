package net

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

func SendHello(log *slog.Logger, port int) error {
	log = log.With(slog.String("Component", "UDP"))
	msg := "HELLO?"
	err := sendBroadcastMessage(log, port, msg)
	if err != nil {
		log.Error(fmt.Sprintf("First err: %s", err))
		time.Sleep(1 * time.Second)
		err = sendBroadcastMessage(log, port, msg)
		if err != nil {
			log.Error(fmt.Sprintf("Second err: %s", err))
		}
	}
	return err
}

func sendBroadcastMessage(log *slog.Logger, port int, msg string) error {
	log = log.With(slog.String("Component", "UDP"))
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: port,
	})
	if err != nil {
		log.Error(err.Error())
		return err
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			log.Error(err.Error())
			return
		}
	}(conn)

	message := []byte(msg)
	_, err = conn.Write(message)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	log.Info("Broadcast message sent")
	return nil
}

func ListenForBroadcastMessage(
	ctx context.Context,
	wg *sync.WaitGroup,
	log *slog.Logger,
	port int,
	channel chan string,
) {
	wg.Add(1)
	defer wg.Done()
	log = log.With(slog.String("Component", "UDP"))
	// Send to channel:
	// channel <- "Test"
}
