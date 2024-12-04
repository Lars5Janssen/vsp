package component

import (
	"context"
	"encoding/json"
	"net"

	n "github.com/Lars5Janssen/vsp/connection"

	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Lars5Janssen/vsp/utils"
)

var component = Component{
	ComUUID:  0,
	ComIP:    "",
	SolUUID:  0,
	StarUUID: "",
	SolIP:    "",
	SolPort:  0,
}

var urlSolPräfix = ""
var runComponentThread = true

var log slog.Logger

var client = &http.Client{}

func StartComponent(
	ctx context.Context,
	logger *slog.Logger,
	commands chan string,
	restIn chan n.RestIn,
	restOut chan n.RestOut,
	response string,
) {
	logger = logger.With(slog.String("Component", "Component"))
	logger.Info("Starting as Component")
	log = *logger

	initializeComponent(logger, ctx, response)

	// TODO Hier scheint eine Loop logic sein zu müssen damit die Ports available bleiben
	go n.AttendHTTP(logger, restIn, restOut, endpoints) // Will Handle endpoints in this thread

	registerByStar()

	logger.Info("Componenten values",
		slog.Int("ComUUID", component.ComUUID),
		slog.Int("SolUUID", component.SolUUID),
		slog.String("StarUUID", component.StarUUID),
		slog.String("ComIP", component.ComIP),
		slog.Int("ComPort", component.ComPort),
		slog.String("SolIP", component.SolIP),
		slog.Int("SolPort", component.SolPort),
	)

	// Send Heartbeat to Sol
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for runComponentThread {
			select {
			case <-ticker.C:
				if !sendHeartBeatToSol(logger) {
					time.Sleep(10 * time.Second)
					if !sendHeartBeatToSol(logger) {
						time.Sleep(20 * time.Second)
						if !sendHeartBeatToSol(logger) {
							logger.Error("Failed to send heartbeat to SOL three time. Exiting Component")
							setRunComponentThread(false)
						}
					}
				}
			}
		}
	}()

	for runComponentThread {
		// Retrieve from user input
		select {
		case command := <-commands:
			if command == "exit" {
				logger.Info("Exiting Component")
				disconnectAfterExit()
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
	StarUUID string `json:"StarUUID"`
	ComPort  int    `json:"ComPort"`
	SolIP    string `json:"SolIP"`
	SolPort  int    `json:"SolPort"`
	Status   int    `json:"Status"`
}

func initializeComponent(log *slog.Logger, ctx context.Context, response string) {
	component.ComIP = ctx.Value("ip").(string)
	component.ComPort = ctx.Value("port").(int)
	component.Status = http.StatusOK // TODO ist das zu beginn wirklich so?

	parseResponseIntoComponent(response, log)

	urlSolPräfix = "http://" + component.SolIP + ":" + strconv.Itoa(component.SolPort)

	// Create a custom DialContext function
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{Port: component.ComPort},
		Timeout:   30 * time.Second,
	}

	customDialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}

	// Custom Transport with our DialContext
	transport := &http.Transport{
		DialContext: customDialContext,
	}

	// HTTP Client with custom Transport
	client = &http.Client{
		Transport: transport,
	}
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
			component.StarUUID = value.(string)
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
	url := urlSolPräfix + "/vs/v1/system"

	var reqisterRequestModel = utils.RegisterRequestModel{
		STAR:      component.StarUUID,
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

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonRegisterRequest)))
	if err != nil {
		log.Error("Failed to create POST request", slog.String("error", err.Error()))
	}

	resp, err := client.Do(req)
	if err == nil && resp != nil {
		switch resp.StatusCode {
		case http.StatusOK:
			log.Info("Successfully registered by Sol")
		case http.StatusUnauthorized:
			log.Error("Unauthorized to register by Sol")
		case http.StatusForbidden:
			log.Error("No room left")
		case http.StatusConflict:
			log.Error("The request was invalid")
		}
		return
	} else {
		if err != nil {
			log.Error("Failed to send request to SOL: ", slog.String("error", err.Error()))
		}
	}
}

/*
sendHeartBeatBackToSol 1.1 Pflege des Sterns – Kontrolle der Komponenten
sendHeartBeatToSol 1.1 Pflege des Sterns – Kontrolle der Komponenten

Wenn SOL für eine aktive Komponente 60 Sekunden nach der letzten Meldung keine neue Meldung mit einem Status von „200“
erhält, baut SOL zum <STARPORT>/tcp der Komponente eine UNICAST-Verbindung auf und kontrolliert selbst, ob die
Komponente noch aktiv und funktionsfähig ist. Diese Kontrollmöglichkeit muss SOL auch für sich selbst unterstützen!
Auch hier kommt eine REST-API zum Einsatz.
*/
func sendHeartBeatBackToSol(response n.RestIn) n.RestOut {
	log.Info("Received Heartbeat from SOL")
	model := utils.HeartBeatRequestModel{
		STAR:      component.StarUUID,
		SOL:       component.SolUUID,
		COMPONENT: component.ComUUID,
		COMIP:     component.ComIP,
		COMTCP:    component.ComPort,
		STATUS:    component.Status,
	}

	if response.Context.Query("star") != component.StarUUID {
		return n.RestOut{StatusCode: http.StatusUnauthorized}
	}

	comUUID := response.Context.Param("comUUID?star=starUUID")

	if comUUID != "" || comUUID != strconv.Itoa(component.ComUUID) {
		return n.RestOut{StatusCode: http.StatusUnauthorized}
	}
	return n.RestOut{StatusCode: http.StatusOK, Body: model}
}

