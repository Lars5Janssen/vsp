package component

import (
	"context"
	"encoding/json"
	"net"

	con "github.com/Lars5Janssen/vsp/connection"

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

var urlSolPraefix = ""
var runComponentThread = true

var log slog.Logger

var client = &http.Client{}

func StartComponent(
	ctx context.Context,
	logger *slog.Logger,
	commands chan string,
	restIn chan con.RestIn,
	restOut chan con.RestOut,
	response string,
) {
	logger = logger.With(slog.String("Component", "Component"))
	logger.Info("Starting as Component")
	log = *logger

	initializeComponent(logger, ctx, response)

	// TODO Hier scheint eine Loop logic sein zu müssen damit die Ports available bleiben
	go con.AttendHTTP(logger, restIn, restOut, endpoints) // Will Handle endpoints in this thread

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

	urlSolPraefix = "http://" + component.SolIP + ":" + strconv.Itoa(component.SolPort)

	// Create a custom DialContext function
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: net.ParseIP(component.ComIP)},
		// Port: component.ComPort da hier schon der eingangsport ist führt das zu:
		// dial tcp :8006->172.17.0.2:8006: bind: address already in use nur beim heartbeat
		Timeout: 30 * time.Second,
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
	var parsedData utils.ResponseModel
	err := json.Unmarshal([]byte(cleanedInput), &parsedData)
	if err != nil {
		log.Error("Error while parsing response")
		return
	}
	err = parsedData.Validate()
	if err != nil {
		log.Error("Error while validating response")
		return
	}

	// Daten in struct schreiben
	component.ComUUID = parsedData.COMPONENT
	component.SolPort = parsedData.SOLTCP
	component.SolUUID = parsedData.SOL
	component.StarUUID = parsedData.STAR
	component.SolIP = parsedData.SOLIP
}

