package net

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/gin-gonic/gin"
)

func StartServer(wg *sync.WaitGroup, logger *slog.Logger, port int, channel chan *gin.Context) {
	wg.Add(1)
	defer wg.Done()
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		channel <- c
		//c.JSON(200, gin.H{
		//	"message": "Hello, World!",
		//})
	})
	// TODO evtl. nicht ip hardcoden
	addr := fmt.Sprintf("127.0.0.1:%d", port) // Nimmt localhost als IP
	err := router.Run(addr)
	if err != nil {
		logger.Error(err.Error())
	}
}
