package utils

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// RequestModel
/*
Benutzt für registering a Component and sending a heart beat to Sol
*/
type RequestModel struct { // 1.0
	STAR      string `json:"star" validate:"required"`
	SOL       int    `json:"sol" validate:"required,min=1000,max=9999"`
	COMPONENT int    `json:"component" validate:"required,min=1000,max=9999"`
	COMIP     string `json:"com-ip" validate:"required,ip"`
	COMTCP    int    `json:"com-tcp" validate:"required,min=0,max=65536"`
	STATUS    string `json:"status" validate:"required"`
}

// ResponseModel
/*
Response Model für die Antwort von SOL via Broadcast.
*/
type ResponseModel struct { // Beispiel
	STAR      string `json:"star" validate:"required"`                        // "irgendeinString"
	SOL       int    `json:"sol" validate:"required,min=1000,max=9999"`       // 3842
	SOLIP     string `json:"sol-ip" validate:"required,ip"`                   // "192.118.0.1"
	SOLTCP    int    `json:"sol-tcp" validate:"required,min=0,max=65536"`     // 8139
	COMPONENT int    `json:"component" validate:"required,min=1000,max=9999"` // 7999
}

// MessageRequestModel
/* 2.1
POST /vs/v1/messages/<MSG-UUID> → gibt es die nummer schon 404
 Wenn nachricht akzeptiert wird → setzen von werten
 Hierbei werden ggf. übergebene Werte nicht verwendet:
 „status“ ::= „active“
 „created“ ::= aktueller Zeitstempel in UNIX-Notation6
 „changed“ ::= gleicher Zeitstempel wie „created“
*/
type MessageRequestModel struct { // 2.1
	STAR   string `json:"star" validate:"required"`
	ORIGIN string `json:"origin" validate:"required,origin"`
	// Entweder die COM-UUID die hier eingetragen ist oder der vom Sender
	// wenn leer => 412
	SENDER string `json:"sender" validate:"required"`
	MSGID  string `json:"msg-id"`
	// setzt dann bei der Speicherung die <MSG-ID> auf einen noch nicht vergebenen Wert.
	// Die Version einer Nachricht beginnt immer bei „1“.
	// MSG-UUID := <NUMBER> . „@“ . <COM-UUID> // NUMBER ist ein zähler der Hochgezählt wird.
	VERSION string `json:"version"`
	CREATED string `json:"created"` // Zeitstempel in UNIX notation
	CHANGED string `json:"changed"` // Zeitstempel in UNIX notation
	SUBJECT string `json:"subject"` // UTF-8 wenn leer => 412
	// Dieses wird zwar in beliebiger Länge angenommen, aber bei der Weiterverarbeitung
	// (Weiterleiten, Speicherung, ...) gekürzt, und zwar bis zum ersten NEWLINE-Zeichen
	// Alle „CARRIAGE RETURN“-Zeichen werden vor der weiteren Verarbeitung aus dem Betreff gelöscht.
	MESSAGE string `json:"message"` // UTF-8
}

// Custom validation function to check if a string is either a number or an email
func validateOrigin(fl validator.FieldLevel) bool {
	origin := fl.Field().String()
	numberRegex := regexp.MustCompile(`^\d+$`)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	return numberRegex.MatchString(origin) || emailRegex.MatchString(origin)
}

/*
Dies ist das Model für das Format wie Sol Nachrichten abspeichert.
*/
type MessageModel struct { // 2.1
	MSGID   string `json:"msg-id"  validate:"required"`
	STAR    string `json:"star"  validate:"required"`
	ORIGIN  string `json:"origin"  validate:"required"`
	SENDER  string `json:"sender"  validate:"required"`
	VERSION int    `json:"version"  validate:"required,min=1"`
	CREATED string `json:"created"  validate:"required"` // Zeitstempel in UNIX notation
	CHANGED string `json:"changed"  validate:"required"` // Zeitstempel in UNIX notation
	SUBJECT string `json:"subject"  validate:"required"`
	MESSAGE string `json:"message"`
	STATUS  string `json:"status" validate:"required,oneof=active deleted"`
}

type MessageModelId struct {
	MSGID  string `json:"msg-id" validate:"required,min=1"`
	STATUS string `json:"status" validate:"required,oneof=active deleted"`
}

type MessageListHeader struct {
	STAR         string         `json:"star" validate:"required"`
	TOTALRESULTS int            `json:"total-results" validate:"required,min=0"`
	SCOPE        string         `json:"scope" validate:"required,oneof=active all"`
	VIEW         string         `json:"view" validate:"required,oneof=id header"`
	MESSAGES     []MessageModel `json:"messages" validate:"required"`
}

type MessageListId struct {
	STAR         string           `json:"star" validate:"required"`
	TOTALRESULTS int              `json:"total-results" validate:"required,min=0"`
	SCOPE        string           `json:"scope" validate:"required,oneof=active all"`
	VIEW         string           `json:"view" validate:"required,oneof=id header"`
	MESSAGES     []MessageModelId `json:"messages" validate:"required"`
}

// Validate the struct
func (r *ResponseModel) Validate() error {
	validate := validator.New()
	err := validate.RegisterValidation("origin", validateOrigin)
	if err != nil {
		return err
	}
	return validate.Struct(r)
}
