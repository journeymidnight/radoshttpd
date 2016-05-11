#!/bin/bash
filepath=$(cd "$(dirname "$0")"; pwd)
cd $filepath

repo='10.140.75.88'
remotedir="/letv/yum-repo/ceph/el6/update"
sshpasspath=`which sshpass`
if [[ $? -ne  0 ]] ;then
        echo "try to install sshpass first"
        exit
fi
$sshpasspath -p TVLEhp800g.com scp -o StrictHostKeyChecking=no $(find .. -type f -name "*.rpm" -a ! -name "*src*") root@$repo:$remotedir/x86_64
$sshpasspath -p TVLEhp800g.com ssh root@$repo createrepo --update $remotedir
$sshpasspath -p TVLEhp800g.com ssh root@$repo /letv/yum-repo/sync.ceph
rm ~/rpmbuild/RPMS/x86_64/*.rpm -rf