func registerByStar() {
	url := urlSolPraefix + "/vs/v1/system"

	var reqisterRequestModel = utils.RequestModel{
		STAR:      component.StarUUID,
		SOL:       component.SolUUID,
		COMPONENT: component.ComUUID,
		COMIP:     component.ComIP,
		COMTCP:    component.ComPort,
		STATUS:    strconv.Itoa(http.StatusOK),
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
func sendHeartBeatBackToSol(response con.RestIn) con.RestOut {
	log.Info("Received Heartbeat from SOL")
	model := utils.RequestModel{
		STAR:      component.StarUUID,
		SOL:       component.SolUUID,
		COMPONENT: component.ComUUID,
		COMIP:     component.ComIP,
		COMTCP:    component.ComPort,
		STATUS:    strconv.Itoa(component.Status),
	}

	if response.Context.Query("star") != component.StarUUID {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	}

	comUUID := response.Context.Param("comUUID?star=starUUID")

	if comUUID != "" || comUUID != strconv.Itoa(component.ComUUID) {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	}
	return con.RestOut{StatusCode: http.StatusOK, Body: model}
}

func sendHeartBeatToSol(log *slog.Logger) bool {
	log.Info("Sending Heartbeat to SOL")
	url := urlSolPraefix + "/vs/v1/system/" + strconv.Itoa(component.ComUUID)

	RequestModel := utils.RequestModel{
		STAR:      component.StarUUID,
		SOL:       component.SolUUID,
		COMPONENT: component.ComUUID,
		COMIP:     component.ComIP,
		COMTCP:    component.ComPort,
		STATUS:    strconv.Itoa(component.Status),
	}

	jsonHeartBeatRequest, err := json.Marshal(RequestModel)
	if err != nil {
		log.Error("Error while marshalling data", slog.String("error", err.Error()))
		return false
	}

	req, err := http.NewRequest("PATCH", url, strings.NewReader(string(jsonHeartBeatRequest)))
	if err != nil {
		log.Error("Failed to create PATCH request", slog.String("error", err.Error()))
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to send heartbeat to SOL:"+component.SolIP+":"+strconv.Itoa(component.SolPort), slog.String("error", err.Error()))
		return false
	}
	if resp.StatusCode != http.StatusOK {
		log.Error("Failed to send heartbeat to SOL:"+component.SolIP+":"+strconv.Itoa(component.SolPort)+", Wrong Status: ", slog.Int("status", resp.StatusCode))
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
	url := urlSolPraefix + "/vs/v1/system/" + strconv.Itoa(component.ComUUID) + "?star=" + component.StarUUID

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
}

// TODO Diese Methode soll nur eine Message auf Basis der Eingaben des Users erstellen?
// TODO not implemented yet
func createMessage(userInput string) {
	messages := utils.MessageRequestModel{
		STAR:    "",
		ORIGIN:  "",
		SENDER:  "",
		MSGID:   "",
		VERSION: "",
		CREATED: "",
		CHANGED: "",
		SUBJECT: "",
		MESSAGE: "",
	}

	sendMessageToSol(messages)
}

/*
forwardMessage nutzt das MessageRequestModel 2.1
TODO Soll es möglich sein eine Liste von Messages zu übergeben oder nur eine?
*/
func forwardMessage(response con.RestIn) con.RestOut {
	var message utils.MessageRequestModel
	err := response.Context.BindJSON(&message)
	if err != nil {
		return con.RestOut{StatusCode: http.StatusBadRequest}
	}

	err = message.Validate() // validierung ueber die json tags siehe models.go
	if err != nil {
		return con.RestOut{StatusCode: http.StatusPreconditionFailed}
	}

	if message.STAR != component.StarUUID {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
		// TODO Soll schon auf den korrekten Star schon bei der Komponente abgefangen werden?
	}

	return sendMessageToSol(message)
}

/*
Aufgabe 2.3 getListOfAllMessages
*/
func getListOfAllMessages(response con.RestIn) con.RestOut {
	body := gin.H{"message": "test"}
	return con.RestOut{StatusCode: http.StatusOK, Body: body}
}

/*
Aufgabe 2.3 getMessageByUUID
*/
func getMessageByUUID(response con.RestIn) con.RestOut {
	body := gin.H{"message": "test"}
	return con.RestOut{StatusCode: http.StatusOK, Body: body}
}

/*
2.2: Weiterleiten von DELETE Requests von Komponente an Sol
*/
func forwardDeletingMessages(response con.RestIn) con.RestOut {
	log.Info("Forwarding DELETE Request to SOL")

	if response.Context.Query("star") != component.StarUUID {
		// TODO Soll schon auf den korrekten Star schon bei der Komponente abgefangen werden?
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	}

	url := urlSolPraefix + "/vs/v1/messages/" +
		response.Context.Param("msgUUID") +
		"?star=" + response.Context.Query("star")

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Error("Failed to create DELETE request", slog.String("error", err.Error()))
		return con.RestOut{StatusCode: http.StatusInternalServerError}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to send DELETE request to SOL: ", slog.String("error", err.Error()))
		return con.RestOut{StatusCode: http.StatusInternalServerError}
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	}
	if resp.StatusCode == http.StatusNotFound {
		return con.RestOut{StatusCode: http.StatusNotFound}
	}

	return con.RestOut{StatusCode: http.StatusOK}
}

func sendMessageToSol(message utils.MessageRequestModel) con.RestOut {
	log.Info("Sending Message to SOL")
	url := urlSolPraefix + "/vs/v1/messages"

	messageToSend, err := json.Marshal(message)
	if err != nil {
		log.Error("Error while marshalling data", slog.String("error", err.Error()))
		return con.RestOut{StatusCode: http.StatusPreconditionFailed, Body: gin.H{"error": "Error while marshalling data"}}
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(messageToSend)))
	if err != nil {
		log.Error("Failed to create POST request", slog.String("error", err.Error()))
		return con.RestOut{StatusCode: http.StatusConflict, Body: gin.H{"error": err.Error()}}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to send Message to SOL with id: "+component.StarUUID+".\n Meaby the star is not reachable anymore.", slog.String("error", err.Error()))
		return con.RestOut{StatusCode: http.StatusInternalServerError, Body: gin.H{"error": err.Error()}}
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	}
	if resp.StatusCode == http.StatusNotFound {
		return con.RestOut{StatusCode: http.StatusNotFound}
	}
	if resp.StatusCode == http.StatusPreconditionFailed {
		return con.RestOut{StatusCode: http.StatusPreconditionFailed}
	}

	log.Info("Message was successfully sent to SOL")
	return con.RestOut{StatusCode: http.StatusOK}
}

/**
	Helper Methods
 **/

func notAvailable(_ con.RestIn) con.RestOut {
	return con.RestOut{StatusCode: http.StatusNotFound}
}

func iAmNotSol(_ con.RestIn) con.RestOut {
	return con.RestOut{StatusCode: http.StatusUnauthorized}
}

func setRunComponentThread(value bool) {
	runComponentThread = value
}
