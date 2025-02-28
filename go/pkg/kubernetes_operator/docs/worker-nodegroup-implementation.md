<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
# Worker Nodegroup Implementation
This document provides an overview of the implementation code for the operations required to manage worker nodegroups and its worker nodes.

## Add a worker node
When the `count` field of a worker nodegroup is modified, and desired count > current count. The nodegroup controller will initiate the creation of new worker nodes.
Worker nodes are being created by following the [Kubelet TLS Bootstrapping](https://kubernetes.io/docs/reference/access-authn-authz/kubelet-tls-bootstrapping/) documentation.

### Process
1. When desired count > current count.
2. The nodegroup controller creates an instance of the kubernetes and node provider based on the Cluster spec configuration.
3. Using the kubernetes provider, nodegroup controller gets the worker bootstrap script.
4. Nodegroup controller gets from the cluster secret, the worker registration command and the management kubeconfig file.
5. Using the management kubeconfig file, nodegroup controller creates a bootstrap token secret in the managed cluster, for the new node that will be created.
  - Bootstrap tokens are created one per worker node.
  - Expiration of bootstrap tokens is hardcoded to 24h.
6. Using the node provider, Nodegroup controller creates a new node, passing required information like network data and user data (bootstrap script and registration command).
7. Nodegroup controller reconciles until all nodes of the worker nodegroup are created.
8. Using management kubeconfig file, nodegroup controller approves all kubelet-serving certificate signing requests created by worker nodes.
  - TODO: We should continue monitoring for kubelet-serving csrs and approve them since they are automatically created by kubelet when they need to be renewed.
9. Nodegroup controller reconciles until all nodes of the worker nodegroup are active.

## Delete a worker node
When the `count` field of an existing worker nodegroup is modified, and desired count < current count. The nodegroup controller will initiate the deletion of worker nodes.

### Process
1. When desired count < current count.
2. The nodegroup controller creates an instance of the kubernetes and node provider based on the Cluster spec configuration.
3. Nodegroup controller gets from the cluster secret, the management kubeconfig file.
4. TODO: Nodegroup controller gets node candidates for deletion based on specific criteria.
5. TODO: Using the management kubeconfig file, nodegroup controller drains the node that will be deleted.
6. Using the management kubeconfig file, nodegroup controller deletes the worker node from the cluster.
7. Using the node provider, Nodegroup controller deletes the worker node.
