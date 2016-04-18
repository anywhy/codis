#!/bin/sh
nohup ../bin/codis-config -c ../etc/config.ini -L ../logs/dashboard.log dashboard --addr=:60010 --http-log=../logs/requests.log &>/dev/null &

