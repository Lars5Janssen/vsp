package connection

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
)

type UDP struct {
	Addr    net.UDPAddr
	Message string
}

func SendHello(log *slog.Logger, port int) error {
	log = log.With(slog.String("LogFrom", "UDP"))
	msg := "HELLO?"
	err := SendMessage(log, net.UDPAddr{}, port, msg)
	if err != nil {
		log.Error(err.Error())
	}
	return err
}

func SendMessage(log *slog.Logger, addr net.UDPAddr, port int, msg string) error {
	log = log.With(slog.String("LogFrom", "UDP"))
	var conn *net.UDPConn
	var err error
	if addr.IP == nil {
		conn, err = net.DialUDP("udp", nil, &net.UDPAddr{
			IP:   net.IPv4bcast,
			Port: port,
		})
	} else {
		// TODO HERE LOOK HERE
		// TODO HERE IS YOUR MISTAKE
		conn, err = net.DialUDP("udp", nil, &net.UDPAddr{
			IP:   addr.IP, // YOU ARE SENDING TO YOURSELF
			Port: port,
		})
	}

	if err != nil {
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

	log.Info("Message sent", slog.String("Message", msg))
	return nil
}

func ListenForBroadcastMessage(
	log *slog.Logger,
	port int,
	channel chan UDP,
) {
	log = log.With(slog.String("LogFrom", "UDP"))

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

		isOwn := OwnAddrCheck(*log, addr.String())
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

func OwnAddrCheck(log slog.Logger, addr string) bool {
	interfaceAddrs, errAddr := net.InterfaceAddrs()
	if errAddr != nil {
		log.Error("Error in getting own Ip")
		os.Exit(2)
	}

	ipAddr := convertNetAddrToIp(log, addr)

	for _, x := range interfaceAddrs {
		aConvert := convertNetAddrToIp(log, x.String())
		if aConvert == ipAddr {
			return true
		}

	}
	// aConvert := convertNetAddrToIp(log, interfaceAddrs[1].String())
	// if aConvert == ipAddr {
	// 	return true
	// }
	return false
}

func convertNetAddrToIp(log slog.Logger, addr string) string {
	addrString := addr
	i := strings.LastIndex(addrString, "/")
	if i != -1 {
		addrString = addrString[0:i]
	}
	i2 := strings.LastIndex(addrString, ":")
	if i2 != -1 {
		addrString = addrString[0:i2]
	}
	if i == -1 || i2 == -1 {
		// log.Error("Conv", "i1", i, "i2", i2, "Pre", addr, "Post", addrString)
	}
	// log.Debug("Conv", "Pre", addr, "Post", addrString)
	return addrString
}
