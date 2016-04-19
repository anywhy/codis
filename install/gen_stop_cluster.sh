#! /bin/bash

cat << EOF > ./tem/stop_cluster.sh
#install path
install_path=$1

read -p "please choose a module(all/proxy/redis):" module
hostlist=../etc/managelist
if [ \$module = "proxy" ]; then
    echo "stop cluster proxy ..."
    cat \${hostlist} | while read host
    do
         echo "stop redis: \$host"
         ssh -f \${host} "cd \$install_path/sbin; sh stop_proxy.sh > /dev/null 2>&1;"
    done
    echo "====proxy done======"
elif [ \$module = "redis" ]; then
    echo "stop cluster redis......"
    cat \${hostlist} | while read host
    do
         echo "stop redis: \$host"
         ssh -f \${host} "cd \$install_path/sbin; sh stop_redis.sh > /dev/null 2>&1;"
    done
    echo "=====redis done====="
else
    cat \${hostlist} | while read host
    do
         echo "stop \$host"
         ssh -f -n  \${host} "cd \$install_path/sbin; sh stopall.sh > /dev/null 2>&1;"
    done
    echo "=====stopped cluster===="
    echo "warn 'all' not stop redis, please use 'redis' stop redis!!!"
fi
EOF
