#!/bin/bash
#
# You should get the 'hfToken' secret from the MaaS team,
# or set it using a valid HuggingFace token of your own.
#
# You should get the 'awsAccessKeyId' and 'awsSecreteAccessKey' secrets from the MaaS team

echo "Creating 'benchmark' namespace..."
kubectl create namespace benchmark

echo "creating HF token secret..."
kubectl create secret generic hf-api-token-secret --from-literal=HF_API_TOKEN=$hfToken -n benchmark

echo "creating AWS secrets..."
kubectl create secret generic aws-access-secrets --from-literal=AWS_ACCESS_KEY_ID=$awsAccessKeyId --from-literal=AWS_SECRET_ACCESS_KEY=$awsSecreteAccessKey -n benchmark