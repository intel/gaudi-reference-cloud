#!/usr/bin/env bash
# Checking if the tenant statefulset pod is created

wait-for-tenant-pool-pod() {
    # wait for 'minio-idc-pool-0-0` pod, it would show 0/1 until we enable it
    for (( ; ; ))
    do
        kubectl get pods --namespace minio-idc-tenant | grep minio-idc-pool-0-0
        if [[ $? = 0 ]]
        then
          break
        fi
        echo "waiting for minio tenant pod..."
        sleep 4
    done

}

wait-for-tenant-pool-pod
