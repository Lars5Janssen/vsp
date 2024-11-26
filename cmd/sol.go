package cmd

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	n "github.com/Lars5Janssen/vsp/net"
	"github.com/Lars5Janssen/vsp/utils"
)

type Sol struct {
	StarUUID  string
	SolUUID   int
	IPAddress string
	Port      int
}

type ComponentEntry struct {
	ComUUID         int
	IPAddress       string
	Port            int
	TimeIntegration time.Time
	TimeInteraktion time.Time
	Status          ComponentStatus
	ActiveStatus    ActiveStatus
}

type ComponentStatus int

const (
	OK                 ComponentStatus = 200
	Unauthorized       ComponentStatus = 401
	Forbidden          ComponentStatus = 403
	NotFound           ComponentStatus = 404
	Conflict           ComponentStatus = 409
	PreconditionFailed ComponentStatus = 412
)

type ActiveStatus string

const (
	Active       ActiveStatus = "active"
	Left         ActiveStatus = "left"
	Disconnected ActiveStatus = "disconnected"
)

var log *slog.Logger

var sol Sol

var solList = map[int]ComponentEntry{}

var lenOfSolList int

func StartSol(
	ctx context.Context,
	logger *slog.Logger,
	commands chan string,
	udp chan n.UDP,
	restIn chan n.RestIn,
	restOut chan n.RestOut,
) {
	log = logger
	log = log.With(slog.String("Component", "SOL"))
	log.Info("Starting as SOL")

	// SOL Logic
	initializeSol(log, ctx)

	// Add to solList
	solList[sol.SolUUID] = ComponentEntry{
		ComUUID:         sol.SolUUID,
		IPAddress:       sol.IPAddress,
		Port:            sol.Port,
		TimeIntegration: time.Now(),
		TimeInteraktion: time.Now(),
		Status:          OK,
		ActiveStatus:    Active,
	}

	// Max active components
	temp := ctx.Value("maxActiveComponents")
	lenOfSolList = temp.(int)

	// has to be done outside for loop
	go n.AttendHTTP(log, restIn, restOut, solEndpoints)

	// Check if the components are still active
	ticker := time.NewTicker(5 * time.Second) // TODO check every 5 seconds or 1 second?
	go func() {
		for {
			select {
			case <-ticker.C:
				checkInteractionTimes()
			}
		}
	}()

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
				response := utils.Response{
					STAR:      sol.StarUUID,
					SOL:       sol.SolUUID,
					COMPONENT: intValue,
					SOLIP:     sol.IPAddress,
					SOLTCP:    sol.Port,
				}

				response.COMPONENT = 2000 // TODO only for test purposes

				marshal, err := json.Marshal(response)
				if err != nil {
					log.Error("Error marshalling response", slog.String("error", err.Error()))
					return
				}
				log.Info("Sending response to HELLO?", slog.String("response", response.STAR))
				err = n.SendMessage(log, udpInput.Addr, sol.Port, string(marshal))
				if err != nil {
					return
				}
			}
		default:
		}

	}
}

func checkInteractionTimes() {
	for uuid, entry := range solList {
		if entry.ActiveStatus == Active {
			if time.Since(entry.TimeInteraktion) > 60*time.Second {
				err := sendRequestsToActiveComponents(uuid)
				if err != nil {
					return
				}
				break
			}
		}
	}
}

func sendRequestsToActiveComponents(uuid int) error {
	entry := solList[uuid]
	url := "http://" + entry.IPAddress + ":" + strconv.Itoa(entry.Port) + "/vs/v1/system/" + strconv.Itoa(
		uuid) + "?star=" + sol.StarUUID
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error("Failed to create GET request", slog.String("error", err.Error()))
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Failed to send GET request", slog.String("error", err.Error()))
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error("Failed to close response body", slog.String("error", err.Error()))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Error("Received non-OK response", slog.Int("statusCode", resp.StatusCode))
		entry.ActiveStatus = Disconnected
		entry.Status = ComponentStatus(resp.StatusCode)
	} else {
		log.Info("Successfully sent GET request", slog.Int("uuid", uuid))
		var heartBeatRequestModel utils.HeartBeatRequestModel
		err := json.NewDecoder(resp.Body).Decode(&heartBeatRequestModel)
		if err != nil {
			log.Error("Failed to decode response body", slog.String("error", err.Error()))
			entry.Status = ComponentStatus(resp.StatusCode)
		} else {
			entry.Status = ComponentStatus(heartBeatRequestModel.STATUS)
			log.Info("Successfully parsed response body", slog.Any("heartBeatRequestModel", heartBeatRequestModel))
		}
	}
	entry.TimeInteraktion = time.Now()
	solList[uuid] = entry
	return err
}

