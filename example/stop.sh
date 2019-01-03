#!/bin/bash

echo "kill old process ..."
kill -QUIT `cat run/gomsg.pid`

echo "sleep 5 second ..."
sleep 5