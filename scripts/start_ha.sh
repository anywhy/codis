#!/bin/sh
nohup ../bin/codis-config -c ../etc/config.ini -L ../logs/codis-ha.log codis-ha --interval=3 &>/dev/null &

