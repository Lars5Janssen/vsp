# Generell

- Port als 8000+Arbeitsgruppen ID
- Befehele
  - Crash
  - Exit
- Syslog aus altem Ding

# Was muss wer machen?

## Component

- Rest API callen
- JSON

## Star/SOL Entity

- die eigene <COM-UUID> und ja, die muss selbst berechnet werden
- Zeitstempel der Initiierung des Sterns
- die <STAR-UUID>, die selbst berechnet werden muss
- Anzahl der aktiven Komponenten ::= „1“
- Maximale Anzahl der aktiven Komponenten ::= „4“ (auch das muss ein
Aufrufparameter werden)
<COM-UUID> ::= int( rand( 9000 ) + 1000 ) -> Immer prüfen ob schon verwendet
<STAR-UUID> ::= md5( IP-Adresse von SOL, <ID>, <COM-UUID> von SOL  (<ID> meint gruppen id)

## Sol

- Rest API bereitstellen
