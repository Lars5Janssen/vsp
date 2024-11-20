package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
)

// Starts the evaluation of User Input
func StartUserInput(
	log *slog.Logger,
	channelToWorker chan string,
	crashFunc context.CancelFunc,
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
		input := scanner.Text()

		// Evaluate User Input
		if input == "CRASH" {
			log.Info("Received Crash Command")
			crashFunc()
			// os.Exit(-4)
			return
		} else if input == "EXIT" {
			log.Info("Received EXIT Command")
			channelToWorker <- input
			return
		} else {
			fmt.Println("Command not recognized")
		}
	}
}
