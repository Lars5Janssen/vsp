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
	"time"

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
}

type ComponentStatus int

var log *slog.Logger

var sol Sol

var solList = map[int]ComponentEntry{}

// TODO only active components count
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
		Status:          200,
	}

	// Max active components
	temp := ctx.Value("maxActiveComponents")
	lenOfSolList = temp.(int)

	log.Info("Data on SOl",
		slog.String("STARUUID", sol.StarUUID),
		slog.Int("SOLUUID", sol.SolUUID),
		slog.String("SOLIP", sol.IPAddress),
		slog.Int("SOLPORT", sol.Port),
	)

	// has to be done outside for loop
	go n.AttendHTTP(log, restIn, restOut, solEndpoints)

	// forever loop for commands and udp messages
	for {
		// Retrieve from channels:
		select {
		case command := <-commands:
			if command == "exit" {
				sendDeleteRequests()
				return
			}
		default:

		}
		select {
		case udpInput := <-udp:
			// to test: echo HELLO? | ncat -u 255.255.255.255 8006
			log.Debug("Received UDP message", slog.String("message", udpInput.Message))
			if udpInput.Message == "HELLO?" {
				// TODO only for test purposes
				/*intValue, err := generateComUUID()
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
				}*/
				response := utils.Response{
					STAR:      sol.StarUUID,
					SOL:       sol.SolUUID,
					COMPONENT: 2000,
					SOLIP:     sol.IPAddress,
					SOLTCP:    sol.Port,
				}
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
	if checkConflict(registerRequestModel) != 200 {
		return n.RestOut{http.StatusConflict, nil}
	} else if checkUnauthorized(registerRequestModel) != 200 {
		return n.RestOut{http.StatusUnauthorized, nil}
	} else if checkNoRoomLeft() != 200 {
		return n.RestOut{http.StatusForbidden, nil}
	}

	// Add the component to the list
	solList[registerRequestModel.COMPONENT] = ComponentEntry{
		ComUUID:         registerRequestModel.COMPONENT,
		IPAddress:       registerRequestModel.COMIP,
		Port:            registerRequestModel.COMTCP,
		TimeIntegration: time.Now(),
		TimeInteraktion: time.Now(),
		Status:          200,
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
	if checkConflict(registerRequestModel) != 200 {
		return n.RestOut{http.StatusConflict, nil}
	} else if checkUnauthorized(registerRequestModel) != 200 {
		return n.RestOut{http.StatusUnauthorized, nil}
	} else if checkNotFound(registerRequestModel) != 200 {
		return n.RestOut{http.StatusNotFound, nil}
	}

	// Update the time of interaction
	if entry, exists := solList[registerRequestModel.COMPONENT]; exists {
		entry.TimeInteraktion = time.Now()
		solList[registerRequestModel.COMPONENT] = entry
	}

	return n.RestOut{http.StatusOK, nil}
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

func initializeSol(log *slog.Logger, ctx context.Context) {
	// ComUUID from Sol
	// TODO only for test purposes
	/*number, err := generateComUUID()
	if err != nil {
		log.Error("Error generating comUUID for Sol")
		return
	}*/
	// sol.SolUUID = number
	sol.SolUUID = 1000

	// IpAddress and Port from Sol
	sol.IPAddress = ctx.Value("ip").(string)
	intValue := ctx.Value("port").(int)
	sol.Port = intValue

	// StarUUID from Sol
	// TODO only for test purposes
	/*hashNumber, err := generateStarUUID(*log)
	if err != nil {
		log.Error("Error generating starUUID")
		return
	}
	sol.StarUUID = hashNumber*/
	sol.StarUUID = "testStarUUID"

	return
}

func generateComUUID() (int, error) {
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(9000))
	if err != nil {
		return 0, err
	}
	return int(randomNumber.Int64() + 1000), nil
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

func checkUnauthorized(r utils.RegisterRequestModel) int {
	if r.STAR != sol.StarUUID || r.SOL != sol.SolUUID {
		// Return 401 Unauthorized
		return 401
	}
	// Return 200 OK
	return 200
}

func checkNoRoomLeft() int {
	if len(solList) >= lenOfSolList {
		// Return 403 No room left
		return 403
	}
	// Return 200 OK
	return 200
}

func checkNotFound(r utils.RegisterRequestModel) int {
	if !listContains(r.COMPONENT) {
		// Return 404 Not Found
		return 404
	}
	// Return 200 OK
	return 200
}

func checkConflict(r utils.RegisterRequestModel) int {
	if r.COMIP != sol.IPAddress || r.COMTCP != sol.Port || r.STATUS != 200 {
		// Return 409 Conflict
		return 409
	}
	// Return 200 OK
	return 200
}

func sendDeleteRequests() {
	for uuid, entry := range solList {
		if uuid == sol.SolUUID {
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
			log.Error("Failed to send DELETE request", slog.String("error", err.Error()))
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
