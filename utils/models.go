package utils

// TODO better name and only one model
type RegisterRequestModel struct { // 1.0
	STAR      string `json:"star"`
	SOL       int    `json:"sol"`
	COMPONENT int    `json:"component"`
	COMIP     string `json:"comip"`
	COMTCP    int    `json:"comtcp"`
	STATUS    int    `json:"status"`
}

type HeartBeatRequestModel struct { // 1.1
	STAR      string `json:"star"`
	SOL       int    `json:"sol"`
	COMPONENT int    `json:"component"`
	COMIP     string `json:"comip"`
	COMTCP    int    `json:"comtcp"`
	STATUS    int    `json:"status"`
}

// POST /vs/v1/messages/<MSG-UUID> => gibt es die nummer schon 404
// Wenn nachricht akzeptiert wird => setzen von werten
// Hierbei werden ggf. übergebene Werte nicht verwendet:
// „status“ ::= „active“
// „created“ ::= aktueller Zeitstempel in UNIX-Notation6
// „changed“ ::= gleicher Zeitstempel wie „created“
type MessageRequestModel struct { // 2.1
	STAR    string "STAR-UUID"
	ORIGIN  string "COM-UUID | EMAIL"
	SENDER  string "SENDER-UUID | '' "
	MSGID   string "MSG-UUID | '' "
	VERSION string "'1' | '' "
	CREATED string "<TIMESTAMP>"
	CHANGED string "<TIMESTAMP>"
	SUBJECT string "<STRING> | UTF-8"
	MESSAGE string "<STRING> | UTF-8"
}

type MessageModel struct { // 2.1
	STAR    string "STAR-UUID"
	ORIGIN  string "COM-UUID | EMAIL"
	SENDER  string "SENDER-UUID | '' "
	VERSION string "'1' | '' "
	CREATED string "<TIMESTAMP>"
	CHANGED string "<TIMESTAMP>"
	SUBJECT string "<STRING> | UTF-8"
	MESSAGE string "<STRING> | UTF-8"
	STATUS  string "'active' | ''"
}

type Response struct {
	STAR      string "STAR-UUID"
	SOL       int    "COM-UUID"
	COMPONENT int    "COM-UUID"
	SOLIP     string "IP"
	SOLTCP    int    "PORT"
}
