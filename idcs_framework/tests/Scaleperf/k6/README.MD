# Performance and Scale Testing RunBook:

## Executing concurrent Tests on Local:

```To be Filled```

## Executing load tests and publishing report to grafana
    ```k6 run --out influxdb=<<influxdb-instance:8086/k6>> scriptname -e env_var1=value1 -e env_var2=value2 ... ```

## Executing concurrent Tests on CICD:
1) Pull the image from harbor.

    ```docker pull k6-scaleperf:latest```

2) Bring up the container and execute concurrent tests with default parameters.

    ```docker run k6-scaleperf:latest```

3) Bring up the container and execute concurrent tests with customized parameters.

    ```docker run k6-scaleperf:latest "-e" "vuCount=<<concurrent users count>>"```

