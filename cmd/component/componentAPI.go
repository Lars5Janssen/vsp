package component

import (
	con "github.com/Lars5Janssen/vsp/connection"
)

func GetComponentEndpoints() []con.Endpoint {
	return endpoints
}

var endpoints = []con.Endpoint{
	{
		Name: []string{"/vs/v1/system"}, // component an sol
		AcceptedMethods: map[con.Method]con.Handler{
			con.POST: iAmNotSol, // unauthorized
		},
	},
	{
		// ?star=starUUID
		// die query's müssen herausgefiltert werden.
		Name: []string{"/vs/v1/system/:comUUID"}, // TODO :comUUID muss definiert werden
		AcceptedMethods: map[con.Method]con.Handler{
			con.PATCH:  iAmNotSol, // unauthorized
			con.GET:    sendHeartBeatBackToSol,
			con.DELETE: disconnectFromStar, // unauthorized, when star is not right
		},
	},
	{
		// ?star=starUUID&scope=scope&view=view
		// sind query's die der Client mitschickt welche dann herausgefiltert werden
		Name: []string{"/vs/v1/messages"},
		AcceptedMethods: map[con.Method]con.Handler{
			con.GET:  getListOfAllMessages,
			con.POST: forwardMessage, // komponenten können messages erstellen und auch an sol weiterleiten
		},
	},
	{
		// ?star=starUUID
		// sind query's die der Client mitschickt welche dann herausgefiltert werden
		Name: []string{"/vs/v1/messages/:msgUUID"},
		AcceptedMethods: map[con.Method]con.Handler{
			con.DELETE: forwardDeletingMessages, // auch für SolAPI
			con.GET:    getMessageByUUID,
		},
	},
}
