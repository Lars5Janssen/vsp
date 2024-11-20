package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"sync"

	"github.com/Lars5Janssen/vsp/cmd"
	"github.com/Lars5Janssen/vsp/net"
)

// TODO better logging currently all is manually set = bad (component string in every file but main.go)
// TODO Buffered Channels for commands
// TODO test relation between "ctx.WithCancel" and "defer wg.Done()". Does a cancellation still execute the wg.Done() function?
// TODO Maybe make the TCP channel a map (Endpoint -> gin.Context)
// TODO Better words to differentiate between components in the program and component as a thing in the networkstructure
func main() {
	ip := "127.0.0.1" // nimmt localhost als IP-Adresse

	// Parse command-line arguments
	port := flag.Int("port", 8006, "Port to run the server on")                     // -port=8006
	rerun := flag.Bool("rerun", false, "Enable this flag to automatically restart") // -rerun
	maxActiveComponents := flag.Int("maxActiveComponents", 4,
		"Maximum number of active components") // -maxActiveComponents=4
	flag.Parse()

	// Logger
	// It may be better for every component to modify this logger for themselves
	// group := slog.Group("UUID", utils.getUUID())
	log := slog.Default() // .With(group)

	log.Info(
		"Start of program",
		slog.String("Component", "Main"),
		slog.Int("Port", *port),
		slog.Bool("ReRun?", *rerun),
		slog.Int("MaxActiveComponents", *maxActiveComponents),
	)

	// Channels, Contexts & WaitGroup (Thread Stuff)
	// Channels:
	inputWorker := make(chan string) // Input -> Worker
	udpMainSol := make(chan string)  // UDP -> SOL/Main
	restIn := make(chan net.RestIn)
	restOut := make(chan net.RestOut)
	var wg sync.WaitGroup

	// Contexts:
	_, udpCancel := context.WithCancel(context.Background())
	workerCTX, workerCancel := context.WithCancel(context.Background())

	/*	go net.StartTCPServer(log, *port, cmd.GetComponentEndpoints(), restIn, restOut)*/
	workerCTX = context.WithValue(workerCTX, "ip", ip)
	workerCTX = context.WithValue(workerCTX, "port", *port)
	workerCTX = context.WithValue(workerCTX, "maxActiveComponents", *maxActiveComponents)

	go cmd.StartUserInput(log, inputWorker, workerCancel, udpCancel)

	firstRun := true
	for *rerun || firstRun {
		firstRun = false
		err := net.SendHello(log, *port)
		if err != nil {
			return
		}
		go net.ListenForBroadcastMessage(log, *port, udpMainSol) // udpCTX?
		// response := <-udpMainSol  // blocking (on both ends)
		response := ""      // TODO remove this line
		if response == "" { // "" might be a bad idea, as this may be sent by someone, so someone could force us to be sol
			log.Info("Start SolTCP")
			workerCancel()
			wg.Add(1)
			go net.StartTCPServer(log, ip, *port, cmd.GetSolEndpoints(), restIn, restOut)
			go func() {
				defer wg.Done()
				cmd.StartSol(workerCTX, log, inputWorker, udpMainSol, restIn, restOut)
			}()
		} else {
			log.Info("Start ComponentTCP")
			udpCancel()
			workerCancel()
			wg.Add(1)
			go net.StartTCPServer(log, ip, *port, cmd.GetComponentEndpoints(), restIn, restOut)
			go func() {
				defer wg.Done()
				cmd.StartComponent(workerCTX, log, inputWorker, restIn, restOut)
			}()
		}
		wg.Wait()
	}
	log.Info("Exiting")
	os.Exit(0)
}
