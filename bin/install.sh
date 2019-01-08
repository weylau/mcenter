#!/bin/bash
echo "kill old process ..."
kill -QUIT `cat ./run/mcenter.pid`
echo "sleep 5 second ..."
sleep 5


echo "start install..."
cd ../src/mcenter
go install
echo "install success!"

echo "start new process ..."
nohup ./mcenter -pidfile ./run/mcenter.pid &
exit 0
