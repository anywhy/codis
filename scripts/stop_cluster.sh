#! /bin/bash
#install path
install_path=$1

hostlist=../etc/managelist
cat ${hostlist} | while read host
do
     echo "stop $host"
     ssh -f -n  ${host} "cd $install_path/sbin; sh stopall.sh > /dev/null 2>&1;"
done
echo "=====stop cluster===="


