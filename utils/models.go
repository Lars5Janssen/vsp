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

// RegisterSolModel
/*
Wird genutzt um ein SOL wenn <STAR-UUID> aus dem UDP Paket noch nicht bekannt ist.
Dies ist das Model das von allen Empfangenden SOL's geschickt wird damit die neue SOL alle Anderen kennt.
*/
type RegisterSolModel struct {
	STAR   string `json:"starUUID" validate:"required"`
	SOL    int    `json:"comUUID" validate:"required,min=1000,max=9999"`
	SOLIP  string `json:"sol-ip" validate:"required,ip"`
	SOLTCP int    `json:"sol-tcp" validate:"required,min=0,max=65536"`
	NOCOM  int    `json:"no-com" validate:"required"`
	STATUS string `json:"status" validate:"required"`
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
	STAR    string `json:"star" validate:"required"`
	ORIGIN  string `json:"origin" validate:"required,origin"`
	SENDER  string `json:"sender" validate:"required"`
	MSGID   string `json:"msg-id"`
	VERSION string `json:"version"`
	CREATED string `json:"created"`                             // Zeitstempel in UNIX notation
	CHANGED string `json:"changed"`                             // Zeitstempel in UNIX notation
	SUBJECT string `json:"subject,subject" validate:"required"` // UTF-8 wenn leer => 412
	MESSAGE string `json:"message"`                             // UTF-8
}

/*
Dies ist das Model für das Format wie Sol Nachrichten abspeichert.
*/
type MessageModel struct { // 2.1
	MSGID   string `json:"msg-id"  validate:"required"`
	STAR    string `json:"star"  validate:"required"`
	ORIGIN  string `json:"origin"  validate:"required"`
	SENDER  string `json:"sender"  validate:"required"`
	VERSION string `json:"version"  validate:"required,min=1"`
	CREATED string `json:"created"  validate:"required"` // Zeitstempel in UNIX notation
	CHANGED string `json:"changed"  validate:"required"` // Zeitstempel in UNIX notation
	SUBJECT string `json:"subject"  validate:"required"`
	MESSAGE string `json:"message"`
	STATUS  string `json:"status" validate:"required,oneof=active deleted"`
}

type MessageId struct {
	MSGID string `json:"msg-id" validate:"required"`
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

type GeneralResponse struct {
	STATUSCODE   int         `json:"http-status-code" validate:"required"`
	RESPONSEBODY interface{} `json:"interface" validate:"required"`
}

// Validate the ResponseModel struct
func (r *ResponseModel) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate the MessageRequestModel struct
func (r *MessageRequestModel) Validate() error {
	validate := validator.New()
	err := validate.RegisterValidation("origin", validateOrigin)
	if err != nil {
		return err
	}
	err = validate.RegisterValidation("subject", validateSubject)
	if err != nil {
		return err
	}
	return validate.Struct(r)
}

// Custom validation function to check if a string is either a number or an email => 412
func validateOrigin(fl validator.FieldLevel) bool {
	origin := fl.Field().String()
	numberRegex := regexp.MustCompile(`^\d+$`)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	return numberRegex.MatchString(origin) || emailRegex.MatchString(origin)
}

// Sind „Origin“ oder das Subject der Nachricht leer... => 412
func validateSubject(fl validator.FieldLevel) bool {
	subject := fl.Field().String()
	return subject == ""
}
