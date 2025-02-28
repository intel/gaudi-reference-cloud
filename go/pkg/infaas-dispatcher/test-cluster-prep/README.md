# Test Cluster Preparation

This directory contains scripts and configuration files to set up and test the Kubernetes cluster for the infaas-dispatcher project.

## Files

- `create-apples.sh`: A script to deploy and test the `apple-app` in the Kubernetes cluster.
- `apple.yaml`: Kubernetes manifest for deploying the `apple-app` pod and service.
- `gateway.yaml`: Istio Gateway configuration for the `apple-app`.
- `virtualservice.yaml`: Istio VirtualService configuration for routing traffic to the `apple-app`.

## Usage

### Prerequisites

- Kubernetes cluster
- `kubectl` command-line tool
- `istioctl` command-line tool (for Istio configurations)
- Successful execution of `prepare-cluster.sh`

### Steps

1. **Set up the Kubernetes cluster**:
   Ensure your Kubernetes cluster is up and running, and you have the necessary context set in your `kubeconfig`.

2. **Run the cluster preparation script**:
   Execute the `prepare-cluster.sh` script to set up the necessary resources and configurations.

   ```sh
   ./prepare-cluster.sh <env_name>


```markdown
### Deploy the apple-app

Run the `create-apples.sh` script to deploy the `apple-app` and its associated services and configurations.

```sh
./create-apples.sh -k <path-to-kubeconfig>
```

**Options:**
- `-k`: Path to the kubeconfig file.
- `-d`: Enable debug mode.
- `-s`: Skip termination of resources after the script completes.

**Verify the deployment:** The script will apply the deployment manifests and perform a `curl` request to verify the response from the `apple-app`.

**Clean up:** The script will automatically clean up the deployed resources unless the `-s` option is used.

### Configuration Details

- `apple.yaml`: Defines the `apple-app` pod and service. The pod runs a simple HTTP server that responds with "apple".
- `gateway.yaml`: Configures an Istio Gateway to expose the `apple-app` service over HTTPS.
- `virtualservice.yaml`: Configures an Istio VirtualService to route traffic to the `apple-app` service based on the URI prefix `/apples`.

### Notes

- Ensure that the Istio ingress gateway is properly configured and running in your cluster.
- The script assumes that the `ingress-certs` secret is already created in the cluster for HTTPS.
```