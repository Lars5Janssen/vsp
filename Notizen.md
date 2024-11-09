# Libs fuer Go

[slog: Libary fuer structured logging](https://betterstack.com/community/guides/logging/logging-in-go/) \
slog ist glaube ich fuer uns super geignet. Ist in der Standard Libary drinne. Und kann automatisch an jede logmessage sowas wie die IP des servers, die unterschiedlichen UUIDs der Prozesse oder den Status des Prozesses im Stern anhaengen, ohne das wir das bei jedem Aufruf von logger selber machen muessen
mit [der libary](https://github.com/samber/slog-multi#broadcast-slogmultifanout) um multilog zu ermoeglichen

[Gin: HTTP Framework und JSON parser](https://gin-gonic.com/) \
scheint gut zu sein und arbeit abzunehmen, ohne dass es spring ist. Es ist zwar moeglich einen HTTP Server mit REST API nur mithilfe der stdlib zu erstellen. Aber ich fand den Syntax mit Gin einwenig netter. Es _soll_ woll auch direkt JSON parsen/validieren. Zum Thema JSON hat uns Kossakowski ja allen stark ans herz gelegt schon was bestehendes fuer JSON zu nutzen

# TODO

- [ ] Gin nutzen
- [ ] Auf slog umstellen!
- [ ] Refactor?
- [ ] Workspaces/Go module structure?

# Generell

- Port als 8000+Arbeitsgruppen ID
- Befehele
  - Crash
  - Exit

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
  <STAR-UUID> ::= md5( IP-Adresse von SOL, <ID>, <COM-UUID> von SOL (<ID> meint gruppen id)

## Sol

- Rest API bereitstellen
