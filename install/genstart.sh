#!/bin/bash

if [ $1 = true ]; then
cat << EOF > ./tem/startall.sh
#!/bin/bash
rm -rf ../logs/*
./start_dashboard.sh PID=\$!;wait \$PID;
./add_group.py PID=\$!;wait \$PID;
./initslot.sh PID=\$!;wait \$PID;
./start_redis.sh PID=\$!;wait \$PID;
./start_proxy.sh PID=\$!;wait \$PID;
EOF
else
cat << EOF > ./tem/startall.sh
#!/bin/bash
rm -rf ../logs/*
./start_redis.sh
./start_proxy.sh
EOF
fi
