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
		// ?star=starUUID
		// die query's müssen herausgefiltert werden.
		Name: []string{"/vs/v1/system/:comUUID"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET:    sendHeartBeatBack,
			n.PATCH:  checkAvailabilityFromComponent, // siehe methodenkommentar
			n.DELETE: disconnectComponentFromStar,
		},
	},
	/*{
		Name: []string{"/vs/v1/system/:comUUID?star=starUUID"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET: sendHeartBeatBack,
			// Endpunkt fuer Komponente meldet sich von Stern ab
			n.DELETE: disconnectComponentFromStar, // Exit befehl von einer Komponente
		},
	},*/
	{
		// ?star=starUUID&scope=scope&view=view
		// sind query's die der Client mitschickt welche dann herausgefiltert werden
		Name: []string{"/vs/v1/messages"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.POST: createAndSaveMessage,
			n.GET:  getListOfAllMessages,
		},
	},
	{
		// ?star=starUUID
		// sind query's die der Client mitschickt welche dann herausgefiltert werden
		Name: []string{"/vs/v1/messages/:msgUUID"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET:    getMessageByUUID,
			n.DELETE: deleteMessage,
		},
	},
	/*{
		Name: []string{"/vs/v1/messages?star=starUUID&scope=scope&view=view"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET: getListOfAllMessages,
		},
	},*/
}
