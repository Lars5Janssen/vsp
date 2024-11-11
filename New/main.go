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
// TODO Buffered Channels for commands
// TODO test relation between "ctx.WithCancel" and "defer wg.Done()". Does a cancelation still execute the wg.Done() function?
// TODO Maybe make the TCP channel a map (Endpoint -> gin.Context)
// TODO Better words to differentiate between components in the program and component as a thing in the networkstructure
func main() {
	// Parse command-line arguments
	port := flag.Int("port", 8006, "Port to run the server on")
	rerun := *flag.Bool("rerun", false, "Enable this flag to automatically restart")
	flag.Parse()

	// Logger
	// It may be better for every component to modify this logger for themselfs
	//group := slog.Group("UUID", utils.getUUID())
	log := slog.Default() //.With(group)

	log.Info(
		"Start of program",
		slog.String("Component", "Main"),
		slog.Int("Port", *port),
		slog.Bool("ReRun?", rerun),
	)

	// Channels, Contexts & WaitGroup (Thread Stuff)
	// Channels:
	InputWorker := make(chan string)     // Input -> Worker
	udpMainSol := make(chan string)      // UDP -> SOL/Main
	tcpWorker := make(chan *gin.Context) // TCP -> Worker.
	workerTCP := make(chan *gin.Context) // TCP -> Worker.
	//                                      TODO Maybe make the TCP channel a map (Endpoint -> gin.Context)
	//                                      TODO Make some of the channels buffered?
	// Contexts:
	udpCTX, udpCancel := context.WithCancel(context.Background())
	workerCTX, workerCancel := context.WithCancel(context.Background())
	// Wait Group:
	wg := new(sync.WaitGroup)

	go net.StartTCPServer(wg, log, *port, tcpWorker, workerTCP)
	go cmd.StartUserInput(wg, log, InputWorker, workerCancel)

	for rerun {
		go net.ListenForBroadcastMessage(udpCTX, wg, log, *port, udpMainSol)
		net.SendHello(log, *port)
		response := <-udpMainSol // blocking (on both ends)
		if response == "" {      // "" might be a bad idea, as this may be sent by someone, so someone could force us to be sol
			go cmd.StartSol(workerCTX, wg, log, InputWorker, tcpWorker, udpMainSol)
		} else {
			udpCancel()
			go cmd.StartComponent(workerCTX, wg, log, InputWorker, tcpWorker)
		}
		workerCTX.Done()
		wg.Wait()
	}
	log.Info("Exiting")
}
