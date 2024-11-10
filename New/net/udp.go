package net

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"
)

func SendHello(logger *slog.Logger, port int) error {
	msg := "HELLO?"
	err := sendBroadcastMessage(logger, port, msg)
	if err != nil {
		logger.Error(fmt.Sprintf("First err: %s", err))
		time.Sleep(1 * time.Second)
		err = sendBroadcastMessage(logger, port, msg)
		if err != nil {
			logger.Error(fmt.Sprintf("Second err: %s", err))
		}
	}
	return err
}

func sendBroadcastMessage(logger *slog.Logger, port int, msg string) error {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: port,
	})
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			logger.Error(err.Error())
			return
		}
	}(conn)

	message := []byte(msg)
	_, err = conn.Write(message)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	logger.Info("Broadcast message sent")
	return nil
}

func ListenForBroadcastMessage(
	wg *sync.WaitGroup,
	ctx context.Context,
	logger *slog.Logger,
	port int,
	channel chan string,
) {
	wg.Add(1)
	defer wg.Done()
}
