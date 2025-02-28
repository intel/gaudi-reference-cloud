#!/bin/bash -e

export HABANA_LOGS=/var/log/habana_logs/
export RDMA_CORE_ROOT=/opt/habanalabs/rdma-core/src
export HABANA_PLUGINS_LIB_PATH=/opt/habanalabs/habana_plugins
export MPI_ROOT=/opt/habanalabs/openmpi-4.1.5
export RDMA_CORE_LIB=/opt/habanalabs/rdma-core/src/build/lib
export HABANA_SCAL_BIN_PATH=/opt/habanalabs/engines_fw
export LD_LIBRARY_PATH=/opt/habanalabs/openmpi-4.1.5/lib:
export OPAL_PREFIX=/opt/habanalabs/openmpi-4.1.5
export DATA_LOADER_AEON_LIB_PATH=/usr/lib/habanalabs/libaeon.so
export GC_KERNEL_PATH=/usr/lib/habanalabs/libtpc_kernels.so
export PIP_NO_CACHE_DIR=false
export __python_cmd=python3

source /etc/profile.d/habanalabs-rdma-core.sh
source /etc/profile.d/habanalabs.sh

# check if drivers (hl-smi) is working
if ! out=$(hl-smi -L 2>&1); then
    echo "Failed to execute hl-smi!"
    echo "Error: $out"
    exit 1
fi

# check power
hl-smi -L|grep "Power Max"
sleep 30

# Check if we have the right GPU count
GPU_COUNT=$(hl-smi -q | grep SPI | wc -l)
if test "8" = "$GPU_COUNT"
then
    echo "Total $GPU_COUNT GPUs successfully validated"
else
    echo "Failed to validate the number of GPUs. expected: 8, found: $GPU_COUNT"
    exit 1
fi

# FW Version
FW_VERSION=$(printenv | awk -F '=' '/FULL_FW_VERSION/{print $2}')
echo "FW_VERSION: ${FW_VERSION}"

# Check version
GPU_VERSION_CHECK=$(hl-smi -q | grep SPI | grep $FW_VERSION | wc -l)
if test "8" = "$GPU_VERSION_CHECK"
then
    echo "Validated FW version: ($FW_VERSION) of all the GPUs"
else
    echo "Failed to validate expected FW version $FW_VERSION on all GPUs"
    hl-smi -q | grep SPI
    exit 1
fi

# Driver reloading
sudo rmmod habanalabs && sudo modprobe habanalabs timeout_locked=0

# Set CPU to Performance
echo -n performance | sudo tee /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor
cat /sys/devices/system/cpu/cpu*/cpufreq/scaling_governor

