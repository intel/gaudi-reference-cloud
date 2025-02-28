# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
from jinja2 import Environment, FileSystemLoader
import kubespawner

import asyncio
import yaml

async def create_policy(
    log,
    jinja_env,
    policy_name: str,
    body: dict,
    k8s_api_name: str,
    applied_namespace: str,
    group="",
    version="",
    plural="",
    timeout=3
):
    policy = jinja_env.get_template(f"{policy_name}.j2")
    rendered = yaml.safe_load(
        policy.render(body)
    )

    k8s_client = kubespawner.clients.shared_client(k8s_api_name)

    match k8s_api_name:
        case "CustomObjectsApi":
            if group == "" or version == "" or plural == "":
                log.error(f"Invalid arguments while creating {policy_name}, resource could not be created")
                return
            await asyncio.wait_for(
                k8s_client.create_namespaced_custom_object(
                    group=group,
                    version=version,
                    namespace=applied_namespace,
                    plural=plural,
                    body=rendered,
                ),
                timeout
            )
        case "NetworkingV1Api":
            await asyncio.wait_for(
                k8s_client.create_namespaced_network_policy(
                    namespace=applied_namespace,
                    body=rendered,
                ),
                timeout
            )
        case _:
            log.warning(f"Invalid K8s API name {k8s_api_name}, resource could not be created")
            return

async def create_user_policies(spawner, pod):
    spawner.log.info("Creating user policies")
    
    env = Environment(
        loader=FileSystemLoader("/usr/local/share/jupyterhub/extra_files"),
        variable_start_string="{-",
        variable_end_string="-}"
    )

    username = spawner.user.name
    user_namespace = f"jupyterhub-{username}"
    
    # create internet policy
    await create_policy(
        jinja_env=env,
        policy_name="internet",
        body={"namespace": user_namespace},
        k8s_api_name="CustomObjectsApi",
        group="projectcalico.org",
        version="v3",
        plural="networkpolicies",
        applied_namespace=user_namespace,
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # create istio policy
    await create_policy(
        jinja_env=env,
        policy_name="istio",
        body={"namespace": user_namespace},
        k8s_api_name="CustomObjectsApi",
        group="projectcalico.org",
        version="v3",
        plural="networkpolicies",
        applied_namespace=user_namespace,
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # create jupyterhub policy
    await create_policy(
        jinja_env=env,
        policy_name="jupyterhub",
        body={"namespace": user_namespace},
        k8s_api_name="CustomObjectsApi",
        group="projectcalico.org",
        version="v3",
        plural="networkpolicies",
        applied_namespace=user_namespace,
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # create kube-apiserver policy
    await create_policy(
        jinja_env=env,
        policy_name="kube-apiserver",
        body={"namespace": user_namespace},
        k8s_api_name="CustomObjectsApi",
        group="projectcalico.org",
        version="v3",
        plural="networkpolicies",
        applied_namespace=user_namespace,
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # create kube-dns policy
    await create_policy(
        jinja_env=env,
        policy_name="kube-dns",
        body={"namespace": user_namespace},
        k8s_api_name="CustomObjectsApi",
        group="projectcalico.org",
        version="v3",
        plural="networkpolicies",
        applied_namespace=user_namespace,
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # create registrationapi policy
    await create_policy(
        jinja_env=env,
        policy_name="registrationapi",
        body={"username": username, "namespace": user_namespace},
        k8s_api_name="CustomObjectsApi",
        group="security.istio.io",
        version="v1",
        plural="authorizationpolicies",
        applied_namespace="registration",
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # create registrationdb policy
    await create_policy(
        jinja_env=env,
        policy_name="registrationdb",
        body={"username": username, "namespace": user_namespace},
        k8s_api_name="CustomObjectsApi",
        group="security.istio.io",
        version="v1",
        plural="authorizationpolicies",
        applied_namespace="registration",
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # create weka policy
    weka_ips_to_block = c.JupyterHub.template_vars["wekaIpsToBlock"][0].split()
    await create_policy(
        jinja_env=env,
        policy_name="weka",
        body={"namespace": user_namespace, "weka_ips": weka_ips_to_block},
        k8s_api_name="CustomObjectsApi",
        group="projectcalico.org",
        version="v3",
        plural="networkpolicies",
        applied_namespace=user_namespace,
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

c.KubeSpawner.after_pod_created_hook = create_user_policies
