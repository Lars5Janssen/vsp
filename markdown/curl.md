# Anfragen über curl

## !!!BITTE IMMER ALLE STARUUIDS ANPASSEN!!! (STRG + SHIFT + R)

## Aufgabe 1
### DELETE
```bash
curl -X DELETE http://172.18.0.2:8006/vs/v1/system/1000
```

#### Curl to Sol
```bash
curl -X POST http://172.17.0.2:8006/vs/v1/system -H "Content-Type: application/json" -d '{
  "STAR": "5877f963452750641d09b72f1f9bd163",
  "SOL": 4187,
  "COMPONENT": 1133,
  "COMIP": "172.17.0.3",
  "COMTCP": 8006,
  "STATUS": "200"
}' -i
```

## Aufgabe 2
### POST message
#### Curl to Component
```bash
curl -X POST http://172.18.0.3:8006/vs/v1/messages -H "Content-Type: application/json" -d '{
"STAR": "6ceeabb1e9e70bc53ca69ebfb076dd8b",
"ORIGIN": "1382",
"SENDER": "1382",
"MSGID": "",
"VERSION": "",
"CREATED": "",
"SUBJECT": "Test",
"MESSAGE": "Dies ist ein Test, den ich performe. Mal gucken, ob ich so eine lange Nachricht schreiben darf."
}'
```

#### Curl to Sol
```bash
curl -X POST http://172.18.0.2:8006/vs/v1/messages -H "Content-Type: application/json" -d '{
"STAR": "6ceeabb1e9e70bc53ca69ebfb076dd8b",
"ORIGIN": "me@you.com",
"SENDER": "5161",
"MSGID": "",
"VERSION": "",
"CREATED": "",
"SUBJECT": "Test",
"MESSAGE": "Dies ist ein Test, den ich performe."
}' -i
```

### GET messages

#### Curl to Component
```bash
curl -X GET "http://172.18.0.2:8006/vs/v1/messages?star=6ceeabb1e9e70bc53ca69ebfb076dd8b&scope=all&view=header"
```

#### Curl to Sol
```bash
curl -X GET "http://172.18.0.2:8006/vs/v1/messages?star=6ceeabb1e9e70bc53ca69ebfb076dd8b&scope=all&view=id"
```

### GET message

#### Curl to Component
```bash
curl -X GET http://172.18.0.3:8006/vs/v1/messages/2@2040?star=6ceeabb1e9e70bc53ca69ebfb076dd8b
```

#### Curl to Sol
```bash
curl -X GET http://172.18.0.2:8006/vs/v1/messages/1@5161?star=6ceeabb1e9e70bc53ca69ebfb076dd8b
```

### DELETE message
```bash
curl -X DELETE http://172.18.0.2:8006/vs/v1/messages/1@5161?star=6ceeabb1e9e70bc53ca69ebfb076dd8b
```
