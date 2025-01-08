#!/bin/bash

FILE=/usr/src/app/main
if test -f "$FILE"; then
    echo "found PREcompiled binary" > /app/logs/compile.log
    cp /usr/src/app/main /usr/local/bin/app
else
    echo "found OWN compiled binary" > /app/logs/compile.log
    go build -v -o /usr/local/bin/app main.go
fi
