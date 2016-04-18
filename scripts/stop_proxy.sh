#!/bin/sh

pid=`ps -ef|grep codis |grep codis-proxy |grep -v 'grep' |awk '{print $2}'`
echo "stop proxy, pid:$pid"
kill ${pid}
