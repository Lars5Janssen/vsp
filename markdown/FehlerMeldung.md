So ich habe nochmal rumprobiert.

## Docker:

Die *Registrierung* einer Componente bei Sol funktioniert, die Componente gibt aus:

time=2024-12-04T10:25:24.900Z level=INFO msg="Successfully registered by Sol" Component=Component

Der *HeartBeat* Von Sol *an* Sol = läuft

Der *HeartBeat* Von Componente *an* Sol
Hier liegt unser Problem. Wir bekommen erst von GIN-debug die Meldung:

		Direkt nach der Endpunkt erstellung
	[GIN-debug] Listening and serving HTTP on 172.17.0.3:8006
    [GIN-debug] [ERROR] listen tcp 172.17.0.3:8006: bind: address already in use
	level=ERROR msg="listen tcp 172.17.0.3:8006: bind: address already in use" Component=TCP
	
		Dann kommt die Erfolgreiche Registrierungs Nachricht:
	time=2024-12-04T10:25:24.900Z level=INFO msg="Successfully registered by Sol" Component=Component
	
	Und nach dem versuch einen Heartbeat zu senden:
	
	level=ERROR msg="Failed to send heartbeat to SOL:172.17.0.2:8006" Component=Component 
	error="Patch \"http://172.17.0.2:8006/vs/v1/system/2840\": dial tcp :8006->172.17.0.2:8006: bind: address already in use"

## Zwei Geräte im gleichen Netzwerk:

Sol started wie gewohnt.
Componente startet wie gewohnt wie in der Docker umgebung (gleiche meldungen)

Die *Registrierung* einer Componente bei Sol funktioniert *nicht*.

		Es kommt:
	evel=ERROR msg="Failed to send request to SOL: " Component=Component 
	error="Post \"http://192.168.178.167:8006/vs/v1/system\": dial tcp :8006->192.168.178.167:8006: bind: address already in use"
		
	Das kommt uns doch irgendwie bekannt vor. 
	Kann mir aber jemand beantworten warum die Registrierung unter Docker geht, 
	aber mit zwei Geräten in einem Netzwerk nicht?

Der *Heartbeat* verhält sich wie ihr es euch vorstellen könnt:

	Von Sol zu Sol alles fein
	
	Von Componente zu Sol:
	
	level=ERROR msg="Failed to send heartbeat to SOL:192.168.178.167:8006" Component=Component 
	error="Patch \"http://192.168.178.167:8006/vs/v1/system/4317\": dial tcp :8006->192.168.178.167:8006: bind: address already in use"

	
	
 
 
