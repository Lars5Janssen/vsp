package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
	"vsp/api"
	"vsp/utils"
)

func main() {
	// Parse command-line arguments
	port := flag.Int("port", 8006, "Port to run the server on")
	flag.Parse()

	// Get the local IP address
	ip, err := getLocalIP()
	if err != nil {
		fmt.Printf("Error getting local IP address: %v\n", err)
		os.Exit(1)
	}

	go checkUserInput()

	go func() {
		err := utils.ListenForBroadcastMessage(*port)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}()

	time.Sleep(2 * time.Second) // Wait for 2 seconds

	err = utils.SendBroadcastMessage(*port)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	time.Sleep(1 * time.Second) // Wait for 1 second
	api.StartServer(ip, *port)

	os.Exit(0)
}

func checkUserInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		if text == "crash\n" {
			os.Exit(-1)
		}
		// TODO "exit" command
	}
}

func getLocalIP() (net.IP, error) {
	conn, err := net.Dial("net", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(conn)

	localAddress := conn.LocalAddr().(*net.UDPAddr)

	return localAddress.IP, err
}
