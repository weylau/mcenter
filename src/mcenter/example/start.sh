#!/bin/bash
echo "start new process ..."
nohup ./mcenter -pidfile ./run/mcenter.pid &
exit 0
