#!/bin/bash

echo "kill old process ..."
kill -QUIT `cat ./run/mcenter.pid`

echo "sleep 5 second ..."
sleep 5

echo "start new process ..."
nohup ./mcenter -pidfile ./run/mcenter.pid &
exit 0