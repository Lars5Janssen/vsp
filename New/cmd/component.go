package cmd

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	n "github.com/Lars5Janssen/vsp/net"
)

func notAvailable(_ n.RestIn) n.RestOut {
	return n.RestOut{http.StatusNotFound, nil}
}

/*
createOrForwardMessage nutzt das MessageRequestModel 2.1
*/
func createOrForwardMessage(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{http.StatusOK, body}
}

/*
Aufgabe 2.3 getListOfAllMessages
*/
func getListOfAllMessages(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{http.StatusOK, body}
}

/*
Aufgabe 2.3 getMessageByUUID
*/
func getMessageByUUID(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{http.StatusOK, body}
}

/*
Aufgabe 1.3 disconnectFromStar Pflege des Sterns – Abmelden von SOL

Wenn die Komponente, die gerade aktiv den Stern „managed“ (also SOL) den „EXIT“Befehl bekommt, werden von ihr alle
aktiven Komponenten im Stern einzeln kontaktiert:
*/
func disconnectFromStar(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{http.StatusOK, body}
}

/*
sendHeartBeatBackToSol 1.1 Pflege des Sterns – Kontrolle der Komponenten

Wenn SOL für eine aktive Komponente 60 Sekunden nach der letzten Meldung keine neue Meldung mit einem Status von „200“
erhält, baut SOL zum <STARPORT>/tcp der Komponente eine UNICAST-Verbindung auf und kontrolliert selbst, ob die
Komponente noch aktiv und funktionsfähig ist. Diese Kontrollmöglichkeit muss SOL auch für sich selbst unterstützen!
Auch hier kommt eine REST-API zum Einsatz.
*/
func sendHeartBeatBackToSol(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{http.StatusOK, body}
}

/*
2.2: Weiterleiten von DELETE Requests von Komponente an Sol
*/
func forwardDeletingMessages(response n.RestIn) n.RestOut {
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
	// TODO Hier scheint eine Loop logic sein zu müssen damit die Ports available bleiben
	n.AttendHTTP(log, restIn, restOut, endpoints) // Will Handle endpoints in this thread
}
