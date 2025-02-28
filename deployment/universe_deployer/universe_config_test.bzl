# Unit tests for Universe Deployer.
# See https://bazel.build/rules/testing#testing-starlark-utilities.
# Run with: make test-universe-deployer

load("@bazel_skylib//lib:unittest.bzl", "asserts", "unittest")
load(":universe_config.bzl", 
    "get_commit_to_components_dict_from_universe_config",
    "trim_universe_config_for_manifests",
    "trim_universe_config_for_push",
    "COMPONENT_ALL",
)

FILTERED_COMMIT = "b966d5f3a3aa2762532d82a06e5ea0435fbc89d4"
FILTERED_COMPONENT = "filteredComponent"

GET_COMMIT_TO_COMPONENTS_DICT_FROM_UNIVERSE_CONFIG_TEST_TABLE = [
    {
        "test_name": "Basic",
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                        "cloudaccount": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "computeApiServer": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        "computeBmInstanceOperator": {
                                            "commit": "b966d5f3a3aa2762532d82a06e5ea0435fbc89d4",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "a5ddeaa02a04c61e5090fae7f6981e471d89d54b": {
                "billing": True,
                "cloudaccount": True,
                "compute": True,
                "computeApiServer": True,
            },
            "b966d5f3a3aa2762532d82a06e5ea0435fbc89d4": {
                "computeBmInstanceOperator": True,
            },
        },
    },
    {
        "test_name": "When global components does not exist, regions should be processed",
        "universe_config": {
            "environments": {
                "staging": {
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "computeApiServer": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        "computeBmInstanceOperator": {
                                            "commit": "b966d5f3a3aa2762532d82a06e5ea0435fbc89d4",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "a5ddeaa02a04c61e5090fae7f6981e471d89d54b": {
                "compute": True,
                "computeApiServer": True,
            },
            "b966d5f3a3aa2762532d82a06e5ea0435fbc89d4": {
                "computeBmInstanceOperator": True,
            },
        },
    },
]

TRIM_UNIVERSE_CONFIG_FOR_PUSH_TEST_TABLE = [
    {
        "test_name": "Include environment with a matching commit only in AZ",
        "commit": FILTERED_COMMIT,
        "component": COMPONENT_ALL,
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                        "cloudaccount": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                                "computeApiServer": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        "computeBmInstanceOperator": {
                                            "commit": FILTERED_COMMIT,
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "environments": {
                "staging": {
                    "components": {
                        "all": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                },
            },
        },
    },
    {
        "test_name": "Include environment with a matching commit only in region",
        "commit": FILTERED_COMMIT,
        "component": COMPONENT_ALL,
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                        "cloudaccount": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                                "computeApiServer": {
                                    "commit": FILTERED_COMMIT,
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        "computeBmInstanceOperator": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "environments": {
                "staging": {
                    "components": {
                        "all": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                },
            },
        },
    },
    {
        "test_name": "Include environment with a matching commit only in global",
        "commit": FILTERED_COMMIT,
        "component": COMPONENT_ALL,
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                        "cloudaccount": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                                "computeApiServer": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        "computeBmInstanceOperator": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "environments": {
                "staging": {
                    "components": {
                        "all": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                },
            },
        },
    },
    {
        "test_name": "When an environment has no matching commits, remove it",
        "commit": FILTERED_COMMIT,
        "component": COMPONENT_ALL,
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": FILTERED_COMMIT,
                        },
                        "cloudaccount": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": FILTERED_COMMIT,
                                },
                                "computeApiServer": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        "computeBmInstanceOperator": {
                                            "commit": FILTERED_COMMIT,
                                        },
                                    },
                                    "us-staging-1b": {
                                        "components": {
                                            "compute": {
                                                "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                                            },
                                            "computeBmInstanceOperator": {
                                                "commit": FILTERED_COMMIT,
                                            },
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
                "prod": {
                    "components": {
                        "billing": {
                            "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                        },
                    },
                    "regions": {
                        "us-region-1": {
                            "components": {
                                "compute": {
                                    "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                                },
                            },
                            "availabilityZones": {
                                "us-region-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "environments": {
                "staging": {
                    "components": {
                        "all": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                },
            },
        },
    },
    {
        "test_name": "When there are two environments with matching commits, include both",
        "commit": FILTERED_COMMIT,
        "component": COMPONENT_ALL,
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": FILTERED_COMMIT,
                        },
                        "cloudaccount": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": FILTERED_COMMIT,
                                },
                                "computeApiServer": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        "computeBmInstanceOperator": {
                                            "commit": FILTERED_COMMIT,
                                        },
                                    },
                                    "us-staging-1b": {
                                        "components": {
                                            "compute": {
                                                "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                                            },
                                            "computeBmInstanceOperator": {
                                                "commit": FILTERED_COMMIT,
                                            },
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
                "prod": {
                    "components": {
                        "billing": {
                            "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                        },
                    },
                    "regions": {
                        "us-region-1": {
                            "components": {
                                "compute": {
                                    "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                                },
                            },
                            "availabilityZones": {
                                "us-region-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": FILTERED_COMMIT,
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "environments": {
                "staging": {
                    "components": {
                        "all": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                },
                "prod": {
                    "components": {
                        "all": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                },
            },
        },
    },
    {
        "test_name": "When there are two environments (staging and testing1) with matching commits but they share Harbor, include only staging",
        "commit": FILTERED_COMMIT,
        "component": COMPONENT_ALL,
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": FILTERED_COMMIT,
                        },
                        "cloudaccount": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": FILTERED_COMMIT,
                                },
                                "computeApiServer": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        "computeBmInstanceOperator": {
                                            "commit": FILTERED_COMMIT,
                                        },
                                    },
                                    "us-staging-1b": {
                                        "components": {
                                            "compute": {
                                                "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                                            },
                                            "computeBmInstanceOperator": {
                                                "commit": FILTERED_COMMIT,
                                            },
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
                "testing1": {
                    "components": {
                        "billing": {
                            "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                        },
                    },
                    "hint_push_to_environment": "staging",
                    "regions": {
                        "us-testing1-1": {
                            "components": {
                                "compute": {
                                    "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                                },
                            },
                            "availabilityZones": {
                                "us-region-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": FILTERED_COMMIT,
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "environments": {
                "staging": {
                    "components": {
                        "all": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                },
            },
        },
    },
    {
        "test_name": "When a specific component is specified, only that component should be included",
        "commit": FILTERED_COMMIT,
        "component": FILTERED_COMPONENT,
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                        "cloudaccount": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                                "computeApiServer": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        FILTERED_COMPONENT: {
                                            "commit": FILTERED_COMMIT,
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "environments": {
                "staging": {
                    "components": {
                        FILTERED_COMPONENT: {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                },
            },
        },
    },
    {
        "test_name": "When a specific component is specified, but component does not have commit, return an empty set",
        "commit": FILTERED_COMMIT,
        "component": FILTERED_COMPONENT,
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                        "cloudaccount": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                                "computeApiServer": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        FILTERED_COMPONENT: {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "environments": {
            },
        },
    },
]

TRIM_UNIVERSE_CONFIG_FOR_MANIFESTS_TEST_TABLE = [
    {
        "test_name": "Remove unmatched commits from global, region, AZ",
        "commit": FILTERED_COMMIT,
        "component": COMPONENT_ALL,
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": FILTERED_COMMIT,
                        },
                        "cloudaccount": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": FILTERED_COMMIT,
                                },
                                "computeApiServer": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        "computeBmInstanceOperator": {
                                            "commit": FILTERED_COMMIT,
                                        },
                                    },
                                    "us-staging-1b": {
                                        "components": {
                                            "compute": {
                                                "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                                            },
                                            "computeBmInstanceOperator": {
                                                "commit": FILTERED_COMMIT,
                                            },
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
                "prod": {
                    "components": {
                        "billing": {
                            "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                        },
                    },
                    "regions": {
                        "us-region-1": {
                            "components": {
                                "compute": {
                                    "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                                },
                            },
                            "availabilityZones": {
                                "us-region-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "c44c863a711dd281a4cbb4f305b66895e34e7e8a",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": FILTERED_COMMIT,
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "computeBmInstanceOperator": {
                                            "commit": FILTERED_COMMIT,
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    },
    {
        "test_name": "When all AZ components are filtered out, remove AZ",
        "commit": FILTERED_COMMIT,
        "component": COMPONENT_ALL,
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": FILTERED_COMMIT,
                        },
                        "cloudaccount": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": FILTERED_COMMIT,
                                },
                                "computeApiServer": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        "computeBmInstanceOperator": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": FILTERED_COMMIT,
                                },
                            },
                            "availabilityZones": {
                            },
                        },
                    },
                },
            },
        },
    },
    {
        "test_name": "When all region components and AZs are filtered out, remove region",
        "commit": FILTERED_COMMIT,
        "component": COMPONENT_ALL,
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": FILTERED_COMMIT,
                        },
                        "cloudaccount": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                                "computeApiServer": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        "computeBmInstanceOperator": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                    "regions": {
                    },
                },
            },
        },
    },
    {
        "test_name": "When all global components and regions are filtered out, remove environment",
        "commit": FILTERED_COMMIT,
        "component": COMPONENT_ALL,
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        "billing": {
                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                "compute": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
                "prod": {
                    "components": {
                        "billing": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                    "regions": {
                        "us-region-1": {
                            "components": {
                                "compute": {
                                    "commit": FILTERED_COMMIT,
                                },
                            },
                            "availabilityZones": {
                                "us-region-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": FILTERED_COMMIT,
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "environments": {
                "prod": {
                    "components": {
                        "billing": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                    "regions": {
                        "us-region-1": {
                            "components": {
                                "compute": {
                                    "commit": FILTERED_COMMIT,
                                },
                            },
                            "availabilityZones": {
                                "us-region-1a": {
                                    "components": {
                                        "compute": {
                                            "commit": FILTERED_COMMIT,
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    },
    {
        "test_name": "When a specific component is specified, only that component should be included",
        "commit": FILTERED_COMMIT,
        "component": FILTERED_COMPONENT,
        "universe_config": {
            "environments": {
                "staging": {
                    "components": {
                        FILTERED_COMPONENT: {
                            "commit": FILTERED_COMMIT,
                        },
                        "cloudaccount": {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                FILTERED_COMPONENT: {
                                    "commit": FILTERED_COMMIT,
                                },
                                "computeApiServer": {
                                    "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                },
                            },
                            "availabilityZones": {
                                "us-staging-1a": {
                                    "components": {
                                        FILTERED_COMPONENT: {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                        "computeBmInstanceOperator": {
                                            "commit": "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
        "expected": {
            "environments": {
                "staging": {
                    "components": {
                        FILTERED_COMPONENT: {
                            "commit": FILTERED_COMMIT,
                        },
                    },
                    "regions": {
                        "us-staging-1": {
                            "components": {
                                FILTERED_COMPONENT: {
                                    "commit": FILTERED_COMMIT,
                                },
                            },
                            "availabilityZones": {
                            },
                        },
                    },
                },
            },
        },
    },
]

def _get_commit_to_components_dict_from_universe_config_test_impl(ctx):
    env = unittest.begin(ctx)
    for test in GET_COMMIT_TO_COMPONENTS_DICT_FROM_UNIVERSE_CONFIG_TEST_TABLE:
        if not test.get("pending", False):
            asserts.equals(
                env,
                expected = test["expected"],
                actual = get_commit_to_components_dict_from_universe_config(test["universe_config"], False),
                msg = test["test_name"],
            )
    return unittest.end(env)

def _trim_universe_config_for_push_test_impl(ctx):
    env = unittest.begin(ctx)
    for test in TRIM_UNIVERSE_CONFIG_FOR_PUSH_TEST_TABLE:
        if not test.get("pending", False):
            asserts.equals(
                env,
                expected = test["expected"],
                actual = trim_universe_config_for_push(test["universe_config"], test["commit"], test["component"]),
                msg = test["test_name"],
            )
    return unittest.end(env)

def _trim_universe_config_for_manifests_test_impl(ctx):
    env = unittest.begin(ctx)
    for test in TRIM_UNIVERSE_CONFIG_FOR_MANIFESTS_TEST_TABLE:
        if not test.get("pending", False):
            asserts.equals(
                env,
                expected = test["expected"],
                actual = trim_universe_config_for_manifests(test["universe_config"], test["commit"], test["component"]),
                msg = test["test_name"],
            )
    return unittest.end(env)

get_commit_to_components_dict_from_universe_config_test = unittest.make(_get_commit_to_components_dict_from_universe_config_test_impl)
trim_universe_config_for_push_test = unittest.make(_trim_universe_config_for_push_test_impl)
trim_universe_config_for_manifests_test = unittest.make(_trim_universe_config_for_manifests_test_impl)

def universe_config_test_suite(name):
    unittest.suite(
        name,
        trim_universe_config_for_manifests_test,
        trim_universe_config_for_push_test,
        get_commit_to_components_dict_from_universe_config_test,
    )
