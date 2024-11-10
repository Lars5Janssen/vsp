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

// TODO better logging currently all is manually set = bad (component string in every file but main.go)
func main() {
	// Parse command-line arguments
	port := flag.Int("port", 8006, "Port to run the server on")
	rerun := *flag.Bool("rerun", false, "Enable this flag to automatically restart")
	flag.Parse()

	// Logger
	// It may be better for every component to modify this logger for themselfs
	//group := slog.Group("UUID", utils.getUUID())
	log := slog.Default() //.With(group)

	// Channels, Contexts & WaitGroup (Thread Stuff)
	// Channels:
	//inputMain := make(chan bool)         // Input -> Main (wegen Loop) -> change to ctx
	InputWorker := make(chan string)     // Input -> Worker
	udpMainSol := make(chan string)      // UDP -> SOL/Main
	tcpWorker := make(chan *gin.Context) // TCP -> Worker.
	//                                      Maybe make the TCP channel a map (Endpoint -> gin.Context)
	// Contexts:
	udpCTX, udpCancel := context.WithCancel(context.Background())
	workerCTX, workerCancel := context.WithCancel(context.Background())
	// Wait Group:
	wg := new(sync.WaitGroup)
	//test := <-tcpWorker

	go net.StartServer(wg, log, *port, tcpWorker)
	go cmd.StartInput(wg, log, InputWorker, workerCancel)

	for rerun {
		go net.ListenForBroadcastMessage(udpCTX, wg, log, *port, udpMainSol)
		net.SendHello(log, *port)
		response := <-udpMainSol // blocking (on both ends)
		if response == "" {      // "" might be a bad idea, as this may be sent by someone, so someone could force us to be sol
			go cmd.StartSol(workerCTX, wg, log)
		} else {
			udpCancel()
			go cmd.StartComponent(workerCTX, wg, log)
		}
		wg.Wait()
		workerCancel() // TODO This is currently redundant
	}
}
