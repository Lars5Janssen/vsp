package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"github.com/Lars5Janssen/vsp/cmd"
	"github.com/Lars5Janssen/vsp/cmd/component"
	"github.com/Lars5Janssen/vsp/cmd/sol"
	"github.com/Lars5Janssen/vsp/connection"
)

// TODO better logging currently all is manually set = bad (component string in every file but main.go)
// TODO Buffered Channels for commands
// TODO test relation between "ctx.WithCancel" and "defer wg.Done()". Does a cancellation still execute the wg.Done() function?
// TODO Maybe make the TCP channel a map (Endpoint -> gin.Context)
// TODO Better words to differentiate between components in the program and component as a thing in the networkstructure
func main() {
	ip := "127.0.0.1" // nimmt localhost als IP-Adresse

	// Parse command-line arguments
	portSolUdp := flag.Int("port", 8006, "Port to run the server on")                // -port=8006
	portComp := flag.Int("portComp", 8007, "Port to run the server on as component") // -portComp=8007
	rerun := flag.Bool("rerun", false, "Enable this flag to automatically restart")  // -rerun
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

	// fmt.Println(string(cmdOut))
	/*	adLs, _ := n.InterfaceAddrs()
		fmt.Println(adLs[1])
		ip = adLs[1].String()
		ip = strings.Split(ip, "/")[0]*/

	ip, err := getFirstIPv4Addr()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("First IPv4 Address:", ip)
	}

	/*	log.Info(
		"Start of program",
		slog.String("Component", "Main"),
		slog.Int("Port", *port),
		slog.Bool("ReRun?", *rerun),
		slog.Bool("Sleep?", *sleep),
		slog.Bool("killSol?", *stopIfSol),
		slog.Int("MaxActiveComponents", *maxActiveComponents),
	)*/

	// Channels, Contexts & WaitGroup (Thread Stuff)
	// Channels:
	inputWorker := make(chan string)           // Input -> Worker
	udpMainSol := make(chan connection.UDP, 1) // UDP -> SOL/Main
	restIn := make(chan connection.RestIn)
	restOut := make(chan connection.RestOut)
	var wg sync.WaitGroup

	// Contexts:
	_, udpCancel := context.WithCancel(context.Background())
	workerCTX, workerCancel := context.WithCancel(context.Background())

	/*	go connection.StartTCPServer(log, *port, cmd.GetComponentEndpoints(), restIn, restOut)*/
	workerCTX = context.WithValue(workerCTX, "ip", ip)
	workerCTX = context.WithValue(workerCTX, "port", *portSolUdp)
	workerCTX = context.WithValue(workerCTX, "maxActiveComponents", *maxActiveComponents)

	go cmd.StartUserInput(log, inputWorker, workerCancel, udpCancel)

	/* Nur um organisch zwei Docker Container auf sol und component zuzuteilen.
	Ist die Flag hierzu gesetzt tendiert der Container dazu Component zu werden. */
	if *sleep {
		sleepTime := 5
		log.Info(fmt.Sprintf("Sleep flag was set. Waiting %v Seconds", sleepTime))
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

	firstRun := true
	// TODO please refactor too much code in one loop
	for *rerun || firstRun {
		firstRun = false

		go connection.ListenForBroadcastMessage(log, *portSolUdp, udpMainSol) // udpCTX?

		var response connection.UDP
		noMessage := true

		// TODO Timeout verstellbar machen
		for i := 0; i < 3; i++ {
			if !noMessage {
				continue
			}
			err := connection.SendHello(log, *portSolUdp)
			if err != nil {
				log.Error("Could not Send Hello")
				return
			}

			time.Sleep(1 * time.Second)

			if len(udpMainSol) == cap(udpMainSol) {
				noMessage = false
				response = <-udpMainSol
			} else {
				log.Debug("No UDP message received, timing out")
			}
		}

		if noMessage && !*stopIfSol {
			log.Info("Starting as Sol")
			wg.Add(1)
			go connection.StartTCPServer(log, ip, *portSolUdp, sol.GetSolEndpoints(), restIn, restOut)
			go func() {
				defer wg.Done()
				sol.StartSol(workerCTX, log, inputWorker, udpMainSol, restIn, restOut)
			}()
		} else if noMessage && *stopIfSol {
			log.Info("Would be sol, but flag is set, stopping")
		} else {
			log.Info("Starting as Component")
			udpCancel()
			wg.Add(1)
			workerCTX = context.WithValue(workerCTX, "port", *portComp)
			go connection.StartTCPServer(log, ip, *portComp, component.GetComponentEndpoints(), restIn, restOut)
			go func() {
				defer wg.Done()
				component.StartComponent(workerCTX, log, inputWorker, restIn, restOut, response.Message)
			}()
		}
		wg.Wait()
	}

	log.Info("Exiting")
	os.Exit(0)
}

func getFirstIPv4Addr() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no IPv4 address found")
}
