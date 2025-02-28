#!/usr/bin/env python3
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation

import argparse
import sys
import logging
import os
import requests
import yaml
import ipaddress

STORAGE_SUBNET='100.96.0.0/12'

def main():
    logging.basicConfig(level=logging.INFO)
    if not sys.version_info >= (3, 7):
        logging.info("Python 3.7 or newer version is required")
    parser = argparse.ArgumentParser(
        description="Extract Men & Mice ranges into IDC subnet files used by IP Resource Manager. Output files can be loaded with Git to GRPC Synchronizer."
    )
    parser.add_argument("--mmws-url", default="https://internal-placeholder.com")
    parser.add_argument("--mmws-username", default="idc")
    parser.add_argument("--mmws-password")
    parser.add_argument("--output-dir", default="build/environments/staging/Subnet")
    parser.add_argument("--region", default="us-staging-1")
    parser.add_argument("--availability-zone", default="us-staging-1a")
    parser.add_argument("--log-level", type=int, default=logging.INFO, help="10=DEBUG,20=INFO")
    args = parser.parse_args()

    logging.basicConfig(level=args.log_level)
    logging.debug("args=%s" % str(args))

    auth = (args.mmws_username, args.mmws_password)
    url = "%s/mmws/api/Ranges" % args.mmws_url
    logging.info("url=%s" % url)
    response = requests.get(url, auth=auth)
    logging.info("response=%s" % response)
    assert response.status_code == 200
    ranges = response.json()
    logging.info("Found %d ranges" % len(ranges["result"]["ranges"]))
    for range in ranges["result"]["ranges"]:
        if "customProperties" in range:
            customProperties = range["customProperties"]
            if "ConsumerID" in customProperties:
                consumerID = customProperties["ConsumerID"]
                if consumerID == args.region:
                    subnet = dict(
                        region=args.region,
                        availabilityZone=args.availability_zone,
                        subnet=range["name"],
                        vlanId=int(customProperties["VLAN"]),
                    )
                    storage_network = ipaddress.ip_network(STORAGE_SUBNET, strict= False)
                    current_network = ipaddress.ip_network(range["name"], strict= False)
                    if current_network.subnet_of(storage_network) and subnet:
                        subnet['addressSpace'] = "storage"
                    subnet_yaml = yaml.dump(subnet)
                    output_filename = os.path.join(args.output_dir, "%s-%s.yaml" % (range["from"], customProperties["Bitmask"]))
                    with open(output_filename, "w") as output_file:
                        output_file.write(subnet_yaml)
                    logging.info("Wrote " + output_filename)


if __name__ == "__main__":
    main()
