#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
set -ex
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
source "${SCRIPT_DIR}/defaults.sh"

cat <<EOF | \
curl ${CURL_OPTS} \
-H 'Content-type: application/json' \
-H "Origin: http://localhost:3001/" \
-H "Authorization: Bearer ${TOKEN}" \
-X POST \
${IDC_REGIONAL_URL_PREFIX}/v1/cloudaccounts/${CLOUDACCOUNT}/instances --data-binary @- \
| jq .
{
  "metadata": {
    "name": "${NAME}"
  },
  "spec": {
    "availabilityZone": "${AZONE}",
    "instanceType": "${INSTANCE_TYPE}",
    "machineImage": "${MACHINE_IMAGE}",
    "runStrategy": "RerunOnFailure",
    "sshPublicKeyNames": [
      "${KEYNAME}"
    ],
    "interfaces": [
      {
        "name": "eth0",
        "vNet": "${VNETNAME}"
      }
    ],
    "userData": "\
#cloud-init\n\
packages:\n\
  - qemu-guest-agent\n\
  - socat\n\
  - conntrack\n\
  - ipset\n\
ssh_authorized_keys:\n\
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCcO4pgGXXz/IDOOk0QcK4n45bkwfhr8TgCLLN1e2Qm5Zpda6egvpeI+ZrYpYNnEvCeIrZHFjwCL0JvkqY2xvf0EF5BiOa1dWc/eDj9csgW0xihQUETcUbgEQDCy8Ph3t7DGqw+h5yh6CwT2oIe9jcNnQmd+097X8aYvxk3zVx8/E7QqBmUDUH23U1VDdHOiB4ie+QUsUrmsKVxI3zZhpDvToY7maRS2TfJe0wucGrGrpvqOx+YF82lFtZRWVxKG+LOBUUTA560+O3XVf6BRCPTK/uvs9KAJsLqGbAyg+NgmbgPvixM19jTaJ7mJstwsLdvvxtcYJ+uVjnxDiR2eEB4fBb5nD/4hIxjzwstYk3EPt2Z08iHA30N29XMbcsecwqZEJqYELLrOrwJOeBB3A6sYqQYv0jxm9GytC4TNB9u63RcJ/tQpzYYauVcRszppAjuE4F4oWGolALqEpcyIczHA+bhKUyAGvy1AXZr5psoC1Dzs6qJKiNmOdUb3YspP0CCHGasNU3hzF10Lja+RvtEUb4DEKnM4D9eAQ7Frky+f/LYYkmbEqs8RkaDQQ745R/9SZnbWfdyLj/vrdXnj7XXR7OzGucqHRZyq/U/3C/D9aADiReIf9Z0V4syTG4xlkpqgBdpOy6pauMQso448tYa+AyNKd3tttYGOKhDcJXHYQ== test1\n\
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCwNSnjr+G3PcwY6RmFSAtZ8D5U9P4jha50btd9gToviKauufgzd13MrvdXFhWAsBo12p8YFT/r9y213yZuyClEqi3Tq7iEJZqlq6aeYo5vRJXQxzc7TN2ON4TBYXbXg1rZb5qLz8rUknTGBU0JrzLAQTSmxWYPH+7snU9yi9kfgc1L3D1735NeOpwydt3itV/LaayV2qu9Xy7q5SyOo7qbhxhgzAMcKbAcMpcKZtdnLaSCb20HKxf7fzqJpOWws15Z4rc7PMD+xU19+PsuJAS9FV8K7rWxRmhA40AbrdIHPeI+bnCk0MLqQ1LAkPRmfmxX5EKQ6mwP/ykCnRnjyiFSE1g8ejEo/p7cfTi8g6kw8qVdJy1CgnhobZ23QqSBJGWC1OvVY3M8McDdztJ2vVLyiGFMKGvn8y9bYGayiH1fwwAl92YvqBLCWkV1iU0dDdvKWD+opYMOqR/Y+a3aZAmUZwO8M30IKA4WpalyE6lKfGa8RM4mgyMYy9k4s65Uw5PVa8Q1hOMWDlqjAcFKUZqeG210U1A8if21KVEt6SECZfljsLgIOOnTjBY2mx6oeFPIeEnN2dXu/WWtSb0POXHDra0ksh/Yg3eqyutk6wBPfdwlo70ARq2EfmqFYFDlbtgALQkJBn10KlZoiH0505tayH8zsBojyCVaQsqNmWgygw== test2\n\
write_files:\n\
  - path: /usr/local/bin/idc_get_network_info\n\
    permissions: '755'\n\
    content: |\n\
      #!/usr/bin/env python3\n\
      import sys\n\
      import yaml\n\
      input_file = sys.argv[1]\n\
      output_file = sys.argv[2]\n\
      eth_info = {}\n\
      with open(input_file) as f:\n\
          input_data = yaml.safe_load(f)\n\
          eth_name = list(input_data['network']['ethernets'].keys())[0]\n\
          eth_info = input_data['network']['ethernets'][eth_name]\n\
          eth_info['interfaces'] = [eth_name]\n\
          eth_info['version'] = input_data['network']['version']\n\
      with open(output_file, 'w') as f:\n\
          yaml.dump(eth_info, f)\n\
\n\
  - path: /usr/local/bin/idc_add_iface_to_bridge\n\
    permissions: '755'\n\
    content: |\n\
      #!/usr/bin/env python3\n\
      import os\n\
      import sys\n\
      import yaml\n\
      bridge_name = sys.argv[1]\n\
      input_file = sys.argv[2]\n\
      output_file = sys.argv[3]\n\
      with open(input_file) as f:\n\
          input_data = yaml.safe_load(f)\n\
          new_netplan = {\n\
              'network': {\n\
                  'version': input_data.get('version'),\n\
                  'bridges': {\n\
                      bridge_name: {\n\
                          'addresses': input_data.get('addresses'),\n\
                          'mtu': input_data.get('mtu'),\n\
                          'nameservers': input_data.get('nameservers'),\n\
                          'routes': input_data.get('routes'),\n\
                          'interfaces': input_data.get('interfaces'),\n\
                      }\n\
                  },\n\
                  'ethernets': {\n\
                      input_data.get('set-name'): {\n\
                          'match': input_data.get('match'),\n\
                          'mtu': input_data.get('mtu'),\n\
                          'set-name': input_data.get('set-name'),\n\
                      }\n\
                  }\n\
              }\n\
          }\n\
      with open(output_file, 'w') as f:\n\
          yaml.dump(new_netplan, f)\n\
          os.chmod(output_file, 0o600)\n\
      print(input_data.get('set-name'))\n\
\n\
  - path: /usr/local/bin/idc_config_network_dispatcher\n\
    permissions: '755'\n\
    content: |\n\
      #!/usr/bin/env python3\n\
      from jinja2 import Template\n\
      from pathlib import Path\n\
      import sys\n\
      import os\n\
      br_name = sys.argv[1]\n\
      eth_name = sys.argv[2]\n\
      fileTemplates = {}\n\
      dirpath=\"/etc/networkd-dispatcher/configured.d/\"\n\
      fileTemplates[dirpath+eth_name] =\"\"\"#!/bin/bash\n\
      [[ \"\${IFACE}\" == \"{{ ETH_NAME }}\" ]] || exit 0\n\
      ip link set dev {{ BR_NAME }} address \$(ip link show dev \${IFACE} | grep link/ether | awk \"{print \$2}\")\n\
      bridge vlan add vid 2-4094 dev \${IFACE}\n\
  \n\
      \"\"\"\n\
  \n\
      fileTemplates[dirpath+br_name]=\"\"\"#!/bin/bash\n\
      [[ \"\${IFACE}\" == \"{{ BR_NAME }}\" ]] || exit 0\n\
      ip link set \${IFACE} type bridge vlan_filtering 1\n\
      bridge vlan add vid 2-4094 dev \${IFACE} self\n\
  \n\
      \"\"\"\n\
  \n\
      Path(dirpath).mkdir(parents=True, exist_ok=True)\n\
      for file in fileTemplates:\n\
          tm = Template(fileTemplates[file])\n\
          file_content = tm.render(ETH_NAME=eth_name, BR_NAME=br_name)\n\
          with open(file, 'w') as f:\n\
              f.write(file_content)\n\
          os.chmod(file, 0o755)\n\
  \n\
  - path: /usr/local/bin/idc_setup_network\n\
    permissions: '755'\n\
    content: |\n\
      #!/bin/bash\n\
      set -xe\n\
      /usr/local/bin/idc_get_network_info /etc/netplan/50-cloud-init.yaml /tmp/idc_net_variables.yaml\n\
      mkdir -p ~/idc_net_old\n\
      mv /etc/netplan/* ~/idc_net_old\n\
      iface_name=\$(/usr/local/bin/idc_add_iface_to_bridge mgmt-br /tmp/idc_net_variables.yaml /etc/netplan/100-idc_net.yaml)\n\
      /usr/local/bin/idc_config_network_dispatcher mgmt-br \${iface_name}\n\
      netplan generate\n\
      netplan apply\n\
runcmd:\n\
- /usr/local/bin/idc_setup_network\n\
"
  }
}
EOF
