package main

import (
	"flag"
	"github.com/Lars5Janssen/vsp/net"
	"github.com/gin-gonic/gin"
	"log/slog"
)

// Gin, slog initialize
// channels fuer thread-com init
func main() {
	// Parse command-line arguments
	port := flag.Int("port", 8006, "Port to run the server on")
	flag.Parse()
	//group := slog.Group("UUID", utils.getUUID())
	logger := slog.Default() //.With(group)
	// Channels
	inputMain := make(chan bool)
	InputWorker := make(chan string)
	tcpWorker := make(chan *gin.Context)
	udpMainSol := make(chan string)

	test := <-tcpWorker
	go net.StartServer(logger, *port, tcpWorker)
	// logik fuer sol oder com
	// empfaengt client befehle (crash und exit) aus anderem thread
	// alles im loop

	// Thread Modell
	// 0: Main()
	// 1: Sol ODER Com (Worker)
	// 1.5: UDP-Server (Nur fuer Sol)
	// 2: TCP-Server
	// 3: Input
	// GET CLI INPUT
	// IF "CRASH"
	// ZU MAIN ein TRUE
	// ELSE IF "EXIT"
	// ZUM WORKER THREAD

	// Channels:
	// Input -> Main (wegen Loop)
	// Input -> Worker
	// TCP -> Worker
	// UDP -> SOL/Main

	// LOOP?
	// START UDP WENN NICHT AN
	// SEND HELLO
	// WARTEN
	// IF HABEN ANTWORT
	// 		go comp.start(antwort, ...) // innerhalb von comp start von tcp
	//		udp.close
	// IF NOT
	// 		go sol.start(udpchannel)
}
