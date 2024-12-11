FROM golang:1.23

WORKDIR /usr/src/app

# Create the logs directory
RUN mkdir -p /app/logs

# RUN apt update -y
# RUN apt upgrade -y
# RUN apt install iproute2 -y

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app ./main.go

EXPOSE 8006

CMD ["app"]
