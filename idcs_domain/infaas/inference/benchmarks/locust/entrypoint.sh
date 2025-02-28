#!/bin/bash

# Default to throughput test if not set
LOCUST_FILE=${TEST_OBJECTIVE:-load}

# Map the chosen mode to the correct file
if [ "$LOCUST_FILE" = "load" ]; then
    cp /home/locust/load.py /home/locust/locustfile.py
else
    cp /home/locust/throughput.py /home/locust/locustfile.py
fi

# Execute Locust with the chosen file
exec locust "$@"