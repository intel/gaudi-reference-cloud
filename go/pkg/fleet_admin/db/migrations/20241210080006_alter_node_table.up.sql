-- INTEL CONFIDENTIAL
-- Copyright (C) 2023 Intel Corporation

ALTER TABLE node
ADD CONSTRAINT unique_region_az_cluster_namespace_node
UNIQUE (region, availability_zone, cluster_id, namespace, node_name);
