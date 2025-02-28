{# INTEL CONFIDENTIAL #}
{# Copyright (C) 2023 Intel Corporation #}
[slurmservers]
{{ .SlurmControllerHosts }}

[slurmdbdservers]
{{ .SlurmControllerHosts }}

[slurmexechosts]
{{ .SlurmComputeHosts }}
{{ .SlurmLoginHosts }}

[jupyterhub_nodes]
{{ .SlurmJupyterHubHosts }}

[all:vars]
cluster_id={{ .ClusterId }}
cluster_name={{ .ClusterId }}
num_workers={{ .NumWorkers }}
cluster_subnet="172.16.0.0/16"