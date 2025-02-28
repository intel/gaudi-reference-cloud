#!/bin/bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
user=$1

#sudo adduser --disabled-password --gecos ""  -e $(date -d "20 days" +"%Y-%m-%d") $user
#sudo adduser --disabled-password --gecos "" $user
sudo useradd -m $user -e $(date -d "20 days" +"%Y-%m-%d") -s /bin/bash -d /home/$user

# hard code either freetier or premium
sudo usermod -a -G premium,render $user

# these are the "premium" settings from the user.sh script
MaxSubmitJobs=5
MaxWall=0-04:00:00

sacctmgr create account $user -i
sacctmgr create user $user account=$user -i
sacctmgr modify user $user set defaultaccount=$user -i
sacctmgr modify account where name=$user set MaxSubmitJobs=$MaxSubmitJobs MaxWall=$MaxWall MaxJobs=1 -i
sacctmgr modify user $user set MaxJobs=1 -i
scontrol reconfigure
echo 'export PS1="\$USER@$HOSTNAME:~$ "' >> /home/$user/.bashrc


rsync -a --chown=$user:users /localhome/devcloud/learning_paths/oneapi-essentials-training /home/$user/

#export PS1="\$user@$HOSTNAME:~$ "' >> .bashrc