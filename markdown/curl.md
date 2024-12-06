# Anfragen über curl

## Aufgabe 1
### DELETE
```bash
curl -X DELETE http://172:17:0:2:8006/vs/v1/system/1000
```

## Aufgabe 2
### POST message
```bash
curl -X POST http://172.17.0.2:8006/vs/v1/messages -H "Content-Type: application/json" -d '{
"STAR": "4a1a1c5a0a7fe7a1ea7d2754b516e9fd",
"ORIGIN": "msg@you.com",
"SENDER": "1000",
"MSGID": "",
"VERSION": "",
"CREATED": "",
"SUBJECT": "test",
"MESSAGE": "Dies ist ein Test, den ich performe. Mal gucken, ob ich so eine lange Nachricht schreiben darf."
}'
```

### GET messages
```bash
 curl -X GET http://172.17.0.2:8006/vs/v1/messages?star=f55998bb998fa316ee82a6dc3245bd42&scope=all&view=id
```

### GET message
```bash
curl -X GET http://172.17.0.2:8006/vs/v1/messages/1@1000
```

### DELETE message
```bash
curl -X DELETE http://172.17.0.2:8006/vs/v1/messages/1@1000
```