func sendHeartBeatToSol(log *slog.Logger) bool {
	log.Info("Sending Heartbeat to SOL")
	url := urlSolPräfix + "/vs/v1/system/" + strconv.Itoa(component.ComUUID)

	var heartBeatRequestModel = utils.HeartBeatRequestModel{
		STAR:      component.StarUUID,
		SOL:       component.SolUUID,
		COMPONENT: component.ComUUID,
		COMIP:     component.ComIP,
		COMTCP:    component.ComPort,
		STATUS:    component.Status,
	}

	jsonHeartBeatRequest, err := json.Marshal(heartBeatRequestModel)
	if err != nil {
		log.Error("Error while marshalling data", slog.String("error", err.Error()))
		return false
	}

	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(jsonHeartBeatRequest)))
	if err != nil {
		log.Error("Failed to create PATCH request", slog.String("error", err.Error()))
	}

	// LOOP for Time maybe here?
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to send heartbeat to SOL:"+component.SolIP+":"+strconv.Itoa(component.SolPort), slog.String("error", err.Error()))
		return false
	}
	if resp.StatusCode != http.StatusOK {
		log.Error("Failed to send heartbeat to SOL:"+component.SolIP+": "+strconv.Itoa(component.SolPort)+", Wrong Status: ", slog.Int("status", resp.StatusCode))
		return false
	}

	return true
}

/*
Aufgabe 1.3 disconnectFromStar Pflege des Sterns – Abmelden von SOL

Wenn die Komponente, die gerade aktiv den Stern „managed“ (also SOL) den „EXIT“Befehl bekommt, werden von ihr alle
aktiven Komponenten im Stern einzeln kontaktiert:
*/
func disconnectFromStar() bool {
	log.Info("Disconnect From Star")
	url := urlSolPräfix + "/vs/v1/system/" + strconv.Itoa(component.ComUUID) + "?star=" + component.StarUUID

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Error("Failed to create DELETE request", slog.String("error", err.Error()))
	}

	// LOOP for Time meaby here?
	_, err = client.Do(req)
	if err != nil {
		log.Error("Failed to send request to SOL: ", slog.String("error", err.Error()))
		/*
			Die sich abmeldende Komponente beendet sich selbst, auch bei einem Statuscode, der
			einen Fehler signalisiert. - Zitat aus der Aufgabe 1.3
		*/
		setRunComponentThread(false)
		return true
	}
	return false
}

// TODO Ab hier nicht implementiert
/*
disconnectAfterExit 1.2 Pflege des Sterns – Abmelden von SOL

Eine aktive Komponente, die sich nach einem „EXIT“-Befehl bei SOL abmeldet, baut
eine UNICAST-Verbindung auf. Wenn SOL nicht erreichbar ist, wird es nach 10 bzw. 20
Sekunden nochmal versucht. Wenn dann immer noch keine Verbindung zustande
kommt, beendet sich die Komponente selbst.
*/
func disconnectAfterExit() {
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for runComponentThread {
			select {
			case <-ticker.C:
				if !disconnectFromStar() {
					log.Error("Failed to disconnect from star")
					time.Sleep(10 * time.Second)
					if !disconnectFromStar() {
						time.Sleep(20 * time.Second)
						if !disconnectFromStar() {
							log.Error("Failed to disconnect from star three time. Exiting Component")
							setRunComponentThread(false)
						}
					}
				}
			}
		}
	}()

	// TODO LOOP
}

/*
createOrForwardMessage nutzt das MessageRequestModel 2.1
*/
func createOrForwardMessage(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{StatusCode: http.StatusOK, Body: body}
}

/*
Aufgabe 2.3 getListOfAllMessages
*/
func getListOfAllMessages(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{StatusCode: http.StatusOK, Body: body}
}

/*
Aufgabe 2.3 getMessageByUUID
*/
func getMessageByUUID(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{StatusCode: http.StatusOK, Body: body}
}

/*
2.2: Weiterleiten von DELETE Requests von Komponente an Sol
*/
func forwardDeletingMessages(response n.RestIn) n.RestOut {
	body := gin.H{"message": "test"}
	return n.RestOut{StatusCode: http.StatusOK, Body: body}
}

/**
	Helper Methods
 **/

func notAvailable(_ n.RestIn) n.RestOut {
	return n.RestOut{StatusCode: http.StatusNotFound}
}

func iAmNotSol(_ n.RestIn) n.RestOut {
	return n.RestOut{StatusCode: http.StatusUnauthorized}
}

func setRunComponentThread(value bool) {
	runComponentThread = value
}
