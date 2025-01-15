## Rest Endpoint von Componenten
- Post /vs/v2/messages/:msgUUID"
Message RequestModel als Json als übergabe.
```json
{
"msgid": "123@COMUUID:STARUUID",
"status": "delivered",
"delivered": "UNIX-Timestamp",
"from": "COMUUID:STARUUID",
"to": "COMUUID:STARUUID",
"received": "UNIX-Timestamp"
}
```
Response von component
200 bei erfolg
400 bei fehler

- Delete /vs/v2/messages/:msgUUID"
MsgUUID als Parameter ausreichen

response 200 bei erfolg
400 bei fehler

- Patch /vs/v2/messages/:msgUUID"
```json
{
"msgid": "123@COMUUID:STARUUID",
"status": "delivered",
"delivered": "UNIX-Timestamp",
"from": "COMUUID:STARUUID",
"to": "COMUUID:STARUUID",
"received": "UNIX-Timestamp"
}
```

Des weiteren muss jede Componente eine Liste über die Messages halten und diese Consistent halten je nach Endpoint Aufruf.