// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
package universe_config

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Universe Config Tests", func() {
	It("NewUniverseConfigFromFile should succeed", func() {
		ctx := context.Background()
		_, err := NewUniverseConfigFromFile(ctx, "go/pkg/universe_deployer/universe_config/testdata/basic.json")
		Expect(err).Should(Succeed())
	})

	It("ReplaceCommits should succeed", func() {
		ctx := context.Background()
		universeConfig := UniverseConfig{
			Environments: map[string]*UniverseEnvironment{
				"staging": {
					Components: map[string]*UniverseComponent{
						"billing": {
							Commit: "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
						},
						"cloudaccount": {
							Commit: "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
						},
					},
					Regions: map[string]*UniverseRegion{
						"us-staging-1": {
							Components: map[string]*UniverseComponent{
								"computeApiServer": {
									Commit: "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
								},
							},
							AvailabilityZones: map[string]*UniverseAvailabilityZone{
								"us-staging-1a": {
									Components: map[string]*UniverseComponent{
										"compute": {
											Commit: "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
										},
										"computeBmInstanceOperator": {
											Commit: "b966d5f3a3aa2762532d82a06e5ea0435fbc89d4",
										},
									},
								},
							},
						},
					},
				},
			},
		}
		universeConfig.ReplaceCommits(ctx, map[string]string{
			"b966d5f3a3aa2762532d82a06e5ea0435fbc89d4": "58c821c1919f28e6e3779f78d940e06ddb36a3b8",
		})
		expected := UniverseConfig{
			Environments: map[string]*UniverseEnvironment{
				"staging": {
					Components: map[string]*UniverseComponent{
						"billing": {
							Commit: "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
						},
						"cloudaccount": {
							Commit: "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
						},
					},
					Regions: map[string]*UniverseRegion{
						"us-staging-1": {
							Components: map[string]*UniverseComponent{
								"computeApiServer": {
									Commit: "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
								},
							},
							AvailabilityZones: map[string]*UniverseAvailabilityZone{
								"us-staging-1a": {
									Components: map[string]*UniverseComponent{
										"compute": {
											Commit: "a5ddeaa02a04c61e5090fae7f6981e471d89d54b",
										},
										"computeBmInstanceOperator": {
											Commit: "58c821c1919f28e6e3779f78d940e06ddb36a3b8", // replaced
										},
									},
								},
							},
						},
					},
				},
			},
		}
		Expect(universeConfig).Should(Equal(expected))
	})
})