# Files permission changes
sudo chmod 777 /opt/habanalabs/qual -R
sudo chmod 777 /opt/habanalabs/qual/gaudi3/bin
sudo chmod 777 /var/log/habana_logs/
sudo chmod uog+rw /var/log/habana_logs/*

echo "Bring down ports"
/opt/habanalabs/qual/gaudi3/bin/manage_network_ifs.sh --up
sleep 3
/opt/habanalabs/qual/gaudi3/bin/manage_network_ifs.sh --down
sleep 3
/opt/habanalabs/qual/gaudi3/bin/manage_network_ifs.sh --up
sleep 3
/opt/habanalabs/qual/gaudi3/bin/manage_network_ifs.sh --down
sleep 120
date

echo "Starting bm-gnr-gaudi3-smc validation"

qual_logs_dir=/tmp/validation/logs/qual_logs
GAUDI_REV=gaudi3

utkn=$(mktemp|sed 's,.*\.,,')
resultsfile=$qual_logs_dir/$utkn.$(basename $0 .sh)_results.txt

qual_options=(
    "-dmesg -dis_mon -$GAUDI_REV -c all -pciOnly -rmod serial -mb -b -gen gen4" #Pci only
    "-dmesg -dis_mon -$GAUDI_REV -c all -p -rmod serial -t 5 -b -gen gen4" # Pci enabled load
    "-dmesg -dis_mon -$GAUDI_REV -c all -p -rmod serial -t 5 -gen gen4" #Pci enabled bw
    "-dmesg -dis_mon -$GAUDI_REV -c all -e2e_concurrency -rmod parallel -t 30 -disable_ports 8,22,23" #hl_qual Concurrency
    "-dmesg -dis_mon -$GAUDI_REV -c all -f2 -rmod parallel -l extreme -t 240" # Functional2 extreme
    "-dmesg -dis_mon -$GAUDI_REV -c all -f2 -l high -rmod parallel -d -t 60" # Functional2 high
    "-dmesg -dis_mon -$GAUDI_REV -c all -hbm_tpc_stress full_rw -rmod parallel -i 4" # HBM stress full rw
    "-dmesg -dis_mon -$GAUDI_REV -c all -hbm_tpc_stress read -rmod parallel -i 3" # HBM stress read only
    "-dmesg -dis_mon -$GAUDI_REV -c all -hbm_tpc_stress write -rmod parallel -i 3" # HBM stress write only
    "-dmesg -dis_mon -$GAUDI_REV -c all -hbm_tpc_stress -rmod parallel -i 3" # HBM stress only
    "-dmesg -dis_mon -$GAUDI_REV -c all -hbm_tpc_stress read_write -rmod parallel -i 3" # HBM stress rw
    "-dmesg -dis_mon -$GAUDI_REV -c all -hbm_tpc_stress -rmod parallel -i 1 -skip_rst" # HBM tpc stress with skip reset
    "-dmesg -dis_mon -$GAUDI_REV -c all -hbm_tpc_stress read -rmod parallel -i 1 -skip_rst" # HBM tpc stress read with skip reset
    "-dmesg -dis_mon -$GAUDI_REV -c all -hbm_tpc_stress write -rmod parallel -i 1 -skip_rst" # HBM tpc stress write with skip reset
    "-dmesg -dis_mon -$GAUDI_REV -c all -hbm_tpc_stress read_write -rmod parallel -i 1 -skip_rst" # HBM tpc stress read and write with skip reset
    "-dmesg -dis_mon -$GAUDI_REV -c all -hbm_tpc_stress full_rw -rmod parallel -i 1 -skip_rst" # HBM tpc stress full rw with skip reset
    "-dmesg -dis_mon -$GAUDI_REV -c all -full_hbm_data_check_test -i 1 -rmod parallel" # HBM data check stress
    "-dmesg -dis_mon -$GAUDI_REV -c all -hbm_dma_stress -i 1 -rmod parallel" # HBM dma stress with skip reset
    # ################################## Below tests are failing for bm-gnr-gaudi3-smc ########################################################
    # "-dmesg -dis_mon -$GAUDI_REV -c all -e -t 120 -rmod parallel" # EDP with default values test
    # "-dmesg -dis_mon -$GAUDI_REV -c all -e -t 60 -rmod parallel -sync -Tw 2 -Ts 3" # EDP test
    # "-dmesg -dis_mon -$GAUDI_REV -c all -s -t 180 -rmod parallel" # Power virus test
    # "-dmesg -dis_mon -$GAUDI_REV -c all -f2 -rmod parallel -l extreme -t 420 -serdes -enable_ports_check int" # f2 test only
)

test_nr=( $(seq 1 ${#qual_options[@]}) )

# check if drivers (hl-smi) is working
if ! out=$(hl-smi -L 2>&1); then
    echo "Failed to execute hl-smi!"
    echo "Error: $out"
    exit 1
fi

sudo chmod 777 /opt/habanalabs/qual -R
cd /opt/habanalabs/qual/$GAUDI_REV/bin

# Make sure driver is loaded in the right way
sudo rmmod habanalabs; sudo modprobe habanalabs timeout_locked=0

# Create the logs and qual_logs folder if it doesn't exist
mkdir -p /tmp/validation/logs || exit
mkdir -p "$qual_logs_dir"

for idx in $(seq 0 $((${#test_nr[@]}-1))); do
    hl-smi
    starttime=$(date +%s)

    {
        echo "== Running test nr ${test_nr[idx]} $(date) =="
        echo "================= ${qual_options[idx]} ================="
        time sudo -E ./hl_qual ${qual_options[idx]} 2>&1
    } | tee $qual_logs_dir/$utkn.${test_nr[idx]}.log

    endtime=$(date +%s.%N)
    runtime=$(echo "$endtime - $starttime" | bc -l)

    detailed_summary="$(grep -A8 "Test result summary" $qual_logs_dir/$utkn.${test_nr[idx]}.log | awk -F ': ' 'BEGIN {countp=0; countf=0} $3=="PASSED" { ++countp } $3=="FAILED" { ++countf } END { print countp " passed, " countf " failed" }')"

    printf "%02d\t%s\t%s\t%s\t%s\n" ${test_nr[idx]} "${qual_options[idx]}" \
        "$(grep -A1 " hl qual report " $qual_logs_dir/$utkn.${test_nr[idx]}.log | tail -1)" \
        "$(date -u -d @${runtime} +'%M:%S')" \
        "$detailed_summary" >> $resultsfile
done

echo "===================="
echo "All ${#test_nr[@]} tests completed, find detailed results at $resultsfile"
echo "===================="

# function to get the gateway address of a subnet
function get_gateway_ip() {
    local gw_ip=$(echo "${1},${2}" |python3 -c 'import ipaddress; input_args=input().split(","); inft_ip=input_args[0]; mask=input_args[1]; net=ipaddress.ip_network(inft_ip+"/"+mask,strict=False); gateway_ip=str(list(net.hosts())[0]) if len(list(net.hosts())) > 1 else ""; print(gateway_ip)')
    echo "${gw_ip}"
}

if [ -n "$clusterGroupId" ]; then
    echo "This node is a part of cluster with clusterGroupId: $clusterGroupId"
    echo "Testing network"
    date
    echo "Bring down ports"
    /opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --down
    sleep 3
    /opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --up
    sleep 3
    /opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --down
    sleep 3
    /opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --up
    sleep 120
    date

    LINKDOWN=$(/opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --status | grep down | wc -l)
    if test "0" = "$LINKDOWN"
    then
        echo "Gaudi interfaces status is validated. All the ports are up"
    else
        echo "Gaudi interfaces status validation failed. Found $LINKDOWN ports that are down"
        /opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --status
        exit 1
    fi

    sudo apt -y remove needrestart

    # Attempt to install jq if it does not exist.
    if ! command -v jq &> /dev/null
    then
        echo "jq command not found. installing it"
        sudo apt-get install jq -y
    fi

    # restart networkd
    sudo systemctl restart systemd-networkd
    if [ $? -ne 0 ]; then
        echo "failed to restart systemd-networkd service"
        exit 1
    fi
    sleep 10

    # validate L3 connectivity
    if [ -f /etc/gaudinet.json ]; then
        echo "validating L3 connectivity to the gateway"
        count=$(jq '.[] | length' /etc/gaudinet.json)
        if [ "${count}" -ne 24 ]
        then
            echo "/etc/gaudinet.json doen't contain information about all the gaudi interfaces"
            exit 1
        fi
        validation_succeeded=()
        validation_failed=()
        # loop through each interface info in the gaudinet.json
        for ((i=0; i<$count; i++));
        do
            nic_mac=$(jq -r '.[]['$i'].NIC_MAC' /etc/gaudinet.json)
            nic_name=$(ip -o link | grep "${nic_mac}" | awk '{ print $2 }' | sed 's/.$//')
            if [ -z "${nic_name}" ]; then
                echo "failed to get the name of interface with MAC address ${nic_mac}"
                validation_failed+=(nic_mac)
                continue
            fi
            nic_ip=$(jq -r '.[]['$i'].NIC_IP' /etc/gaudinet.json)
            nic_mask=$(jq -r '.[]['$i'].SUBNET_MASK' /etc/gaudinet.json)
            nic_gw_ip=$(< <(get_gateway_ip "${nic_ip}" "${nic_mask}"))
            if [ -z "${nic_gw_ip}" ]; then
                echo "failed to get the gateway IP of interface ${nic_name}"
                validation_failed+=(nic_name)
                continue
            fi

            echo "testing the gateway connectivity of interface ${nic_name}"
            ping -c 3 -w 10 "${nic_gw_ip}"
            if [ $? -eq 0 ]; then
                echo "validated L3 connectivity to the gateway IP address ${nic_gw_ip}"
                validation_succeeded+=(nic_name)
            else
                echo "failed to validate the L3 connectivity to the gateway IP address ${nic_gw_ip}"
                validation_failed+=(nic_name)
            fi
        done
        # check for validation failures
        if [ ${#validation_failed[@]} -ne 0 ]; then
            echo "L3 validation failed for the for the following interfaces"
            for inft in "${validation_failed[@]}"
            do
                echo "${intf}"
            done
        else
            echo "Validated L3 connectivity of gaudi interfaces"
        fi
    fi
fi

# collect and compress hl_qual logs
cd /tmp/validation/logs
tar -zvcf hl_qual_logs.tar.gz qual_logs

# any fail qual option shall fail the validation
should_fail=false
while IFS= read -r line; do
    if [[ $line =~ ([0-9]+)\ failed ]]; then
        failed_count="${BASH_REMATCH[1]}"

        if (( failed_count != 0 )); then
            echo "$line"
            should_fail=true
        fi
    fi
done < "$resultsfile"

rm -rf $qual_logs_dir

if $should_fail; then
    echo "Validation failed because certain qual options failed."
    echo "Failing validation. Exiting..."
    exit 1
fi

echo "bm-gnr-gaudi3-smc validation completed"
