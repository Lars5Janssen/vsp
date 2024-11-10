package utils

import (
	"fmt"
	"net"
	"time"
)

func SendBroadcastMessage(port int) error {
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: port,
	})
	if err != nil {
		return fmt.Errorf("failed to dial UDP: %w", err)
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			fmt.Printf("failed to close connection: %v\n", err)
		}
	}(conn)

	message := []byte("HELLO?")
	_, err = conn.Write(message)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	fmt.Println("Broadcast message sent")
	return err
}

func ListenForBroadcastMessage(port int) error {
	addr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: port,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP port: %w", err)
	} else {
		fmt.Printf("Listening for UDP broadcast...\n")
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			fmt.Printf("failed to close connection: %v\n", err)
		}
	}(conn)

	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			return fmt.Errorf("failed to read from UDP connection: %w", err)
		}

		message := string(buffer[:n])
		if message == "HELLO?" {
			time.Sleep(100 * time.Millisecond)
			fmt.Printf("Received 'HELLO?' message from %s\n", remoteAddr)
		}
	}
}
