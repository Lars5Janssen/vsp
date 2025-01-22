package sol

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	con "github.com/Lars5Janssen/vsp/connection"
	"github.com/Lars5Janssen/vsp/utils"
)

type Sol struct {
	StarUUID  string
	SolUUID   int
	IPAddress string
	Port      int
}

type ComponentEntry struct {
	// TODO Comuuid rausnehmen?
	ComUUID         int
	IPAddress       string
	Port            int
	TimeIntegration time.Time
	TimeInteraktion time.Time
	Status          string
	ActiveStatus    utils.ActiveStatus
}

var log *slog.Logger

var sol Sol

var solList = map[int]ComponentEntry{}

var lenOfSolList int

var nonce = 1

var msgList = map[string]utils.MessageModel{}

var tmpComUuidSet = utils.NewIntSet()

func StartSol(
	ctx context.Context,
	logger *slog.Logger,
	commands chan string,
	udp chan con.UDP,
	restIn chan con.RestIn,
	restOut chan con.RestOut,
) {
	log = logger
	log = log.With(slog.String("LogFrom", "SOL"))
	log.Info("Starting as SOL")
	fmt.Sprintf("Starting as SOL id: %s", sol.SolUUID)

	// SOL Logic
	initializeSol(log, ctx)

	// Add to solList
	solList[sol.SolUUID] = ComponentEntry{
		ComUUID:         sol.SolUUID,
		IPAddress:       sol.IPAddress,
		Port:            sol.Port,
		TimeIntegration: time.Now(),
		TimeInteraktion: time.Now(),
		Status:          strconv.Itoa(http.StatusOK),
		ActiveStatus:    utils.Active,
	}

	// Max active components
	temp := ctx.Value("maxActiveComponents")
	lenOfSolList = temp.(int)

	// has to be done outside for loop
	go con.AttendHTTP(log, restIn, restOut, solEndpoints)

	// Check if the components are still active
	go monitorComponents()

	// forever loop for commands and udp messages
	for {
		// Retrieve from user input
		select {
		case command := <-commands:
			if command == "exit" {
				sendDeleteRequests()
				return
			}
		default:
		}

		// Retrieve from UDP server
		select {
		case udpInput := <-udp:
			// to test: echo HELLO? | ncat -u 255.255.255.255 8006
			log.Info("Received UDP message", slog.String("message", udpInput.Message))
			if udpInput.Message == "HELLO?" {
				intValue, err := generateComUUID()
				if err != nil {
					log.Error("Error generating comUUID")
					return
				}
				response := utils.ResponseModel{
					STAR:      sol.StarUUID,
					SOL:       sol.SolUUID,
					COMPONENT: intValue,
					SOLIP:     sol.IPAddress,
					SOLTCP:    sol.Port,
				}

				// response.COMPONENT = 2000 // TODO only for test purposes

				marshal, err := json.Marshal(response)
				if err != nil {
					log.Error("Error marshalling response", slog.String("error", err.Error()))
					return
				}
				log.Info("Sending response to: "+response.STAR, slog.String("response", response.STAR))
				if con.OwnAddrCheck(*log, udpInput.Addr.IP.String()) {
					log.Debug("Would send message to own. Bad")
				}
				err = con.SendMessage(log, udpInput.Addr, sol.Port, string(marshal))
				if err != nil {
					log.Error("Error sending msg", "Error", err, "Addr", udpInput.Addr.IP)
					// return
				}
				// Set um zu prüfen welche ComUUIDs schon vergeben wurden
				tmpComUuidSet.Add(intValue)
			}
		default:
		}

	}
}

func monitorComponents() {
	for {
		// Sort solList by TimeInteraktion
		sortedList := sortSolListByTimeInteraktion()

		if len(sortedList) == 0 {
			time.Sleep(60 * time.Second)
			continue
		}

		// Get the earliest TimeInteraktion and add 60 seconds
		earliestTime := sortedList[0].TimeInteraktion
		waitTime := earliestTime.Add(60 * time.Second)
		log.Info("Next check in 60 seconds",
			slog.Time("time", waitTime),
			slog.String("comUUID", strconv.Itoa(sortedList[0].ComUUID)),
		)

		// Wait until the waitTime
		time.Sleep(time.Until(waitTime))

		// Check if the earliest TimeInteraktion has changed
		if sortedList[0].TimeInteraktion != earliestTime {
			continue
		}

		// Send request to the component
		err := sendRequestsToActiveComponents(sortedList[0].ComUUID)
		if err != nil {
			log.Error("Error sending request to component", slog.String("error", err.Error()))
		}
	}
}

