package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

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
	port := flag.Int("port", 8006, "Port to run the server on")                     // -port=8006
	rerun := flag.Bool("rerun", false, "Enable this flag to automatically restart") // -rerun
	flag.Parse()

	// Logger
	// It may be better for every component to modify this logger for themselfs
	//group := slog.Group("UUID", utils.getUUID())
	log := slog.Default() //.With(group)

	log.Info(
		"Start of program",
		slog.String("Component", "Main"),
		slog.Int("Port", *port),
		slog.Bool("ReRun?", *rerun),
	)

	// Channels, Contexts & WaitGroup (Thread Stuff)
	// Channels:
	InputWorker := make(chan string) // Input -> Worker
	udpMainSol := make(chan string)  // UDP -> SOL/Main
	restIn := make(chan net.RestIn)
	restOut := make(chan net.RestOut)

	// Contexts:
	udpCTX, udpCancel := context.WithCancel(context.Background())
	workerCTX, workerCancel := context.WithCancel(context.Background())

	go net.StartTCPServer(log, *port, cmd.GetEndpoints(), restIn, restOut)
	go cmd.StartUserInput(log, InputWorker, workerCancel)

	firstRun := true
	for *rerun || firstRun {
		firstRun = false
		go net.ListenForBroadcastMessage(udpCTX, log, *port, udpMainSol)
		net.SendHello(log, *port)
		response := <-udpMainSol // blocking (on both ends)
		if response == "" {      // "" might be a bad idea, as this may be sent by someone, so someone could force us to be sol
			go cmd.StartSol(workerCTX, log, InputWorker, udpMainSol, restIn, restOut)
		} else {
			udpCancel()
			go cmd.StartComponent(workerCTX, log, InputWorker, restIn, restOut)
		}
		time.Sleep(1 * time.Hour)
	}
	log.Info("Exiting")
	os.Exit(0)
}
