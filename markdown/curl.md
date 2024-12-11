# Anfragen über curl

## Aufgabe 1
### DELETE
```bash
curl -X DELETE http://172:17:0:2:8006/vs/v1/system/1000
```

## Aufgabe 2
### POST message
#### Curl to Component
```bash
curl -X POST http://172.17.0.3:8006/vs/v1/messages -H "Content-Type: application/json" -d '{
"STAR": "13b327bca6763d08822bb8c159d96da8",
"ORIGIN": "2040",
"SENDER": "2040",
"MSGID": "",
"VERSION": "",
"CREATED": "",
"SUBJECT": "Test",
"MESSAGE": "Dies ist ein Test, den ich performe. Mal gucken, ob ich so eine lange Nachricht schreiben darf."
}'
```

#### Curl to Sol
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

#### Curl to Component
```bash
 curl -X GET http://172.17.0.3:8006/vs/v1/messages?star=f55998bb998fa316ee82a6dc3245bd42&scope=all&view=id
```

#### Curl to Sol
```bash
 curl -X GET http://172.17.0.2:8006/vs/v1/messages?star=f55998bb998fa316ee82a6dc3245bd42&scope=all&view=id
```

### GET message

#### Curl to Component
```bash
curl -X GET http://172.17.0.3:8006/vs/v1/messages/2@2040?star=13b327bca6763d08822bb8c159d96da8
```

#### Curl to Sol
```bash
curl -X GET http://172.17.0.2:8006/vs/v1/messages/1@1000
```

### DELETE message
```bash
curl -X DELETE http://172.17.0.2:8006/vs/v1/messages/1@1000
```