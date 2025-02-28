// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import DeployedServices from '../../components/userSummary/DeployedServices'
import useInstanceGroupsStore from '../../store/instancesStore/instanceGroupStore'
import useInstancesStore from '../../store/instancesStore/InstancesStore'
import moment from 'moment'
import useClusterStore from '../../store/clusterStore/ClusterStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'

const dateFormat = 'MM/DD/YYYY hh:mm a'

const DeployedServicesContainer = (props: any): JSX.Element => {
  const userId = props.userId
  const userEmail = props.userEmail
  // local variables
  const instancesColumns = [
    {
      columnName: 'Instance Name',
      targetColumn: 'instanceName'
    },
    {
      columnName: 'IP',
      targetColumn: 'ip'
    },
    {
      columnName: 'Instance Type',
      targetColumn: 'instanceType'
    },
    {
      columnName: 'Created at',
      targetColumn: 'createdAt'
    },
    {
      columnName: 'Status',
      targetColumn: 'status'
    }
  ]

  const instanceGroupsColumns = [
    {
      columnName: 'Instance Group Name',
      targetColumn: 'instanceGroupName'
    },
    {
      columnName: 'Instance Type',
      targetColumn: 'instanceType'
    },
    {
      columnName: 'Instance Count',
      targetColumn: 'instanceCount'
    },
    {
      columnName: 'Ready Count',
      targetColumn: 'readyCount'
    }
  ]

  const iksColumns = [
    {
      columnName: 'Cluster Id',
      targetColumn: 'cluster-id',
      columnConfig: {
        behaviorType: 'hyperLink',
        behaviorFunction: 'setDetails'
      }
    },
    {
      columnName: 'Cluster Name',
      targetColumn: 'cluster-name'
    },
    {
      columnName: 'K8s Version',
      targetColumn: 'k8sversion'
    },
    {
      columnName: 'Status',
      targetColumn: 'clusterStatus'
    },
    {
      columnName: 'Created at',
      targetColumn: 'creationTimestamp'
    },
    {
      columnName: 'IMI Upgrade Available',
      targetColumn: 'cpupgradeavailable'
    },
    {
      columnName: 'K8s Upgrade Available',
      targetColumn: 'k8supgradeavailable'
    }
  ]

  const instancesEmptyGrid = {
    title: 'No instances found',
    subTitle: 'The selected account currently has no instances'
  }

  const instanceGroupsEmptyGrid = {
    title: 'No instance groups found',
    subTitle: 'The selected account currently has no instance groups'
  }

  const clustersEmptyGrid = {
    title: 'No clusters found',
    subTitle: 'The selected account currently has no clusters'
  }
  const throwError = useErrorBoundary()

  // global states
  // Global State for Instances
  const loadingInstances = useInstancesStore((state) => state.loading)
  const instancesState = useInstancesStore((state) => state.instances)
  const setInstancesState = useInstancesStore((state) => state.setInstances)

  // Global State for Instance Groups
  const loadingInstanceGroup = useInstanceGroupsStore((state) => state.loading)
  const instanceGroupsState = useInstanceGroupsStore((state) => state.instanceGroups)
  const setInstanceGroupsState = useInstanceGroupsStore((state) => state.setInstanceGroups)

  // Global State for Clusters and SC Clusters
  const loadingClusters = useClusterStore((state) => state.loading)
  const clustersData: any = useClusterStore((state) => state.clustersData)
  const setClustersData = useClusterStore((state) => state.setClustersData)

  // states
  const [instances, setInstances] = useState<any[]>([])
  const [instanceGroups, setInstanceGroups] = useState<any[]>([])
  const [clusters, setClusters] = useState<any[]>([])
  const [scClusters, setSCClusters] = useState<any[]>([])

  // hooks
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        const promises = []
        promises.push(setInstancesState(userId))
        promises.push(setInstanceGroupsState(userId))
        promises.push(setClustersData(false))
        await Promise.allSettled(promises)
      } catch (error) {
        throwError(error)
      }
    }
    if (userId || userEmail) fetch().catch(() => {})
  }, [userId, userEmail])

  useEffect(() => {
    if (userId != null) {
      setInstancesGridInfo()
      setInstanceGroupsGridInfo()
      setClustersGridInfo()
    }
  }, [instancesState, instanceGroupsState, clustersData, userId, userEmail])

  // functions
  // Set instances grid.
  function setInstancesGridInfo(): void {
    const instancesGridInfo = []
    // Initializing the states to verify the status.
    for (const index in instancesState.items) {
      const inst = { ...instancesState.items }

      instancesGridInfo.push({
        instanceName: inst[index]?.metadata?.name,
        ip: inst[index]?.status?.interfaces[0]?.addresses[0],
        instanceType: inst[index]?.spec?.instanceType,
        createdAt: moment(inst[index]?.metadata?.creationTimestamp).format(dateFormat),
        status: inst[index]?.status?.phase
      })
    }
    setInstances(instancesGridInfo)
  }

  // Set instance groups grid.
  function setInstanceGroupsGridInfo(): void {
    const instanceGroupGridInfo = []
    for (const index in instanceGroupsState.items) {
      const instG = { ...instanceGroupsState.items }
      const instanceCountVar = instG[index]?.spec?.instanceCount
      const readyCountVar = instG[index]?.status?.readyCount

      instanceGroupGridInfo.push({
        instanceName: instG[index]?.metadata?.name,
        instanceType: instG[index]?.spec?.instanceSpec?.instanceType,
        instanceCount: instanceCountVar,
        status: readyCountVar
      })
    }
    setInstanceGroups(instanceGroupGridInfo)
  }

  // Set clusters and SC clusters grids.
  function setClustersGridInfo(): void {
    const clustersGridInfo = []
    const scClustersGridInfo = []

    const userClusters = clustersData?.filter((cluster: any) => cluster.account === userId)
    for (const item in userClusters) {
      const clusterItem = { ...userClusters[item] }
      clusterItem.value = clusterItem.state

      const gridInfoData = {
        'cluster-id': clusterItem.uuid,
        'cluster-name': clusterItem.name,
        k8sversion: clusterItem.k8sversion,
        clusterstatus: clusterItem.state,
        creationTimestamp: {
          showField: true,
          type: 'date',
          toLocalTime: true,
          value: clusterItem.createddate,
          format: 'MM/DD/YYYY h:mm a'
        },
        cpupgradeavailable: clusterItem.cpupgradeversions.toString(),
        k8supgradeavailable: clusterItem.k8supgradeversions.toString()
      }
      if (clusterItem.clustertype === 'supercompute') scClustersGridInfo.push(gridInfoData)
      else clustersGridInfo.push(gridInfoData)
    }

    setClusters(clustersGridInfo)
    setSCClusters(scClustersGridInfo)
  }

  return (
    <DeployedServices
      instances={instances}
      instancesColumns={instancesColumns}
      loadingInstances={loadingInstances}
      instancesEmptyGrid={instancesEmptyGrid}
      instanceGroups={instanceGroups}
      instanceGroupsColumns={instanceGroupsColumns}
      loadingInstanceGroup={loadingInstanceGroup}
      instanceGroupsEmptyGrid={instanceGroupsEmptyGrid}
      clusters={clusters}
      scClusters={scClusters}
      iksColumns={iksColumns}
      loadingClusters={loadingClusters}
      clustersEmptyGrid={clustersEmptyGrid}
    />
  )
}

export default DeployedServicesContainer
