package net

import (
	"fmt"
	"log/slog"
	"slices"

	"github.com/gin-gonic/gin"
)

type RestIn struct {
	EndpointAddr string
	IpAndPort    string
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
	PATCH  = "PATCH"
	POST   = "POST"
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
	} else if s == "PATCH" {
		return PATCH
	} else if s == "POST" {
		return POST
	} else {
		return ""
	}
}

func AttendHTTP(
	log *slog.Logger,
	receiveChannel chan RestIn,
	sendChannel chan RestOut,
	handlers []Endpoint,
) {
	for {
		received := <-receiveChannel
		var handler Handler
		foundHandler := false
		for _, v := range handlers {
			i := slices.Index(v.Name, received.EndpointAddr)
			if i != -1 {
				foundHandler = true
				h, exists := v.AcceptedMethods[stringToMethod(received.Context.Request.Method)]
				if exists {
					handler = h
				}
			}
		}
		if !foundHandler {
			log.Error("Did not find Handler")
			return
		}
		sendChannel <- handler(received)
	}
}

func StartTCPServer(
	log *slog.Logger,
	ip string,
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
					ipAndPort := c.Request.RemoteAddr
					inputChannel <- RestIn{vv, ipAndPort, c}
					o := <-outputChannel
					// TODO use plaintext for some requests
					c.JSON(o.StatusCode, o.Body)
				}

				switch k {
				case GET:
					router.GET(vv, f)
				case PUT:
					router.PUT(vv, f)
				case DELETE:
					router.DELETE(vv, f)
				case PATCH:
					router.PATCH(vv, f)
				case POST:
					router.POST(vv, f)
				default:
					log.Error("Unrecognized Method")

				}
			}
		}
	}

	addr := fmt.Sprintf("%s:%d", ip, port)
	log.Info("Starting TCP Server", slog.String("Address", addr))
	err := router.Run(addr)
	if err != nil {
		log.Error(err.Error())
	}
}
