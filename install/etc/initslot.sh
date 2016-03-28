#!/bin/sh
echo "slots initializing..."
../bin/codis-config -c config.ini slot init -f
echo "done"

echo "set slot ranges to server groups..."
for ((i=0;i<2;i++)); do
    let beg="512*i"
    let end="512*i + 511"
    let group="1+i"
    ../bin/codis-config -c  config.ini slot range-set $beg $end $group online
done
echo "done"

