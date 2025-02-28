# Gaudi2 Validation - Distributed run using Deepspeed in Docker Containers

This Deepspeed test requires the passwordless SSH across the BM nodes on the cluster

Directory Structure:

```bash
.
├── deepspeed
│   ├── Makefile
│   ├── README.md
│   └── playbooks
│       ├── ansible.cfg
│       ├── deepspeed-e2e.yml
│       ├── docker-image.yml
│       ├── inventory
│       ├── roles
│       │   ├── build_image
│       │   │   ├── files
│       │   │   └── tasks
│       │   │       ├── build_docker_image.yml
│       │   │       ├── main.yml
│       │   │       └── save_docker_image.yml
│       │   ├── create_hostfile
│       │   │   ├── tasks
│       │   │   │   └── main.yml
│       │   │   └── templates
│       │   │       ├── bthostfile.j2
│       │   │       └── hostfile.j2
│       │   ├── create_val_dir
│       │   │   ├── tasks
│       │   │   │   └── main.yml
│       │   │   └── templates
│       │   │       ├── create-dockerfile.j2
│       │   │       ├── deepspeed-env.j2
│       │   │       └── deepspeed-run.j2
│       │   ├── deepspeed_test
│       │   │   ├── files
│       │   │   │   ├── hostfile
│       │   │   │   ├── launch_container.sh
│       │   │   │   └── limits.conf
│       │   │   └── tasks
│       │   │       ├── main.yml
│       │   │       └── setup_deepspeed.yml
│       │   ├── gen_ssh_key
│       │   │   └── tasks
│       │   │       └── main.yml
│       │   ├── load_image
│       │   │   └── tasks
│       │   │       └── main.yml
│       │   └── setup_network
│       │       ├── files
│       │       │   └── setup_network.sh
│       │       └── tasks
│       │           └── main.yml
│       ├── setup-network.yml
│       ├── trigger_cleanup.yml
│       ├── val-dir.yml
│       └── vars
│           ├── common.yml
│           └── val-dir.yml
├── start.sh
└── validation.tar.gz
```

### Gaudi2 Cluster Validation using Deepspeed Test through Ansible by Validation Operator

**Note:** 
- `validation-0.0.1.tar.gz` contains `start.sh` and `deepspeed` which is extracted by the Validation Operator.
- The Deepspeed Test is executed as a part of Cluster Validation by Validation Operator by triggering the execution of `start.sh` which has the commands to execute the entire test.

**Upon the execution of `start.sh`, the Test is executed in the following steps:**

1. All the dependencies like `make` and `ansible` are installed by start.sh once it is executed by Validation Operator.
-  **Note**: Password-less SSH for connection between nodes is added by PR: https://github.com/intel-innersource/frameworks.cloud.devcloud.services.idc/pull/4141

2. Both the Deepspeed Test and the ansible need the hostfile to run the Deepspeed Test and the Playbooks respectively.
   These hostfile(s) are generated through `start.sh` which runs the following command:
   ```bash
   ansible-playbook "/tmp/validation/deepspeed/playbooks/create-hostfile.yml" --extra-vars "printenv_output='$output'"
   ```

- The hostfile `bt/hostfile` (refer directory structure) will be created the list of all the IP(s) of the Master Node and Members Nodes. The First IP is of the Master Node.

  For example:
  
  ```bash
  100.83.155.75 slots=8
  100.83.155.86 slots=8
  ```
  Here, 100.83.155.75 is the IP of Master Node

- The `inventory/hosts.yaml` which is the ansible inventory will be created with the list of all the hosts.

  For example:

  ```bash
  hosts:
  hosts:
    node-0:
      ansible_host: "100.83.155.72"

    node-1:
      ansible_host: "100.83.155.86"
  ```
  Here, `node-0` is the MASTER_IP (Master Node).
  
3. SSH Keys are generated on `Master Node` to SSH to other nodes from inside the Docker Container on Master Node during the execution of Deepspeed Test.
	```bash
	make generate
	```
	This generates SSH key pair in `playbooks/roles/gen_ssh_key/files/ssh_keys` directory which are also copied into the `playbooks/roles/deepspeed_test/files/.ssh_keys` directory. Please note that the key pair is copied and not moved for backup purposes.

4. Copy all the files needed to run Deepspeed Tests to Member nodes
   
	```bash
	make create-validation-directory
	```
	This command does the following:
	
	- Creates a folder `tmp` with SSH Keys, deepspeed scripts
	- Generates the Deepspeed run script from `run.j2` template
	- Generates the DockerFile from `create-dockerfile.j2` template
	- Copies the `tmp` folder to a newly created directory `$HOME/validation` on all the nodes including Master node and delete it once copying is complete
	- Clones the `optimum-habana` **(v1.10-release)** inside the `$HOME/validation` from [huggingface/optimum-habana](https://github.com/huggingface/optimum-habana/tree/v1.10-release) github repo
 	- Update the `transformers` to `4.37.0`

5. Build Docker Image and Run Container
	```bash
 	make build
	```
	This creates Docker Image `btower` with passwordless SSH and custom SSH port on all nodes and launches docker container using `launch_container.sh`
	
	- `launch_container.sh` also makes sure that no other container is running, thus avoiding any SSH port problems.
	- --network=host ensures that host and container IPs are the same

6. After successful execution of `hl-qual` test, we need to setup the network again to allow the successful execution of Deepspeed Test. This can be done by running following command:
   ```bash
   make setup-network
   ```

7. Finally, Run the Deepspeed Test:
   
   ```bash
   make deepspeed
   ```
   This will docker exec into the Master Node and run the `run.sh` which will initiate the Deepspeed Test.
   Validation is completed once the test is completed.

8. To ensure all allocated directories, images, etc are remove from all the hosts, use the `clean` target.

   ```bash
   make clean
   ```

8. The Deepspeed Test is executed E2E through the make target `validate`
 
   ```bash
   make validate
   ```
   Once the test is completed, trigger cleanup using `clean` target.
