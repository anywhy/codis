#install path
install_path=/opt/beh/core/codis3

echo "start path is: $install_path"
hostlist=../etc/managelist
cat ${hostlist} | while read host
do
     echo "start $host"
     ssh -f -n  ${host} "cd $install_path/sbin; sh stopall.sh > /dev/null 2>&1;"
done
echo "=====started cluster===="
