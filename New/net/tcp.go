package net

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/gin-gonic/gin"
)

func StartTCPServer(
	wg *sync.WaitGroup,
	log *slog.Logger,
	port int,
	inputChannel chan *gin.Context,
	outputChannel chan *gin.Context,
) {
	wg.Add(1)
	defer wg.Done()
	log = log.With(slog.String("Component", "TCP"))

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		inputChannel <- c   // Blocking
		c = <-outputChannel // Blocking
		// but its just a reference so may be unnecceary
		//c.JSON(200, gin.H{
		//	"message": "Hello, World!",
		//})
	})

	// TODO evtl. nicht ip hardcoden
	addr := fmt.Sprintf("127.0.0.1:%d", port) // Nimmt localhost als IP
	log.Info("Starting TCP Server", slog.String("Address", addr))
	err := router.Run(addr)
	if err != nil {
		log.Error(err.Error())
	}
}
