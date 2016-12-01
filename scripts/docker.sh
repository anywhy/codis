#!/bin/bash

hostip=`ifconfig eth0 | grep "inet " | awk -F " " '{print $2}'`

if [ "x$hostip" == "x" ]; then
    echo "cann't resolve host ip address"
    exit 1
fi

mkdir -p log

case "$1" in
dashboard)
    docker rm -f      "Codis-D28080" &> /dev/null
    docker run --name "Codis-D28080" -d \
        --read-only -v `realpath ../config/dashboard.toml`:/codis/dashboard.toml \
                    -v `realpath log`:/codis/log \
        -p 28080:18080 \
        codis-image \
        codis-dashboard -l log/dashboard.log -c dashboard.toml --host-admin ${hostip}:28080
    ;;

proxy)
    docker rm -f      "Codis-P29000" &> /dev/null
    docker run --name "Codis-P29000" -d \
        --read-only -v `realpath ../config/proxy.toml`:/codis/proxy.toml \
                    -v `realpath log`:/codis/log \
        -p 29000:19000 -p 21080:11080 \
        codis-image \
        codis-proxy -l log/proxy.log -c proxy.toml --host-admin ${hostip}:29000 --host-proxy ${hostip}:21080
    ;;

server)
    for ((i=0;i<4;i++)); do
        let port="26379 + i"
        docker rm -f      "Codis-S${port}" &> /dev/null
        docker run --name "Codis-S${port}" -d \
            -v `realpath log`:/codis/log \
            -p $port:6379 \
            codis-image \
            codis-server --logfile log/${port}.log
    done
    ;;

cleanup)
    docker rm -f      "Codis-D28080" &> /dev/null
    docker rm -f      "Codis-P29000" &> /dev/null
    for ((i=0;i<4;i++)); do
        let port="26379 + i"
        docker rm -f      "Codis-S${port}" &> /dev/null
    done
    ;;

*)
    echo "wrong argument(s)"
    ;;

esac