func sortSolListByTimeInteraktion() []ComponentEntry {
	var sortedList []ComponentEntry
	for _, entry := range solList {
		if entry.ActiveStatus == utils.Active {
			sortedList = append(sortedList, entry)
		}
	}

	sort.Slice(sortedList, func(i, j int) bool {
		return sortedList[i].TimeInteraktion.Before(sortedList[j].TimeInteraktion)
	})

	return sortedList
}

func sendRequestsToActiveComponents(uuid int) error {
	entry := solList[uuid]
	url := "http://" + entry.IPAddress + ":" + strconv.Itoa(entry.Port) + "/vs/v1/system/" + strconv.Itoa(uuid) + "?star=" + sol.StarUUID
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error("Failed to create GET request", slog.String("error", err.Error()))
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to send GET request", slog.String("error", err.Error()))
		log.Info("Component is disconnected", slog.Int("uuid", uuid))
		entry.ActiveStatus = utils.Disconnected
	}
	if resp != nil {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Error("Failed to close response body", slog.String("error", err.Error()))
			}
		}(resp.Body)

		if resp.StatusCode != http.StatusOK {
			log.Error("Received non-OK response", slog.Int("statusCode", resp.StatusCode))
			log.Info("Component is disconnected", slog.Int("uuid", uuid))
			entry.ActiveStatus = utils.Disconnected
			entry.Status = strconv.Itoa(resp.StatusCode)
		} else {
			log.Info("Successfully sent GET request", slog.Int("uuid", uuid))
			var RequestModel utils.RequestModel
			err := json.NewDecoder(resp.Body).Decode(&RequestModel)
			if err != nil {
				log.Error("Failed to decode response body", slog.String("error", err.Error()))
				entry.Status = strconv.Itoa(resp.StatusCode)
				log.Info("Component is disconnected", slog.Int("uuid", uuid))
				entry.ActiveStatus = utils.Disconnected
			} else {
				entry.Status = RequestModel.STATUS
				log.Info("Successfully parsed response body", slog.Any("RequestModel", RequestModel))
			}
		}
	}

	entry.TimeInteraktion = time.Now()
	solList[uuid] = entry
	return err
}

/*
createComponent - Seite 4 im Aufgabenblatt Aufgabe 1.0

Bitte das RequestModel nutzen.

Damit SOL den auch kennt, gibt die neue Komponente selbst ihren derzeitigen Status an. Außerdem werden Angaben zum Stern
übermittelt, die von SOL vor der Integration auch geprüft werden, damit sichergestellt ist, dass auch der „richtige“
Stern gemeint ist.
*/
func registerComponentBySol(request con.RestIn) con.RestOut {
	var registerRequestModel utils.RequestModel
	// err := request.Context.BindJSON(&registerRequestModel)
	err := request.Context.ShouldBindJSON(&registerRequestModel)
	// If the JSON is not valid, return 401 Unauthorized
	if err != nil {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	}

	// Check if all the info from the component is correct, and hold tmpComUuidSet consistent.
	if checkTmpComUuidSet(registerRequestModel) != http.StatusOK {
		return con.RestOut{StatusCode: http.StatusConflict}
	} else if checkConflict(registerRequestModel, request.IpAndPort) != http.StatusOK {
		return con.RestOut{StatusCode: http.StatusConflict}
	} else if checkUnauthorized(registerRequestModel) != http.StatusOK {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	} else if checkNoRoomLeft() != http.StatusOK {
		return con.RestOut{StatusCode: http.StatusForbidden}
	} else if checkNotFound(registerRequestModel) == http.StatusOK { // If the component is already in the list, return 409 Conflict
		return con.RestOut{StatusCode: http.StatusConflict}
	}

	// Add the component to the list
	solList[registerRequestModel.COMPONENT] = ComponentEntry{
		ComUUID:         registerRequestModel.COMPONENT,
		IPAddress:       registerRequestModel.COMIP,
		Port:            registerRequestModel.COMTCP,
		TimeIntegration: time.Now(),
		TimeInteraktion: time.Now(),
		Status:          strconv.Itoa(http.StatusOK),
		ActiveStatus:    utils.Active,
	}

	return con.RestOut{StatusCode: http.StatusOK}
}

