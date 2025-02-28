#!/bin/bash

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
    echo "all links are up"
else
    echo "found $LINKDOWN that are down"
    /opt/habanalabs/qual/gaudi2/bin/manage_network_ifs.sh --status
    exit 1
fi
