#! /bin/bash

cat << EOF > ./tem/start_cluster.sh
#install path
install_path=$1
master_index=$2

read -p "please choose a module(all/proxy/redis):" module
hostlist=../etc/managelist
if [ \$module = "proxy" ]; then
    echo "start cluster proxy ..."
    cat \${hostlist} | while read host
    do
         echo "start redis: \$host"
         ssh -f \${host} "cd \$install_path/sbin; sh start_proxy.sh > /dev/null 2>&1;"
    done
    echo "====cluster proxy done====="
elif [ \$module = "redis" ]; then
    echo "start cluster redis......"
    cat \${hostlist} | while read host
    do
         echo "start redis: \$host"
         ssh -f \${host} "cd \$install_path/sbin; sh start_redis.sh > /dev/null 2>&1;"
    done
    echo "====cluster redis done====="
else
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
fi
EOF
