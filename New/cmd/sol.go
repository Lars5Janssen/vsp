package cmd

import (
	"context"
	"log/slog"

	"github.com/Lars5Janssen/vsp/net"
)

func StartSol(
	ctx context.Context,
	log *slog.Logger,
	commands chan string,
	udp chan string,
	restIn chan net.RestIn,
	restOut chan net.RestOut,
) {
	log = log.With(slog.String("Component", "SOL"))
	log.Info("Starting as SOL")
	// SOL Logic
	// Retrieve from channels:
	// command := <-commands
	// udpInput := <-udp

}
