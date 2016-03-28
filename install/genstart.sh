#!/bin/bash

if [ $1 = true ]; then
cat ./etc/startall_master.sh > ./tem/startall.sh
else
cat ./etc/startall_slave.sh > ./tem/startall.sh
fi
