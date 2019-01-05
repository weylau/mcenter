#!/bin/bash

echo "kill old process ..."
kill -QUIT `cat ./run/gomsg.pid`

echo "sleep 5 second ..."
sleep 5

echo "start new process ..."
nohup ./gomsg -pidfile ./run/gomsg.pid &
exit 0