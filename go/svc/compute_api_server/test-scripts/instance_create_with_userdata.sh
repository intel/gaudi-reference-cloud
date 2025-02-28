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
  - path: /etc/environment\n\
    append: true\n\
    content: |\n\
      HTTP_PROXY=http://internal-placeholder.com:912/\n\
      HTTPS_PROXY=http://internal-placeholder.com:912/\n\
      NO_PROXY=127.0.0.1,127.0.1.1,localhost,.intel.com\n\
"
  }
}
EOF
