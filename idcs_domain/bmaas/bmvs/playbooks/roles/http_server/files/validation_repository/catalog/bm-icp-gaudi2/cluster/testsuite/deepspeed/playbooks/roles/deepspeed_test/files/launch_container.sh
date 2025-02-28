#! /bin/bash

docker stop btower
docker rm btower

if [ -f /etc/gaudinet.json ]; then
    docker run \
        --env=HABANA_VISIBLE_DEVICES=all \
        --env=OMPI_MCA_btl_vader_single_copy_mechanism=none \
        --cap-add=sys_nice \
        --network=host \
        --restart=no \
        --device=/dev/infiniband \
        --runtime=habana \
        --shm-size=64g \
        --name btower \
        --volume /etc/gaudinet.json:/etc/habanalabs/gaudinet.json:ro \
        --volume /tmp/habana_logs:/var/log/habana_logs \
        -t -d btower:latest
else
    docker run \
        --env=HABANA_VISIBLE_DEVICES=all \
        --env=OMPI_MCA_btl_vader_single_copy_mechanism=none \
        --cap-add=sys_nice \
        --network=host \
        --restart=no \
        --device=/dev/infiniband \
        --runtime=habana \
        --shm-size=64g \
        --name btower \
        --volume /tmp/habana_logs:/var/log/habana_logs \
        -t -d btower:latest
fi
