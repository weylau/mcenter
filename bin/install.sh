#!/bin/bash
bin_dir=$PWD
dir=${bin_dir}/../
pid_file=${bin_dir}/run/mcenter.pid
main_dir=${dir}src/mcenter

if [ -f $pid_file ];then
echo "kill old process ..."
kill -QUIT `cat $pid_file`
echo "sleep 5 second ..."
sleep 5
rm -f $pid_file
fi

echo "start install"
cd $main_dir
go install
echo "install success!"
cd $bin_dir
echo "start new process"
nohup ${bin_dir}/mcenter -pidfile $pid_file > nohup.out 2>&1 &
echo "done"
exit 0
