# Prio
- [X] User Input für Componenten
- [X] Fertige Implementierung der Aufgabe 2
- [X] Log Files
- [X] Überprüfung Aufgabenblatt 1
  - [ ] Watchdog problem. Irgendeine Lösung?
  - [ ] logs dürfen nicht appended werden @Lars
  - [X] disconnect from Sol anschauen. Ticker scheint das Problem zu sein. @Leo
  - [ ] Überprüfung der Funktion mit 2 geräten in einem Netzwerk @Leo
- [ ] Durchlesen Aufgabenblatt 2
  - [ ] vorbereiten Architektonische überlegungen @zis

# Fragen
- Soll es eigene Statusmeldungen geben oder können wir die classy one von http nehmen

# Aufgaben zu 2: SOL
- [ ] Sowohl UDP als auch TCP über GALAXYPORT zuhören und antworten
  - [ ] UDP
  - [ ] REST API
    - POST /vs/v1/star
    - PATCH /vs/v1/star/{staruuid}
    - GET /vs/v1/star/{staruuid}
    - GET /vs/v1/star
    - DELETE /vs/v1/star/{staruuid}

    - POST /vs/v**2**/messages/{msgid}
    - GET /vs/v**2**/messages/{msgid}
    - DELETE /vs/v**2**/messages/{msgid}?star={staruuid}
    - GET /vs/v**2**/messages?star={staruuid|all}&scope={all|active}&view={header|id}
      - heißt es hier "view" statt "info"?
      - "all" => alle nachrichten von allen sternen?
        - nur die die von anderen Sternen eingegangen sind.
      - bei staruuid NUR welche die bei dem Stern eingegangen sind?
      - 
      
  - [ ] GALAXYPORT wird dem Programm übergeben
- [ ] SOL muss nach der Initialisierung einen Broadcast mit "HELLO? | I AM STARUUID" an GALAXYPORT per UDP senden
- [ ] Beim Abmelden werden auch andere Sterne benachrichtigt wie Komponenten

- [ ] Nachrichten müssen leicht verändert werden, bevor sie weitergeleitet werden an andere Sterne
  - ORIGIN ::= COMUUID + ":" + STARUUID | EMAIL
  - MSGUUID ::= NUMBER + "@" COMUUID + ":" + STARUUID
- [ ] Nachrichten bekommen vier neue Angaben bei der Speicherung
  - from-star: STARUUID
  - to-star: STARUUID
  - received: aktueller Zeitstempel in UNIX-Notation
  - delivered: aktueller Zeitstempel in UNIX-Notation
- [ ] Änderungen der Nachricht bei Löschung
  - status: "deleted by STARUUID" | "deleted by us from COMUUID"
  - delivered: Zeitstempel des Auftrags
  - removed: aktueller Zeitstempel in UNIX-Notation
- [ ] Nachrichten werden beim GET Listen Aufruf anders zurückgegeben
  - [ ] Kurzformat:
    - msgid
    - status
    - delivered als Array von STARUUIDs
  - [ ] Langformat:
    - msgid
    - status
    - delivered als Array von STARUUIDs und Zeitstempeln
    - from
    - to-star
    - received