/*
Das soll die checkAvailabilityFromComponent Methode übernehmen - Aufgabe 1.1 Pflege des Sterns

Jede aktive Komponente baut alle 30 Sekunden eine UNICAST-Verbindung zum
<STARPORT>/tcp von SOL auf.2 Wenn SOL nicht erreichbar ist, wird es nach 10 bzw. 20
Sekunden nochmal versucht. Wenn dann immer noch keine Verbindung zustande
kommt, beendet sich die Komponente selbst.
*/
func checkAvailabilityFromComponent(request con.RestIn) con.RestOut {
	var registerRequestModel utils.RequestModel
	err := request.Context.ShouldBindJSON(&registerRequestModel)
	if err != nil {
		// Return 400 Bad Request if JSON is not valid
		return con.RestOut{StatusCode: http.StatusBadRequest}
	}

	log.Info("RequestModel",
		slog.String("Star", registerRequestModel.STAR),
		slog.String("Sol", strconv.Itoa(registerRequestModel.SOL)),
		slog.String("LogFrom", strconv.Itoa(registerRequestModel.COMPONENT)),
		slog.String("ComIP", registerRequestModel.COMIP),
		slog.String("ComTcp", strconv.Itoa(registerRequestModel.COMTCP)),
		slog.String("Status", registerRequestModel.STATUS),
	)

	// Check if info correct
	if checkUnauthorized(registerRequestModel) != http.StatusOK {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	} else if checkNotFound(registerRequestModel) != http.StatusOK {
		return con.RestOut{StatusCode: http.StatusNotFound}
	} else if checkConflict(registerRequestModel, request.IpAndPort) != http.StatusOK {
		return con.RestOut{StatusCode: http.StatusConflict}
	}

	// Update the time of interaction
	if entry, exists := solList[registerRequestModel.COMPONENT]; exists {
		entry.ActiveStatus = utils.Active
		entry.TimeInteraktion = time.Now()
		solList[registerRequestModel.COMPONENT] = entry
	}

	return con.RestOut{StatusCode: http.StatusOK}
}

func sendHeartBeatBack(request con.RestIn) con.RestOut {
	if sol.StarUUID != request.Context.Query("star") {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	}
	uuid := request.Context.Param("comUUID")
	if uuid == "" {
		return con.RestOut{StatusCode: http.StatusConflict}
	}
	comUuid, _ := strconv.Atoi(uuid)
	if sol.SolUUID != comUuid {
		return con.RestOut{StatusCode: http.StatusConflict}
	}

	RequestModel := utils.RequestModel{
		STAR:      sol.StarUUID,
		SOL:       sol.SolUUID,
		COMPONENT: sol.SolUUID,
		COMIP:     sol.IPAddress,
		COMTCP:    sol.Port,
		STATUS:    strconv.Itoa(http.StatusOK),
	}

	return con.RestOut{StatusCode: http.StatusOK, Body: RequestModel}
}

