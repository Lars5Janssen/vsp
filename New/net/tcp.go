package net

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/gin-gonic/gin"
)

type RestIn struct {
	EndpointAddr string
	Context      *gin.Context
}
type RestOut struct {
	StatusCode int
	Body       any
}
type Handler func(RestIn) RestOut

type Method string

// TODO might need export
const (
	GET    = "GET"
	PUT    = "PUT"
	DELETE = "DELETE"
)

type Endpoint struct {
	Name            []string
	AcceptedMethods map[Method]Handler
}

func stringToMethod(s string) Method {
	if s == "GET" {
		return GET
	} else if s == "PUT" {
		return PUT
	} else if s == "DELETE" {
		return DELETE
	} else {
		return ""
	}
}

func AttendHTTP(
	log *slog.Logger,
	reciveChannel chan RestIn,
	sendChannel chan RestOut,
	handlers []Endpoint,
) {
	recived := <-reciveChannel
	var handler Handler
	foundHandler := false
	for _, v := range handlers {
		i := slices.Index(v.Name, recived.EndpointAddr)
		if i != -1 {
			foundHandler = true
			h, exists := v.AcceptedMethods[stringToMethod(recived.Context.Request.Method)]
			if exists {
				handler = h
			}
		}
	}
	if !foundHandler {
		log.Error("Did not find Handler")
		return
	}
	sendChannel <- handler(recived)
}

func StartTCPServer(
	log *slog.Logger,
	port int,
	endpoints []Endpoint,
	inputChannel chan RestIn,
	outputChannel chan RestOut,
) {
	log = log.With(slog.String("Component", "TCP"))

	router := gin.Default()
	for _, v := range endpoints {
		for _, vv := range v.Name {
			for k := range v.AcceptedMethods {
				f := func(c *gin.Context) {
					inputChannel <- RestIn{vv, c}
					o := <-outputChannel
					c.JSON(o.StatusCode, o.Body)
				}

				switch k {
				case GET:
					router.GET(vv, f)
				case PUT:
					router.PUT(vv, f)
				case DELETE:
					router.DELETE(vv, f)
				default:
					log.Error("Unrecognized Method")

				}
			}
		}
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port) // Nimmt localhost als IP
	log.Info("Starting TCP Server", slog.String("Address", addr))
	err := router.Run(addr)
	if err != nil {
		log.Error(err.Error())
	}
}
