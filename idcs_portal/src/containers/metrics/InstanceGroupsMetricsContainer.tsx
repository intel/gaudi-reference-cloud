// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import Metrics from '../../components/metrics/Metrics'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'

const InstanceGroupsMetricsContainer = (): JSX.Element => {
  const throwError = useErrorBoundary()

  const mainTitle = 'Instance Groups Cloud Monitor - telemetry and logging'
  const allowedInstanceCategories = ['VirtualMachine']

  if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_METRICS_BARE_METAL)) {
    allowedInstanceCategories.push('BareMetalHost')
  }

  const emptyGrid = {
    title: 'No instance groups found',
    subTitle: 'Your account currently has no instance groups',
    action: {
      type: 'redirect',
      href: '/compute-groups/reserve',
      label: 'Launch instance groups'
    }
  }

  // Local State
  const [instanceGroupFilteredValues, setInstanceGroupFilteredValues] = useState<any[]>([])

  // Global State
  const instanceGroups = useCloudAccountStore((state) => state.instanceGroups)
  const loading = useCloudAccountStore((state) => state.loading)
  const setInstanceGroups = useCloudAccountStore((state) => state.setInstanceGroups)

  // Hooks
  useEffect(() => {
    void loadInstanceGroups()
  }, [])

  useEffect(() => {
    getInstanceGroupValues()
  }, [instanceGroups])

  // functions
  const loadInstanceGroups = async (): Promise<void> => {
    const fetchInstances = async (): Promise<void> => {
      try {
        await setInstanceGroups(false)
      } catch (error) {
        throwError(error)
      }
    }
    await fetchInstances()
  }

  const getInstanceGroupValues = (): void => {
    if (instanceGroups.length > 0) {
      const instancesValues = instanceGroups
        .filter(
          (x) =>
            allowedInstanceCategories.includes(x.instanceTypeDetails?.instanceCategory as string) && x.readyCount > 0
        )
        .map((instanceGroup) => {
          return {
            name: instanceGroup.name + ` (${instanceGroup.instanceType})`,
            value: instanceGroup.name,
            instanceName: instanceGroup.name,
            instanceCategory: instanceGroup.instanceTypeDetails?.instanceCategory
          }
        })
      setInstanceGroupFilteredValues(instancesValues)
    }
  }

  return (
    <Metrics
      loading={loading}
      items={instanceGroups}
      emptyGrid={emptyGrid}
      filteredItems={instanceGroupFilteredValues}
      mainTitle={mainTitle}
      itemType="instance-groups"
    />
  )
}

export default InstanceGroupsMetricsContainer
