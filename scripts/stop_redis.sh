#!/bin/sh

for i in 638{0..1}; do
    pid=`ps -ef|grep codis |grep $i |grep -v 'grep' |awk '{print $2}'`
    echo "stop redis ip:$i, pid:$pid"
    kill ${pid}
done


