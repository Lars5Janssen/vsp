package sol

import (
	n "github.com/Lars5Janssen/vsp/net"
)

func GetSolEndpoints() []n.Endpoint {
	return solEndpoints
}

var solEndpoints = []n.Endpoint{
	{
		Name: []string{"/vs/v1/system"}, // component an sol
		AcceptedMethods: map[n.Method]n.Handler{
			n.POST: registerComponentBySol,
		},
	},
	{
		Name: []string{"/vs/v1/system/:comUUID"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.PATCH: checkAvailabilityFromComponent, // siehe methodenkommentar
		},
	},
	{
		Name: []string{"/vs/v1/system/:comUUID?star=starUUID"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET: sendHeartBeatBack,
			// Endpunkt fuer Komponente meldet sich von Stern ab
			n.DELETE: disconnectComponentFromStar, // Exit befehl von einer Komponente
		},
	},
	{
		Name: []string{"/vs/v1/messages"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.POST: createAndSaveMessage,
		},
	},
	{
		Name: []string{"/vs/v1/messages/:msgUUID?star=starUUID"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET:    getMessageByUUID,
			n.DELETE: deleteMessage,
		},
	},
	{
		Name: []string{"/vs/v1/messages?star=starUUID&scope=scope&view=view"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET: getListOfAllMessages,
		},
	},
}
