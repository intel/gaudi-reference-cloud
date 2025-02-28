// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import ChartsWidget from '../../../components/homePage/secondUseDashboard/ChartsWidget'
import useDarkModeStore from '../../../store/darkModeStore/DarkModeStore'
import useCloudAccountStore from '../../../store/cloudAccountStore/CloudAccountStore'
import useClusterStore from '../../../store/clusterStore/ClusterStore'
import useBucketStore from '../../../store/bucketStore/BucketStore'
import useUserStore from '../../../store/userStore/UserStore'
import useSuperComputerStore from '../../../store/superComputer/SuperComputerStore'
import useLoadBalancerStore from '../../../store/loadBalancerStore/LoadBalancerStore'

export interface ChartSection {
  title: string
  charts: any
}

const ChartsWidgetContainer = (): JSX.Element => {
  const computeSectionBase = {
    title: 'Compute',
    charts: {
      instances: {
        title: 'Instances',
        total: 0,
        running: 0,
        runningDefinition: 'ready',
        url: '/compute',
        launchUrl: '/compute/reserve'
      },
      instanceGroups: {
        title: 'Instance groups',
        total: 0,
        running: 0,
        runningDefinition: 'ready',
        url: '/compute-groups',
        launchUrl: '/compute-groups/reserve'
      },
      loadBalancers: {
        title: 'Load balancers',
        total: 0,
        running: 0,
        runningDefinition: 'ready',
        url: '/load-balancer',
        launchUrl: '/load-balancer/reserve'
      }
    }
  }

  const storageSectionBase = {
    title: 'Storage',
    charts: {
      buckets: {
        title: 'Buckets',
        total: 0,
        running: 0,
        runningDefinition: 'ready',
        url: '/buckets',
        launchUrl: '/buckets/reserve'
      },
      volumes: {
        title: 'Volumes',
        total: 0,
        running: 0,
        runningDefinition: 'ready',
        url: '/storage',
        launchUrl: '/storage/reserve'
      }
    }
  }

  const kubernetesSectionBase = {
    title: 'Intel Kubernetes',
    charts: {
      clusters: {
        title: 'Clusters',
        total: 0,
        running: 0,
        runningDefinition: 'active',
        url: '/cluster',
        launchUrl: '/cluster/reserve'
      }
    }
  }

  const superComputerSectionBase = {
    title: 'Supercomputer',
    charts: {
      clusters: {
        title: 'Clusters',
        total: 0,
        running: 0,
        runningDefinition: 'active',
        url: '/supercomputer',
        launchUrl: '/supercomputer/launch'
      }
    }
  }

  const isDarkMode = useDarkModeStore((state) => state.isDarkMode)
  const [computeSection, setComputeSection] = useState<ChartSection>(computeSectionBase)
  const [storageSection, setStorageSection] = useState<ChartSection>(storageSectionBase)
  const [kubernetesSection, setKubernetesSection] = useState<ChartSection>(kubernetesSectionBase)
  const [superComputerSection, setSuperComputerSection] = useState<ChartSection>(superComputerSectionBase)
  const instances = useCloudAccountStore((state) => state.instances)
  const setInstances = useCloudAccountStore((state) => state.setInstances)
  const instanceGroups = useCloudAccountStore((state) => state.instanceGroups)
  const setInstanceGroups = useCloudAccountStore((state) => state.setInstanceGroups)
  const storages = useCloudAccountStore((state) => state.storages)
  const setStorages = useCloudAccountStore((state) => state.setStorages)
  const objectStorages = useBucketStore((state) => state.objectStorages)
  const setObjectStorages = useBucketStore((state) => state.setObjectStorages)
  const iksClusters = useClusterStore((state) => state.clustersData)
  const setIksClusters = useClusterStore((state) => state.setClustersData)
  const isOwnCloudAccount = useUserStore((state) => state.isOwnCloudAccount)
  const superClusters = useSuperComputerStore((state) => state.clusters)
  const setSuperClusters = useSuperComputerStore((state) => state.setClusters)
  const setLoadBalancers = useLoadBalancerStore((state) => state.setLoadBalancers)
  const loadBalancers = useLoadBalancerStore((state) => state.loadBalancers)

  const fetchInstances = async (): Promise<void> => {
    try {
      await setInstances(false)
    } catch {
      setComputeSection((state) => ({
        title: state.title,
        charts: {
          ...state.charts,
          instances: {
            ...state.charts.instances,
            hasError: true
          }
        }
      }))
    }
  }

  const fetchInstanceGroups = async (): Promise<void> => {
    try {
      await setInstanceGroups(false)
    } catch {
      setComputeSection((state) => ({
        title: state.title,
        charts: {
          ...state.charts,
          instanceGroups: {
            ...state.charts.instanceGroups,
            hasError: true
          }
        }
      }))
    }
  }

  const fetchStorages = async (): Promise<void> => {
    try {
      await setStorages(false)
    } catch {
      setStorageSection((state) => ({
        title: state.title,
        charts: {
          ...state.charts,
          volumes: {
            ...state.charts.volumes,
            hasError: true
          }
        }
      }))
    }
  }

  const fetchBuckets = async (): Promise<void> => {
    try {
      await setObjectStorages(false)
    } catch {
      setStorageSection((state) => ({
        title: state.title,
        charts: {
          ...state.charts,
          buckets: {
            ...state.charts.buckets,
            hasError: true
          }
        }
      }))
    }
  }

  const fetchIksClusters = async (): Promise<void> => {
    try {
      await setIksClusters(false)
    } catch {
      setKubernetesSection((state) => ({
        title: state.title,
        charts: {
          ...state.charts,
          clusters: {
            ...state.charts.clusters,
            hasError: true
          }
        }
      }))
    }
  }

  const fetchsuperComputerClusters = async (): Promise<void> => {
    try {
      await setSuperClusters(false)
    } catch {
      setSuperComputerSection((state) => ({
        title: state.title,
        charts: {
          ...state.charts,
          instances: {
            ...state.charts.instances,
            hasError: true
          }
        }
      }))
    }
  }

  const fetchLoadBalancers = async (): Promise<void> => {
    try {
      await setLoadBalancers(false)
    } catch {
      setComputeSection((state) => ({
        title: state.title,
        charts: {
          ...state.charts,
          loadBalancers: {
            ...state.charts.loa,
            hasError: true
          }
        }
      }))
    }
  }

  useEffect(() => {
    const fetch = async (): Promise<void> => {
      const promiseArray = []
      if (instances.length === 0) promiseArray.push(fetchInstances())
      if (instanceGroups.length === 0) promiseArray.push(fetchInstanceGroups())
      if (storages.length === 0) promiseArray.push(fetchStorages())
      if (objectStorages.length === 0) promiseArray.push(fetchBuckets())
      if (iksClusters?.length === 0) promiseArray.push(fetchIksClusters())
      if (superClusters?.length === 0) promiseArray.push(fetchsuperComputerClusters())
      if (loadBalancers?.length === 0) promiseArray.push(fetchLoadBalancers())
      await Promise.allSettled(promiseArray)
    }
    void fetch()
  }, [])

  useEffect(() => {
    setComputeSection((state) => ({
      title: state.title,
      charts: {
        ...state.charts,
        instances: {
          ...state.charts.instances,
          hasError: false,
          total: instances.length,
          running: countProperty(instances, 'status', 'Ready')
        }
      }
    }))
  }, [instances])

  useEffect(() => {
    setComputeSection((state) => ({
      title: state.title,
      charts: {
        ...state.charts,
        instanceGroups: {
          ...state.charts.instanceGroups,
          hasError: false,
          total: instanceGroups.length,
          running: instanceGroups.filter((x) => x.instanceCount === x.readyCount).length
        }
      }
    }))
  }, [instanceGroups])

  useEffect(() => {
    setStorageSection((state) => ({
      title: state.title,
      charts: {
        ...state.charts,
        volumes: {
          ...state.charts.volumes,
          hasError: false,
          total: storages.length,
          running: countProperty(storages, 'status', 'Ready')
        }
      }
    }))
  }, [storages])

  useEffect(() => {
    setStorageSection((state) => ({
      title: state.title,
      charts: {
        ...state.charts,
        buckets: {
          ...state.charts.buckets,
          hasError: false,
          total: objectStorages.length,
          running: countProperty(objectStorages, 'status', 'Ready')
        }
      }
    }))
  }, [objectStorages])

  useEffect(() => {
    setKubernetesSection((state) => ({
      title: state.title,
      charts: {
        ...state.charts,
        clusters: {
          ...state.charts.clusters,
          hasError: false,
          total: iksClusters?.length,
          running: countProperty(iksClusters ?? [], 'clusterstate', 'Active')
        }
      }
    }))
  }, [iksClusters])

  useEffect(() => {
    setSuperComputerSection((state) => ({
      title: state.title,
      charts: {
        ...state.charts,
        clusters: {
          ...state.charts.clusters,
          hasError: false,
          total: superClusters?.length,
          running: countProperty(superClusters ?? [], 'clusterstate', 'Active')
        }
      }
    }))
  }, [superClusters])

  useEffect(() => {
    setComputeSection((state) => ({
      title: state.title,
      charts: {
        ...state.charts,
        loadBalancers: {
          ...state.charts.loadBalancers,
          hasError: false,
          total: loadBalancers.length,
          running: countProperty(loadBalancers ?? [], 'status', 'Active')
        }
      }
    }))
  }, [loadBalancers])

  const countProperty = (array: any[], property: string, value: string): number => {
    return array.reduce((acc, obj) => (obj[property] === value ? ++acc : acc), 0)
  }

  return (
    <ChartsWidget
      isDarkMode={isDarkMode}
      isOwnCloudAccount={isOwnCloudAccount}
      computeSection={computeSection}
      storageSection={storageSection}
      kubernetesSection={kubernetesSection}
      superComputeSection={superComputerSection}
    />
  )
}

export default ChartsWidgetContainer
