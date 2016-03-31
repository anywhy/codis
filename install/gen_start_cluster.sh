#! /bin/bash

cat << EOF > ./tem/start_cluster.sh
#install path
install_path=$1
master_index=$2

echo "install path: \$install_path"
hostlist=../etc/managelist
echo "starting cluster redis......"
cat \${hostlist} | while read host
do
     echo "start redis: \$host"
     ssh -f \${host} "cd \$install_path/sbin; sh start_redis.sh > /dev/null 2>&1;"
done

sleep 3
echo "starting cluster..."
let loop_index=1
cat \${hostlist} | while read host
do
     echo "start node:\$host"
     if [ \$master_index = \$loop_index ]; then
       ssh -f \${host} "cd \$install_path/sbin; ./startall.sh > /dev/null 2>&1;"
     echo "wait 30's to do starting dashboard & init slots info...."
     sleep 30
     else
       ssh -f \${host} "cd \$install_path/sbin; ./startall.sh > /dev/null 2>&1;"
     fi
     let loop_index="\$loop_index + 1"
done

echo "=====started cluster===="
EOF