/*
1.2 Pflege des Sterns – Abmelden einer Komponente

Eine aktive Komponente, die sich nach einem „EXIT“-Befehl bei SOL abmeldet, baut eine UNICAST-Verbindung auf.
Wenn SOL nicht erreichbar ist, wird es nach 10 bzw. 20 Sekunden nochmal versucht. Wenn dann immer noch keine Verbindung
zustande kommt, beendet sich die Komponente selbst.
*/
func disconnectComponentFromStar(request con.RestIn) con.RestOut {
	// TODO check if component is already deleted
	var out con.RestOut
	var registerRequestModel utils.RequestModel

	registerRequestModel.STAR = request.Context.Query("star")
	stringValue := request.Context.Param("comUUID")
	comUUid, err := strconv.Atoi(stringValue)
	if err != nil {
		out.StatusCode = http.StatusBadRequest
	} else {
		registerRequestModel.COMPONENT = comUUid
		if checkNotFound(registerRequestModel) != http.StatusOK {
			out.StatusCode = http.StatusNotFound
		} else {
			comEntry := solList[registerRequestModel.COMPONENT]
			registerRequestModel.COMIP = comEntry.IPAddress
			registerRequestModel.COMTCP = comEntry.Port
			registerRequestModel.STATUS = comEntry.Status
			registerRequestModel.SOL = sol.SolUUID

			if checkUnauthorized(registerRequestModel) != http.StatusOK {
				out.StatusCode = http.StatusUnauthorized
			} else {
				out.StatusCode = http.StatusOK
			}
		}
	}
	// Update the active status of the component
	if entry, exists := solList[registerRequestModel.COMPONENT]; exists {
		log.Info("Component has left", slog.Int("uuid", entry.ComUUID))
		entry.ActiveStatus = utils.Left
		entry.TimeInteraktion = time.Now()
		solList[registerRequestModel.COMPONENT] = entry
	}

	out.Body = nil
	return out
}

/*
createAndSaveMessage Aufgabe 2.1
*/
func createAndSaveMessage(request con.RestIn) con.RestOut {
	// TODO warum soll SOL eine Sonderbehandlung bekommen?
	var message utils.MessageRequestModel
	err := request.Context.ShouldBindJSON(&message)
	if err != nil {
		return con.RestOut{StatusCode: http.StatusBadRequest}
	}
	if message.STAR != sol.StarUUID {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	} else if message.ORIGIN == "" || message.SUBJECT == "" || !utf8.ValidString(message.ORIGIN) || !utf8.ValidString(message.SUBJECT) {
		return con.RestOut{StatusCode: http.StatusPreconditionFailed}
	}
	subject := strings.Split(message.SUBJECT, "\n")[0]
	subject = strings.ReplaceAll(subject, "\r", "")

	// Regular expression for email validation
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	var msgId string
	if re.MatchString(message.ORIGIN) {
		msgId = strconv.Itoa(nonce) + "@" + message.SENDER
	} else {
		msgId = strconv.Itoa(nonce) + "@" + message.ORIGIN
	}
	nonce++

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	// Add the message to the list
	msgList[msgId] = utils.MessageModel{
		MSGID:   msgId,
		STAR:    message.STAR,
		ORIGIN:  message.ORIGIN,
		SENDER:  message.SENDER,
		VERSION: "1", // TODO Methode für die versionsgabe
		CREATED: timestamp,
		CHANGED: timestamp,
		SUBJECT: subject,
		MESSAGE: message.MESSAGE,
		STATUS:  "active",
	}

	body := utils.MessageId{MSGID: msgId}
	return con.RestOut{StatusCode: http.StatusOK, Body: body}
}

/*
deleteMessage Aufgabe 2.2
*/
func deleteMessage(request con.RestIn) con.RestOut {
	starUuid := request.Context.Query("star")
	msgId := request.Context.Param("msgUUID")

	if starUuid != sol.StarUUID {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	} else if msgId == "" {
		return con.RestOut{StatusCode: http.StatusNotFound}
	} else if _, exists := msgList[msgId]; !exists {
		return con.RestOut{StatusCode: http.StatusNotFound}
	}

	msg := msgList[msgId]
	msg.STATUS = "deleted"
	msg.CHANGED = strconv.FormatInt(time.Now().Unix(), 10)
	msgList[msgId] = msg

	return con.RestOut{StatusCode: http.StatusOK}
}

