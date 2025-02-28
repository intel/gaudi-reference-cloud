<!--INTEL CONFIDENTIAL-->
<!--Copyright (C) 2023 Intel Corporation-->
This directory contains Instance Types for development and staging environments.
For production, see [../../../../../build/environments/prod/InstanceType/](../../../../../build/environments/prod/InstanceType/).

When adding VM instance types, you must also add labels to Harvester worker nodes indicating the supported instance types.

For example:

    - instance-type.cloud.intel.com/tiny: true
    - instance-type.cloud.intel.com/small: true
