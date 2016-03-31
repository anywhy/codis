#! /bin/bash
#install path
install_path=/opt/beh/core/codis3
# master index
master_index=1

echo "install path is: $install_path"

hostlist=./etc/managelist
install_index=1
cat ${hostlist} | while read host
do
     echo "install $host"
     scp -r ./source/* ${host}:${install_path}/

     # gen config
     rm -rf ./tem/*
     cp -r ./etc/initslot.sh ./tem/initslot.sh
     cp -r ./etc/group.json ./tem/group.json
     
     echo "generate config file, proxy_id: $install_index"
     sh genconf.sh ${install_index}
     #gen startall
     if [ "$master_index" = "$install_index" ]; then
	echo "generate master startall"
        sh genstart.sh true
     else
	echo "generate slave startall"
        sh genstart.sh false
     fi
     
     sh gen_start_cluster.sh ${install_path} ${master_index}
     sh gen_stop_cluster.sh ${install_path}
     sh gen_start_proxy.sh ${host}

     echo "copy config file...."
     scp -r ./tem/*.sh ${host}:${install_path}/sbin/
     scp -r ./tem/*.json ${host}:${install_path}/etc/
     scp -r ./tem/*.ini ${host}:${install_path}/etc/
     scp -r ./etc/redis_conf/* ${host}:${install_path}/etc/redis_conf/
     scp -r ./etc/managelist ${host}:${install_path}/etc/

     # update +x
     ssh -f -n  ${host} "cd $install_path; chmod -R +x *;"

    let install_index="$install_index + 1"
done
echo "=====installed===="


