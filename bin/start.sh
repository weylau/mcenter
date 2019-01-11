#!/bin/bash
bin_dir=$PWD
dir=${bin_dir}/../
pid_file=${bin_dir}/run/mcenter.pid
main_dir=${dir}src/mcenter
if [ -f $pid_file ];then
echo "the process is already started"
else
echo "start new process ..."
nohup ${bin_dir}/mcenter -pidfile $pid_file > nohup.out 2>&1 &
fi
echo "done"
exit 0
