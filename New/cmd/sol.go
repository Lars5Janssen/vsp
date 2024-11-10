package cmd

import (
	"context"
	"log/slog"
	"sync"
)

func StartSol(ctx context.Context, wg *sync.WaitGroup, log *slog.Logger) {
	wg.Add(1)
	defer wg.Done()
	log = log.With(slog.String("Component", "SOL"))
}
