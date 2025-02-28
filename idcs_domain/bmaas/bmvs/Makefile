ifneq ($(GUEST_HOST_DEPLOYMENTS),)
PLAYBOOK_EXTRAS += -e guest_host_deployments=$(GUEST_HOST_DEPLOYMENTS)
endif
ifneq ($(GUEST_HOST_MEMORY_MB),)
PLAYBOOK_EXTRAS += -e guest_host_memory_mb=$(GUEST_HOST_MEMORY_MB)
endif
ifneq ($(GUEST_HOST_DISK_GB),)
PLAYBOOK_EXTRAS += -e guest_host_disk_gb=$(GUEST_HOST_DISK_GB)
endif
ifneq ($(GUEST_HOST_VCPU),)
PLAYBOOK_EXTRAS += -e guest_host_vcpu=$(GUEST_HOST_VCPU)
endif
ifneq ($(GUEST_HOST_CPUSET),)
PLAYBOOK_EXTRAS += -e guest_host_cpuset=$(GUEST_HOST_CPUSET)
endif
ifneq ($(GUEST_TEST_HOST_DEPLOYMENTS),)
PLAYBOOK_EXTRAS += -e guest_test_host_deployments=$(GUEST_TEST_HOST_DEPLOYMENTS)
endif
ifneq ($(HTTP_SERVER_PORT),)
PLAYBOOK_EXTRAS += -e http_port=$(HTTP_SERVER_PORT)
endif




all: status

status:
	@echo "============================================================"
	@echo "Checking for active Nodes ..."
	@sudo virsh list --all || echo "No Nodes Active"
	@echo "============================================================"
	@echo "Checking for active Networks ..."
	@sudo virsh net-list --all || echo "No Networks Active"
	@echo "============================================================"
	@echo "Checking for active vBMCs ..."
	@pgrep -f "sushy-gh-[0-9]+-emulator.conf" -a || echo "No vBMCs Active"
	@echo ""
	@echo "============================================================"
	@echo "Checking for active HTTP Server ..."
	@pgrep -f "python3 -m http.server --directory .*8888" -a | grep -v pgrep || echo "No Active HTTP Server"
	@echo ""
	@if [ -n "${PLAYBOOK_EXTRAS}" ] ; then \
		echo "============================================================" ; \
		echo "PLAYBOOK_EXTRAS = '${PLAYBOOK_EXTRAS}'" ; \
		echo "" ; \
	fi

bifrost: setup
	sudo install -o root -g root -m 644 zzz_intel_proxy.sh /etc/profile.d/
	./install_bifrost.sh

start: setup
startup: setup
setup:
	cd playbooks ; ansible-playbook metal3-setup.yml ${PLAYBOOK_EXTRAS}

bmvs-setup:
	cd playbooks ; ansible-playbook bmvs-setup.yml ${PLAYBOOK_EXTRAS}

stop: teardown
destroy: teardown
teardown:
	cd playbooks ; ansible-playbook metal3-teardown.yml --tags teardown ${PLAYBOOK_EXTRAS}
	cd playbooks ; ansible-playbook guest-test.yml --tags teardown ${PLAYBOOK_EXTRAS}
	sudo pkill -f /opt/stack/bifrost

bmvs-teardown:
	cd playbooks ; ansible-playbook bmvs-teardown.yml --tags teardown ${PLAYBOOK_EXTRAS}
	cd playbooks ; ansible-playbook guest-test.yml --tags teardown ${PLAYBOOK_EXTRAS}

clean:
	cd playbooks ; ansible-playbook metal3-teardown.yml --tags teardown,clean ${PLAYBOOK_EXTRAS}
	cd playbooks ; ansible-playbook guest-test.yml --tags teardown,clean ${PLAYBOOK_EXTRAS}
	cd playbooks ; ansible-playbook download-cache.yml --tags teardown,clean ${PLAYBOOK_EXTRAS}
	sudo rm -rf /opt/stack/
	rm -rf ./bifrost/

bmvs-clean:
	cd playbooks ; ansible-playbook bmvs-teardown.yml --tags teardown,clean ${PLAYBOOK_EXTRAS}
	cd playbooks ; ansible-playbook guest-test.yml --tags teardown,clean ${PLAYBOOK_EXTRAS}
	cd playbooks ; ansible-playbook download-cache.yml --tags teardown,clean ${PLAYBOOK_EXTRAS}

start-guest-test: setup-guest-test
startup-guest-test: setup-guest-test
setup-guest-test:
	cd playbooks ; ansible-playbook guest-test.yml ${PLAYBOOK_EXTRAS}

stop-guest-test: teardown-guest-test
destroy-guest-test: teardown-guest-test
teardown-guest-test:
	cd playbooks ; ansible-playbook guest-test.yml --tags teardown ${PLAYBOOK_EXTRAS}

clean-guest-test:
	cd playbooks ; ansible-playbook guest-test.yml --tags teardown,clean ${PLAYBOOK_EXTRAS}

download-cache: cache
cache:
	cd playbooks ; ansible-playbook download-cache.yml ${PLAYBOOK_EXTRAS}

clean-cache:
	cd playbooks ; ansible-playbook download-cache.yml --tags teardown ${PLAYBOOK_EXTRAS}

lint:
	ansible-lint playbooks/metal3-setup.yml
	cd playbooks ; ansible-playbook --syntax-check metal3-setup.yml ${PLAYBOOK_EXTRAS}
	ansible-lint playbooks/metal3-teardown.yml
	cd playbooks ; ansible-playbook --syntax-check metal3-teardown.yml ${PLAYBOOK_EXTRAS}
	ansible-lint playbooks/bmvs-setup.yml
	cd playbooks ; ansible-playbook --syntax-check bmvs-setup.yml ${PLAYBOOK_EXTRAS}
	ansible-lint playbooks/bmvs-teardown.yml
	cd playbooks ; ansible-playbook --syntax-check bmvs-teardown.yml ${PLAYBOOK_EXTRAS}
	ansible-lint playbooks/download-cache.yml
	cd playbooks ; ansible-playbook --syntax-check download-cache.yml ${PLAYBOOK_EXTRAS}
	ansible-lint playbooks/guest-test.yml
	cd playbooks ; ansible-playbook --syntax-check guest-test.yml ${PLAYBOOK_EXTRAS}

# First requirement may be to install 'make'
# sudo apt install make
install-requirements:
	sudo apt install -y\
		openssh-client \
		libvirt-clients libvirt-daemon-system virtinst virt-manager \
		ansible-lint \
		python3-pip \
		cloud-image-utils \
		git
	sudo pip3 uninstall -y sushy-tools || true
	sudo rm -rf /tmp/sushy-tools
	git clone https://github.com/keedyandre/sushy-tools.git /tmp/sushy-tools
	sudo --preserve-env=HTTPS_PROXY pip3 install --ignore-installed /tmp/sushy-tools/
	sudo --preserve-env=HTTPS_PROXY pip3 install ansible
