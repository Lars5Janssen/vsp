# Anfragen über curl

## Aufgabe 1
### DELETE
```bash
curl -X DELETE http://172:17:0:2:8006/vs/v1/system/1000
```

## Aufgabe 2
### POST message
#### Curl for Component
```bash
curl -X POST http://172.17.0.3:8006/vs/v1/messages -H "Content-Type: application/json" -d '{
"STAR": "9343e3029e602dc3e2336a48b17d3d3c",
"ORIGIN": "8230",
"SENDER": "8230",
"MSGID": "",
"VERSION": "",
"CREATED": "",
"SUBJECT": "Test",
"MESSAGE": "Dies ist ein Test, den ich performe. Mal gucken, ob ich so eine lange Nachricht schreiben darf."
}'
```

#### Curl for Sol
```bash
curl -X POST http://172.17.0.2:8006/vs/v1/messages -H "Content-Type: application/json" -d '{
"STAR": "9343e3029e602dc3e2336a48b17d3d3c",
"ORIGIN": "7165",
"SENDER": "7165",
"MSGID": "",
"VERSION": "",
"CREATED": "",
"SUBJECT": "Test",
"MESSAGE": "Dies ist ein Test, den ich performe. Mal gucken, ob ich so eine lange Nachricht schreiben darf."
}' -i
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