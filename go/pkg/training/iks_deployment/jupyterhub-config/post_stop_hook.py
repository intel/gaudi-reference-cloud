# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
import kubespawner
    
import asyncio

async def delete_policy(
    log,
    policy_name: str,
    k8s_api_name: str,
    applied_namespace: str,
    group="",
    version="",
    plural="",
    timeout=3,
):
    k8s_client = kubespawner.clients.shared_client(k8s_api_name)

    match k8s_api_name:
        case "CustomObjectsApi":
            if group == "" or version == "" or plural == "":
                log.error(f"Invalid arguments while creating {policy_name}, resource could not be created")
                return
            await asyncio.wait_for(
                k8s_client.delete_namespaced_custom_object(
                    group=group,
                    version=version,
                    namespace=applied_namespace,
                    plural=plural,
                    name=policy_name,
                ),
                timeout
            )
        case "NetworkingV1Api":
            await asyncio.wait_for(
                k8s_client.delete_namespaced_network_policy(
                    namespace=applied_namespace,
                    name=policy_name,
                ),
                timeout
            )
        case _:
            log.warning(f"Invalid K8s API name {k8s_api_name}, resource could not be created")
            return
    

async def delete_user_policies(spawner):

    username = spawner.user.name
    user_namespace = f"jupyterhub-{username}"

    # delete internet policy
    await delete_policy(
        policy_name="internet",
        k8s_api_name="CustomObjectsApi",
        group="projectcalico.org",
        version="v3",
        plural="networkpolicies",
        applied_namespace=user_namespace,
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # delete istio policy
    await delete_policy(
        policy_name="istio",
        k8s_api_name="CustomObjectsApi",
        group="projectcalico.org",
        version="v3",
        plural="networkpolicies",
        applied_namespace=user_namespace,
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # delete jupyterhub policy
    await delete_policy(
        policy_name="jupyterhub",
        k8s_api_name="CustomObjectsApi",
        group="projectcalico.org",
        version="v3",
        plural="networkpolicies",
        applied_namespace=user_namespace,
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # delete kube-apiserver policy
    await delete_policy(
        policy_name="kube-apiserver",
        k8s_api_name="CustomObjectsApi",
        group="projectcalico.org",
        version="v3",
        plural="networkpolicies",
        applied_namespace=user_namespace,
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # delete kube-dns policy
    await delete_policy(
        policy_name="kube-dns",
        k8s_api_name="CustomObjectsApi",
        group="projectcalico.org",
        version="v3",
        plural="networkpolicies",
        applied_namespace=user_namespace,
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # delete registrationapi policy
    await delete_policy(
        policy_name=f"registrationapi-{username}",
        k8s_api_name="CustomObjectsApi",
        group="security.istio.io",
        version="v1",
        plural="authorizationpolicies",
        applied_namespace="registration",
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # delete registrationdb policy
    await delete_policy(
        policy_name=f"registrationdb-{username}",
        k8s_api_name="CustomObjectsApi",
        group="security.istio.io",
        version="v1",
        plural="authorizationpolicies",
        applied_namespace="registration",
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

    # delete weka policy
    await delete_policy(
        policy_name="weka",
        k8s_api_name="CustomObjectsApi",
        group="projectcalico.org",
        version="v3",
        plural="networkpolicies",
        applied_namespace=user_namespace,
        timeout=spawner.k8s_api_request_timeout,
        log=spawner.log
    )

c.KubeSpawner.post_stop_hook = delete_user_policies
