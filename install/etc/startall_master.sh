#!/bin/bash
./start_dashboard.sh
sleep 3
./add_group.sh
./initslot.sh
./start_redis.sh
./start_proxy.sh


