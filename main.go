package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/Lars5Janssen/vsp/cmd"
	"github.com/Lars5Janssen/vsp/net"
)

// TODO better logging currently all is manually set = bad (component string in every file but main.go)
// TODO Buffered Channels for commands
// TODO test relation between "ctx.WithCancel" and "defer wg.Done()". Does a cancellation still execute the wg.Done() function?
// TODO Maybe make the TCP channel a map (Endpoint -> gin.Context)
// TODO Better words to differentiate between components in the program and component as a thing in the networkstructure
func main() {
	// Parse command-line arguments
	port := flag.Int("port", 8006, "Port to run the server on")                     // -port=8006
	rerun := flag.Bool("rerun", false, "Enable this flag to automatically restart") // -rerun
	flag.Parse()

	// Logger
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelDebug)
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	}))

	log.Info(
		"Start of program",
		slog.String("Component", "Main"),
		slog.Int("Port", *port),
		slog.Bool("ReRun?", *rerun),
	)

	// Channels:
	InputWorker := make(chan string) // Input -> Worker
	udpMainSol := make(chan string)  // UDP -> SOL/Main
	restIn := make(chan net.RestIn)
	restOut := make(chan net.RestOut)

	// Contexts:
	udpCTX, udpCancel := context.WithCancel(context.Background())
	workerCTX, workerCancel := context.WithCancel(context.Background())

	go cmd.StartUserInput(log, InputWorker, workerCancel)

	firstRun := true
	for *rerun || firstRun { // TODO hier würde ich das an Crash oder Exit anpassen
		firstRun = false
		go net.ListenForBroadcastMessage(udpCTX, log, *port, udpMainSol)
		/*err := net.SendHello(log, *port) // TODO SendHello muss 2 x stattfinden (alle 20 sekunden) und anschließend werden wir soll sofern wir eine Antwort erhalten haben
		if err != nil {
			return
		}*/
		response := <-udpMainSol // blocking (on both ends)
		log.Info("Received response", slog.String("Response", response))
		if len(strings.TrimLeft(response, "\n")) == 1 {
			go net.StartTCPServer(log, *port, cmd.GetSolEndpoints(), restIn, restOut)
			go cmd.StartSol(workerCTX, log, InputWorker, udpMainSol, restIn, restOut)
		} else {
			udpCancel()
			go net.StartTCPServer(log, *port, cmd.GetComponentEndpoints(), restIn, restOut)
			go cmd.StartComponent(workerCTX, log, InputWorker, restIn, restOut, response)
		}
		time.Sleep(1 * time.Hour)
	}
	log.Info("Exiting")
	os.Exit(0)
}
