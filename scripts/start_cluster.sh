#! /bin/bash
#install path
install_path=$1

echo "start path is: $install_path"
hostlist=../etc/managelist
cat ${hostlist} | while read host
do
     echo "start $host"
     ssh -f -n  ${host} "cd $install_path/sbin; sh startall.sh > /dev/null 2>&1;"
done
echo "=====started cluster===="


