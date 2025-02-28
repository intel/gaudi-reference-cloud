#!/usr/bin/env bash
# INTEL CONFIDENTIAL
# Copyright (C) 2023 Intel Corporation
# Show differences from base kube-scheduler code.

cd "$(dirname "$0")/.."

gitdiff() {
    # Below excludes deleted lines.
    git --no-pager diff --ignore-all-space --unified=0 $1 -- "${@:2}" | egrep -v '^\-'
    # Alternative views:
    # git --no-pager diff --ignore-all-space --compact-summary $1 -- "${@:2}"
    # git --no-pager diff --color --ignore-all-space --unified=0 $1 -- "${@:2}"
    # git --no-pager diff --ignore-all-space --unified=0 $1 -- "${@:2}"
}

gitdiff 82efe35763a8a208fee83353a902fc5373c0b9d7 apis/config/{types,types_pluginargs}.go
gitdiff 3d8cd257324c6b943d17dd2c323aa5fd99fb8b06 apis/config/zz_generated.deepcopy.go apis/config/deepcopy.go
gitdiff 33688a40fc531686ff31b3993faaf325833ed996 framework/parallelize/{error_channel,parallelism}.go
gitdiff 68c2f7ecc5717d62a2506e38544d60e34464a39a framework/plugins/helper/normalize_score.go
gitdiff aedb605aca34a1908e4c3204d10c2b000b3eccd5 framework/plugins/helper/shape_score.go
gitdiff 35e732e4564a31042c5a571d614ef42b3be5e82b framework/plugins/interpodaffinity/{filtering,plugin,scoring}.go
gitdiff 49ee0f9c61e91715a7543a3904ea4fe5c42d825e framework/plugins/names/*.go
gitdiff 58b82e28dc549c2f9e25e002c874f11c723b883d framework/plugins/nodeaffinity/node_affinity.go
gitdiff 49ee0f9c61e91715a7543a3904ea4fe5c42d825e framework/plugins/noderesources/fit.go
gitdiff d138b6ca11b889ad8385d9456094f8771f170df6 framework/plugins/noderesources/{least_allocated,resource_allocation}.go
gitdiff aedb605aca34a1908e4c3204d10c2b000b3eccd5 framework/plugins/noderesources/{balanced_allocation,most_allocated,requested_to_capacity_ratio}.go
gitdiff 8599abcf9b1e2649a40f74c0cb95f2286a303667 framework/plugins/nodeunschedulable/node_unschedulable.go
gitdiff 194b4dac913d1f0fbc9bff5a20b1c702fc1d3d90 framework/plugins/podtopologyspread/*.go
gitdiff 58b82e28dc549c2f9e25e002c874f11c723b883d framework/plugins/tainttoleration/taint_toleration.go
gitdiff 712586fc5dd56fed8a7991dff316d633abcaa3e1 framework/plugins/*.go
gitdiff 3b09c18490c5abacc0e65d3d3e9de37f420322f7 framework/runtime/{framework,metrics_recorder,registry}.go
gitdiff 0296a697f6261dfd37e210ea92d459e0b0d324d7 framework/{cycle_state,interface,listers,types}.go
gitdiff beb6470a83cab7062936c84381a5e32d4c3c27c5 internal/cache/{cache,interface,node_tree,snapshot}.go
gitdiff bd2ddd5e393214cb42567ce1524912d6e4771d4e metrics/*.go
gitdiff 077a2cddc702e4fb3b23016258b9deb5ddb2e005 profile/profile.go
gitdiff 53fee3e5c7418ffa311cb644270ea49cf8aa5dd4 scheduler/{eventhandlers,schedule_one,scheduler}.go
gitdiff 02e56f6f4f47ce92941c858704586f9197fa0878 util/{pod_resources,utils}.go

# adding test files for scheduler
gitdiff f59b6158a3f2769aa5128dd33f1394120027cdc5 apis/config/types_test.go
gitdiff f59b6158a3f2769aa5128dd33f1394120027cdc5 framework/parallelize/{error_channel_test,parallelism_test}.go
gitdiff a311bbbd2e042867a7726769f4f04300b1c09752 framework/{cycle_state_test,interface_test}.go
gitdiff d77b6e804f23ebab1ab09ab52936439475ac301b framework/plugins/helper/normalize_score_test.go
gitdiff d77b6e804f23ebab1ab09ab52936439475ac301b framework/plugins/interpodaffinity/{filtering_test,scoring_test}.go
gitdiff d77b6e804f23ebab1ab09ab52936439475ac301b framework/plugins/nodeaffinity/node_affinity_test.go
gitdiff d77b6e804f23ebab1ab09ab52936439475ac301b framework/plugins/noderesources/{fit_test,least_allocated_test,balanced_allocation_test,most_allocated_test,requested_to_capacity_ratio_test,test_util}.go
gitdiff d77b6e804f23ebab1ab09ab52936439475ac301b framework/plugins/nodeunschedulable/node_unschedulable_test.go
gitdiff d77b6e804f23ebab1ab09ab52936439475ac301b framework/plugins/tainttoleration/taint_toleration_test.go
gitdiff d77b6e804f23ebab1ab09ab52936439475ac301b framework/plugins/testing/*.go
gitdiff d77b6e804f23ebab1ab09ab52936439475ac301b framework/runtime/{framework_test,registry_test}.go
gitdiff d77b6e804f23ebab1ab09ab52936439475ac301b framework/types_test.go
gitdiff cf3bb2af25f9702e2a115e7ae561de8f3bff7a60 testing/*.go
gitdiff 23463eeface89d63d2ec4beeeaf791433e8679fd internal/cache/{cache_test,node_tree_test,snapshot_test}.go
gitdiff 23463eeface89d63d2ec4beeeaf791433e8679fd profile/profile_test.go
gitdiff 23463eeface89d63d2ec4beeeaf791433e8679fd scheduler/eventhandlers_test.go
gitdiff 23463eeface89d63d2ec4beeeaf791433e8679fd util/pod_resources_test.go










