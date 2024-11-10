package cmd

import (
	"context"
	"log/slog"
	"sync"

	"github.com/gin-gonic/gin"
)

func StartComponent(
	ctx context.Context,
	wg *sync.WaitGroup,
	log *slog.Logger,
	commands chan string,
	tcp chan *gin.Context,
) {
	wg.Add(1)
	defer wg.Done()
	log = log.With(slog.String("Component", "Component"))
	log.Info("Starting as Component")
	// Component Logic
	// Retrive from channel:
	// command := <-commands
	// tcpRequest := <-tcp
}
