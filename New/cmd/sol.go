package cmd

import (
	"context"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"

	n "github.com/Lars5Janssen/vsp/net"
)

func StartSol(
	ctx context.Context,
	log *slog.Logger,
	commands chan string,
	udp chan string,
	restIn chan n.RestIn,
	restOut chan n.RestOut,
) {
	log = log.With(slog.String("Component", "SOL"))
	log.Info("Starting as SOL")
	// SOL Logic
	// Retrieve from channels:
	// command := <-commands
	// udpInput := <-udp

}

/*
createComponent - Seite 4 im Aufgabenblatt Aufgabe 1.0

Bitte das RegisterRequestModel nutzen.

Damit SOL den auch kennt, gibt die neue
Komponente selbst ihren derzeitigen Status an. Außerdem werden Angaben zum Stern
übermittelt, die von SOL vor der Integration auch geprüft werden, damit sichergestellt
ist, dass auch der „richtige“ Stern gemeint ist.
*/
func registerComponentBySol(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{http.StatusOK, body}
}

/*
Das soll die checkAvailabilityFromComponent Methode übernehmen - Aufgabe 1.1 Pflege des Sterns

Jede aktive Komponente baut alle 30 Sekunden eine UNICAST-Verbindung zum
<STARPORT>/tcp von SOL auf.2 Wenn SOL nicht erreichbar ist, wird es nach 10 bzw. 20
Sekunden nochmal versucht. Wenn dann immer noch keine Verbindung zustande
kommt, beendet sich die Komponente selbst.
*/
func checkAvailabilityFromComponent(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{http.StatusOK, body}
}

/*
1.2 Pflege des Sterns – Abmelden einer Komponente

Eine aktive Komponente, die sich nach einem „EXIT“-Befehl bei SOL abmeldet, baut eine UNICAST-Verbindung auf.
Wenn SOL nicht erreichbar ist, wird es nach 10 bzw. 20 Sekunden nochmal versucht. Wenn dann immer noch keine Verbindung
zustande kommt, beendet sich die Komponente selbst.
*/
func disconnectComponentFromStar(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{http.StatusOK, body}
}

/*
createAndSaveMessage Aufgabe 2.1
*/
func createAndSaveMessage(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{http.StatusOK, body}
}

/*
deleteMessage Aufgabe 2.2
*/
func deleteMessage(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{http.StatusOK, body}
}
