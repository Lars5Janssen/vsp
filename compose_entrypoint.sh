#!/bin/bash

cat /app/logs/compile.log
if [[ "$1" == "NotSol" ]]; then
    screen -dmS app app -sleep -killSol
else
    screen -dmS app app
fi

sleep 1
tail --follow /app/logs/app.log
