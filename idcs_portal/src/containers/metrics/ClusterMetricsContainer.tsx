// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import Metrics from '../../components/metrics/Metrics'
import useClusterStore from '../../store/clusterStore/ClusterStore'

const ClusterMetricsContainer = (): JSX.Element => {
  const throwError = useErrorBoundary()

  const mainTitle = 'Clusters Cloud Monitor - telemetry and logging'
  const allowedStatus = ['Active']

  const emptyGrid = {
    title: 'No cluster found',
    subTitle: 'Your account currently has no clusters',
    action: {
      type: 'redirect',
      href: '/cluster/reserve',
      label: 'Launch cluster'
    }
  }

  // Local State
  const [clustersFilteredValues, setClustersFilteredValues] = useState<any[]>([])

  // Global State
  const clusters = useClusterStore((state) => state.clustersData)
  const loading = useClusterStore((state) => state.loading)
  const setClusters = useClusterStore((state) => state.setClustersData)

  // Hooks
  useEffect(() => {
    void loadClusters()
  }, [])

  useEffect(() => {
    getClustersValues()
  }, [clusters])

  // functions
  const loadClusters = async (): Promise<void> => {
    const fetchClusters = async (): Promise<void> => {
      try {
        await setClusters(false)
      } catch (error) {
        throwError(error)
      }
    }
    await fetchClusters()
  }

  const getClustersValues = (): void => {
    if (clusters && clusters.length > 0) {
      const clustersValues = clusters
        .filter((x) => {
          let isAllow = false

          if (allowedStatus.includes(x.clusterstate)) {
            const clusterAnnotations = x.annotations

            if (clusterAnnotations.length > 0) {
              const metricsEnable = clusterAnnotations.filter(
                (x: any) => x.key === 'cloudmonitorEnable' && String(true).toLowerCase() === x.value.toLowerCase()
              )
              isAllow = metricsEnable.length > 0
            }
          }

          return isAllow
        })
        .map((cluster) => {
          return {
            name: cluster.name,
            value: cluster.uuid
          }
        })
      setClustersFilteredValues(clustersValues)
    }
  }

  return (
    <Metrics
      loading={loading}
      items={clusters}
      emptyGrid={emptyGrid}
      filteredItems={clustersFilteredValues}
      mainTitle={mainTitle}
      itemType="cluster"
    />
  )
}

export default ClusterMetricsContainer
