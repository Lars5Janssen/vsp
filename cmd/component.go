package cmd

import (
	"context"
	"encoding/json"
	n "github.com/Lars5Janssen/vsp/net"
	"github.com/Lars5Janssen/vsp/utils"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var component = Component{
	ComUUID:  0,
	ComIP:    "",
	SolUUID:  0,
	StarUUID: 0,
	SolIP:    "",
	SolPort:  0,
}

var urlPräfix = ""

func StartComponent(
	ctx context.Context,
	log *slog.Logger,
	commands chan string,
	restIn chan n.RestIn,
	restOut chan n.RestOut,
	response string,
) {
	log = log.With(slog.String("Component", "Component"))
	log.Info("Starting as Component")

	initializeComponent(log, ctx, response)

	// TODO Hier scheint eine Loop logic sein zu müssen damit die Ports available bleiben
	go n.AttendHTTP(log, restIn, restOut, endpoints) // Will Handle endpoints in this thread

	log.Info("Componenten values",
		slog.Int("ComUUID", component.ComUUID),
		slog.Int("SolUUID", component.SolUUID),
		slog.Int("StarUUID", component.StarUUID),
		slog.String("ComIP", component.ComIP),
		slog.Int("ComPort", component.ComPort),
		slog.String("SolIP", component.SolIP),
		slog.Int("SolPort", component.SolPort),
	)

	// Send Heartbeat to Sol
	ticker := time.NewTicker(30 * time.Second) // TODO check every 5 seconds or 1 second?
	go func() {
		for {
			select {
			case <-ticker.C:
				sendHeartBeatToSol()
			}
		}
	}()

	for {
		// Retrieve from user input
		select {
		case command := <-commands:
			if command == "exit" {
				log.Info("Exiting Component")
				return
			}
		default:
		}
	}
}

type Component struct {
	ComUUID  int    `json:"ComUUID"`
	ComIP    string `json:"ComIP"`
	SolUUID  int    `json:"SolUUID"`
	StarUUID int    `json:"StarUUID"`
	ComPort  int    `json:"ComPort"`
	SolIP    string `json:"SolIP"`
	SolPort  int    `json:"SolPort"`
}

func initializeComponent(log *slog.Logger, ctx context.Context, response string) {
	component.ComIP = ctx.Value("ip").(string)

	parseResponseIntoComponent(response, log)

	urlPräfix = "http://" + component.SolIP + ":" + strconv.Itoa(component.SolPort)

	registerByStar()
}

func parseResponseIntoComponent(response string, log *slog.Logger) {
	// Bereinigen des Strings, falls nötig (z. B. Ersetzen einzelner Anführungszeichen)
	cleanedInput := strings.ReplaceAll(response, "\\", "")

	// JSON-Daten unmarshallen
	var parsedData map[string]interface{}
	err := json.Unmarshal([]byte(cleanedInput), &parsedData)

	if err != nil {
		log.Error("Error while parsing response")
		return
	}

	// Daten in struct schreiben
	for key, value := range parsedData {
		switch strings.ToLower(key) {
		case "star":
			component.StarUUID = int(value.(float64))
		case "sol":
			component.SolUUID = int(value.(float64)) // Com UUID des stars?
		case "solip":
			component.SolIP = value.(string)
		case "soltcp":
			component.SolPort = int(value.(float64))
		case "component":
			component.ComUUID = int(value.(float64))
		}
	}
}

func registerByStar() {
	url := urlPräfix + "/vs/v1/system"

	var reqisterRequestModel = utils.RegisterRequestModel{
		STAR:      strconv.Itoa(component.StarUUID),
		SOL:       component.SolUUID,
		COMPONENT: component.ComUUID,
		COMIP:     component.ComIP,
		COMTCP:    component.ComPort,
		STATUS:    http.StatusOK,
	}

	jsonRegisterRequest, err := json.Marshal(reqisterRequestModel)
	if err != nil {
		log.Error("Error while marshalling data", slog.String("error", err.Error()))
		return
	}

	_, err = http.NewRequest("POST", url, strings.NewReader(string(jsonRegisterRequest)))
	if err != nil {
		log.Error("Failed to create POST request", slog.String("error", err.Error()))
	}
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

func sendHeartBeatToSol() {

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
