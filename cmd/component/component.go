package component

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	con "github.com/Lars5Janssen/vsp/connection"
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
	logger = logger.With(slog.String("LogFrom", "Component"))
	logger.Info("Starting as Component")
	log = *logger

	initializeComponent(logger, ctx, response)

	// TODO Hier scheint eine Loop logic sein zu müssen damit die Ports available bleiben
	go con.AttendHTTP(logger, restIn, restOut, endpoints) // Will Handle endpoints in this thread

	registerByStar()

	logger.Info("Component values",
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
				if !sendHeartBeatToSol() {
					time.Sleep(10 * time.Second)
					if !sendHeartBeatToSol() {
						time.Sleep(20 * time.Second)
						if !sendHeartBeatToSol() {
							logger.Error(
								"Failed to send heartbeat to SOL three time. Exiting Component",
							)
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

	// TODO Muss man hier noch aufräumen? Context leeren oder so?
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

	resp := sendMessageToSol(reqisterRequestModel, url, "POST")

	if reflect.TypeOf(resp) == reflect.TypeOf(con.RestOut{}) {
		return
	}

	respGeneralRequest := resp.(utils.GeneralResponse)

	switch respGeneralRequest.STATUSCODE {
	case http.StatusOK:
		log.Info("Successfully registered by Sol")
		fmt.Printf("Sucessfully registered by Sol with id: %d \n", component.ComUUID)
		return
	case http.StatusUnauthorized:
		log.Error("Unauthorized to register by Sol")
		fmt.Printf("Unauthorized to register by Sol with id: %d \n", component.ComUUID)
		runComponentThread = false
		return
	case http.StatusForbidden:
		log.Error("No room left")
		fmt.Printf("No room left to register by Sol with id: %d \n", component.ComUUID)
		runComponentThread = false
		return
	case http.StatusConflict:
		log.Error("The request was invalid")
		fmt.Printf("The request was invalid to register by Sol with id: %d \n", component.ComUUID)
		runComponentThread = false
		return
	}
	return
}

/*
sendHeartBeatBackToSol 1.1 Pflege des Sterns – Kontrolle der Komponenten
sendHeartBeatToSol 1.1 Pflege des Sterns – Kontrolle der Komponenten
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
	fmt.Printf("Sending HeartBeatBackToSol\n")
	return con.RestOut{StatusCode: http.StatusOK, Body: model}
}

func sendHeartBeatToSol() bool {
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

	resp := sendMessageToSol(RequestModel, url, "PATCH")

	if reflect.TypeOf(resp) == reflect.TypeOf(con.RestOut{}) {
		log.Info("Failed to send heartbeat to SOL: " + component.SolIP +
			":" + strconv.Itoa(component.SolPort))
		return false
	}

	response := resp.(utils.GeneralResponse)

	if response.STATUSCODE != http.StatusOK {
		log.Error("Failed to send heartbeat to SOL:"+component.SolIP+
			":"+strconv.Itoa(component.SolPort)+", Wrong Status: ", slog.Int("status", response.STATUSCODE))
		return false
	}
	fmt.Printf("Sending Heartbeat to Sol. Sol Id: %d \n", component.SolUUID)
	return true
}

/*
*
disconnectFromStar 1.3 Pflege des Sterns – Abmelden von SOL nach Aufruf von SOL
*/
func disconnectFromStar(response con.RestIn) con.RestOut {
	if response.Context.Query("star") != component.StarUUID {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	}

	log.Info("Disconnecting from Star")
	fmt.Printf("Disconnecting from Star after request from Sol.\n")
	return con.RestOut{StatusCode: http.StatusOK}
}

/*
disconnectAfterExit 1.2 Pflege des Sterns – Abmelden von SOL
*/
func disconnectAfterExit() {
	ticker := time.NewTicker(10 * time.Second)
	failedString := "Failed to disconnect from Star. Star not reachable. Try it again"

	for runComponentThread {
		select {
		case <-ticker.C:
			if !disconnectAfterExitHelper() {
				log.Error(failedString)
				fmt.Println(failedString)
				time.Sleep(10 * time.Second)
				if !disconnectAfterExitHelper() {
					log.Error(failedString)
					fmt.Println(failedString)
					time.Sleep(20 * time.Second)
					if !disconnectAfterExitHelper() {
						log.Error(failedString)
						fmt.Println(failedString)
						setRunComponentThread(false)
					}
				}
			}
			setRunComponentThread(false)
		}
	}
}

func disconnectAfterExitHelper() bool {
	log.Info("Disconnect from Star")
	url := urlSolPraefix + "/vs/v1/system/" + strconv.Itoa(
		component.ComUUID,
	) + "?star=" + component.StarUUID

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Error("Failed to create DELETE request", slog.String("error", err.Error()))
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to send request to Star: ", slog.String("error", err.Error()))
		/*
			Die sich abmeldende Komponente beendet sich selbst, auch bei einem Statuscode, der
			einen Fehler signalisiert. - Zitat aus der Aufgabe 1.3
		*/
		return false
	}

	if resp.StatusCode != http.StatusOK {
		log.Error("Failed to send request to Star: ", slog.String("StatusCode", resp.Status))
		return false
	}
	fmt.Printf("Disconnect after exit.")
	return true
}

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

	sendMessageToSol(messages, urlSolPraefix+"/vs/v1/messages", "POST")
}

/*
forwardMessage nutzt das MessageRequestModel 2.1
*/
func forwardMessage(response con.RestIn) con.RestOut {
	log.Info("Forwarding Message to Star")

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
	}

	url := urlSolPraefix + "/vs/v1/messages"

	resp := sendMessageToSol(message, url, "POST")

	if reflect.TypeOf(resp) == reflect.TypeOf(con.RestOut{}) {
		return resp.(con.RestOut)
	}

	respGeneralRequest := resp.(utils.GeneralResponse)

	if respGeneralRequest.STATUSCODE != http.StatusOK {
		return con.RestOut{StatusCode: respGeneralRequest.STATUSCODE}
	}

	fmt.Printf("Forward message to Star. \n")
	return con.RestOut{StatusCode: http.StatusOK, Body: respGeneralRequest.RESPONSEBODY}
}

