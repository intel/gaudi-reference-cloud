# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
import batchspawner

from tornado import gen
from subprocess import Popen, PIPE
import subprocess
import uuid
import hashlib

c = get_config()  #noqa

c.Authenticator.auto_login = True

@gen.coroutine
def create_system_user(authenticator, handler, authentication):
    eid = authentication['name']
    hashed = hashlib.sha256(eid.encode()).hexdigest()
    username = 'u' + hashed[:31]
    existing_users = subprocess.check_output(['cut', '-d:', '-f1', '/etc/passwd']).decode('utf-8').split('\n')
    if username not in existing_users:
        process = Popen(["/srv/jupyter/bin/jupyteruser.sh", username], stdout=PIPE, stderr=PIPE)
        stdout, stderr = process.communicate()
        if process.returncode != 0:
            raise Exception(f"Failed to create user {username}. Error: {stderr}")
        print(f"Successfully added user")
    else:
        print(f"User {username} already exists.")
    authentication['name'] = username
    return authentication

# Create user if it does not exist in the system.
c.Authenticator.post_auth_hook = create_system_user

#Azure AD Authentication B2C Prepprod
c.JupyterHub.authenticator_class = "azuread"
c.Application.log_level = 'INFO'
#
# redacted
#
c.JupyterHub.bind_url = 'http://127.0.0.1:8000'
c.JupyterHub.hub_bind_url = 'http://127.0.0.1:8081'

#
# security recomendation
#
#c.JupyterHub.internal_ssl = True


c.Application.log_level = 'INFO'
c.ConfigurableHTTPProxy.debug = 0


c.JupyterHub.spawner_class = 'batchspawner.SlurmSpawner'
c.Spawner.http_timeout = 120

c.Spawner.default_url = '/lab/tree/oneapi-essentials-training/Welcome.ipynb'

#
# address security issues
# https://jupyterhub.readthedocs.io/en/stable/explanation/websecurity.html
#
c.Spawner.disable_user_config = True

c.SlurmSpawner.batch_script = """#!/bin/bash

 #SBATCH --output=/dev/null
 #SBATCH --job-name=spawner-jupyterhub
 #SBATCH --chdir=/home/${USER}
 #SBATCH --export={{keepvars}}
 #SBATCH --get-user-env=L
 {% if partition  %}#SBATCH --partition={{partition}}
 {% endif %}{% if runtime    %}#SBATCH --time={{runtime}}
 {% endif %}{% if memory     %}#SBATCH --mem={{memory}}
 {% endif %}{% if gres       %}#SBATCH --gres={{gres}}
 {% endif %}{% if nprocs     %}#SBATCH --cpus-per-task={{nprocs}}
 {% endif %}{% if reservation%}#SBATCH --reservation={{reservation}}
 {% endif %}{% if options    %}#SBATCH {{options}}{% endif %}

 echo "running"
 #set -euo pipefail
env
 echo ${USER}
 cd /home/${USER}
 source /opt/intel/oneapi/setvars.sh

 echo {{cmd}}
 trap 'echo SIGTERM received' TERM
 echo  {{prologue}}
 which jupyterhub-singleuser
 {% if srun %}{{srun}} {% endif %} /srv/jupyter/python-venv/bin/batchspawner-singleuser /srv/jupyter/python-venv/bin/jupyterhub-singleuser 
 echo "jupyterhub-singleuser ended gracefully"
 {{epilogue}}
 """