# Ansible playbooks

This folder contains the playbooks that automate the nodes configurations (for Jenkins, Docker, Kind deployment, RKE2 nodes, etc).


> ### **Note**
> Current status: Available to use from your own machine (local).
>
> Next steps: Include them in Jenkins to fully complete the automation.

## How to run them?

1. Make sure you have Ansible installed. If not, follow the instructions [here](https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html).
1. As we are using encryption for secrets, Ansible require a vault password to decrypt them. So, create in this directory a file called *vault_file.txt* and copy the content of the password vault stored in PAM.
1. To run a playbook, use the following command: 
`ansible-playbook <playbook-name>`.

### RKE2 node playbook

This playbook requires an extra variable for the token:

`ansible-playbook --extra-vars 'token=<node registration token>' rke2_node.yml`
