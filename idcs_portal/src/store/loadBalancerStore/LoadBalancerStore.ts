// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import LoadBalancerService from '../../services/LoadBalancerService'
import type Product from '../models/Product/Product'
import moment from 'moment'
import PublicService from '../../services/PublicService'

export interface LoadBalancerReservation {
  name: string
  cloudAccountId: string
  resourceId: string
  sourceips: string[] | []
  listeners: Listeners[] | []
  status: string
  vip: string
  firewallRuleCreated: boolean
  creationTimestamp: string
  message: string
}

export interface Listeners {
  name: string
  vipID: number
  message: string
  poolID: number
  externalPort: string
  internalPort: string
  monitor: string
  loadBalancingMode: string
  poolCreated: boolean
  vipPoolLinked: boolean
  instanceSelector: string
  instanceSelectors: string[] | instanceSelectors
  poolMembers: PoolMembers[] | []
}

interface PoolMembers {
  instanceRef: string
  ip: string
}

export type instanceSelectors = Record<string, string>

interface NetworkProduct extends Product {
  maxListeners: string
  maxSourceIps: string
}

export interface LoadBalancer {
  loading: boolean
  productLoading: boolean
  loadBalancers: LoadBalancerReservation[] | []
  setLoadBalancers: (isBackground: boolean) => Promise<void>
  shouldRefreshLoadBalancers: boolean
  setShouldRefreshLoadBalancers: (value: boolean) => void
  loadBalancerActiveTab: number
  setLoadBalancerActiveTab: (tabNumber: number) => void
  currentSelectedBalancer: LoadBalancerReservation | null
  setCurrentSelectedBalancer: (balancer: LoadBalancerReservation | null) => void
  networkProducts: NetworkProduct[] | []
  setNetworkProducts: () => Promise<void>
  reset: () => void
}

const initialState = {
  loadBalancers: [],
  loading: false,
  productLoading: true,
  shouldRefreshLoadBalancers: false,
  loadBalancerActiveTab: 0,
  currentSelectedBalancer: null,
  networkProducts: []
}

const useLoadBalancerStore = create<LoadBalancer>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setLoadBalancers: async (isBackground = false) => {
    try {
      if (!isBackground) {
        set({ loading: true })
      }

      const loadBalancerResponse = await LoadBalancerService.getLoadBalancers()
      const loadBalancers: LoadBalancerReservation[] = []

      for (const balancerItem of loadBalancerResponse.data.items) {
        loadBalancers.push(buildLoadBalancer(balancerItem))
      }
      loadBalancers.sort((p1, p2) => (p1.name < p2.name ? 1 : p1.name > p2.name ? -1 : 0))
      set({ loading: false, loadBalancers })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setShouldRefreshLoadBalancers: (value: boolean) => {
    set({ shouldRefreshLoadBalancers: value })
  },
  setLoadBalancerActiveTab: (tabNumber: number) => {
    set({ loadBalancerActiveTab: tabNumber })
  },
  setCurrentSelectedBalancer: (balancer: LoadBalancerReservation | null) => {
    set({ currentSelectedBalancer: balancer })
  },

  setNetworkProducts: async () => {
    set({ productLoading: true })

    try {
      const response = await PublicService.getNetworkCatalog()
      const productDetails = { ...response.data.products }

      const networkProducts: NetworkProduct[] = []

      for (const index in productDetails) {
        const productDetail = {
          ...productDetails[index]
        }

        const metadata = {
          ...productDetail.metadata
        }

        const rates = productDetail.rates
        let accountType = null
        let unit = null
        let rateValue = null
        let usageExpr = null

        if (rates.length > 0) {
          const rate = rates[0]

          if (rate) {
            accountType = filterValue(rate, 'accountType')
            unit = filterValue(rate, 'unit')
            rateValue = filterValue(rate, 'rate')
            usageExpr = filterValue(rate, 'usageExpr')
          }
        }

        const product: NetworkProduct = {
          name: filterValue(productDetail, 'name'),
          id: filterValue(productDetail, 'id'),
          created: moment(filterValue(productDetail, 'created')).format('MM/DD/YYYY h:mm a'),
          vendorId: filterValue(productDetail, 'vendorId'),
          familyId: filterValue(productDetail, 'familyId'),
          description: filterValue(productDetail, 'description'),
          category: filterValue(metadata, 'category'),
          recommendedUseCase: filterValue(metadata, 'recommendedUseCase'),
          cpuSockets: filterValue(metadata, 'cpu.sockets'),
          cpuCores: filterValue(metadata, 'cpu.cores'),
          diskSize: filterValue(metadata, 'disks.size'),
          displayName: filterValue(metadata, 'displayName'),
          familyDisplayDescription: filterValue(metadata, 'family.displayDescription'),
          releaseStatus: filterValue(metadata, 'releaseStatus'),
          familyDisplayName: filterValue(metadata, 'family.displayName'),
          information: filterValue(metadata, 'information'),
          instanceType: filterValue(metadata, 'instanceType'),
          instanceCategories: filterValue(metadata, 'instanceCategories'),
          memorySize: filterValue(metadata, 'memory.size'),
          nodesCount: filterValue(metadata, 'nodesCount'),
          processor: filterValue(metadata, 'processor'),
          region: filterValue(metadata, 'region'),
          service: filterValue(metadata, 'service'),
          eccn: filterValue(productDetail, 'eccn'),
          pcq: filterValue(productDetail, 'pcq'),
          accountType,
          unit,
          rate: rateValue,
          usageExpr,
          imageSource: filterValue(metadata, 'family.displayName'),
          maxListeners: filterValue(metadata, 'listener.count.max'),
          maxSourceIps: filterValue(metadata, 'sourceIp.count.max')
        }

        networkProducts.push(product)
      }

      set({ productLoading: false })
      set({ networkProducts })
    } catch (error) {
      set({ productLoading: false })
      throw error
    }
  }
}))

