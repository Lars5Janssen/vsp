# Befehle fuer das Benutzen von Docker
If you want GIN Debug logs, comment the env line in the dockerfile out
```dockerfile
# ENV GIN_MODE="release"
```

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
