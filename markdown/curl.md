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
curl -X POST http://172.20.10.5:8006/vs/v1/messages -H "Content-Type: application/json" -d '{
"STAR": "79280dda40b7bd3a03a2020ab9ab221f",
"ORIGIN": "4956",
"SENDER": "4956",
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
"STAR": "c89cf30b5c8ebdd811736f65896b1a0f",
"ORIGIN": "5374",
"SENDER": "5374",
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
 curl -X GET http://172.17.0.3:8006/vs/v1/messages?star=13b327bca6763d08822bb8c159d96da8&scope=all&view=id
```

#### Curl to Sol
```bash
 curl -X GET http://172.17.0.2:8006/vs/v1/messages?star=c89cf30b5c8ebdd811736f65896b1a0f&scope=all&view=id
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