const needToRefreshLoadBalancers = (): boolean => {
  const shouldRefreshLoadBalancers = useLoadBalancerStore.getState().shouldRefreshLoadBalancers
  return shouldRefreshLoadBalancers
}

setInterval(() => {
  if (needToRefreshLoadBalancers()) {
    void (async () => {
      await useLoadBalancerStore.getState().setLoadBalancers(true)
    })()
  }
}, 5000)

const buildLoadBalancer = (loadBalancer: any): LoadBalancerReservation => {
  const getAllSecurityIps = (security: any): string[] => {
    return security?.sourceips && security.sourceips.length > 0 ? security.sourceips : []
  }

  const { metadata, spec, status } = loadBalancer
  const { listeners, security } = spec

  return {
    name: metadata.name,
    cloudAccountId: metadata.cloudAccountId,
    resourceId: metadata.resourceId,
    sourceips: getAllSecurityIps(security),
    listeners: buildListeners(listeners, status),

    status: status?.state ? status.state : '',
    vip: status?.vip ? status.vip : '',
    firewallRuleCreated: !!(status?.conditions?.firewallRuleCreated && status.conditions.firewallRuleCreated === true),

    creationTimestamp: metadata.creationTimestamp,
    message: status?.message ? status.message : ''
  }
}

const buildListeners = (listeners: any, status: any): any => {
  let listenersResult: Listeners[] = []

  const listenersCondition = status?.conditions?.listeners
  const statusListeners = status?.listeners

  if (listeners.length > 0) {
    const getFromCondition = (port: string): any => {
      let response = {
        poolCreated: false,
        vipPoolLinked: false
      }
      if (listenersCondition && listenersCondition.length > 0) {
        const listener = listenersCondition.find((x: any) => x.port === port)
        response = {
          poolCreated: listener?.poolCreated === 'true',
          vipPoolLinked: listener?.vipPoolLinked === 'true'
        }
      }

      return response
    }

    const getFromStatus = (port: string): any => {
      let response = {
        name: '',
        vipID: '',
        message: '',
        poolID: '',
        poolMembers: []
      }
      if (statusListeners.length > 0) {
        const listener = statusListeners.find((x: any) => x.name.split('-').pop() === String(port))
        response = {
          name: listener?.name,
          vipID: listener?.vipID,
          message: listener?.message,
          poolID: listener?.poolID,
          poolMembers: listener?.poolMembers || []
        }
      }
      return response
    }

    const getInstanceSelectors = (selector: string, pool: any): any => {
      if (selector === 'labels') {
        return pool.instanceSelectors
      } else {
        return pool.instanceResourceIds
      }
    }

    listenersResult = [
      ...listeners.map((listener: any): Listeners => {
        const { name, vipID, message, poolID, poolMembers } = getFromStatus(listener.port)
        const { poolCreated, vipPoolLinked } = getFromCondition(listener.port)
        const { pool } = listener

        let instanceSelector = 'labels'

        if (pool.instanceResourceIds.length > 0) {
          instanceSelector = 'instances'
        }

        return {
          name,
          vipID,
          message,
          poolID,

          externalPort: listener.port,
          internalPort: pool.port,
          monitor: pool.monitor,
          loadBalancingMode: pool.loadBalancingMode,

          poolCreated,
          vipPoolLinked,
          instanceSelector,
          instanceSelectors: getInstanceSelectors(instanceSelector, pool),
          poolMembers
        }
      })
    ]
  }

  return listenersResult
}

const filterValue = (object: any, property: any): any => {
  return object[property] ? object[property] : ''
}

export default useLoadBalancerStore
