package cmd

import (
	"log/slog"
	"sync"
)

func StartInput(
	wg *sync.WaitGroup,
	logger *slog.Logger,
	channelToMain chan bool,
	channelToWorker chan string,
) {
	wg.Add(1)
	defer wg.Done()
}
