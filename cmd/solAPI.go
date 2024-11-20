package cmd

import n "github.com/Lars5Janssen/vsp/net"

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
		Name: []string{"/vs/v1/system/<comUUID>"}, // TODO :comUUID muss definiert werden
		AcceptedMethods: map[n.Method]n.Handler{
			n.PATCH: checkAvailabilityFromComponent, // TODO siehe methodenkommentar
		},
	},
	{
		Name: []string{"/vs/v1/system/<comUUID>?star=<starUUID>"}, // TODO :starUUID muss definiert werden
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET: notAvailable, // TODO 1.2 GET funktion
			// Endpunkt fuer Komponente meldet sich von Stern ab
			n.DELETE: disconnectComponentFromStar, // Exit befehl von einer Komponente
		},
	},
	{
		Name: []string{"/vs/v1/messages/<msgUUID>"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.POST: createAndSaveMessage,
		},
	},
	{
		Name: []string{"/vs/v1/messages/<msgUUID>?star=<starUUID>"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET:    getMessageByUUID,
			n.DELETE: deleteMessage,
		},
	},
	{
		Name: []string{"/vs/v1/messages/<msgUUID>?star=<starUUID>&scope=<scope>&info=<info>"},
		AcceptedMethods: map[n.Method]n.Handler{
			n.GET: getListOfAllMessages,
		},
	},
}
