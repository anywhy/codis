#!/bin/bash
echo "start new proxy..."
nohup ../bin/codis-proxy --log-level info -c ../etc/config.ini -L ../logs/proxy.log  --cpu=8 --addr=10.161.35.73:8099 --http-addr=10.161.35.73:11000 &>/dev/null &
echo "done"
