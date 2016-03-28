#!/bin/bash

#copy
cat ./etc/config.ini > ./tem/config.ini
cat << EOF >> ./tem/config.ini
proxy_id=proxy_$1
EOF
