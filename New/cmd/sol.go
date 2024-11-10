package cmd

import (
	"log/slog"
	"sync"
)

func StartSol(wg *sync.WaitGroup, logger *slog.Logger) {
	wg.Add(1)
	defer wg.Done()
}
