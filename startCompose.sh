#!/bin/bash
go build ./main.go
docker compose up --build 
