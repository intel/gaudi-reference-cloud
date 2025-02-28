#!/bin/bash

SLEEP=180
# Edit the sever name to run tgi/vllm server
SERVER="tgi"
SCRIPT_DIR=$(dirname "$(realpath "$0")")
PARENT_DIR=$(dirname "$(realpath "$0")")/..
NAMESPACE="benchmark"
LOCUST_POD_NAME="locust-benchmark"
LOCUST_FILE_PATH="$SCRIPT_DIR/locust.yaml"
SERVER_FILE_PATH="$PARENT_DIR/$SERVER.yaml"

echo "Deploying $SERVER server..."
kubectl apply -f $SERVER_FILE_PATH -n $NAMESPACE
# Sleep after server creation to make sure
# we start testing after server warmup
echo "Sleeping during model warmup ($SLEEP [s])"
sleep $SLEEP

# Create ClusterIP service
echo "Creating ClusterIP service for $SERVER server..."
cat <<EOF | kubectl apply -n $NAMESPACE -f -
apiVersion: v1
kind: Service
metadata:
  name: ${SERVER}-server
  namespace: $NAMESPACE
spec:
  selector:
    app: $SERVER-server
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
  type: ClusterIP
EOF

# Apply the pod file
echo "Deploying locust tester..."
# Check if the pod exists before deleting it
if kubectl get pod $LOCUST_POD_NAME -n $NAMESPACE > /dev/null 2>&1; then
    echo "Pod '$LOCUST_POD_NAME' exists in namespace '$NAMESPACE', deleting it..."
    kubectl delete pod "$LOCUST_POD_NAME" -n $NAMESPACE
else
    echo "Pod '$LOCUST_POD_NAME' does not exist in namespace '$NAMESPACE'."
fi
kubectl apply -f $LOCUST_FILE_PATH -n $NAMESPACE