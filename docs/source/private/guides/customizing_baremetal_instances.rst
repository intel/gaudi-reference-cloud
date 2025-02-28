.. _customizing_baremetal_instances:

BaremetalHost intance customization with userData
#################################################

Baremetal instances can be customized by using the spec.userData field of the BaremetalHost CRD.

.. code:: bash

    apiVersion: metal3.io/v1alpha1
    kind: BareMetalHost
    metadata:
      name: host-0
      namespace: my-cluster
    spec:
      online: true
      bootMACAddress: 80:c1:6e:7a:e8:10
      bmc:
        address: ipmi://192.168.1.13
        credentialsName: host-0-bmc
      image:
        checksum: http://192.168.0.150/SHA256SUMS
        url: http://192.168.0.150/jammy-server-cloudimg-amd64.img
      userData:    <-------------------------------
        name: user-data-example

where the ``user-data-example`` is a  secret, which can be created in a couple of ways.

Example 1: create secret from a cloud-config.yaml file.
-------------------------------------------------------

Start with a cloud-init config file which creates a simple python script. The file
name is ``cloud-config-example.yaml``, ad its contents look like:

.. code:: bash

    #cloud-config
        write_files:
          - content: |
              #!/usr/bin/env python3
              print("This is an example of a python script written using the spec.userData of the BaremetalHost CRD")
            path: /usr/bin/uer-data-example
            owner: root:root
            permissions: '755'

with the appropriate access defined in our KUBECONFIG to reach the cluster, the secret maybe created by using he command:

.. code:: bash

    kubectl create secret generic user-data-secret-example --from-file=userData=cloud-config-example.yaml

The result is a created secret in the cluster:

.. code:: bash

    kubectl get secret user-data-secret-example -o yaml

.. code:: bash

    apiVersion: v1
    data:
      userData: I2Nsb3VkLWNvbmZpZwp3cml0ZV9maWxlczoKICAtIGNvbn...
    kind: Secret
    metadata:
      creationTimestamp: "2024-10-04T03:47:23Z"
      name: user-data-secret-example
      namespace: default
      resourceVersion: "3091886"
      uid: 0a33591c-920b-4824-b974-82c5ca90d35e
    type: Opaque


to verify the contents you can use the following command

.. code:: bash

    kubectl get secret user-data-secret-example -o jsonpath={.data.userData} | base64 -d

.. code:: bash

    #cloud-config
    write_files:
      - content: |
          #!/usr/bin/env python3
          print("This is an example of a python script written using the spec.userData of the BaremetalHost CRD")
        path: /usr/bin/uer-data-example
        owner: root:root


Example 2: create secret from a kubernetes secret yaml file
-----------------------------------------------------------
include the contents of the cloud-config.yaml file content as port of the
stringData  of a kubernetes secret definition.

For example, the contents of the cloud-config.yaml file look like:

.. code:: bash

    #cloud-config
    write_files:
      - content: |
          #!/usr/bin/env python3
          print("This is an example of a python script written using the spec.userData of the BaremetalHost CRD")
        path: /usr/bin/uer-data-example
        owner: root:root
        permissions: '755'

Create a secret yaml file that looks like the following. (filename ``user-data-secret-example.yaml``):

.. code:: bash

    apiVersion: v1
    kind: Secret
    metadata:
      name: user-data-secret-example
    type: Opaque
    stringData:
      userData: |
        #cloud-config
        write_files:
          - content: |
              #!/usr/bin/env python3
              print("This is an example of a python script written using the spec.userData of the BaremetalHost CRD")
            path: /usr/bin/uer-data-example
            owner: root:root
            permissions: '755'


and apply it

.. code:: bash

    kubectl apply -f user-data-secret-example.yaml

once the secret is created succesfully, you can look at it's contents the usual way.

.. code:: bash

    kubectl get secret user-data-secret-example -o jsonpath={.data.userData} | base64 -d

.. code:: bash

    #cloud-config
    write_files:
      - content: |
          #!/usr/bin/env python3
          print("This is an example of a python script written using the spec.userData of the BaremetalHost CRD")
        path: /usr/bin/uer-data-example
        owner: root:root

Example 3: using userDatat to reconfigure instance host networking
------------------------------------------------------------------

Using one of the methods in the examples above and the following content for ``cloud-config-example.yaml``, the host
networking is configured  with the eth0 port attached to the br0 linux bridge.

.. code:: bash

    #cloud-config

    write_files:
    - content: |
        #!/usr/bin/env python3
        from jinja2 import Template
        from netifaces import ifaddresses, AF_LINK
        from os import chmod
        import sys
        ifname = sys.argv[1]
        brname = sys.argv[2]
        mtu = sys.argv[3]
        dst_file = sys.argv[4]
        mac = ifaddresses(ifname)[AF_LINK][0].get('addr')
        template = Template("""network:
            bridges:
                {{  bridge_name }}:
                    dhcp4: true
                    interfaces:
                    - {{ ethernet_name }}
            ethernets:
                {{ ethernet_name }}:
                    match:
                        macaddress: '{{ ethernet_mac }}'
                    mtu: {{ ethernet_mtu }}
                    set-name: {{ ethernet_name }}
            renderer: networkd
            version: 2
            """)
        with open(dst_file, 'w') as f:
            result = template.render(ethernet_name=ifname, bridge_name=brname,ethernet_mac=mac, ethernet_mtu=mtu)
            f.write(result)
        chmod(dst_file, 0o600)

      path: /usr/bin/create_idc_net_plan
      owner: root:root
      permissions: '755'

    - content: 'network: {config: disabled}'
      path: /etc/cloud/cloud.cfg.d/99-custom-networking.cfg
      permissions: "0644"

    - content: |-
        #!/bin/bash
        [[ "${IFACE}" == "eth0" ]] || exit 0
        ip link set dev br0 address $(ip link show dev ${IFACE} | grep link/ether | awk '{print $2}')
        bridge vlan add vid 2-4094 dev ${IFACE}
      path: /etc/networkd-dispatcher/configured.d/eth0
      permissions: "0755"

    - content: |-
        #!/bin/bash
        [[ "${IFACE}" == "br0" ]] || exit 0
        ip link set ${IFACE} type bridge vlan_filtering 1
        bridge vlan add vid 2-4094 dev ${IFACE} self
      path: /etc/networkd-dispatcher/configured.d/br0
      permissions: "0755"


    runcmd:
     - rm /etc/netplan/*
     - /usr/bin/create_idc_net_plan eth0 br0  9000  /etc/netplan/100-idc.yaml
     - netplan generate
     - netplan apply
     - /etc/networkd-dispatcher/configured.d/eth0
     - /etc/networkd-dispatcher/configured.d/br0
