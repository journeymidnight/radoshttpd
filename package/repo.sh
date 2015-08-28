#!/bin/bash
if [ -z $1 ] ; then
	echo "no rpm found"
	exit
fi


#$1 is rpm
#$2 is project
if [[ $1 == '' ]];then
	echo "no rpm"
	exit
fi
if [[ $2 != 'uts' && $2 != 'ceph' ]]; then
	echo "no project"
	exit
fi
repo='115.182.93.170'
remotedir="/letv/yum-repo/"$2"/el6/update"
sshpasspath=`which sshpass`
if [[ $? -ne  0 ]] ;then
	echo "try to install sshpass first"
	exit
fi
$sshpasspath -p Myiaas.chensh.net scp $1 chensh@$repo:$remotedir/x86_64
$sshpasspath -p Myiaas.chensh.net ssh chensh@$repo createrepo --update $remotedir
