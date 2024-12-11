FROM golang:1.23

WORKDIR /usr/src/app

RUN apt update -y
RUN apt upgrade -y
RUN apt install screen -y

COPY screenrc /etc/.screenrc

# Create the logs directory
RUN mkdir -p /app/logs

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY cmd/ ./cmd/
COPY connection/ ./connection/
COPY utils/ ./utils/
COPY main.go ./

COPY --chmod=755 dockerbuild.sh mai[n] ./
# RUN chmod +x dockerbuild.sh
RUN ./dockerbuild.sh

COPY --chmod=755 compose_entrypoint.sh ./
# RUN chmod +x compose_entrypoint.sh

# RUN go build -v -o /usr/local/bin/app ./main.go

EXPOSE 8006

ENTRYPOINT ["app"]
