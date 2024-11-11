package cmd

import (
	"context"
	"log/slog"
	"sync"

	"github.com/gin-gonic/gin"
)

func StartSol(
	ctx context.Context,
	wg *sync.WaitGroup,
	log *slog.Logger,
	commands chan string,
	tcp chan *gin.Context,
	udp chan string,
) {
	wg.Add(1)
	defer wg.Done()
	log = log.With(slog.String("Component", "SOL"))
	log.Info("Starting as SOL")
	// SOL Logic
	// Retrieve from channels:
	// command := <-commands
	// udpInput := <-udp

}
