// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect } from 'react'
import HomePage from '../../components/homePage/HomePage'
import useUserStore from '../../store/userStore/UserStore'
import { useNavigate } from 'react-router'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import useLoadBalancerStore from '../../store/loadBalancerStore/LoadBalancerStore'
import useBucketStore from '../../store/bucketStore/BucketStore'
import useClusterStore from '../../store/clusterStore/ClusterStore'
import useUsagesReport from '../../store/billingStore/UsagesStore'
import { initialCardState } from './Home.types'

const HomePageContainer = (): JSX.Element => {
  // Store
  const isOwnCloudAccount = useUserStore((state) => state.isOwnCloudAccount)

  const instances = useCloudAccountStore((state) => state.instances)
  const setInstances = useCloudAccountStore((state) => state.setInstances)

  const instanceGroups = useCloudAccountStore((state) => state.instanceGroups)
  const setInstanceGroups = useCloudAccountStore((state) => state.setInstanceGroups)

  const storages = useCloudAccountStore((state) => state.storages)
  const setStorages = useCloudAccountStore((state) => state.setStorages)

  const objectStorages = useBucketStore((state) => state.objectStorages)
  const setObjectStorages = useBucketStore((state) => state.setObjectStorages)

  const loadBalancers = useLoadBalancerStore((state) => state.loadBalancers)
  const setLoadBalancers = useLoadBalancerStore((state) => state.setLoadBalancers)

  const iksClusters = useClusterStore((state) => state.clustersData)
  const setIksClusters = useClusterStore((state) => state.setClustersData)

  const totalAmount = useUsagesReport((state) => state.totalAmount)
  const setUsage = useUsagesReport((state) => state.setUsage)

  // navigate
  const navigate = useNavigate()

  // Hooks
  useEffect(() => {
    const currentPath = window.location.pathname
    if (currentPath === '/') navigate('/home')
  }, [window.location.pathname])

  useEffect(() => {
    const fetchUsage = async (): Promise<void> => {
      await setUsage()
    }

    if (!totalAmount && isOwnCloudAccount) {
      fetchUsage().catch(() => {})
    }
  }, [])

  useEffect(() => {
    const fetchInstances = async (): Promise<void> => {
      await setInstances(false)
    }

    const fetchInstanceGroups = async (): Promise<void> => {
      await setInstanceGroups(false)
    }

    const fetchStorages = async (): Promise<void> => {
      await setStorages(false)
    }

    const fetchBuckets = async (): Promise<void> => {
      await setObjectStorages(false)
    }

    const fetchLoadBalancers = async (): Promise<void> => {
      await setLoadBalancers(false)
    }

    const fetchIksClusters = async (): Promise<void> => {
      await setIksClusters(false)
    }

    const fetch = async (): Promise<void> => {
      if (!totalAmount) {
        const promiseArray = []
        if (instances?.length === 0) promiseArray.push(fetchInstances())
        if (instanceGroups?.length === 0) promiseArray.push(fetchInstanceGroups())
        if (storages?.length === 0) promiseArray.push(fetchStorages())
        if (objectStorages.length === 0) promiseArray.push(fetchBuckets())
        if (loadBalancers?.length === 0) promiseArray.push(fetchLoadBalancers())
        if (iksClusters?.length === 0) promiseArray.push(fetchIksClusters())
        await Promise.allSettled(promiseArray)
      }
    }
    fetch().catch(() => {})
  }, [totalAmount])

  // Functions

  const onRedirectTo = (route: string): void => {
    navigate(route)
  }

  return <HomePage cardState={initialCardState} onRedirectTo={onRedirectTo} />
}

export default HomePageContainer
