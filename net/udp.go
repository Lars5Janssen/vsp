package net

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type UDP struct {
	Addr    net.UDPAddr
	Message string
}

func SendHello(log *slog.Logger, port int) error {
	log = log.With(slog.String("Component", "UDP"))
	msg := "HELLO?"
	err := SendMessage(log, net.UDPAddr{}, port, msg)
	if err != nil {
		log.Error(fmt.Sprintf("First err: %s", err))
		time.Sleep(1 * time.Second)
		err = SendMessage(log, net.UDPAddr{}, port, msg)
		if err != nil {
			log.Error(fmt.Sprintf("Second err: %s", err))
		}
	}
	return err
}

func SendMessage(log *slog.Logger, addr net.UDPAddr, port int, msg string) error {
	log = log.With(slog.String("Component", "UDP"))
	conn, err := net.DialUDP("udp", &addr, &net.UDPAddr{
		IP:   addr.IP,
		Port: port,
	})
	if addr.IP == nil {
		conn, err = net.DialUDP("udp", nil, &net.UDPAddr{
			IP:   net.IPv4bcast,
			Port: port,
		})
	}

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
	channel chan UDP,
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
	// udpServer.SetReadDeadline(<-time.After(time.Second))
	for {
		buf := make([]byte, 1024)
		numOfBytes, addr, err := udpServer.ReadFrom(buf)
		if err != nil {
			log.Error(fmt.Sprintf("Error during Reading from buffer. Error: %s", err.Error()))
		}

		isOwn := ownAddrCheck(*log, addr)
		if isOwn {
			continue
		}

		toSize := make([]byte, numOfBytes)
		for i := range toSize {
			toSize[i] = buf[i]
		}
		received := string(toSize)
		// TODO if string should end with \r\n then problem is here
		received = strings.TrimRight(received, "\r\n")
		log.Debug(
			"Received Packet",
			slog.Int("Length", numOfBytes),
			slog.String("From", addr.String()),
			slog.String("Content", received),
		)
		// to send back to
		udpAddr, ok := addr.(*net.UDPAddr)
		if !ok {
			log.Error("Address is not a UDP address")
			return
		}
		receivedUdp := UDP{
			Addr:    *udpAddr,
			Message: received,
		}
		channel <- receivedUdp
	}
	// udpServer.Close()
}

func ownAddrCheck(log slog.Logger, addr net.Addr) bool {
	interfaceAddrs, errAddr := net.InterfaceAddrs()
	if errAddr != nil {
		log.Error("Error in getting own Ip")
		os.Exit(2)
	}

	ipAddr := convertNetAddrToIp(log, addr)

	for _, a := range interfaceAddrs {
		aConvert := convertNetAddrToIp(log, a)

		if aConvert.Equal(ipAddr) {
			log.Debug("Own IP found", "Addr", ipAddr.String(), "From Interface", aConvert.String())
			return true
		}
	}
	return false
}

func convertNetAddrToIp(log slog.Logger, addr net.Addr) net.IP {
	addrString := addr.String()
	indexOfPort := strings.LastIndex(addrString, ":")
	if indexOfPort != -1 {
		addrString = addrString[0:indexOfPort]

	}
	indexOfMask := strings.LastIndex(addrString, "/")
	if indexOfMask != -1 {
		addrString = addrString[0:indexOfMask]

	}
	conv := net.ParseIP(addrString)
	// log.Debug("Parsed IP", "Imput", addr.String(), "Out", conv.String())
	return conv
}
