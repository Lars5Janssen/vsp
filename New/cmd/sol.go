package cmd

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
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

var sol Sol

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
	sol = initializeSol(log, ctx)

	// Retrieve from channels:
	// command := <-commands
	// udpInput := <-udp

	// forever loop for API
	for {
		n.AttendHTTP(log, restIn, restOut, solEndpoints)
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
	err := response.Context.BindJSON(&registerRequestModel)
	// If the JSON is not valid, return 401 Unauthorized
	if err != nil {
		return n.RestOut{http.StatusUnauthorized, nil}
	}

	// Check if all the info from the component is correct
	checkInfoCorrect(registerRequestModel)

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

func initializeSol(log *slog.Logger, ctx context.Context) Sol {
	var sol Sol

	// ComUUID from Sol
	number, err := generateComUUID()
	if err != nil {
		log.Error("Error generating comUUID for Sol")
		return sol
	}
	sol.SolUUID = number

	// IpAddress and Port from Sol
	sol.IPAddress = ctx.Value("ip").(string)
	intValue, ok := ctx.Value("port").(int)
	if !ok {
		log.Error("Type conversion failed for port")
		return sol
	}

	// StarUUID from Sol
	sol.Port = intValue
	hashNumber, err := generateStarUUID(sol.IPAddress, sol.Port, strconv.Itoa(sol.SolUUID), *log)
	if err != nil {
		log.Error("Error generating starUUID")
		return sol
	}
	sol.StarUUID = hashNumber

	return sol
}

func generateComUUID() (int, error) {
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(9000))
	if err != nil {
		return 0, err
	}
	return int(randomNumber.Int64() + 1000), nil
}

func generateStarUUID(ip string, port int, solUUID string, log slog.Logger) (string, error) {
	// TODO ID equals last two digits of port?
	// Get the last two digits of the port
	portStr := strconv.Itoa(port)
	if len(portStr) < 2 {
		log.Error("port number is too short")
		return "", nil
	}
	lastTwoDigits := portStr[len(portStr)-2:]

	// Concatenate the IP address, last two digits of the port, and solUUID
	concatenated := ip + lastTwoDigits + solUUID

	// Generate the MD5 hash
	hash := md5.Sum([]byte(concatenated))

	// Convert the hash to a hexadecimal string
	hashStr := hex.EncodeToString(hash[:])

	return hashStr, nil
}

func checkInfoCorrect(registerRequestModel utils.RegisterRequestModel) int {
	if registerRequestModel.STAR != sol.StarUUID { // || registerRequestModel.SOL != sol.SolUUID {
		// Return 401 Unauthorized
		return 401
	}
	return 200
}
