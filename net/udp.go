package net

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
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

	udpServer, err := net.ListenPacket("udp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Error("Could not start UDP Server")
		os.Exit(2)
	}
	log.Info("Started UDP Server",
		slog.String("Addr of Server", udpServer.LocalAddr().String()),
		slog.String("Netw of Server", udpServer.LocalAddr().Network()),
	)
	for {
		buf := make([]byte, 1024)
		numOfBytes, addr, err := udpServer.ReadFrom(buf)
		if err != nil {
			log.Error(fmt.Sprintf("Error during Reading from buffer. Error: %s", err.Error()))
		}
		toSize := make([]byte, numOfBytes)
		for i, _ := range toSize {
			toSize[i] = buf[i]
		}
		recieved := string(toSize)
		log.Debug(
			"Recieved Packet",
			slog.Int("Length", numOfBytes),
			slog.String("From", addr.String()),
			slog.String("Content", recieved),
		)
		channel <- recieved
	}
	udpServer.Close()

}