/*
createComponent - Seite 4 im Aufgabenblatt Aufgabe 1.0

Bitte das RegisterRequestModel nutzen.

Damit SOL den auch kennt, gibt die neue Komponente selbst ihren derzeitigen Status an. Außerdem werden Angaben zum Stern
übermittelt, die von SOL vor der Integration auch geprüft werden, damit sichergestellt ist, dass auch der „richtige“
Stern gemeint ist.
*/
func registerComponentBySol(response n.RestIn) n.RestOut {
	var registerRequestModel utils.RegisterRequestModel
	// err := response.Context.BindJSON(&registerRequestModel)
	err := response.Context.ShouldBindJSON(&registerRequestModel)
	// If the JSON is not valid, return 401 Unauthorized
	if err != nil {
		return n.RestOut{http.StatusUnauthorized, nil}
	}

	// Check if all the info from the component is correct
	if checkConflict(registerRequestModel, response.IpAndPort) != OK {
		return n.RestOut{http.StatusConflict, nil}
	} else if checkUnauthorized(registerRequestModel) != OK {
		return n.RestOut{http.StatusUnauthorized, nil}
	} else if checkNoRoomLeft() != OK {
		return n.RestOut{http.StatusForbidden, nil}
	} else if checkNotFound(registerRequestModel) == OK { // If the component is already in the list, return 409 Conflict
		return n.RestOut{http.StatusConflict, nil}
	}

	// Add the component to the list
	solList[registerRequestModel.COMPONENT] = ComponentEntry{
		ComUUID:         registerRequestModel.COMPONENT,
		IPAddress:       registerRequestModel.COMIP,
		Port:            registerRequestModel.COMTCP,
		TimeIntegration: time.Now(),
		TimeInteraktion: time.Now(),
		Status:          OK,
		ActiveStatus:    Active,
	}

	return n.RestOut{http.StatusOK, nil}
}

/*
Das soll die checkAvailabilityFromComponent Methode übernehmen - Aufgabe 1.1 Pflege des Sterns

Jede aktive Komponente baut alle 30 Sekunden eine UNICAST-Verbindung zum
<STARPORT>/tcp von SOL auf.2 Wenn SOL nicht erreichbar ist, wird es nach 10 bzw. 20
Sekunden nochmal versucht. Wenn dann immer noch keine Verbindung zustande
kommt, beendet sich die Komponente selbst.
*/
func checkAvailabilityFromComponent(response n.RestIn) n.RestOut {
	var registerRequestModel utils.RegisterRequestModel
	err := response.Context.ShouldBindJSON(&registerRequestModel)
	if err != nil {
		// Return 400 Bad Request if JSON is not valid
		return n.RestOut{http.StatusBadRequest, nil}
	}

	// Check if info correct
	if checkNotFound(registerRequestModel) != OK {
		return n.RestOut{http.StatusNotFound, nil}
	} else if checkUnauthorized(registerRequestModel) != OK {
		return n.RestOut{http.StatusUnauthorized, nil}
	} else if checkConflict(registerRequestModel, response.EndpointAddr) != OK {
		return n.RestOut{http.StatusConflict, nil}
	}

	// Update the time of interaction
	if entry, exists := solList[registerRequestModel.COMPONENT]; exists {
		entry.ActiveStatus = Active
		entry.TimeInteraktion = time.Now()
		solList[registerRequestModel.COMPONENT] = entry
	}

	return n.RestOut{http.StatusOK, nil}
}

func sendHeartBeatBack(response n.RestIn) n.RestOut {
	if sol.StarUUID != response.Context.Query("star") {
		return n.RestOut{http.StatusUnauthorized, nil}
	}
	uuid := response.Context.Param("comUUID?star=starUUID")
	if uuid == "" {
		return n.RestOut{http.StatusConflict, nil}
	}
	comUuid, _ := strconv.Atoi(uuid)
	if sol.SolUUID != comUuid {
		return n.RestOut{http.StatusConflict, nil}
	}

	heartBeatRequestModel := utils.HeartBeatRequestModel{
		STAR:      sol.StarUUID,
		SOL:       sol.SolUUID,
		COMPONENT: sol.SolUUID,
		COMIP:     sol.IPAddress,
		COMTCP:    sol.Port,
		STATUS:    int(OK),
	}

	return n.RestOut{http.StatusOK, heartBeatRequestModel}
}

