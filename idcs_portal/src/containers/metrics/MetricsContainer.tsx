// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import Metrics from '../../components/metrics/Metrics'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'

const MetricsContainer = (): JSX.Element => {
  const throwError = useErrorBoundary()

  const mainTitle = 'Cloud Monitor - telemetry and logging'
  const allowedInstanceStatus = ['Ready']
  const allowedInstanceCategories = ['VirtualMachine']

  if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_METRICS_BARE_METAL)) {
    allowedInstanceCategories.push('BareMetalHost')
  }

  const emptyGrid = {
    title: 'No instances found',
    subTitle: 'Your account currently has no instances',
    action: {
      type: 'redirect',
      href: '/compute/reserve',
      label: 'Launch instance'
    }
  }

  // Local State
  const [instancesFilteredValues, setInstancesFilteredValues] = useState<any[]>([])

  // Global State
  const instances = useCloudAccountStore((state) => state.instances)
  const loading = useCloudAccountStore((state) => state.loading)
  const setInstances = useCloudAccountStore((state) => state.setInstances)

  // Hooks
  useEffect(() => {
    void loadInstances()
  }, [])

  useEffect(() => {
    getInstanceValues()
  }, [instances])

  // functions
  const loadInstances = async (): Promise<void> => {
    const fetchInstances = async (): Promise<void> => {
      try {
        await setInstances(false)
      } catch (error) {
        throwError(error)
      }
    }
    await fetchInstances()
  }

  const getInstanceValues = (): void => {
    if (instances.length > 0) {
      const instancesValues = instances
        .filter((x) => {
          const isClusterBM =
            x.nodegroupType === 'worker' && x.instanceTypeDetails?.instanceCategory === 'BareMetalHost'
          if (
            allowedInstanceCategories.includes(x.instanceTypeDetails?.instanceCategory as string) &&
            allowedInstanceStatus.includes(x.status) &&
            !isClusterBM
          ) {
            return true
          } else {
            return false
          }
        })
        .map((instance) => {
          return {
            name: instance.name + ` (${instance.instanceType})`,
            value: instance.resourceId,
            instanceName: instance.name,
            instanceCategory: instance.instanceTypeDetails?.instanceCategory
          }
        })
      setInstancesFilteredValues(instancesValues)
    }
  }

  return (
    <Metrics
      loading={loading}
      items={instances}
      emptyGrid={emptyGrid}
      filteredItems={instancesFilteredValues}
      mainTitle={mainTitle}
      itemType="instance"
    />
  )
}

export default MetricsContainer
