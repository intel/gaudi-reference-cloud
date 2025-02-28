# Enable Weka Storage

## Enable Storage in old IKS Clusters

1. Apply cluster roles with right permissions.

    Using `kubectl`, apply the following cluster roles.

    ```
    kubectl apply -f - <<END
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      labels:
        controller: nodegroup
        service: iks
      name: iks:nodegroup-controller
    rules:
    # Get apiserver health status.
    - nonResourceURLs:
      - "/healthz"
      - "/healthz/*"
      - "/livez"
      - "/livez/*"
      - "/readyz"
      - "/readyz/*"
      verbs:
      - get
    # Create bootstrap token secret and weka storage secret.
    - apiGroups:
      - ""
      resources:
      - secrets
      - namespaces
      verbs:
      - create
      - get
    # Manage nodes.
    - apiGroups:
      - ""
      resources:
      - nodes
      verbs:
      - "*"
    - apiGroups:
      - ""
      resources:
      - nodes/status
      verbs:
      - patch
      - update
    - apiGroups:
      - ""
      resources:
      - pods/eviction
      verbs:
      - create
    - apiGroups:
      - ""
      resources:
      - pods
      verbs:
      - get
      - list
      - watch
      - delete
    - apiGroups:
      - "apps"
      resources:
      - "*"
      verbs:
      - get
    # Approve kubelet serving certificates.
    - apiGroups:
      - certificates.k8s.io
      resources:
      - certificatesigningrequests
      verbs:
      - get
      - list
      - create
    - apiGroups:
      - certificates.k8s.io
      resources:
      - certificatesigningrequests/approval
      verbs:
      - update
    - apiGroups:
      - certificates.k8s.io
      resources:
      - signers
      verbs:
      - approve
    END

    kubectl apply -f - <<END
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: iks:addon-controller
    rules:
    - apiGroups:
      - ""
      resources:
      - serviceaccounts
      verbs:
      - create
    - apiGroups:
      - ""
      resourceNames:
      - kube-proxy
      - tigera-operator
      - konnectivity-agent
      resources:
      - serviceaccounts
      verbs:
      - "*"
    - apiGroups:
      - rbac.authorization.k8s.io
      resources:
      - clusterrolebindings
      verbs:
      - create
    - apiGroups:
      - rbac.authorization.k8s.io
      resourceNames:
      - system:kube-proxy
      - system:coredns
      - tigera-operator
      - system:konnectivity-server
      resources:
      - clusterrolebindings
      verbs:
      - "*"
    - apiGroups:
      - rbac.authorization.k8s.io
      resources:
      - clusterroles
      verbs:
      - create
    - apiGroups:
      - rbac.authorization.k8s.io
      resourceNames:
      - system:kube-proxy
      - system:coredns
      - tigera-operator
      resources:
      - clusterroles
      verbs:
      - "*"
    - apiGroups:
      - authentication.k8s.io #required for konnectivity agents clusterrole
      resources:
      - tokenreviews
      verbs:
      - create
    - apiGroups:
      - authorization.k8s.io #required for konnectivity agents clusterrole
      resources:
      - subjectaccessreviews
      verbs:
      - create
    - apiGroups:
      - "" #required for kube-proxy clusterrole
      resources:
      - endpoints
      - nodes
      verbs:
      - get
      - list
      - watch
    - apiGroups:
      - "" #required for kube-proxy clusterrole
      - "events.k8s.io"
      resources:
      - events
      verbs:
      - create
      - patch
      - update
    - apiGroups:
      - "discovery.k8s.io" #required for kube-proxy clusterrole
      resources:
      - endpointslices
      verbs:
      - list
      - watch
    - apiGroups:
      - ""
      resources:
      - configmaps
      - secrets
      verbs:
      - create
    - apiGroups:
      - ""
      resourceNames:
      - kube-proxy
      - coredns
      resources:
      - configmaps
      verbs:
      - "*"
    - apiGroups:
      - "apps"
      resources:
      - daemonsets
      verbs:
      - create
    - apiGroups:
      - "apps"
      resourceNames:
      - kube-proxy
      - konnectivity-agent
      resources:
      - daemonsets
      verbs:
      - "*"
    - apiGroups:
      - "apps"
      resources:
      - deployments
      verbs:
      - create
    - apiGroups:
      - "apps"
      resourceNames:
      - tigera-operator
      - coredns
      resources:
      - deployments
      verbs:
      - "*"
    - apiGroups:
      - ""
      resources:
      - services
      verbs:
      - create
      - list
      - watch
    - apiGroups:
      - ""
      resourceNames:
      - kube-dns
      resources:
      - services
      verbs:
      - "*"
    - apiGroups:
      - ""
      resources:
      - namespaces
      verbs:
      - create
    - apiGroups:
      - ""
      resourceNames:
      - tigera-operator
      resources:
      - namespaces
      verbs:
      - "*"
    - apiGroups:
      - "apiextensions.k8s.io"
      resources:
      - customresourcedefinitions
      verbs:
      - create
      - get
      - update
    - apiGroups:
      - "crd.projectcalico.org"
      resources:
      - bgpconfigurations
      - bgpfilters
      - bgppeers
      - blockaffinities
      - caliconodestatuses
      - clusterinformations
      - felixconfigurations
      - globalnetworkpolicies
      - globalnetworksets
      - hostendpoints
      - caliconodestatuses
      - ipamblocks
      - ipamconfigs
      - ipamhandles
      - ippools
      - ipreservations
      - kubecontrollersconfigurations
      - networkpolicies
      - networksets
      verbs:
      - "*"
    - apiGroups:
      - "operator.tigera.io"
      resources:
      - "*"
      verbs:
      - "*"
    # The following permissions are required for the calico tigera operator clusterrole
    - apiGroups:
      - ""
      resources:
      - nodes
      verbs:
      - patch
    - apiGroups:
      - ""
      resources:
      - resourcequotas
      verbs:
      - get
      - list
      - watch
    - apiGroups:
      - ""
      resources:
      - endpoints
      verbs:
      - create
      - update
      - delete
    - apiGroups:
      - ""
      resources:
      - services
      verbs:
      - get
      - update
      - delete
    - apiGroups:
      - ""
      resources:
      - events
      verbs:
      - get
      - list
      - delete
      - watch
    - apiGroups:
      - ""
      resources:
      - configmaps
      - namespaces
      - secrets
      - serviceaccounts
      verbs:
      - get
      - list
      - update
      - delete
      - watch
    - apiGroups:
      - ""
      resources:
      - pods
      - podtemplates
      verbs:
      - create
      - get
      - list
      - update
      - delete
      - watch
    - apiGroups:
      - "apiregistration.k8s.io"
      resources:
      - apiservices
      verbs:
      - create
      - list
      - update
      - watch
    - apiGroups:
      - "apps"
      resources:
      - daemonsets
      - deployments
      verbs:
      - get
      - list
      - patch
      - update
      - delete
      - watch
    - apiGroups:
      - "apps"
      resourceNames:
      - tigera-operator
      resources:
      - deployments/finalizers
      verbs:
      - update
    - apiGroups:
      - "apps"
      resources:
      - statefulsets
      verbs:
      - create
      - get
      - list
      - patch
      - update
      - delete
      - watch
    - apiGroups:
      - "certificates.k8s.io"
      resources:
      - certificatesigningrequests
      verbs:
      - list
      - watch
    - apiGroups:
      - "coordination.k8s.io"
      resources:
      - leases
      verbs:
      - create
      - get
      - list
      - update
      - delete
      - watch
    - apiGroups:
      - "networking.k8s.io"
      resources:
      - networkpolicies
      verbs:
      - create
      - get
      - list
      - update
      - delete
      - watch
    - apiGroups:
      - "policy"
      resources:
      - poddisruptionbudgets
      - podsecuritypolicies
      verbs:
      - create
      - get
      - list
      - update
      - delete
      - watch
    - apiGroups:
      - "policy"
      resourceNames:
      - tigera-operator
      resources:
      - podsecuritypolicies
      verbs:
      - use
    - apiGroups:
      - "rbac.authorization.k8s.io"
      resources:
      - clusterrolebindings
      - clusterroles
      verbs:
      - get
      - list
      - update
      - delete
      - watch
      - bind
      - escalate
    - apiGroups:
      - "rbac.authorization.k8s.io"
      resources:
      - rolebindings
      - roles
      verbs:
      - create
      - get
      - list
      - update
      - delete
      - watch
      - bind
      - escalate
    - apiGroups:
      - "scheduling.k8s.io"
      resources:
      - priorityclasses
      verbs:
      - create
      - get
      - list
      - update
      - delete
      - watch
    - apiGroups:
      - "storage.k8s.io"
      resources:
      - csidrivers
      - storageclasses
      verbs:
      - create
      - get
      - list
      - update
      - delete
      - watch
    END
    ```

  2. Install and configure Weka client on each worker node.
      
      On each worker node, ssh into it and run the following.

      ```
      # This should be set based on IDC environmet and instance type.
      # Ensure correct settings are being used before execution.
      STORAGE_AGENT_URL="internal-placeholder.com:14000"
      STORAGE_WEKA_SW_VERSION="4.2.2"
      STORAGE_WEKA_CONTAINER_NAME="client"
      STORAGE_WEKA_NUM_CORES=4
      
      # Get network information.
      INTERFACE_NAME=$(ip -o -4 route show to default | awk '{print $5}')

      # Some nodes will have a dedicated interface for storage configuration named "storage0-tenant".
      # If it is found, it will be used instead of default interface.
      if ip addr | grep "storage0-tenant"; then
        log_message "Dedicated storage interface found"
        INTERFACE_NAME="storage0-tenant"
      fi

      IP_ADDR=$(ip -o -4 addr show dev $INTERFACE_NAME | awk '{print $4}' | cut -d/ -f1 | head -n 1)
      NETMASK=$(ip -o -4 addr show dev $INTERFACE_NAME | awk '{print $4}' | cut -d/ -f2 | head -n 1)
      GATEWAY=$(ip -o -4 route show to default | awk '{print $3}')

      curl -vk "http://$STORAGE_AGENT_URL/dist/v1/install" | sh
      weka version get $STORAGE_WEKA_SW_VERSION
      weka version set $STORAGE_WEKA_SW_VERSION
      weka local setup container --name $STORAGE_WEKA_CONTAINER_NAME --cores $STORAGE_WEKA_NUM_CORES --only-frontend-cores --net "$INTERFACE_NAME/$IP_ADDR/$NETMASK/$GATEWAY"

      if weka local ps | grep -i stem >/dev/null; then
        log_message "The storage agent has been setup successfully"
      else
        log_message "There was an issue installing the required storage agent. Please check the output of 'weka local ps' for additional details"
      fi

      ```

3. In the IDC Console, enable Storage for the cluster.
4. Check the Storage status, it should become `Active`.
5. Check worker status.

    Run `kubectl` command and ensure all `csi-wekafs-*` pods are running.
    ```
    kubectl -n storage-system get pod
    ```