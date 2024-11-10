package net

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
)

func StartServer(logger *slog.Logger, port int, channel chan *gin.Context) {
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
