#!/bin/sh

pid=`ps -ef|grep codis |grep codis-ha |grep -v 'grep' |awk '{print $2}'`
echo "stop codis-ha, pid:$pid"
kill ${pid}
