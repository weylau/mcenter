#!/bin/bash

bin_dir=$PWD
dir=${bin_dir}/../
pid_file=${bin_dir}/run/mcenter.pid
if [ -f $pid_file ];then
echo "kill old process ..."
kill -QUIT `cat $pid_file`
echo "sleep 5 second ..."
sleep 5
rm -f $pid_file
echo "done"
else
echo "pid file not found"
fi
exit 0