/*
Aufgabe 2.3 getListOfAllMessages
*/
func getListOfAllMessages(response con.RestIn) con.RestOut {
	log.Info("Getting List of all Messages")

	starUuid := response.Context.Query("star")
	scope := response.Context.Query("scope")
	view := response.Context.Query("view")

	if starUuid != component.StarUUID {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	}

	url := urlSolPraefix + "/vs/v1/messages?star=" + starUuid + "&scope=" + scope + "&view=" + view

	resp := sendMessageToSol(nil, url, "GET")

	if reflect.TypeOf(resp) == reflect.TypeOf(con.RestOut{}) {
		return resp.(con.RestOut)
	}

	respGeneralRequest := resp.(utils.GeneralResponse)

	return con.RestOut{StatusCode: http.StatusOK, Body: respGeneralRequest.RESPONSEBODY}
}

/*
Aufgabe 2.3 getMessageByUUID
*/
func getMessageByUUID(response con.RestIn) con.RestOut {
	log.Info("Getting Message by UUID")

	starUuid := response.Context.Query("star")
	msgId := response.Context.Param("msgUUID")

	if starUuid != component.StarUUID {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	} else if msgId == "" {
		return con.RestOut{StatusCode: http.StatusNotFound}
	}
	url := urlSolPraefix + "/vs/v1/messages/" + msgId + "?star=" + starUuid

	resp := sendMessageToSol(nil, url, "GET")

	if reflect.TypeOf(resp) == reflect.TypeOf(con.RestOut{}) {
		return resp.(con.RestOut)
	}

	respGeneralRequest := resp.(utils.GeneralResponse)

	if respGeneralRequest.STATUSCODE != http.StatusOK {
		return con.RestOut{StatusCode: respGeneralRequest.STATUSCODE}
	}

	return con.RestOut{StatusCode: http.StatusOK, Body: respGeneralRequest.RESPONSEBODY}
}

/*
2.2: Weiterleiten von DELETE Requests von Komponente an Sol
*/
func forwardDeletingMessages(request con.RestIn) con.RestOut {
	log.Info("Forwarding DELETE Request to Star")

	if request.Context.Query("star") != component.StarUUID {
		// TODO Soll schon auf den korrekten Star schon bei der Komponente abgefangen werden?
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	}

	url := urlSolPraefix + "/vs/v1/messages/" +
		request.Context.Param("msgUUID") +
		"?star=" + request.Context.Query("star")

	resp := sendMessageToSol(nil, url, "DELETE")

	if reflect.TypeOf(resp) == reflect.TypeOf(con.RestOut{}) {
		return resp.(con.RestOut)
	}

	respGeneralRequest := resp.(utils.GeneralResponse)

	if respGeneralRequest.STATUSCODE != http.StatusOK {
		return con.RestOut{StatusCode: respGeneralRequest.STATUSCODE}
	}

	fmt.Printf("Forward deleting message.\n")
	return con.RestOut{StatusCode: http.StatusOK, Body: respGeneralRequest.RESPONSEBODY}
}

func sendMessageToSol(message interface{}, url string, requestType string) interface{} {
	log.Info("Sending Message to SOL")
	var client = &http.Client{}
	var resp *http.Response
	var req *http.Request

	if message != nil {
		messageToSend, err := json.Marshal(message)
		if err != nil {
			log.Error("Error while marshalling data", slog.String("error", err.Error()))
			return con.RestOut{StatusCode: http.StatusConflict, Body: gin.H{"error": "Error while marshalling data"}}
		}

		// Build Json Request
		req, err = http.NewRequest(requestType, url, strings.NewReader(string(messageToSend)))
		if err != nil {
			log.Error("Failed to create "+requestType+" request", slog.String("error", err.Error()))
			return con.RestOut{StatusCode: http.StatusConflict, Body: gin.H{"error": err.Error()}}
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		var err error
		// Build text/plain Request
		req, err = http.NewRequest(requestType, url, nil)
		if err != nil {
			log.Error("Failed to create "+requestType+" request", slog.String("error", err.Error()))
			return con.RestOut{StatusCode: http.StatusConflict, Body: gin.H{"error": err.Error()}}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to send Message to SOL with id: "+component.StarUUID+"."+
			"\n Maybe the star is not reachable anymore.", slog.String("error", err.Error()))
		return con.RestOut{
			StatusCode: http.StatusInternalServerError,
			Body:       gin.H{"error": err.Error()},
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Failed to read response body", slog.String("error", err.Error()))
		return con.RestOut{StatusCode: http.StatusBadRequest, Body: gin.H{"error": err.Error()}}
	}

	var respBody interface{}
	err = json.Unmarshal(body, &respBody)

	return utils.GeneralResponse{
		STATUSCODE:   resp.StatusCode,
		RESPONSEBODY: respBody,
	}
}

/**
	Helper Methods
 **/

func iAmNotSol(_ con.RestIn) con.RestOut {
	return con.RestOut{StatusCode: http.StatusUnauthorized}
}

func setRunComponentThread(value bool) {
	runComponentThread = value
	log.Info("Flag to terminate the Component is set.")
}
