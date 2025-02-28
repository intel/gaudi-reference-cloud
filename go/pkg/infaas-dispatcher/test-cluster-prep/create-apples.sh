#!/bin/bash

# Define the directory containing the additional files
SCRIPT_DIR=$(dirname "$0")

# Initialize variables
KUBECONFIG=""
DEBUG=false
APPLES_PODS_AND_SVC="$SCRIPT_DIR/apple.yaml"
APPLES_GW="$SCRIPT_DIR/gateway.yaml"
APPLES_VS="$SCRIPT_DIR/virtualservice.yaml"

terminate() {
  if [ "$SKIP_TERMINATION" = true ]; then
    echo "Skipping termination as requested."
    exit "$1"
  fi
  echo "Deleting deployment manifests..."
  kubectl delete -f "$APPLES_PODS_AND_SVC" -f "$APPLES_GW" -f "$APPLES_VS" --kubeconfig="$KUBECONFIG"
  exit "$1"
}

# Parse command-line options
while getopts ":k:ds" opt; do
  case ${opt} in
    k )
      KUBECONFIG=$OPTARG
      ;;
    d )
      DEBUG=true
      ;;
    s )
      SKIP_TERMINATION=true
      ;;
    \? )
      echo "Invalid option: -$OPTARG" 1>&2
      exit 1
      ;;
    : )
      echo "Invalid option: -$OPTARG requires an argument" 1>&2
      exit 1
      ;;
  esac
done

# Enable debugging if the -d flag is set
if [ "$DEBUG" = true ]; then
  set -x
fi

# Check if KUBECONFIG is set
if [ -z "$KUBECONFIG" ]; then
  # No flag provided, use KUBECONFIG environment variable
  if [ -z "$KUBECONFIG" ]; then
    echo "Error: No kubeconfig file provided and KUBECONFIG environment variable is not set."
    exit 1
  fi
  echo "No flag provided. Using KUBECONFIG: $KUBECONFIG"
else
  echo "Using provided kubeconfig file: $KUBECONFIG"
fi

# Ask for user approval
read -r -p "Do you want to use this KUBECONFIG to apply the deployment manifests? (Y/N): " approval
if [ "$approval" != "Y" ]; then
  echo "Operation cancelled by user."
  exit 1
fi

# Apply the deployment manifests
echo "Applying deployment manifests..."
# Example of using a file in the same directory

kubectl apply -f "$APPLES_PODS_AND_SVC" -f "$APPLES_GW" -f "$APPLES_VS" --kubeconfig="$KUBECONFIG"


# Verify we have only one
INGRESS_POD=$(kubectl get pods -n ingress-nginx -l app.kubernetes.io/name=ingress-nginx --kubeconfig="$KUBECONFIG" -o jsonpath='{.items[*].metadata.name}')
# Count the number of pods returned
POD_COUNT=$(echo "$INGRESS_POD" | wc -w)
if [ "$POD_COUNT" -ne 1 ]; then
  echo "Error: Expected exactly one ingress-nginx-controller pod, but found $INGRESS_POD_COUNT."
  EXIT_CODE=1
  terminate $EXIT_CODE
fi

# Port-forward to the nginx ingress controller
echo "Port-forwarding to nginx ingress controller..."
kubectl port-forward "$INGRESS_POD" 8443:443 -n ingress-nginx --kubeconfig="$KUBECONFIG" &
PORT_FORWARD_PID=$!
if ! ps -p $PORT_FORWARD_PID > /dev/null; then
  echo "Error: Port-forwarding failed."
  EXIT_CODE=1
  terminate $EXIT_CODE
fi

# Give port-forwarding some time to establish
sleep 5

# Perform the curl request
echo "Performing curl request..."
RESPONSE=$(curl -k https://localhost:8443/apples)

# Verify the response
if [ "$RESPONSE" = "apple" ]; then
  echo "Response is correct: $RESPONSE"
  EXIT_CODE=0
else
  echo "Response is incorrect: $RESPONSE"
  EXIT_CODE=1
fi

# Clean up port-forwarding
kill $PORT_FORWARD_PID
terminate $EXIT_CODE