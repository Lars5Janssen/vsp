package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	n "net"
	"os"
	"os/exec"
	"sync"
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
	ip := "127.0.0.1" // nimmt localhost als IP-Adresse

	// Parse command-line arguments
	port := flag.Int("port", 8006, "Port to run the server on")                     // -port=8006
	rerun := flag.Bool("rerun", false, "Enable this flag to automatically restart") // -rerun
	sleep := flag.Bool("sleep", false, "Enable this flag to sleep once at start")
	stopIfSol := flag.Bool("killSol", false, "Stop if the process would be sol")
	maxActiveComponents := flag.Int("maxActiveComponents", 4,
		"Maximum number of active components") // -maxActiveComponents=4
	flag.Parse()

	// Logger
	lvl := new(slog.LevelVar)
	lvl.Set(slog.LevelDebug)
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: lvl,
	}))

	_, err := exec.Command("ip", "addr").Output()
	if err != nil {
		log.Error(err.Error())
	}
	// fmt.Println(string(cmdOut))
	adLs, _ := n.InterfaceAddrs()
	fmt.Println(adLs[1])

	log.Info(
		"Start of program",
		slog.String("Component", "Main"),
		slog.Int("Port", *port),
		slog.Bool("ReRun?", *rerun),
		slog.Bool("Sleep?", *sleep),
		slog.Bool("killSol?", *stopIfSol),
		slog.Int("MaxActiveComponents", *maxActiveComponents),
	)

	// Channels, Contexts & WaitGroup (Thread Stuff)
	// Channels:
	inputWorker := make(chan string)    // Input -> Worker
	udpMainSol := make(chan net.UDP, 1) // UDP -> SOL/Main
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

	if *sleep {
		sleepTime := 5
		log.Info(fmt.Sprintf("Sleep flag was set. Waiting %v Seconds", sleepTime))
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
	firstRun := true
	for *rerun || firstRun {
		firstRun = false

		go net.ListenForBroadcastMessage(log, *port, udpMainSol) // udpCTX?

		var response net.UDP
		noMessage := true

		// TODO Timeout verstellbar machen
		for i := 0; i < 3; i++ {
			if !noMessage {
				continue
			}
			err := net.SendHello(log, *port)
			if err != nil {
				log.Error("Could not Send Hello")
				return
			}

			time.Sleep(1 * time.Second)

			if len(udpMainSol) == cap(udpMainSol) {
				noMessage = false
			} else {
				log.Debug("No UDP message recived, timing out")
			}
		}

		if noMessage && !*stopIfSol {
			log.Info("Starting as Sol")
			wg.Add(1)
			go net.StartTCPServer(log, ip, *port, cmd.GetSolEndpoints(), restIn, restOut)
			go func() {
				defer wg.Done()
				cmd.StartSol(workerCTX, log, inputWorker, udpMainSol, restIn, restOut)
			}()
		} else if noMessage && *stopIfSol {
			log.Info("Would be sol, but flag is set, stopping")
		} else {
			log.Info("Starting as Component")
			udpCancel()
			wg.Add(1)
			go net.StartTCPServer(log, ip, *port, cmd.GetComponentEndpoints(), restIn, restOut)
			go func() {
				defer wg.Done()
				cmd.StartComponent(workerCTX, log, inputWorker, restIn, restOut, response.Message)
			}()
		}
		wg.Wait()
	}

	log.Info("Exiting")
	os.Exit(0)
}
