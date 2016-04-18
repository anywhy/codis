#!/bin/sh

for i in 638{0..1}; do
    nohup ../bin/codis-server ../etc/redis_conf/${i}.conf &> ../logs/redis_${i}.log &
done


