package component

import (
	n "github.com/Lars5Janssen/vsp/net"
)

func GetComponentEndpoints() []n.Endpoint {
	return endpoints
}

var endpoints = []n.Endpoint{
	{
		Name: []string{"/vs/v1/system"}, // component an sol
		AcceptedMethods: map[n.Method]n.Handler{
			n.POST: iAmNotSol, // unauthorized
		},
	},
	{
		Name: []string{"/vs/v1/system/:comUUID"}, // TODO :comUUID muss definiert werden
		AcceptedMethods: map[n.Method]n.Handler{
			n.PATCH: iAmNotSol, // unauthorized
		},
	},
	{
		Name: []string{"/vs/v1/system/:comUUID?star=starUUID"}, // TODO :starUUID muss definiert werden
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET: sendHeartBeatBackToSol,
			// Endpunkt fuer Sol meldet sich von Stern ab
			n.DELETE: iAmNotSol, // unauthorized
		},
	},
	{
		Name: []string{"/vs/v1/messages/"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.POST: createOrForwardMessage, // komponenten können messages erstellen und auch an sol weiterleiten
		},
	},
	{
		Name: []string{"/vs/v1/messages/:msgUUID?star=starUUID"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.DELETE: forwardDeletingMessages, // auch für SolAPI
			n.GET:    notAvailable,
		},
	},
	{
		Name: []string{"/vs/v1/messages?star=starUUID&scope=scope&view=view"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET: notAvailable,
		},
	},
}
