package utils

// TODO better name and only one model
type RegisterRequestModel struct { // 1.0
	STAR      string `json:"star"`
	SOL       int    `json:"sol"`
	COMPONENT int    `json:"component"`
	COMIP     string `json:"comIp"`
	COMTCP    int    `json:"comTcp"`
	STATUS    int    `json:"status"`
}

type HeartBeatRequestModel struct { // 1.1
	STAR      string `json:"star"`
	SOL       int    `json:"sol"`
	COMPONENT int    `json:"component"`
	COMIP     string `json:"comIp"`
	COMTCP    int    `json:"comTcp"`
	STATUS    int    `json:"status"`
}

// MessageRequestModel
/* 2.1
POST /vs/v1/messages/<MSG-UUID> => gibt es die nummer schon 404
 Wenn nachricht akzeptiert wird => setzen von werten
 Hierbei werden ggf. übergebene Werte nicht verwendet:
 „status“ ::= „active“
 „created“ ::= aktueller Zeitstempel in UNIX-Notation6
 „changed“ ::= gleicher Zeitstempel wie „created“
*/
type MessageRequestModel struct { // 2.1
	STAR   string "STAR-UUID"
	ORIGIN string "COM-UUID | EMAIL"
	// Entweder die COM-UUID die hier eingetragen ist oder der vom Sender
	// wenn leer => 412
	SENDER string "SENDER-UUID | '' "
	MSGID  string "MSG-UUID | '' "
	// setzt dann bei der Speicherung die <MSG-UUID> auf einen noch nicht vergebenen Wert.
	// Die Version einer Nachricht beginnt immer bei „1“.
	// MSG-UUID := <NUMBER> . „@“ . <COM-UUID> // NUMBER ist ein zähler der Hochgezählt wird.
	VERSION string "'1' | '' "
	CREATED string "<TIMESTAMP>"
	CHANGED string "<TIMESTAMP>"
	SUBJECT string "<STRING> | UTF-8" // wenn leer => 412
	// Dieses wird zwar in beliebiger Länge angenommen, aber bei der Weiterverarbeitung
	// (Weiterleiten, Speicherung, ...) gekürzt, und zwar bis zum ersten NEWLINE-Zeichen
	// Alle „CARRIAGE RETURN“-Zeichen werden vor der weiteren Verarbeitung aus dem Betreff gelöscht.
	MESSAGE string "<STRING> | UTF-8"
}

type MessageModel struct { // 2.1
	MSGID   string `json:"msgId"`
	STAR    string `json:"star"`
	ORIGIN  string `json:"origin"`
	SENDER  string `json:"sender"`
	VERSION string `json:"version"`
	CREATED string `json:"created"`
	CHANGED string `json:"changed"`
	SUBJECT string `json:"subject"`
	MESSAGE string `json:"message"`
	STATUS  string `json:"status"`
}

type MessageModelId struct {
	MSGID  string `json:"msgId"`
	STATUS string `json:"status"`
}

type Response struct {
	STAR      string `json:"star"`
	SOL       int    `json:"sol"`
	COMPONENT int    `json:"component"`
	SOLIP     string `json:"solIp"`
	SOLTCP    int    `json:"solTcp"`
}

type MessageListHeader struct {
	STAR         string         `json:"star"`
	TOTALRESULTS int            `json:"totalResults"`
	SCOPE        string         `json:"scope"`
	VIEW         string         `json:"view"`
	MESSAGES     []MessageModel `json:"messages"`
}

type MessageListId struct {
	STAR         string           `json:"star"`
	TOTALRESULTS int              `json:"totalResults"`
	SCOPE        string           `json:"scope"`
	VIEW         string           `json:"view"`
	MESSAGES     []MessageModelId `json:"messages"`
}
