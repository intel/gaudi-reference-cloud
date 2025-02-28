.PHONY: help
help:
	@echo "Available commands:"
	@echo "  setup                     Install ansible"
	@echo "  setup-fw-version          Set firmware versions based on FW_VERSION"
	@echo "  setup-hostfile            Generate and update hostfile with hostnames"
	@echo "  generate                  Generate SSH Key pair"
	@echo "  create-validation-directory Create validation directory with specified parameters"
	@echo "  build                     Build DeepSpeed environment with specified versions"
	@echo "  collect-logs              Collect Gaudi logs"
	@echo "  setup-network             Setup network"
	@echo "  run-deepspeed             Run the DeepSpeed validation script inside Docker"
	@echo "  clean                     Cleanup generated files and inventory"
	@echo "  build-package-offline     Build Docker image for DeepSpeed offline package"
	@echo "  load-docker-image         Load Docker image for DeepSpeed"
	@echo "  setup-deepspeed           Setup DeepSpeed with specified versions"
	@echo "  validate-cluster-e2e      Validate cluster end-to-end, optionally skipping network setup"
	@echo "  run-deepspeed-offline     Run DeepSpeed validation script using offline package"

SOURCE_ENV := $(source .env)

.PHONY: setup
setup:
	sudo pip3 install ansible

.PHONY: setup-fw-version
setup-fw-version:
	@echo "Setting versions based on FW_VERSION"
	cd /tmp/validation/testsuite/fw-version && ansible-playbook main.yml

.PHONY: setup-hostfile
setup-hostfile:
	@$(SOURCE_ENV) cd playbooks && ansible-playbook -i inventory val-dir.yml --tags=create_hostfile --extra-vars "MEMBER_IPS='$(MEMBER_IPS)' MASTER_IP='$(MASTER_IP)'"
	@echo "Updating inventory with hostnames..."
	cd playbooks && ansible-playbook -i inventory set-hostnames.yml

.PHONY: generate
generate:
	cd playbooks && ansible-playbook -i inventory val-dir.yml --tags=gen_ssh_key

.PHONY: create-validation-directory
create-validation-directory:
	cd playbooks && ansible-playbook -i inventory val-dir.yml --tags=create_val_dir --extra-vars "master_ip='$(MASTER_IP)' hf_home='$(HF_HOME)' pt_hpu_pool_mem_acquire_perc='$(PT_HPU_POOL_MEM_ACQUIRE_PERC)' pt_hpu_max_compound_op_size='$(PT_HPU_MAX_COMPOUND_OP_SIZE)' pt_hpu_lazy_acc_par_mode='$(PT_HPU_LAZY_ACC_PAR_MODE)' pt_hpu_recipe_cache_config='$(PT_HPU_RECIPE_CACHE_CONFIG)' layer3_enabled='$(L3_ENABLED)' optimum_habana_version='$(OPTIMUM_HABANA_VERSION)'"

.PHONY: build
build:
	@$(SOURCE_ENV) cd playbooks && ansible-playbook -i inventory deepspeed-e2e.yml --extra-vars "transformers_version='$(TRANSFORMERS_VERSION)', deepspeed_version='${DEEPSPEED_VERSION}' optimum_habana_version='$(OPTIMUM_HABANA_VERSION)'"

.PHONY: collect-logs
collect-logs:
	@$(SOURCE_ENV) cd playbooks && ansible-playbook -i inventory collect_gaudi_logs.yml

.PHONY: setup-network
setup-network:
	cd playbooks && ansible-playbook -i inventory setup-network.yml

.PHONY: run-deepspeed
run-deepspeed:
	docker exec -t btower bash -c "chmod +x /validation/run.sh && /validation/run.sh" ||  ( $(MAKE) collect-logs && exit 1 )

.PHONY: clean
clean:
	cd playbooks && ansible-playbook -i inventory trigger_cleanup.yml
	@echo "Removing hostfiles..."
	@rm -f playbooks/roles/deepspeed_test/files/hostfile
	@echo "Hostfile in deepspeed_test folder removed."
	@rm -rf playbooks/inventory
	@echo "Hostfile in inventory folder removed."

.PHONY: build-package-offline
build-package-offline: setup setup-hostfile generate create-validation-directory
	cd playbooks && ansible-playbook -i inventory docker-image.yml --tags=build_image

.PHONY: load-docker-image
load-docker-image:
	cd playbooks && ansible-playbook -i inventory docker-image.yml --tags=load_image

.PHONY: setup-deepspeed
setup-deepspeed:
	@$(SOURCE_ENV) cd playbooks && ansible-playbook -i inventory roles/deepspeed_test/tasks/setup_deepspeed.yml --extra-vars "transformers_version='$(TRANSFORMERS_VERSION)' deepspeed_version='$(DEEPSPEED_VERSION)' optimum_habana_version='$(OPTIMUM_HABANA_VERSION)'"

.PHONY: validate-cluster-e2e
validate-cluster-e2e: setup setup-fw-version setup-hostfile generate create-validation-directory
ifeq ($(SKIP_NETWORK_SETUP),true)
	@echo "Skipping network setup..."
	$(MAKE) build
	$(MAKE) run-deepspeed
	$(MAKE) collect-logs
	$(MAKE) clean
else
	@echo "Running network setup..."
	$(MAKE) setup-network
endif

.PHONY: run-deepspeed-offline
run-deepspeed-offline: setup build-package-offline load-docker-image setup-network run-deepspeed clean
