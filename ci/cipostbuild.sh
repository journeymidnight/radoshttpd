#!/bin/bash
filepath=$(cd "$(dirname "$0")"; pwd)
cd $filepath

repo='10.200.93.170'
remotedir="/letv/yum-repo/ceph/el6/update"
sshpasspath=`which sshpass`
if [[ $? -ne  0 ]] ;then
        echo "try to install sshpass first"
        exit
fi
$sshpasspath -p Myiaas.chensh.net scp -o StrictHostKeyChecking=no $(find .. -type f -name "*.rpm" -a ! -name "*src*") chensh@$repo:$remotedir/x86_64
$sshpasspath -p Myiaas.chensh.net ssh chensh@$repo createrepo --update $remotedir
