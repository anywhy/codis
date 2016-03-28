#! /bin/bash
#install path
install_path=/backup/bonc/codis-test
# master index
master_index=1

echo "install path is: $install_path"

hostlist=./etc/managelist
install_index=1
cat ${hostlist} | while read host
do
     echo "install $host"
     scp ./source/codis.zip ${host}:${install_path}
     ssh -f -n  ${host} "cd $install_path; unzip codis.zip > /dev/null 2>&1;"

     echo "sleep 5's wait for unzip ......"
     sleep 5     

     # gen config
     rm -rf ./tem/*
     cp -r ./etc/initslot.sh ./tem/initslot.sh
     cp -r ./etc/group.json ./tem/group.json
     
     echo "generate config.ini file"
     sh genconf.sh ${install_index}
     #gen startall
     if [ "$master_index" = "$install_index" ]; then
	echo "generate master startall"
        sh genstart.sh true
     else
	echo "generate slave startall"
        sh genstart.sh false
     fi

     scp -r ./tem/*.sh ${host}:${install_path}/sbin/
     scp -r ./tem/*.json ${host}:${install_path}/etc/
     scp -r ./tem/*.ini ${host}:${install_path}/etc/
     scp -r ./etc/redis_conf/* ${host}:${install_path}/etc/redis_conf/
     scp -r ./etc/managelist ${host}:${install_path}/etc/

    install_index="$install_index + 1"
done
echo "=====installed===="


