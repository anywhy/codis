#!/bin/sh

echo "start new proxy..."
nohup ../bin/codis-proxy --log-level info -c ../etc/config.ini -L ../logs/proxy.log  --cpu=8 --addr=0.0.0.0:19000 --http-addr=0.0.0.0:11000 &>/dev/null &
echo "done"

