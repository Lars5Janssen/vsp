package cmd

import n "github.com/Lars5Janssen/vsp/net"

func GetEndpoints() []n.Endpoint {
	return endpoints
}

var endpoints = []n.Endpoint{
	{
		Name: []string{"/vs/v1/system"}, // component an sol
		AcceptedMethods: map[n.Method]n.Handler{
			n.POST: notAvailable,
		},
	},
	{
		Name: []string{"/vs/v1/system/:comUUID"}, // TODO :comUUID muss definiert werden
		AcceptedMethods: map[n.Method]n.Handler{
			n.PATCH: notAvailable,
		},
	},
	{
		Name: []string{"/vs/v1/system/:comUUID?star=:starUUID"}, // TODO :starUUID muss definiert werden
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET: sendHeartBeatBackToSol,
			// Endpunkt fuer Sol meldet sich von Stern ab
			n.DELETE: disconnectFromStar,
		},
	},
	{
		Name: []string{"/vs/v1/messages/:msgUUID"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.POST: createOrForwardMessage, // komponenten können messages erstellen und auch an sol weiterleiten
		},
	},
	{
		Name: []string{"/vs/v1/messages/:msgUUID?star=:starUUID"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.DELETE: forwardDeletingMessages, // auch für SolAPI
			n.GET:    notAvailable,
		},
	},
	{
		Name: []string{"/vs/v1/messages/:msgUUID?star=:starUUID&scope=:scope&info=:info"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET: notAvailable,
		},
	},
}
