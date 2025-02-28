DEPLOYMENT_NAME="localstack"
NAMESPACE="idcs-system"

POD_NAME=$(kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=$DEPLOYMENT_NAME -o jsonpath="{.items[0].metadata.name}")
if [ -z "$POD_NAME" ]; then
  echo "No pods found for the deployment $DEPLOYMENT_NAME."
  exit 1
fi
while true; do
    READY=$(kubectl get pods $POD_NAME -n $NAMESPACE --output=jsonpath='{.status.containerStatuses[0].ready}')
    if [ "$READY" == "true" ]; then
        echo "Container is ready, Creating AWS resources"
        break
    else
        echo "Waiting for container to be ready..."
        sleep 5
    fi
done

kubectl cp ./deployment/localstack/setup-localstack.sh $NAMESPACE/$POD_NAME:/tmp
kubectl -n $NAMESPACE exec -it $POD_NAME -- sh /tmp/setup-localstack.sh