/*
Aufgabe 2.3 getListOfAllMessages
*/
func getListOfAllMessages(request con.RestIn) con.RestOut {
	starUuid := request.Context.Query("star")
	scope := request.Context.Query("scope")
	view := request.Context.Query("view")

	if starUuid != sol.StarUUID {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	}

	if scope != "all" {
		scope = "active"
	}
	if view != "header" {
		view = "id"
	}

	var resultList []utils.MessageModel
	for _, msg := range msgList {
		if scope == "active" && msg.STATUS == "active" {
			resultList = append(resultList, msg)
		} else if scope == "all" && msg.STATUS == "active" {
			resultList = append(resultList, msg)
		} else if scope == "all" && msg.STATUS != "active" {
			delMsg := utils.MessageModel{
				MSGID:  msg.MSGID,
				STATUS: msg.STATUS,
			}
			resultList = append(resultList, delMsg)
		}
	}

	var body any
	if view == "id" {
		var resultIdList []utils.MessageModelId
		for _, value := range resultList {
			resultIdList = append(resultIdList, utils.MessageModelId{
				MSGID:  value.ORIGIN,
				STATUS: value.STATUS,
			})
		}
		body = utils.MessageListId{
			STAR:         sol.StarUUID,
			SCOPE:        scope,
			VIEW:         view,
			TOTALRESULTS: len(resultList),
			MESSAGES:     resultIdList,
		}
	} else {
		body = utils.MessageListHeader{
			STAR:         sol.StarUUID,
			SCOPE:        scope,
			VIEW:         view,
			TOTALRESULTS: len(resultList),
			MESSAGES:     resultList,
		}
	}

	return con.RestOut{StatusCode: http.StatusOK, Body: body}
}

/*
Aufgabe 2.3 getMessageByUUID
*/
func getMessageByUUID(request con.RestIn) con.RestOut {
	starUuid := request.Context.Query("star")
	msgId := request.Context.Param("msgUUID")

	if starUuid != sol.StarUUID {
		return con.RestOut{StatusCode: http.StatusUnauthorized}
	} else if msgId == "" {
		return con.RestOut{StatusCode: http.StatusNotFound}
		// Dennoch rückgabe leere Liste
		// Auch wenn die <MSG-UUID> nicht existiert, wird
		// eine leere Liste zurückgegeben und die Antwort „404“.
	} else if _, exists := msgList[msgId]; !exists {
		return con.RestOut{StatusCode: http.StatusNotFound}
	}

	return con.RestOut{StatusCode: http.StatusOK, Body: msgList[msgId]}
}

/*
Hilfsmethoden
*/

func initializeSol(log *slog.Logger, ctx context.Context) {
	// ComUUID from Sol
	number, err := generateComUUID()
	if err != nil {
		log.Error("Error generating comUUID for Sol")
		return
	}
	sol.SolUUID = number

	// sol.SolUUID = 1000 // TODO only for test purposes

	// IpAddress and Port from Sol
	sol.IPAddress = ctx.Value("ip").(string)
	intValue := ctx.Value("port").(int)
	sol.Port = intValue

	// StarUUID from Sol
	hashNumber, err := generateStarUUID(*log)
	if err != nil {
		log.Error("Error generating starUUID")
		return
	}
	sol.StarUUID = hashNumber

	// sol.StarUUID = "testStarUUID" // TODO only for test purposes

	return
}

func generateComUUID() (int, error) {
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(9000))
	if err != nil {
		return 0, err
	}
	number := int(randomNumber.Int64() + 1000)
	// If the number was already given, generate a new one
	if listContains(number) {
		return generateComUUID()
	}
	return number, nil
}

func generateStarUUID(log slog.Logger) (string, error) {
	// TODO ID equals last two digits of port?
	// Get the last two digits of the port
	portStr := strconv.Itoa(sol.Port)
	if len(portStr) < 2 {
		log.Error("port number is too short")
		return "", nil
	}
	lastTwoDigits := portStr[len(portStr)-2:]

	// Concatenate the IP address, last two digits of the port, and solUUID
	concatenated := sol.IPAddress + lastTwoDigits + strconv.Itoa(sol.SolUUID)

	// Generate the MD5 hash
	hash := md5.Sum([]byte(concatenated))

	// Convert the hash to a hexadecimal string
	hashStr := hex.EncodeToString(hash[:])

	return hashStr, nil
}

