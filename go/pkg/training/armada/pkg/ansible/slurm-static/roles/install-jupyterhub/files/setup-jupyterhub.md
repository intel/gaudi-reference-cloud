<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Setup JupyterHub and Nginx

## Run these commands

__**NOTE:**__ Run commands below with `sudo -s`

```bash
export http_proxy=http://internal-placeholder.com:912
export https_proxy=http://internal-placeholder.com:912

ln -s /srv/jupyter/python-3.11.5/bin/python3.11 /srv/jupyter/python-3.11.5/bin/python3
ln -s /srv/jupyter/python-3.11.5/bin/pip3.11 /srv/jupyter/python-3.11.5/bin/pip3

export PYTHONPATH=/srv/jupyter/python-3.11.5
export PATH=/srv/jupyter/python-3.11.5/bin:${PATH}

export NODE=/srv/jupyter/n/bin/node-v18.17.1-linux-x64
export PATH=${NODE}/bin:${PATH}

python3 -m venv /srv/jupyter/python-venv
source /srv/jupyter/python-venv/bin/activate

python3 -m pip install jupyterhub
npm install -g configurable-http-proxy
python3 -m pip install jupyterlab notebook
pip3 install batchspawner
pip3 install "oauthenticator[azuread]"
pip3 install ipywidgets
pip3 install jupyterhub_moss

patch /srv/jupyter/python-venv/lib/python3.11/site-packages/batchspawner/singleuser.py /srv/jupyter/patch.txt
```

## Post Setup

* Configure Nginx config file witht the correct properties.
* Start JupyterHub
