FROM golang:1.23

WORKDIR /usr/src/app

RUN apt update -y
RUN apt upgrade -y
RUN apt install screen -y

COPY go.mod go.sum ./
RUN go mod download && go mod verify


COPY . .
RUN cp screenrc /etc/screenrc
# RUN go build -v -o /usr/local/bin/app ./main.go
RUN cp main /usr/local/bin/app

EXPOSE 8006

ENTRYPOINT ["app"]