func checkUnauthorized(r utils.RequestModel) int {
	if r.STAR != sol.StarUUID || r.SOL != sol.SolUUID {
		// Return 401 Unauthorized
		return http.StatusUnauthorized
	}
	// Return 200 OK
	return http.StatusOK
}

func checkNoRoomLeft() int {
	count := 0
	for _, entry := range solList {
		if entry.ActiveStatus == utils.Active {
			count++
		}
	}
	if count >= lenOfSolList {
		// Return 403 No room left
		return http.StatusForbidden
	}
	// Return 200 OK
	return http.StatusOK
}

func checkNotFound(r utils.RequestModel) int {
	if !listContains(r.COMPONENT) {
		// Return 404 Not Found
		return http.StatusNotFound
	}
	// Return 200 OK
	return http.StatusOK
}

func checkTmpComUuidSet(r utils.RequestModel) int {
	if !tmpComUuidSet.Contains(r.COMPONENT) {
		log.Error("ComUUID is not Valid")
		fmt.Printf("ComUUID is not Valid")
		return http.StatusConflict
	}

	// Return 200 OK
	tmpComUuidSet.Remove(r.COMPONENT)
	return http.StatusOK
}

func checkConflict(r utils.RequestModel, addr string) int {
	addrs := strings.Split(addr, ":")
	port, err := strconv.Atoi(addrs[1]) // Port von dem Component schickt nicht auf dem er hört

	// TODO remove port == -1
	if err != nil {
		log.Error("Error converting port to int", slog.String("error", err.Error()))
		fmt.Printf("Error converting port to int. Have a look into error logs.")
		return http.StatusConflict
	}

	if checkNotFound(r) == http.StatusOK && solList[r.COMPONENT].IPAddress != addrs[0] {
		log.Error("IP Address is not the same as the IP from request.")
		return http.StatusConflict
	}

	// r.COMTCP != port führt dazu das der Port von dem aus geschickt mit dem eingangsport der component verglichen wird.
	// Das darf jedoch nicht der gleiche Port sein.
	if r.COMIP != addrs[0] || r.STATUS != strconv.Itoa(200) {
		log.Debug("hier is das problem",
			slog.String("r.COMIP", r.COMIP),
			slog.String("addrs[0]", addrs[0]),
			slog.Int("r.COMTCP", r.COMTCP),
			slog.Int("port", port),
			slog.String("r.STATUS", r.STATUS),
		)
		// Return 409 Conflict
		return http.StatusConflict
	}
	// Return 200 OK
	return http.StatusOK
}

func sendDeleteRequests() {
	for uuid, entry := range solList {
		if uuid == sol.SolUUID {
			continue
		} else if entry.ActiveStatus == utils.Disconnected || entry.ActiveStatus == utils.Left {
			continue
		}

		url := "http://" + entry.IPAddress + ":" + strconv.Itoa(
			entry.Port,
		) + "/vs/v1/system/" + strconv.Itoa(
			uuid,
		) +
			"?star=" + sol.StarUUID
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			log.Error("Failed to create DELETE request", slog.String("error", err.Error()))
			continue
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Error("Failed to send first DELETE request", slog.String("error", err.Error()))
			time.Sleep(10 * time.Second)
			resp, err = client.Do(req)
			if err != nil {
				log.Error("Failed to send second DELETE request", slog.String("error", err.Error()))
				time.Sleep(20 * time.Second)
				resp, err = client.Do(req)
				if err != nil {
					log.Error(
						"Failed to send third DELETE request",
						slog.String("error", err.Error()),
					)
					continue
				}
				continue
			}
			continue
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(resp.Body)

		if resp.StatusCode != http.StatusOK {
			log.Error("Received non-OK response", slog.Int("statusCode", resp.StatusCode))
		} else {
			log.Info("Successfully sent DELETE request", slog.Int("uuid", uuid))
		}
	}
}

func listContains(uuid int) bool {
	// iterate over the array and compare given string to each element
	for _, value := range solList {
		if value.ComUUID == uuid {
			return true
		}
	}
	return false
}
