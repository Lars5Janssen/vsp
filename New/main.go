package main

import (
	"context"
	"flag"
	"log/slog"
	"sync"

	"github.com/gin-gonic/gin"

	"github.com/Lars5Janssen/vsp/cmd"
	"github.com/Lars5Janssen/vsp/net"
)

// Gin, slog initialize
// channels fuer thread-com init
func main() {
	// Parse command-line arguments
	port := flag.Int("port", 8006, "Port to run the server on")
	rerun := *flag.Bool("rerun", false, "Enable this flag to automatically restart")
	flag.Parse()

	// Logger
	//group := slog.Group("UUID", utils.getUUID())
	logger := slog.Default() //.With(group)

	// Channels, Contexts & WaitGroup (Thread Stuff)
	// Channels:
	// Input -> Main (wegen Loop) -> change to ctx
	// Input -> Worker
	// TCP -> Worker
	// UDP -> SOL/Main
	inputMain := make(chan bool)
	InputWorker := make(chan string)
	tcpWorker := make(chan *gin.Context) // Maybe make this a map (Endpoint -> gin.Context)
	udpMainSol := make(chan string)
	testCTX := context.Background()
	ctxWC, cancel := context.WithCancel(testCTX) // To cancel... Needs testing
	wg := new(sync.WaitGroup)

	// LOOP?
	// START UDP WENN NICHT AN
	// SEND HELLO
	// WARTEN
	// IF HABEN ANTWORT
	// 		go comp.start(antwort, ...) // innerhalb von comp start von tcp
	//		udp.close
	// IF NOT
	// 		go sol.start(udpchannel)
	//test := <-tcpWorker
	go net.StartServer(wg, logger, *port, tcpWorker)
	go cmd.StartInput(wg, logger, inputMain, InputWorker)

	// Loop Start
	for rerun {
		go net.ListenForBroadcastMessage(wg, ctxWC, logger, *port, udpMainSol)
		net.SendHello(logger, *port)
		response := <-udpMainSol
		if response == "" {
			cmd.StartSol(wg, logger)
		} else {
			cancel()
			cmd.StartComponent(wg, logger)
		}
	}

	// Loop End

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

}
