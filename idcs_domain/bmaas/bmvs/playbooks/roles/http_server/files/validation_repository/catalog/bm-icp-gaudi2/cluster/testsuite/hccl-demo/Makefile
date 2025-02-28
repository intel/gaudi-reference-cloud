.PHONY: setup-ighs
setup-ighs:
	@echo "Installing dependencies for IGHS and HCCL-Demo"
	sudo pip3 install ansible
	sudo apt install pdsh -y
	
.PHONY: setup-fw-version
setup-fw-version:
	@echo "Setting versions based on FW_VERSION"
	cd /tmp/validation/testsuite/fw-version && ansible-playbook main.yml

.PHONY: generate-hostfile
generate-hostfile:
	cd playbooks && ansible-playbook -i inventory setup-ighs.yml --tags=gen_hostfile --extra-vars "MEMBER_IPS='$(MEMBER_IPS)' MASTER_IP='$(MASTER_IP)'"
	@echo "Updating inventory with hostnames..."
	cd playbooks && ansible-playbook -i inventory hostnames.yml

.PHONY: generate
generate:
	@echo "Generating SSH Key pair for IGHS and HCCL-Demo"
	cd playbooks && ansible-playbook -i inventory setup-ighs.yml --tags=gen_ssh_key

.PHONY: generate-config
generate-config:
	@echo "Generating config file for IGHS and HCCL-Demo"
	cd playbooks && ansible-playbook -i inventory setup-ighs.yml --tags=gen_config_file

.PHONY: collect-logs
collect-logs:
	@echo "Collecting Habana logs and IGHS logs"
	cd playbooks && ansible-playbook -i inventory collect-logs.yml

.PHONY: clean
clean:
	@echo "Triggering cleanup"
	cd playbooks && ansible-playbook -i inventory cleanup.yml
	rm -rf playbooks/inventory
	@echo "Hostfile in inventory folder removed."
	@echo "Cleanup completed successfully."

.PHONY: run-hccl-demo
run-hccl-demo: setup-ighs setup-fw-version generate-hostfile generate generate-config
	@echo "Executing IGHS test for cluster validation..."
	cd playbooks/roles/ighs/files/utils/ighs_hccl_demo/ && bash run_ighs.sh
	$(MAKE) collect-logs
	$(MAKE) clean

.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo "Available targets for IGHS test:"
	@echo "  setup-ighs          Install dependencies for IGHS and HCCL-Demo"
	@echo "  setup-fw-version    Set firmware versions based on FW_VERSION"
	@echo "  generate-hostfile   Generate and update hostfile with hostnames"
	@echo "  generate            Generate SSH Key pair for IGHS and HCCL-Demo"
	@echo "  generate-config     Generate config file for IGHS and HCCL-Demo"
	@echo "  collect-logs        Collect Habana logs and IGHS logs"
	@echo "  clean               Cleanup generated files and inventory"
	@echo "  run-hccl-demo       Run the HCCL demo with setup and cleanup"
