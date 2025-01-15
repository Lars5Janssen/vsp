package sol

import (
	con "github.com/Lars5Janssen/vsp/connection"
)

func GetSolEndpoints() []con.Endpoint {
	return solEndpoints
}

var solEndpoints = []con.Endpoint{
	{
		Name: []string{"/vs/v2/star"}, // component an sol
		AcceptedMethods: map[con.Method]con.Handler{
			con.POST: registerStarByGalaxy, // 3.1
			con.GET:  getListOfAllStars,
		},
	},
	{
		Name: []string{"/vs/v2/star/:starUUID"}, // component an sol
		AcceptedMethods: map[con.Method]con.Handler{
			con.GET:    getStarInGalaxy,  //
			con.DELETE: disconnectStar,   // disconnects the star from all galaxy's
			con.PATCH:  updateStarStatus, // 3.1 patch
		},
	},
	{
		Name: []string{"/vs/v2/system"}, // component an sol
		AcceptedMethods: map[con.Method]con.Handler{
			con.POST: registerComponentBySol,
		},
	},
	{
		// ?star=starUUID
		// die query's müssen herausgefiltert werden.
		Name: []string{"/vs/v2/system/:comUUID"},
		AcceptedMethods: map[con.Method]con.Handler{
			con.GET:    sendHeartBeatBack,
			con.PATCH:  checkAvailabilityFromComponent, // siehe methodenkommentar
			con.DELETE: disconnectComponentFromStar,
		},
	},
	{
		// ?star=starUUID&scope=scope&view=view
		// sind query's die der Client mitschickt welche dann herausgefiltert werden
		Name: []string{"/vs/v2/messages"},
		AcceptedMethods: map[con.Method]con.Handler{
			con.POST: createAndSaveMessage,
			con.GET:  getListOfAllMessages,
		},
	},
	{
		// ?star=starUUID
		// sind query's die der Client mitschickt welche dann herausgefiltert werden
		Name: []string{"/vs/v2/messages/:msgUUID"},
		AcceptedMethods: map[con.Method]con.Handler{
			con.GET:    getMessageByUUID,
			con.DELETE: deleteMessage,
			con.POST:   forwardMessageToStar, // 4.1
		},
	},
}