/*
1.2 Pflege des Sterns – Abmelden einer Komponente

Eine aktive Komponente, die sich nach einem „EXIT“-Befehl bei SOL abmeldet, baut eine UNICAST-Verbindung auf.
Wenn SOL nicht erreichbar ist, wird es nach 10 bzw. 20 Sekunden nochmal versucht. Wenn dann immer noch keine Verbindung
zustande kommt, beendet sich die Komponente selbst.
*/
func disconnectComponentFromStar(response n.RestIn) n.RestOut {
	var out n.RestOut
	var registerRequestModel utils.RegisterRequestModel

	registerRequestModel.STAR = response.Context.Query("star")
	stringValue := response.Context.Param("comUUID?star=starUUID")
	comUUid, err := strconv.Atoi(stringValue)
	if err != nil {
		out.StatusCode = http.StatusBadRequest
	} else {
		registerRequestModel.COMPONENT = comUUid
		if checkNotFound(registerRequestModel) != OK {
			out.StatusCode = http.StatusNotFound
		} else {
			comEntry := solList[registerRequestModel.COMPONENT]
			registerRequestModel.COMIP = comEntry.IPAddress
			registerRequestModel.COMTCP = comEntry.Port
			registerRequestModel.STATUS = int(comEntry.Status)
			registerRequestModel.SOL = sol.SolUUID

			if checkUnauthorized(registerRequestModel) != OK {
				out.StatusCode = http.StatusUnauthorized
			} else {
				out.StatusCode = http.StatusOK
			}
		}
	}
	// Update the active status of the component
	if entry, exists := solList[registerRequestModel.COMPONENT]; exists {
		entry.ActiveStatus = Left
		entry.TimeInteraktion = time.Now()
		solList[registerRequestModel.COMPONENT] = entry
	}

	out.Body = nil
	return out
}

/*
createAndSaveMessage Aufgabe 2.1
*/
func createAndSaveMessage(response n.RestIn) n.RestOut {
	var message utils.MessageRequestModel
	err := response.Context.ShouldBindJSON(&message)
	if err != nil {
		return n.RestOut{http.StatusBadRequest, nil}
	}
	if message.STAR != sol.StarUUID {
		return n.RestOut{http.StatusUnauthorized, nil}
	} else if message.ORIGIN == "" || message.SUBJECT == "" || !utf8.ValidString(message.ORIGIN) || !utf8.ValidString(message.SUBJECT) {
		return n.RestOut{http.StatusPreconditionFailed, nil}
	}
	subject := strings.Split(message.SUBJECT, "\n")[0]
	subject = strings.ReplaceAll(subject, "\r", "")
	// TODO create msgUUID and create nonce counter
	// TODO save message

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

func initializeSol(log *slog.Logger, ctx context.Context) {
	// ComUUID from Sol
	number, err := generateComUUID()
	if err != nil {
		log.Error("Error generating comUUID for Sol")
		return
	}
	sol.SolUUID = number

	sol.SolUUID = 1000 // TODO only for test purposes

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

	sol.StarUUID = "testStarUUID" // TODO only for test purposes

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

func checkUnauthorized(r utils.RegisterRequestModel) ComponentStatus {
	if r.STAR != sol.StarUUID || r.SOL != sol.SolUUID {
		// Return 401 Unauthorized
		return Unauthorized
	}
	// Return 200 OK
	return OK
}

func checkNoRoomLeft() ComponentStatus {
	count := 0
	for _, entry := range solList {
		if entry.ActiveStatus == Active {
			count++
		}
	}
	if count >= lenOfSolList {
		// Return 403 No room left
		return Forbidden
	}
	// Return 200 OK
	return OK
}

func checkNotFound(r utils.RegisterRequestModel) ComponentStatus {
	if !listContains(r.COMPONENT) {
		// Return 404 Not Found
		return NotFound
	}
	// Return 200 OK
	return OK
}

func checkConflict(r utils.RegisterRequestModel, addr string) ComponentStatus {
	// TODO check if the IP address and port are correct (no idea which port is gonna be used) cannot test
	addrs := strings.Split(addr, ":")
	port, err := strconv.Atoi(addrs[1])
	// TODO remove port != 0
	if err != nil || port == -1 {
		return Conflict
	}
	// TODO because i cant test it rn, i will just return 200 OK
	/*if r.COMIP != addrs[0] || solList[r.COMPONENT].IPAddress != addrs[0] || r.COMTCP != port || r.STATUS != 200 {
		// Return 409 Conflict
		return Conflict
	}*/
	// Return 200 OK
	return OK
}

func sendDeleteRequests() {
	for uuid, entry := range solList {
		if uuid == sol.SolUUID {
			continue
		} else if entry.ActiveStatus == Disconnected || entry.ActiveStatus == Left {
			continue
		}

		url := "http://" + entry.IPAddress + ":" + strconv.Itoa(entry.Port) + "/vs/v1/system/" + strconv.Itoa(uuid) +
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
					log.Error("Failed to send third DELETE request", slog.String("error", err.Error()))
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