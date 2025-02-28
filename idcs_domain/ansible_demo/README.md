# Demonstration of using Ansible to provision IDC infrastructure

This directory demonstrates how an IDC user can use Ansible to provision IDC infrastructure (instances)
along with applications.

## Prerequisites

1. Install kubectl.

2. Download your Kubernetes config file from the IDC Cloud Console and place it in ~/.kube/config.

3. Run:

    ```bash
    make install-requirements
    ```

## Deploy Slurm Workload Manager

This will deploy several virtual machine instances on IDC and deploy [Slurm](https://slurm.schedmd.com/).

1. Edit the [Ansible variables file](playbooks/vars/slurm.yml).

2. Run:

    ```bash
    make slurm
    ```

## Deploy a generic cluster (without Slurm)

This will deploy several virtual machine instances on IDC.

1. Edit the [Ansible playbook](playbooks/basic_workers.yml).

2. Run:

    ```bash
    make basic_workers
    ```
