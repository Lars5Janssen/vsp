FROM golang:1.23
LABEL version="1.0"
LABEL description="A chat-programm that builds a star-topology"

WORKDIR /usr/src/app

# Update and install screen
RUN apt update -y
RUN apt upgrade -y
RUN apt install screen -y

# Config for screen so that multiple users cann accses the same screen
COPY screenrc /etc/.screenrc

# Logs directory
RUN mkdir -p /app/logs

# Relevant src files
COPY main.go go.mod go.sum ./
COPY cmd/ ./cmd/
COPY connection/ ./connection/
COPY utils/ ./utils/
COPY --chmod=755 compose_entrypoint.sh dockerbuild.sh mai[n] ./

# Build programm or copy it to /usr/local/bin/ if a precompiled binary has been provided
RUN go mod download && go mod verify
RUN ./dockerbuild.sh

# Configure Environment
EXPOSE 8006
ENV GIN_MODE="release"

# Will be overwritten by compose. But would start the app if only the dockerfile is in use
ENTRYPOINT ["app"]
