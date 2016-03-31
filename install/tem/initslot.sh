#!/bin/sh
echo "slots initializing..."
../bin/codis-config -c ../etc/config.ini slot init -f
echo "done"
sleep 3

echo "set slot ranges to server groups..."
for ((i=0;i<4;i++)); do
    let beg="256*i"
    let end="256*i + 255"
    let group="1+i"
    ../bin/codis-config -c  ../etc/config.ini slot range-set $beg $end $group online
done
echo "done"

