#!/bin/sh

pid=`ps -ef|grep codis |grep codis-config |grep dashboard |grep -v 'grep' |awk '{print $2}'`
echo "stop dashboard, pid:$pid"
kill ${pid}
