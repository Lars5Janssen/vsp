package cmd

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	n "github.com/Lars5Janssen/vsp/net"
)

func GetEndpoints() []n.Endpoint {
	return endpoints
}

var endpoints []n.Endpoint = []n.Endpoint{
	// {[]string{"/exampleEndpoint", "/exampleEndpoint/"}, map[n.Method]n.Handler{
	// 	n.GET:    test,
	// 	n.DELETE: test,
	// }},
	{
		Name: []string{"/exampleEndpoint2", "/exampleEndpoint2/"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET:    test,
			n.DELETE: testDELETE,
			n.PUT:    HAHAHA,
		},
	},
	{
		Name: []string{"/vs/v1/system/:comUUID"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET: getComponentByUUID,
		},
	},
}

func getComponentByUUID(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{http.StatusOK, body}
}

func test(r n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{http.StatusOK, body}
}
func testDELETE(r n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{http.StatusOK, body}
}

func StartComponent(
	ctx context.Context,
	log *slog.Logger,
	commands chan string,
	restIn chan n.RestIn,
	restOut chan n.RestOut,

) {
	log = log.With(slog.String("Component", "Component"))
	log.Info("Starting as Component")

	n.AttendHTTP(log, restIn, restOut, endpoints) // Will Handle endpoints in this thread

}
