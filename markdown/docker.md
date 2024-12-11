# Befehle fuer das Benutzen von Docker

## Starting Docker compose
Under Linux:
```bash
./startCompose.sh
```

Under Windows:
```bash
docker compose up --build
```

## Enter Containers
SOL:
```bash
docker exec -it vsp-sol-1 screen -r app
```

Component:
```bash
docker exec -it vsp-component-1 screen -r app
```
