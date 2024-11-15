package net

import (
	"fmt"
	"log/slog"
	"net"
	"time"
)

func SendHello(log *slog.Logger, port int) error {
	log = log.With(slog.String("Component", "UDP"))
	msg := "HELLO?"
	err := SendBroadcastMessage(log, port, msg)
	if err != nil {
		log.Error(fmt.Sprintf("First err: %s", err))
		time.Sleep(1 * time.Second)
		err = SendBroadcastMessage(log, port, msg)
		if err != nil {
			log.Error(fmt.Sprintf("Second err: %s", err))
		}
	}
	return err
}

func sendMessage(log *slog.Logger, addr net.IPAddr, port int, msg string) {
	// log = log.With(slog.String("Component", "UDP"))
}

func SendBroadcastMessage(log *slog.Logger, port int, msg string) error {
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
	log *slog.Logger,
	port int,
	channel chan string,
) {
	log = log.With(slog.String("Component", "UDP"))

	addr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: port,
	}

	conn, err := net.ListenUDP("net", &addr)
	if err != nil {
		log.Error("Failed to listen on UDP port: %w", err)
		return
	} else {
		log.Info("Listening for UDP broadcast...\n")
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			log.Error("Failed to close connection: %v\n", err)
			return
		}
	}(conn)

	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Error("Failed to read from UDP connection: %w", err)
			return
		}

		message := string(buffer[:n])
		if message == "HELLO?" {
			log.Info("Received 'HELLO?' message from %s\n", remoteAddr)
			channel <- message
		}
	}
}
