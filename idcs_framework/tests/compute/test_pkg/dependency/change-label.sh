#!/bin/bash

# Function to check if system is available
check_system_availability() {
    local kubeconfig_file="$1"
    local namespace="$2"
    local system_name="$3"

    # Fetch the state of the system using kubectl command
    system_state=$(kubectl --kubeconfig="$kubeconfig_file" get bmhost --field-selector=metadata.name="$system_name",metadata.namespace="$namespace" -l cloud.intel.com/verified -Ao jsonpath='{.items..status.provisioning.state}')
}

# Function to toggle label value
toggle_label_value() {
    local kubeconfig_file="$1"
    local namespace="$2"
    local system_name="$3"
    local new_value="$4"  # Accept the new label value as an argument
    local label_key="cloud.intel.com/unschedulable"

    # Check if system is available
    check_system_availability "$kubeconfig_file" "$namespace" "$system_name"

    # Update the label value
    if ! kubectl --kubeconfig="$kubeconfig_file" label --overwrite bmhost "$system_name" -n "$namespace" "$label_key=$new_value"; then
        echo "Failed to update label $label_key on BMHost $system_name in namespace $namespace to $new_value"
        exit 1
    fi

    echo "Label $label_key on BMHost $system_name in namespace $namespace updated to $new_value"
}

main() {

    if [ "$#" -ne 4 ]; then  # Expecting three arguments: namespace, system_name, and desired_label_value
        echo "Usage: $0 <kubeconfig_file> <namespace> <system_name> <desired_label_value>"
        exit 1
    fi

    local kubeconfig_file="$1"
    local namespace="$2"
    local system_name="$3"
    local desired_label_value="$4"  # Get the desired label value from command line argument

    # Toggle label value
    toggle_label_value "$kubeconfig_file" "$namespace" "$system_name" "$desired_label_value"
}

# Run the script
main "$@"
