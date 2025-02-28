// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package util

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudConfig", func() {
	It("BaremetalhostsReconciler.getCloudInit", func() {
		// The expected values were created using the same calls to the previous implementation
		expected := `#cloud-config
runcmd:
- install -D -m 400 /dev/null '/home/sdp/.ssh/id_rsa'
- echo -n AQIDBA== | base64 -d > '/home/sdp/.ssh/id_rsa'
- chown sdp:sdp /home/sdp/.ssh/id_rsa
`
		cloudConfig, err := NewEmptyCloudConfig("ubuntu")
		Expect(err).Should(Succeed())

		cloudConfig.AddRunBinaryFile("/home/sdp/.ssh/id_rsa", []byte{0x01, 0x02, 0x03, 0x04}, 0400)
		cloudConfig.AddRunCmd("chown sdp:sdp /home/sdp/.ssh/id_rsa")
		cloudConfig.SetRunCmd()
		cloudConfig.SetPackages()

		data, err := cloudConfig.RenderYAML()
		Expect(err).Should(Succeed())
		yaml := string(data)

		Expect(expected).To(Equal(yaml))
	})

	// TODO Why does juju insert Juju: in comment field?
	sshPublicKeys := []string{
		"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCcO4pgGXXz/IDOOk0QcK4n45bkwfhr8TgCLLN1e2Qm5Zpda6egvpeI+ZrYpYNnEvCeIrZHFjwCL0JvkqY2xvf0EF5BiOa1dWc/eDj9csgW0xihQUETcUbgEQDCy8Ph3t7DGqw+h5yh6CwT2oIe9jcNnQmd+097X8aYvxk3zVx8/E7QqBmUDUH23U1VDdHOiB4ie+QUsUrmsKVxI3zZhpDvToY7maRS2TfJe0wucGrGrpvqOx+YF82lFtZRWVxKG+LOBUUTA560+O3XVf6BRCPTK/uvs9KAJsLqGbAyg+NgmbgPvixM19jTaJ7mJstwsLdvvxtcYJ+uVjnxDiR2eEB4fBb5nD/4hIxjzwstYk3EPt2Z08iHA30N29XMbcsecwqZEJqYELLrOrwJOeBB3A6sYqQYv0jxm9GytC4TNB9u63RcJ/tQpzYYauVcRszppAjuE4F4oWGolALqEpcyIczHA+bhKUyAGvy1AXZr5psoC1Dzs6qJKiNmOdUb3YspP0CCHGasNU3hzF10Lja+RvtEUb4DEKnM4D9eAQ7Frky+f/LYYkmbEqs8RkaDQQ745R/9SZnbWfdyLj/vrdXnj7XXR7OzGucqHRZyq/U/3C/D9aADiReIf9Z0V4syTG4xlkpqgBdpOy6pauMQso448tYa+AyNKd3tttYGOKhDcJXHYQ== Juju:test1",
		"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCwNSnjr+G3PcwY6RmFSAtZ8D5U9P4jha50btd9gToviKauufgzd13MrvdXFhWAsBo12p8YFT/r9y213yZuyClEqi3Tq7iEJZqlq6aeYo5vRJXQxzc7TN2ON4TBYXbXg1rZb5qLz8rUknTGBU0JrzLAQTSmxWYPH+7snU9yi9kfgc1L3D1735NeOpwydt3itV/LaayV2qu9Xy7q5SyOo7qbhxhgzAMcKbAcMpcKZtdnLaSCb20HKxf7fzqJpOWws15Z4rc7PMD+xU19+PsuJAS9FV8K7rWxRmhA40AbrdIHPeI+bnCk0MLqQ1LAkPRmfmxX5EKQ6mwP/ykCnRnjyiFSE1g8ejEo/p7cfTi8g6kw8qVdJy1CgnhobZ23QqSBJGWC1OvVY3M8McDdztJ2vVLyiGFMKGvn8y9bYGayiH1fwwAl92YvqBLCWkV1iU0dDdvKWD+opYMOqR/Y+a3aZAmUZwO8M30IKA4WpalyE6lKfGa8RM4mgyMYy9k4s65Uw5PVa8Q1hOMWDlqjAcFKUZqeG210U1A8if21KVEt6SECZfljsLgIOOnTjBY2mx6oeFPIeEnN2dXu/WWtSb0POXHDra0ksh/Yg3eqyutk6wBPfdwlo70ARq2EfmqFYFDlbtgALQkJBn10KlZoiH0505tayH8zsBojyCVaQsqNmWgygw== Juju:test2",
	}
	instanceId := "7746dcb5-07e2-4e59-afd9-b723f9b53355"

	// Emulate BmInstanceBackend.setUserData
	setUserData := func(cloudConfig *CloudConfig) string {
		cloudConfig.SetHostName("intfSpec.DnsName")
		cloudConfig.SetDefaultUserGroup("sdp", sshPublicKeys)

		networkLinkFilePath := fmt.Sprintf("/etc/systemd/network/10-eth0.link")
		networkLinkFileContent := fmt.Sprintf(`[Match]
MACAddress=aa:bb:cc:dd:ee:ff
[Link]
Name=eth0
MTUBytes=8000
`)

		networkFilePath := fmt.Sprintf("/etc/systemd/network/10-eth0.network")
		networkFileContent := fmt.Sprintf(`[Match]
MACAddress=aa:bb:cc:dd:ee:ff
[Link]
Name=eth0
MTUBytes=8000
[Network]
LinkLocalAddressing=no
Address=192.168.1.2
`)
		cloudConfig.AddWriteFile(networkFilePath, networkFileContent)
		cloudConfig.AddWriteFile(networkLinkFilePath, networkLinkFileContent)

		out := `{"NIC_NET_CONFIG":[{"NIC_MAC":"abc","NIC_IP":"def","SUBNET_MASK":"ghi","GATEWAY_MAC":"jkl"},{"NIC_MAC":"123","NIC_IP":"456","SUBNET_MASK":"789","GATEWAY_MAC":"012"}]}`
		cloudConfig.AddWriteFile("/etc/gaudinet.json", out)
		cloudConfig.SetDefaultPowerState()
		cloudConfig.AddRunCmd("systemctl enable rc-local")
		cloudConfig.AddRunCmd("systemctl start rc-local")
		cloudConfig.AddRunCmd("systemctl disable ondemand")
		rcLocalContent := `#!/bin/sh
echo -n "performance" | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
/opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --down
/opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --up
exit 0
`
		cloudConfig.AddWriteFileWithPermissions("/etc/rc.local", rcLocalContent, "0755")

		dataDisksScriptData := fmt.Sprintf(`#!/usr/bin/env bash
set -ex
declare -a wwns=(WWN)
count=0
for wwn in "${wwns[@]}"
do
echo "wwn: ${wwn}"
partitions_count=$(lsblk -io wwn,name  | grep $wwn | wc -l)
echo $partitions_count
# Skip the disk with partitions
if [[ $partitions_count -gt 1 ]]
then
	continue
else
	count=$((count+1))
	echo "getting disk identifier"
	disk_id=$(lsblk -io wwn,name  | grep ${wwn} | awk '{print $2}')
	echo "disk id: $disk_id"
	echo "partitioning disk /dev/${disk_id}"
	sudo parted --script /dev/${disk_id} mklabel gpt mkpart primary ext4 1MiB 100%%
	sleep 2
	echo "setting filesystem"
	sudo mkfs.ext4 /dev/${disk_id}p1
	sudo mkdir -p /scratch-${count}
	sudo mount /dev/${disk_id}p1 /scratch-${count}
	uuid=$(lsblk -io UUID /dev/${disk_id}p1 --noheadings)
	echo "UUID=${uuid}     /scratch-${count}   ext4    rw,user,auto    0    0" | sudo tee -a /etc/fstab
fi
done
sudo mount -a
exit 0
`)
		cloudConfig.AddWriteFileWithPermissions("/etc/configuredatadisks", dataDisksScriptData, "0755")
		cloudConfig.AddRunCmd("cd /etc && ./configuredatadisks && cd ..")
		cloudConfig.AddWriteFile("/etc/machine-id", instanceId)
		cloudConfig.AddRunCmd("mkdir -p /opt/weka/data/agent")
		cloudConfig.AddRunCmd("cp /etc/machine-id /opt/weka/data/agent/machine-identifier")
		setWekaMachineID := `#!/usr/bin/env bash
set -e

MACHINE_IDENTIFIER='/etc/machine-id'
WEKA_MACHINE_IDENTIFIER='/opt/weka/data/agent/machine-identifier'

while true
do
# check if weka machine identifier is present
if [ ! -f "${WEKA_MACHINE_IDENTIFIER}" ]; then
	mkdir -p '/opt/weka/data/agent'
	cp ${MACHINE_IDENTIFIER} ${WEKA_MACHINE_IDENTIFIER}
fi
if  ! cmp -s "${MACHINE_IDENTIFIER}" "${WEKA_MACHINE_IDENTIFIER}"; then
	rm -f ${WEKA_MACHINE_IDENTIFIER}
	mkdir -p '/opt/weka/data/agent'
	cp ${MACHINE_IDENTIFIER} ${WEKA_MACHINE_IDENTIFIER}
fi
sleep 10
done
exit 0
`
		cloudConfig.AddWriteFileWithPermissions("/etc/set_weka_id", setWekaMachineID, "0700")
		wekaIdSystemdService := `[Unit]
Description=Weka machine id check
StartLimitIntervalSec=90
StartLimitBurst=10

[Service]
WorkingDirectory=/etc
ExecStart=/bin/bash set_weka_id
Restart=always
RestartSec=10s

[Install]
WantedBy=multi-user.target
`
		cloudConfig.AddWriteFileWithPermissions("/etc/systemd/system/weka-machine-id.service", wekaIdSystemdService, "0600")
		cloudConfig.AddRunCmd("systemctl daemon-reload")
		cloudConfig.AddRunCmd("systemctl  enable weka-machine-id")
		cloudConfig.AddRunCmd("systemctl  start weka-machine-id")

		cloudConfig.SetWriteFile()
		cloudConfig.SetRunCmd()
		cloudConfig.SetPackages()
		data, err := cloudConfig.RenderYAML()
		Expect(err).Should(Succeed())
		return string(data)
	}

	It("BmInstanceBackend.setUserData", func() {
		expected := `#cloud-config
hostname: intfSpec.DnsName
package_update: false
power_state:
  delay: now
  mode: reboot
runcmd:
- systemctl enable rc-local
- systemctl start rc-local
- systemctl disable ondemand
- cd /etc && ./configuredatadisks && cd ..
- mkdir -p /opt/weka/data/agent
- cp /etc/machine-id /opt/weka/data/agent/machine-identifier
- systemctl daemon-reload
- systemctl  enable weka-machine-id
- systemctl  start weka-machine-id
users:
- groups:
  - docker
  - render
  lock_passwd: true
  name: sdp
  shell: /bin/bash
  ssh_authorized_keys:
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCcO4pgGXXz/IDOOk0QcK4n45bkwfhr8TgCLLN1e2Qm5Zpda6egvpeI+ZrYpYNnEvCeIrZHFjwCL0JvkqY2xvf0EF5BiOa1dWc/eDj9csgW0xihQUETcUbgEQDCy8Ph3t7DGqw+h5yh6CwT2oIe9jcNnQmd+097X8aYvxk3zVx8/E7QqBmUDUH23U1VDdHOiB4ie+QUsUrmsKVxI3zZhpDvToY7maRS2TfJe0wucGrGrpvqOx+YF82lFtZRWVxKG+LOBUUTA560+O3XVf6BRCPTK/uvs9KAJsLqGbAyg+NgmbgPvixM19jTaJ7mJstwsLdvvxtcYJ+uVjnxDiR2eEB4fBb5nD/4hIxjzwstYk3EPt2Z08iHA30N29XMbcsecwqZEJqYELLrOrwJOeBB3A6sYqQYv0jxm9GytC4TNB9u63RcJ/tQpzYYauVcRszppAjuE4F4oWGolALqEpcyIczHA+bhKUyAGvy1AXZr5psoC1Dzs6qJKiNmOdUb3YspP0CCHGasNU3hzF10Lja+RvtEUb4DEKnM4D9eAQ7Frky+f/LYYkmbEqs8RkaDQQ745R/9SZnbWfdyLj/vrdXnj7XXR7OzGucqHRZyq/U/3C/D9aADiReIf9Z0V4syTG4xlkpqgBdpOy6pauMQso448tYa+AyNKd3tttYGOKhDcJXHYQ==
    Juju:test1
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCwNSnjr+G3PcwY6RmFSAtZ8D5U9P4jha50btd9gToviKauufgzd13MrvdXFhWAsBo12p8YFT/r9y213yZuyClEqi3Tq7iEJZqlq6aeYo5vRJXQxzc7TN2ON4TBYXbXg1rZb5qLz8rUknTGBU0JrzLAQTSmxWYPH+7snU9yi9kfgc1L3D1735NeOpwydt3itV/LaayV2qu9Xy7q5SyOo7qbhxhgzAMcKbAcMpcKZtdnLaSCb20HKxf7fzqJpOWws15Z4rc7PMD+xU19+PsuJAS9FV8K7rWxRmhA40AbrdIHPeI+bnCk0MLqQ1LAkPRmfmxX5EKQ6mwP/ykCnRnjyiFSE1g8ejEo/p7cfTi8g6kw8qVdJy1CgnhobZ23QqSBJGWC1OvVY3M8McDdztJ2vVLyiGFMKGvn8y9bYGayiH1fwwAl92YvqBLCWkV1iU0dDdvKWD+opYMOqR/Y+a3aZAmUZwO8M30IKA4WpalyE6lKfGa8RM4mgyMYy9k4s65Uw5PVa8Q1hOMWDlqjAcFKUZqeG210U1A8if21KVEt6SECZfljsLgIOOnTjBY2mx6oeFPIeEnN2dXu/WWtSb0POXHDra0ksh/Yg3eqyutk6wBPfdwlo70ARq2EfmqFYFDlbtgALQkJBn10KlZoiH0505tayH8zsBojyCVaQsqNmWgygw==
    Juju:test2
  sudo: ALL=(ALL) NOPASSWD:ALL
write_files:
- content: |
    [Match]
    MACAddress=aa:bb:cc:dd:ee:ff
    [Link]
    Name=eth0
    MTUBytes=8000
    [Network]
    LinkLocalAddressing=no
    Address=192.168.1.2
  path: /etc/systemd/network/10-eth0.network
- content: |
    [Match]
    MACAddress=aa:bb:cc:dd:ee:ff
    [Link]
    Name=eth0
    MTUBytes=8000
  path: /etc/systemd/network/10-eth0.link
- content: '{"NIC_NET_CONFIG":[{"NIC_MAC":"abc","NIC_IP":"def","SUBNET_MASK":"ghi","GATEWAY_MAC":"jkl"},{"NIC_MAC":"123","NIC_IP":"456","SUBNET_MASK":"789","GATEWAY_MAC":"012"}]}'
  path: /etc/gaudinet.json
- content: |
    #!/bin/sh
    echo -n "performance" | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
    /opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --down
    /opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --up
    exit 0
  path: /etc/rc.local
  permissions: "0755"
- content: "#!/usr/bin/env bash\nset -ex\ndeclare -a wwns=(WWN)\ncount=0\nfor wwn
    in \"${wwns[@]}\"\ndo\necho \"wwn: ${wwn}\"\npartitions_count=$(lsblk -io wwn,name
    \ | grep $wwn | wc -l)\necho $partitions_count\n# Skip the disk with partitions\nif
    [[ $partitions_count -gt 1 ]]\nthen\n\tcontinue\nelse\n\tcount=$((count+1))\n\techo
    \"getting disk identifier\"\n\tdisk_id=$(lsblk -io wwn,name  | grep ${wwn} | awk
    '{print $2}')\n\techo \"disk id: $disk_id\"\n\techo \"partitioning disk /dev/${disk_id}\"\n\tsudo
    parted --script /dev/${disk_id} mklabel gpt mkpart primary ext4 1MiB 100%\n\tsleep
    2\n\techo \"setting filesystem\"\n\tsudo mkfs.ext4 /dev/${disk_id}p1\n\tsudo mkdir
    -p /scratch-${count}\n\tsudo mount /dev/${disk_id}p1 /scratch-${count}\n\tuuid=$(lsblk
    -io UUID /dev/${disk_id}p1 --noheadings)\n\techo \"UUID=${uuid}     /scratch-${count}
    \  ext4    rw,user,auto    0    0\" | sudo tee -a /etc/fstab\nfi\ndone\nsudo mount
    -a\nexit 0\n"
  path: /etc/configuredatadisks
  permissions: "0755"
- content: 7746dcb5-07e2-4e59-afd9-b723f9b53355
  path: /etc/machine-id
- content: "#!/usr/bin/env bash\nset -e\n\nMACHINE_IDENTIFIER='/etc/machine-id'\nWEKA_MACHINE_IDENTIFIER='/opt/weka/data/agent/machine-identifier'\n\nwhile
    true\ndo\n# check if weka machine identifier is present\nif [ ! -f \"${WEKA_MACHINE_IDENTIFIER}\"
    ]; then\n\tmkdir -p '/opt/weka/data/agent'\n\tcp ${MACHINE_IDENTIFIER} ${WEKA_MACHINE_IDENTIFIER}\nfi\nif
    \ ! cmp -s \"${MACHINE_IDENTIFIER}\" \"${WEKA_MACHINE_IDENTIFIER}\"; then\n\trm
    -f ${WEKA_MACHINE_IDENTIFIER}\n\tmkdir -p '/opt/weka/data/agent'\n\tcp ${MACHINE_IDENTIFIER}
    ${WEKA_MACHINE_IDENTIFIER}\nfi\nsleep 10\ndone\nexit 0\n"
  path: /etc/set_weka_id
  permissions: "0700"
- content: |
    [Unit]
    Description=Weka machine id check
    StartLimitIntervalSec=90
    StartLimitBurst=10

    [Service]
    WorkingDirectory=/etc
    ExecStart=/bin/bash set_weka_id
    Restart=always
    RestartSec=10s

    [Install]
    WantedBy=multi-user.target
  path: /etc/systemd/system/weka-machine-id.service
  permissions: "0600"
`
		cloudConfigYAML, err := NewCloudConfig("ubuntu", "")
		Expect(err).Should(Succeed())
		yaml := setUserData(cloudConfigYAML)
		Expect(expected).To(Equal(yaml))
	})

	// Emulate VmInstanceBackend.CreateVmSecrets
	createVmSecrets := func(cloudConfig *CloudConfig) string {
		cloudConfig.SetManageEtcHosts("localhost")
		cloudConfig.SetHostName("hostName")
		cloudConfig.SetFqdn("intfStatus.DnsName")
		cloudConfig.SetSshPublicKeys(sshPublicKeys)
		cloudConfig.SetWriteFile()
		cloudConfig.SetRunCmd()
		cloudConfig.SetPackages()

		data, err := cloudConfig.RenderYAML()
		Expect(err).Should(Succeed())
		return string(data)
	}

	It("VmInstanceBackend.CreateVmSecrets", func() {
		expected := `#cloud-config
fqdn: intfStatus.DnsName
hostname: hostName
manage_etc_hosts: localhost
package_update: false
ssh_authorized_keys:
- ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCcO4pgGXXz/IDOOk0QcK4n45bkwfhr8TgCLLN1e2Qm5Zpda6egvpeI+ZrYpYNnEvCeIrZHFjwCL0JvkqY2xvf0EF5BiOa1dWc/eDj9csgW0xihQUETcUbgEQDCy8Ph3t7DGqw+h5yh6CwT2oIe9jcNnQmd+097X8aYvxk3zVx8/E7QqBmUDUH23U1VDdHOiB4ie+QUsUrmsKVxI3zZhpDvToY7maRS2TfJe0wucGrGrpvqOx+YF82lFtZRWVxKG+LOBUUTA560+O3XVf6BRCPTK/uvs9KAJsLqGbAyg+NgmbgPvixM19jTaJ7mJstwsLdvvxtcYJ+uVjnxDiR2eEB4fBb5nD/4hIxjzwstYk3EPt2Z08iHA30N29XMbcsecwqZEJqYELLrOrwJOeBB3A6sYqQYv0jxm9GytC4TNB9u63RcJ/tQpzYYauVcRszppAjuE4F4oWGolALqEpcyIczHA+bhKUyAGvy1AXZr5psoC1Dzs6qJKiNmOdUb3YspP0CCHGasNU3hzF10Lja+RvtEUb4DEKnM4D9eAQ7Frky+f/LYYkmbEqs8RkaDQQ745R/9SZnbWfdyLj/vrdXnj7XXR7OzGucqHRZyq/U/3C/D9aADiReIf9Z0V4syTG4xlkpqgBdpOy6pauMQso448tYa+AyNKd3tttYGOKhDcJXHYQ==
  Juju:test1
- ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCwNSnjr+G3PcwY6RmFSAtZ8D5U9P4jha50btd9gToviKauufgzd13MrvdXFhWAsBo12p8YFT/r9y213yZuyClEqi3Tq7iEJZqlq6aeYo5vRJXQxzc7TN2ON4TBYXbXg1rZb5qLz8rUknTGBU0JrzLAQTSmxWYPH+7snU9yi9kfgc1L3D1735NeOpwydt3itV/LaayV2qu9Xy7q5SyOo7qbhxhgzAMcKbAcMpcKZtdnLaSCb20HKxf7fzqJpOWws15Z4rc7PMD+xU19+PsuJAS9FV8K7rWxRmhA40AbrdIHPeI+bnCk0MLqQ1LAkPRmfmxX5EKQ6mwP/ykCnRnjyiFSE1g8ejEo/p7cfTi8g6kw8qVdJy1CgnhobZ23QqSBJGWC1OvVY3M8McDdztJ2vVLyiGFMKGvn8y9bYGayiH1fwwAl92YvqBLCWkV1iU0dDdvKWD+opYMOqR/Y+a3aZAmUZwO8M30IKA4WpalyE6lKfGa8RM4mgyMYy9k4s65Uw5PVa8Q1hOMWDlqjAcFKUZqeG210U1A8if21KVEt6SECZfljsLgIOOnTjBY2mx6oeFPIeEnN2dXu/WWtSb0POXHDra0ksh/Yg3eqyutk6wBPfdwlo70ARq2EfmqFYFDlbtgALQkJBn10KlZoiH0505tayH8zsBojyCVaQsqNmWgygw==
  Juju:test2
`
		cloudConfig, err := NewCloudConfig("ubuntu", "")
		Expect(err).Should(Succeed())
		yaml := createVmSecrets(cloudConfig)
		Expect(expected).To(Equal(yaml))
	})

	It("instance.Spec.UserData", func() {
		expected := `#cloud-config
package_update: false
packages:
- qemu-guest-agent
- socat
- conntrack
- ipset
runcmd:
- ls -l /
- ls -l /root
- ls -l /tmp
ssh_authorized_keys:
- ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCcO4pgGXXz/IDOOk0QcK4n45bkwfhr8TgCLLN1e2Qm5Zpda6egvpeI+ZrYpYNnEvCeIrZHFjwCL0JvkqY2xvf0EF5BiOa1dWc/eDj9csgW0xihQUETcUbgEQDCy8Ph3t7DGqw+h5yh6CwT2oIe9jcNnQmd+097X8aYvxk3zVx8/E7QqBmUDUH23U1VDdHOiB4ie+QUsUrmsKVxI3zZhpDvToY7maRS2TfJe0wucGrGrpvqOx+YF82lFtZRWVxKG+LOBUUTA560+O3XVf6BRCPTK/uvs9KAJsLqGbAyg+NgmbgPvixM19jTaJ7mJstwsLdvvxtcYJ+uVjnxDiR2eEB4fBb5nD/4hIxjzwstYk3EPt2Z08iHA30N29XMbcsecwqZEJqYELLrOrwJOeBB3A6sYqQYv0jxm9GytC4TNB9u63RcJ/tQpzYYauVcRszppAjuE4F4oWGolALqEpcyIczHA+bhKUyAGvy1AXZr5psoC1Dzs6qJKiNmOdUb3YspP0CCHGasNU3hzF10Lja+RvtEUb4DEKnM4D9eAQ7Frky+f/LYYkmbEqs8RkaDQQ745R/9SZnbWfdyLj/vrdXnj7XXR7OzGucqHRZyq/U/3C/D9aADiReIf9Z0V4syTG4xlkpqgBdpOy6pauMQso448tYa+AyNKd3tttYGOKhDcJXHYQ==
  Juju:test1
users:
- groups:
  - docker
  - render
  lock_passwd: true
  name: test
  shell: /bin/bash
  ssh_authorized_keys:
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCwNSnjr+G3PcwY6RmFSAtZ8D5U9P4jha50btd9gToviKauufgzd13MrvdXFhWAsBo12p8YFT/r9y213yZuyClEqi3Tq7iEJZqlq6aeYo5vRJXQxzc7TN2ON4TBYXbXg1rZb5qLz8rUknTGBU0JrzLAQTSmxWYPH+7snU9yi9kfgc1L3D1735NeOpwydt3itV/LaayV2qu9Xy7q5SyOo7qbhxhgzAMcKbAcMpcKZtdnLaSCb20HKxf7fzqJpOWws15Z4rc7PMD+xU19+PsuJAS9FV8K7rWxRmhA40AbrdIHPeI+bnCk0MLqQ1LAkPRmfmxX5EKQ6mwP/ykCnRnjyiFSE1g8ejEo/p7cfTi8g6kw8qVdJy1CgnhobZ23QqSBJGWC1OvVY3M8McDdztJ2vVLyiGFMKGvn8y9bYGayiH1fwwAl92YvqBLCWkV1iU0dDdvKWD+opYMOqR/Y+a3aZAmUZwO8M30IKA4WpalyE6lKfGa8RM4mgyMYy9k4s65Uw5PVa8Q1hOMWDlqjAcFKUZqeG210U1A8if21KVEt6SECZfljsLgIOOnTjBY2mx6oeFPIeEnN2dXu/WWtSb0POXHDra0ksh/Yg3eqyutk6wBPfdwlo70ARq2EfmqFYFDlbtgALQkJBn10KlZoiH0505tayH8zsBojyCVaQsqNmWgygw==
    Juju:test2
  sudo: ALL=(ALL) NOPASSWD:ALL
write_files:
- content: |
    HTTP_PROXY=http://internal-placeholder.com:912/
    HTTPS_PROXY=http://internal-placeholder.com:912/
    NO_PROXY=127.0.0.1,127.0.1.1,localhost,.intel.com
  path: /etc/environment
  append: true
- content: |
    #!/usr/bin/bash
    echo 'helloworld'
  path: /etc/helloworld
  append: true
  permissions: "0777"
- content: |-
    #!/usr/bin/bash
    echo 'helloworld3'
  path: /etc/helloworld3
  append: true
  permissions: "0700"
`
		userData := "#cloud-init\npackages:\n  - qemu-guest-agent\n  - socat\n  - conntrack\n  - ipset\nssh_authorized_keys:\n  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCcO4pgGXXz/IDOOk0QcK4n45bkwfhr8TgCLLN1e2Qm5Zpda6egvpeI+ZrYpYNnEvCeIrZHFjwCL0JvkqY2xvf0EF5BiOa1dWc/eDj9csgW0xihQUETcUbgEQDCy8Ph3t7DGqw+h5yh6CwT2oIe9jcNnQmd+097X8aYvxk3zVx8/E7QqBmUDUH23U1VDdHOiB4ie+QUsUrmsKVxI3zZhpDvToY7maRS2TfJe0wucGrGrpvqOx+YF82lFtZRWVxKG+LOBUUTA560+O3XVf6BRCPTK/uvs9KAJsLqGbAyg+NgmbgPvixM19jTaJ7mJstwsLdvvxtcYJ+uVjnxDiR2eEB4fBb5nD/4hIxjzwstYk3EPt2Z08iHA30N29XMbcsecwqZEJqYELLrOrwJOeBB3A6sYqQYv0jxm9GytC4TNB9u63RcJ/tQpzYYauVcRszppAjuE4F4oWGolALqEpcyIczHA+bhKUyAGvy1AXZr5psoC1Dzs6qJKiNmOdUb3YspP0CCHGasNU3hzF10Lja+RvtEUb4DEKnM4D9eAQ7Frky+f/LYYkmbEqs8RkaDQQ745R/9SZnbWfdyLj/vrdXnj7XXR7OzGucqHRZyq/U/3C/D9aADiReIf9Z0V4syTG4xlkpqgBdpOy6pauMQso448tYa+AyNKd3tttYGOKhDcJXHYQ== Juju:test1\nusers:\n  - groups: docker,render\n    lock_passwd: true\n    name: test\n    shell: /bin/bash\n    ssh_authorized_keys:\n      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCwNSnjr+G3PcwY6RmFSAtZ8D5U9P4jha50btd9gToviKauufgzd13MrvdXFhWAsBo12p8YFT/r9y213yZuyClEqi3Tq7iEJZqlq6aeYo5vRJXQxzc7TN2ON4TBYXbXg1rZb5qLz8rUknTGBU0JrzLAQTSmxWYPH+7snU9yi9kfgc1L3D1735NeOpwydt3itV/LaayV2qu9Xy7q5SyOo7qbhxhgzAMcKbAcMpcKZtdnLaSCb20HKxf7fzqJpOWws15Z4rc7PMD+xU19+PsuJAS9FV8K7rWxRmhA40AbrdIHPeI+bnCk0MLqQ1LAkPRmfmxX5EKQ6mwP/ykCnRnjyiFSE1g8ejEo/p7cfTi8g6kw8qVdJy1CgnhobZ23QqSBJGWC1OvVY3M8McDdztJ2vVLyiGFMKGvn8y9bYGayiH1fwwAl92YvqBLCWkV1iU0dDdvKWD+opYMOqR/Y+a3aZAmUZwO8M30IKA4WpalyE6lKfGa8RM4mgyMYy9k4s65Uw5PVa8Q1hOMWDlqjAcFKUZqeG210U1A8if21KVEt6SECZfljsLgIOOnTjBY2mx6oeFPIeEnN2dXu/WWtSb0POXHDra0ksh/Yg3eqyutk6wBPfdwlo70ARq2EfmqFYFDlbtgALQkJBn10KlZoiH0505tayH8zsBojyCVaQsqNmWgygw== Juju:test2\n    sudo: ALL=(ALL) NOPASSWD:ALL\nruncmd:\n  - ls -l /\n  - ls -l /root\n  - ls -l /tmp\nwrite_files:\n  - path: /etc/environment\n    append: true\n    content: |\n      HTTP_PROXY=http://internal-placeholder.com:912/\n      HTTPS_PROXY=http://internal-placeholder.com:912/\n      NO_PROXY=127.0.0.1,127.0.1.1,localhost,.intel.com\n  - path: /etc/helloworld\n    permissions: '0777'\n    content: |\n      #!/usr/bin/bash\n      echo 'helloworld'\n  - path: /etc/helloworld3\n    permissions: '0700'\n    content: |\n      #!/usr/bin/bash\n      echo 'helloworld3'"
		cloudConfig, err := NewCloudConfig("ubuntu", userData)
		Expect(err).Should(Succeed())
		cloudConfig.SetWriteFile()
		cloudConfig.SetRunCmd()
		cloudConfig.SetPackages()
		data, err := cloudConfig.RenderYAML()
		Expect(err).Should(Succeed())
		yaml := string(data)
		Expect(expected).To(Equal(yaml))
	})

	It("BmInstanceBackend.setUserData with instance.Spec.UserData", func() {
		expected := `#cloud-config
hostname: intfSpec.DnsName
package_update: false
packages:
- qemu-guest-agent
- socat
- conntrack
- ipset
power_state:
  delay: now
  mode: reboot
runcmd:
- ls -l /
- ls -l /root
- ls -l /tmp
- systemctl enable rc-local
- systemctl start rc-local
- systemctl disable ondemand
- cd /etc && ./configuredatadisks && cd ..
- mkdir -p /opt/weka/data/agent
- cp /etc/machine-id /opt/weka/data/agent/machine-identifier
- systemctl daemon-reload
- systemctl  enable weka-machine-id
- systemctl  start weka-machine-id
ssh_authorized_keys:
- ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCcO4pgGXXz/IDOOk0QcK4n45bkwfhr8TgCLLN1e2Qm5Zpda6egvpeI+ZrYpYNnEvCeIrZHFjwCL0JvkqY2xvf0EF5BiOa1dWc/eDj9csgW0xihQUETcUbgEQDCy8Ph3t7DGqw+h5yh6CwT2oIe9jcNnQmd+097X8aYvxk3zVx8/E7QqBmUDUH23U1VDdHOiB4ie+QUsUrmsKVxI3zZhpDvToY7maRS2TfJe0wucGrGrpvqOx+YF82lFtZRWVxKG+LOBUUTA560+O3XVf6BRCPTK/uvs9KAJsLqGbAyg+NgmbgPvixM19jTaJ7mJstwsLdvvxtcYJ+uVjnxDiR2eEB4fBb5nD/4hIxjzwstYk3EPt2Z08iHA30N29XMbcsecwqZEJqYELLrOrwJOeBB3A6sYqQYv0jxm9GytC4TNB9u63RcJ/tQpzYYauVcRszppAjuE4F4oWGolALqEpcyIczHA+bhKUyAGvy1AXZr5psoC1Dzs6qJKiNmOdUb3YspP0CCHGasNU3hzF10Lja+RvtEUb4DEKnM4D9eAQ7Frky+f/LYYkmbEqs8RkaDQQ745R/9SZnbWfdyLj/vrdXnj7XXR7OzGucqHRZyq/U/3C/D9aADiReIf9Z0V4syTG4xlkpqgBdpOy6pauMQso448tYa+AyNKd3tttYGOKhDcJXHYQ==
  Juju:test1
users:
- groups:
  - docker
  - render
  lock_passwd: true
  name: test
  shell: /bin/bash
  ssh_authorized_keys:
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCwNSnjr+G3PcwY6RmFSAtZ8D5U9P4jha50btd9gToviKauufgzd13MrvdXFhWAsBo12p8YFT/r9y213yZuyClEqi3Tq7iEJZqlq6aeYo5vRJXQxzc7TN2ON4TBYXbXg1rZb5qLz8rUknTGBU0JrzLAQTSmxWYPH+7snU9yi9kfgc1L3D1735NeOpwydt3itV/LaayV2qu9Xy7q5SyOo7qbhxhgzAMcKbAcMpcKZtdnLaSCb20HKxf7fzqJpOWws15Z4rc7PMD+xU19+PsuJAS9FV8K7rWxRmhA40AbrdIHPeI+bnCk0MLqQ1LAkPRmfmxX5EKQ6mwP/ykCnRnjyiFSE1g8ejEo/p7cfTi8g6kw8qVdJy1CgnhobZ23QqSBJGWC1OvVY3M8McDdztJ2vVLyiGFMKGvn8y9bYGayiH1fwwAl92YvqBLCWkV1iU0dDdvKWD+opYMOqR/Y+a3aZAmUZwO8M30IKA4WpalyE6lKfGa8RM4mgyMYy9k4s65Uw5PVa8Q1hOMWDlqjAcFKUZqeG210U1A8if21KVEt6SECZfljsLgIOOnTjBY2mx6oeFPIeEnN2dXu/WWtSb0POXHDra0ksh/Yg3eqyutk6wBPfdwlo70ARq2EfmqFYFDlbtgALQkJBn10KlZoiH0505tayH8zsBojyCVaQsqNmWgygw==
    Juju:test2
  sudo: ALL=(ALL) NOPASSWD:ALL
- groups:
  - docker
  - render
  lock_passwd: true
  name: sdp
  shell: /bin/bash
  ssh_authorized_keys:
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCcO4pgGXXz/IDOOk0QcK4n45bkwfhr8TgCLLN1e2Qm5Zpda6egvpeI+ZrYpYNnEvCeIrZHFjwCL0JvkqY2xvf0EF5BiOa1dWc/eDj9csgW0xihQUETcUbgEQDCy8Ph3t7DGqw+h5yh6CwT2oIe9jcNnQmd+097X8aYvxk3zVx8/E7QqBmUDUH23U1VDdHOiB4ie+QUsUrmsKVxI3zZhpDvToY7maRS2TfJe0wucGrGrpvqOx+YF82lFtZRWVxKG+LOBUUTA560+O3XVf6BRCPTK/uvs9KAJsLqGbAyg+NgmbgPvixM19jTaJ7mJstwsLdvvxtcYJ+uVjnxDiR2eEB4fBb5nD/4hIxjzwstYk3EPt2Z08iHA30N29XMbcsecwqZEJqYELLrOrwJOeBB3A6sYqQYv0jxm9GytC4TNB9u63RcJ/tQpzYYauVcRszppAjuE4F4oWGolALqEpcyIczHA+bhKUyAGvy1AXZr5psoC1Dzs6qJKiNmOdUb3YspP0CCHGasNU3hzF10Lja+RvtEUb4DEKnM4D9eAQ7Frky+f/LYYkmbEqs8RkaDQQ745R/9SZnbWfdyLj/vrdXnj7XXR7OzGucqHRZyq/U/3C/D9aADiReIf9Z0V4syTG4xlkpqgBdpOy6pauMQso448tYa+AyNKd3tttYGOKhDcJXHYQ==
    Juju:test1
  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCwNSnjr+G3PcwY6RmFSAtZ8D5U9P4jha50btd9gToviKauufgzd13MrvdXFhWAsBo12p8YFT/r9y213yZuyClEqi3Tq7iEJZqlq6aeYo5vRJXQxzc7TN2ON4TBYXbXg1rZb5qLz8rUknTGBU0JrzLAQTSmxWYPH+7snU9yi9kfgc1L3D1735NeOpwydt3itV/LaayV2qu9Xy7q5SyOo7qbhxhgzAMcKbAcMpcKZtdnLaSCb20HKxf7fzqJpOWws15Z4rc7PMD+xU19+PsuJAS9FV8K7rWxRmhA40AbrdIHPeI+bnCk0MLqQ1LAkPRmfmxX5EKQ6mwP/ykCnRnjyiFSE1g8ejEo/p7cfTi8g6kw8qVdJy1CgnhobZ23QqSBJGWC1OvVY3M8McDdztJ2vVLyiGFMKGvn8y9bYGayiH1fwwAl92YvqBLCWkV1iU0dDdvKWD+opYMOqR/Y+a3aZAmUZwO8M30IKA4WpalyE6lKfGa8RM4mgyMYy9k4s65Uw5PVa8Q1hOMWDlqjAcFKUZqeG210U1A8if21KVEt6SECZfljsLgIOOnTjBY2mx6oeFPIeEnN2dXu/WWtSb0POXHDra0ksh/Yg3eqyutk6wBPfdwlo70ARq2EfmqFYFDlbtgALQkJBn10KlZoiH0505tayH8zsBojyCVaQsqNmWgygw==
    Juju:test2
  sudo: ALL=(ALL) NOPASSWD:ALL
write_files:
- content: |
    HTTP_PROXY=http://internal-placeholder.com:912/
    HTTPS_PROXY=http://internal-placeholder.com:912/
    NO_PROXY=127.0.0.1,127.0.1.1,localhost,.intel.com
  path: /etc/environment
  append: true
- content: |
    #!/usr/bin/bash
    echo 'helloworld'
  path: /etc/helloworld
  append: true
  permissions: "0777"
- content: |-
    #!/usr/bin/bash
    echo 'helloworld3'
  path: /etc/helloworld3
  append: true
  permissions: "0700"
- content: |
    [Match]
    MACAddress=aa:bb:cc:dd:ee:ff
    [Link]
    Name=eth0
    MTUBytes=8000
    [Network]
    LinkLocalAddressing=no
    Address=192.168.1.2
  path: /etc/systemd/network/10-eth0.network
- content: |
    [Match]
    MACAddress=aa:bb:cc:dd:ee:ff
    [Link]
    Name=eth0
    MTUBytes=8000
  path: /etc/systemd/network/10-eth0.link
- content: '{"NIC_NET_CONFIG":[{"NIC_MAC":"abc","NIC_IP":"def","SUBNET_MASK":"ghi","GATEWAY_MAC":"jkl"},{"NIC_MAC":"123","NIC_IP":"456","SUBNET_MASK":"789","GATEWAY_MAC":"012"}]}'
  path: /etc/gaudinet.json
- content: |
    #!/bin/sh
    echo -n "performance" | tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
    /opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --down
    /opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --up
    exit 0
  path: /etc/rc.local
  permissions: "0755"
- content: "#!/usr/bin/env bash\nset -ex\ndeclare -a wwns=(WWN)\ncount=0\nfor wwn
    in \"${wwns[@]}\"\ndo\necho \"wwn: ${wwn}\"\npartitions_count=$(lsblk -io wwn,name
    \ | grep $wwn | wc -l)\necho $partitions_count\n# Skip the disk with partitions\nif
    [[ $partitions_count -gt 1 ]]\nthen\n\tcontinue\nelse\n\tcount=$((count+1))\n\techo
    \"getting disk identifier\"\n\tdisk_id=$(lsblk -io wwn,name  | grep ${wwn} | awk
    '{print $2}')\n\techo \"disk id: $disk_id\"\n\techo \"partitioning disk /dev/${disk_id}\"\n\tsudo
    parted --script /dev/${disk_id} mklabel gpt mkpart primary ext4 1MiB 100%\n\tsleep
    2\n\techo \"setting filesystem\"\n\tsudo mkfs.ext4 /dev/${disk_id}p1\n\tsudo mkdir
    -p /scratch-${count}\n\tsudo mount /dev/${disk_id}p1 /scratch-${count}\n\tuuid=$(lsblk
    -io UUID /dev/${disk_id}p1 --noheadings)\n\techo \"UUID=${uuid}     /scratch-${count}
    \  ext4    rw,user,auto    0    0\" | sudo tee -a /etc/fstab\nfi\ndone\nsudo mount
    -a\nexit 0\n"
  path: /etc/configuredatadisks
  permissions: "0755"
- content: 7746dcb5-07e2-4e59-afd9-b723f9b53355
  path: /etc/machine-id
- content: "#!/usr/bin/env bash\nset -e\n\nMACHINE_IDENTIFIER='/etc/machine-id'\nWEKA_MACHINE_IDENTIFIER='/opt/weka/data/agent/machine-identifier'\n\nwhile
    true\ndo\n# check if weka machine identifier is present\nif [ ! -f \"${WEKA_MACHINE_IDENTIFIER}\"
    ]; then\n\tmkdir -p '/opt/weka/data/agent'\n\tcp ${MACHINE_IDENTIFIER} ${WEKA_MACHINE_IDENTIFIER}\nfi\nif
    \ ! cmp -s \"${MACHINE_IDENTIFIER}\" \"${WEKA_MACHINE_IDENTIFIER}\"; then\n\trm
    -f ${WEKA_MACHINE_IDENTIFIER}\n\tmkdir -p '/opt/weka/data/agent'\n\tcp ${MACHINE_IDENTIFIER}
    ${WEKA_MACHINE_IDENTIFIER}\nfi\nsleep 10\ndone\nexit 0\n"
  path: /etc/set_weka_id
  permissions: "0700"
- content: |
    [Unit]
    Description=Weka machine id check
    StartLimitIntervalSec=90
    StartLimitBurst=10

    [Service]
    WorkingDirectory=/etc
    ExecStart=/bin/bash set_weka_id
    Restart=always
    RestartSec=10s

    [Install]
    WantedBy=multi-user.target
  path: /etc/systemd/system/weka-machine-id.service
  permissions: "0600"
`
		userData := "#cloud-init\npackages:\n  - qemu-guest-agent\n  - socat\n  - conntrack\n  - ipset\nssh_authorized_keys:\n  - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCcO4pgGXXz/IDOOk0QcK4n45bkwfhr8TgCLLN1e2Qm5Zpda6egvpeI+ZrYpYNnEvCeIrZHFjwCL0JvkqY2xvf0EF5BiOa1dWc/eDj9csgW0xihQUETcUbgEQDCy8Ph3t7DGqw+h5yh6CwT2oIe9jcNnQmd+097X8aYvxk3zVx8/E7QqBmUDUH23U1VDdHOiB4ie+QUsUrmsKVxI3zZhpDvToY7maRS2TfJe0wucGrGrpvqOx+YF82lFtZRWVxKG+LOBUUTA560+O3XVf6BRCPTK/uvs9KAJsLqGbAyg+NgmbgPvixM19jTaJ7mJstwsLdvvxtcYJ+uVjnxDiR2eEB4fBb5nD/4hIxjzwstYk3EPt2Z08iHA30N29XMbcsecwqZEJqYELLrOrwJOeBB3A6sYqQYv0jxm9GytC4TNB9u63RcJ/tQpzYYauVcRszppAjuE4F4oWGolALqEpcyIczHA+bhKUyAGvy1AXZr5psoC1Dzs6qJKiNmOdUb3YspP0CCHGasNU3hzF10Lja+RvtEUb4DEKnM4D9eAQ7Frky+f/LYYkmbEqs8RkaDQQ745R/9SZnbWfdyLj/vrdXnj7XXR7OzGucqHRZyq/U/3C/D9aADiReIf9Z0V4syTG4xlkpqgBdpOy6pauMQso448tYa+AyNKd3tttYGOKhDcJXHYQ== Juju:test1\nusers:\n  - groups: docker,render\n    lock_passwd: true\n    name: test\n    shell: /bin/bash\n    ssh_authorized_keys:\n      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCwNSnjr+G3PcwY6RmFSAtZ8D5U9P4jha50btd9gToviKauufgzd13MrvdXFhWAsBo12p8YFT/r9y213yZuyClEqi3Tq7iEJZqlq6aeYo5vRJXQxzc7TN2ON4TBYXbXg1rZb5qLz8rUknTGBU0JrzLAQTSmxWYPH+7snU9yi9kfgc1L3D1735NeOpwydt3itV/LaayV2qu9Xy7q5SyOo7qbhxhgzAMcKbAcMpcKZtdnLaSCb20HKxf7fzqJpOWws15Z4rc7PMD+xU19+PsuJAS9FV8K7rWxRmhA40AbrdIHPeI+bnCk0MLqQ1LAkPRmfmxX5EKQ6mwP/ykCnRnjyiFSE1g8ejEo/p7cfTi8g6kw8qVdJy1CgnhobZ23QqSBJGWC1OvVY3M8McDdztJ2vVLyiGFMKGvn8y9bYGayiH1fwwAl92YvqBLCWkV1iU0dDdvKWD+opYMOqR/Y+a3aZAmUZwO8M30IKA4WpalyE6lKfGa8RM4mgyMYy9k4s65Uw5PVa8Q1hOMWDlqjAcFKUZqeG210U1A8if21KVEt6SECZfljsLgIOOnTjBY2mx6oeFPIeEnN2dXu/WWtSb0POXHDra0ksh/Yg3eqyutk6wBPfdwlo70ARq2EfmqFYFDlbtgALQkJBn10KlZoiH0505tayH8zsBojyCVaQsqNmWgygw== Juju:test2\n    sudo: ALL=(ALL) NOPASSWD:ALL\nruncmd:\n  - ls -l /\n  - ls -l /root\n  - ls -l /tmp\nwrite_files:\n  - path: /etc/environment\n    append: true\n    content: |\n      HTTP_PROXY=http://internal-placeholder.com:912/\n      HTTPS_PROXY=http://internal-placeholder.com:912/\n      NO_PROXY=127.0.0.1,127.0.1.1,localhost,.intel.com\n  - path: /etc/helloworld\n    permissions: '0777'\n    content: |\n      #!/usr/bin/bash\n      echo 'helloworld'\n  - path: /etc/helloworld3\n    permissions: '0700'\n    content: |\n      #!/usr/bin/bash\n      echo 'helloworld3'"
		cloudConfig, err := NewCloudConfig("ubuntu", userData)
		Expect(err).Should(Succeed())
		yaml := setUserData(cloudConfig)
		Expect(expected).To(Equal(yaml))
	})
})
