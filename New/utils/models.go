package utils

type RegisterRequestModel struct { // 1.0
	STAR      string `json:"STAR-UUID" binding:"required"`
	SOL       string `json:"COM-UUID" binding:"required"`
	COMPONENT string `json:"COM-UUID" binding:"required"`
	COMIP     int    `json:"IP" binding:"required"`
	COMTCP    int    `json:"PORT" binding:"required"`
	STATUS    int    `json:"STATUS" binding:"required"`
}

type HeartBeatRequestModel struct { // 1.1
	STAR      string `json:"STAR-UUID"`
	SOL       string `json:"COM-UUID"`
	COMPONENT string `json:"COM-UUID"`
	COMIP     int    `json:"IP"`
	COMTCP    int    `json:"PORT"`
	STATUS    int    `json:"STATUS"`
}

// POST /vs/v1/messages/<MSG-UUID> => gibt es die nummer schon 404
// Wenn nachricht akzeptiert wird => setzen von werten
// Hierbei werden ggf. übergebene Werte nicht verwendet:
// „status“ ::= „active“
// „created“ ::= aktueller Zeitstempel in UNIX-Notation6
// „changed“ ::= gleicher Zeitstempel wie „created“
type MessageRequestModel struct { // 2.1
	STAR   string `json:"STAR-UUID" binding:"required"`
	ORIGIN string `json:"COM-UUID | EMAIL" binding:"required"`
	// TODO Entweder die COM-UUID die hier eingetragen ist oder der vom Sender
	// TODO wenn leer => 422
	SENDER string `json:"SENDER-UUID | '' " binding:"required"`
	MSGID  string `json:"SENDER-UUID | '' " binding:"required"`
	// TODO setzt dann bei der Speicherung die <MSG-UUID> auf einen noch nicht vergebenen Wert.
	// TODO Die Version einer Nachricht beginnt immer bei „1“.
	// TODO MSG-UUID := <NUMBER> . „@“ . <COM-UUID> // NUMBER ist ein zähler der Hochgezählt wird.
	VERSION string `json:"'1' | '' " binding:"required"`
	CREATED string `json:"<TIMESTAMP>" binding:"required"`
	CHANGED string `json:"<TIMESTAMP>" binding:"required"`
	SUBJECT string `json:"<STRING> | UTF-8" binding:"required"` // TODO wenn leer => 422
	// TODO Dieses wird zwar in beliebiger Länge angenommen, aber bei der Weiterverarbeitung
	// TODO (Weiterleiten, Speicherung, ...) gekürzt, und zwar bis zum ersten NEWLINE-Zeichen
	// TODO Alle „CARRIAGE RETURN“-Zeichen werden vor der weiteren Verarbeitung aus dem Betreff gelöscht.
	MESSAGE string `json:"<STRING> | UTF-8"`
}

type Response struct {
}
