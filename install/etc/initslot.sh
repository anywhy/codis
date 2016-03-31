#!/bin/sh
echo "slots initializing..."
../bin/codis-config -c ../etc/config.ini slot init -f
echo "done"
sleep 3

echo "set slot ranges to server groups..."
#set group num and step
group_num=6; step=170
for ((i=0;i<${group_num};i++)); do
    let beg="$step*i"
    let end="$step*i + $step - 1"
    let group="1+i"
    let num="$group_num - 1"
    if [ "$i" = "$num" -a "$end" != "1023" ]; then
        end=1023
    fi
    echo "begin:$beg to end:$end"
    ../bin/codis-config -c  ../etc/config.ini slot range-set $beg $end $group online
done
echo "done"

