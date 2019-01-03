#!/bin/bash
echo "start new process ..."
nohup ./gomsg -pidfile ./run/gomsg.pid &
exit 0
