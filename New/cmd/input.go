package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Starts the evaluation of User Input
func StartUserInput(
	log *slog.Logger,
	// channelToMain chan bool,
	channelToWorker chan string,
	crashFunc context.CancelFunc,
	crashUdp context.CancelFunc,
) {
	log = log.With(slog.String("Component", "UserInput"))

	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Get User Input
		fmt.Println("User Input: ")
		scanner.Scan()
		err := scanner.Err()
		if err != nil {
			log.Error(err.Error())
			continue
		}
		input := strings.ToLower(scanner.Text())

		// Evaluate User Input
		if input == "crash" {
			log.Info("Received crash Command")
			// channelToMain <- true // Maybe unnecessary, b/c of ctx/CancelFunc/crashFunc
			crashFunc()
			crashUdp() // TODO also needs to crash?
			return
		} else if input == "exit" {
			log.Info("Received exit Command")
			channelToWorker <- input
			return
		} else {
			fmt.Println("Command not recognized")
		}
	}
}
