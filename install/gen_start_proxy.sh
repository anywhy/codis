#!/bin/bash

cat << EOF > ./tem/start_proxy.sh
#!/bin/bash
echo "start new proxy..."
nohup ../bin/codis-proxy --log-level info -c ../etc/config.ini -L ../logs/proxy.log  --cpu=8 --addr=$1:19000 --http-addr=$1:11000 &>/dev/null &
echo "done"
EOF
