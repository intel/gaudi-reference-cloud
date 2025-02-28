// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import {
  type SuperCluster,
  type NodeGroup,
  type InstanceTypes,
  type runtimes,
  type SuperClusterResourceLimit,
  type SecurityRulesTypes
} from '../models/superComputer/SuperComputer'
import SuperComputerService from '../../services/SuperComputerService'
import type Product from '../models/Product/Product'
import type { Storage } from '../models/Storage/Storage'
import PublicService from '../../services/PublicService'
import moment from 'moment'
import { superComputerProductCatalogTypes, iksClusterTypes } from '../../utils/Enums'
import useHardwareStore from '../hardwareStore/HardwareStore'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'

interface SuperComputerStore {
  clusters: SuperCluster[] | null
  clusterResourceLimit: SuperClusterResourceLimit | null
  setClusters: (isBackGround: boolean) => Promise<void>
  clusterDetail: SuperCluster | null
  setClusterDetail: (idCluster: string, isBackGround: boolean) => Promise<void>
  shouldRefreshClusterDetail: boolean
  setShouldRefreshClusterDetail: (value: boolean) => void
  debounceDetailRefresh: boolean
  setDebounceDetailRefresh: (value: boolean) => void
  nodeTabNumber: number | null
  setNodeTabNumber: (tabNumber: number | null) => void
  shouldRefreshClusters: boolean
  setShouldRefreshClusters: (value: boolean) => void
  kaasInstanceTypes: InstanceTypes[] | null
  setKaasInstanceTypes: () => Promise<void>
  clustersRuntimes: runtimes[] | []
  setClustersRuntimes: () => Promise<void>
  coreComputeFamilies: Product[] | []
  coreComputeProducts: Product[] | []
  scProducts: Product[] | []
  scControlPlane: Product[] | []
  aiFamilies: Product[] | []
  isWhitelisted: boolean
  isGeneralComputeAvailable: boolean
  aiProducts: Product[] | []
  fileStorage: Storage[] | []
  setProducts: () => Promise<void>
  reset: () => void
  loading: boolean
  loadingDetail: boolean
  editSecurityRule: SecurityRulesTypes | null
  setEditSecurityRule: (rule: SecurityRulesTypes) => void
}

const initialState = {
  clusters: [],
  clusterResourceLimit: null,
  clusterDetail: null,
  nodeTabNumber: 0,
  shouldRefreshClusterDetail: false,
  shouldRefreshClusters: false,
  debounceDetailRefresh: false,
  kaasInstanceTypes: [],
  clustersRuntimes: [],
  coreComputeFamilies: [],
  coreComputeProducts: [],
  scControlPlane: [],
  isWhitelisted: true,
  isGeneralComputeAvailable: false,
  editSecurityRule: null,
  aiFamilies: [],
  aiProducts: [],
  scProducts: [],
  fileStorage: [],
  loading: false,
  loadingDetail: false
}

const parseState = (state: string, status: any): string => {
  const { errorcode } = status
  if (!errorcode) {
    return state
  }
  return 'Error'
}

const createClustersJson = async (rows: any): Promise<SuperCluster[]> => {
  const clustersData: SuperCluster[] = []
  const socketResponseJson = rows
  for (let i = 0; i < socketResponseJson.length; i++) {
    const clustertype = socketResponseJson[i].clustertype
    if (clustertype !== iksClusterTypes.superCluster) {
      continue
    }
    let securityRules: SecurityRulesTypes[] = []
    const annotations = socketResponseJson[i].annotations
    const clusterstate = parseState(socketResponseJson[i].clusterstate, socketResponseJson[i].clusterstatus)
    const clusterstatus = socketResponseJson[i].clusterstatus
    const createddate = socketResponseJson[i].createddate
    const description = socketResponseJson[i].description
    const k8sversion = socketResponseJson[i].k8sversion
    const name = socketResponseJson[i].name
    const network = socketResponseJson[i].network
    const nodegroups = socketResponseJson[i].nodegroups
    const provisioningLog = socketResponseJson[i].provisioningLog
    const tags = socketResponseJson[i].tags
    const upgradeavailable = socketResponseJson[i].upgradeavailable
    const upgradek8sversionavailable = socketResponseJson[i].upgradek8sversionavailable
    const uuid = socketResponseJson[i].uuid
    if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_SC_SECURITY)) {
      const response = await SuperComputerService.getSecurityRules(uuid)
      securityRules = createSecurityRulesArray(response.data.getfirewallresponse)
    }
    const vips = socketResponseJson[i].vips
    const kubeconfig = socketResponseJson[i].kubeconfig
    const storages = socketResponseJson[i].storages ? socketResponseJson[i].storages : []
    const clusterData: SuperCluster = {
      name,
      annotations,
      clusterstate,
      clusterstatus,
      clustertype,
      createddate,
      description,
      k8sversion,
      network,
      nodegroups,
      provisioningLog,
      tags,
      upgradeavailable,
      upgradek8sversionavailable,
      uuid,
      vips,
      kubeconfig,
      storages,
      securityRules
    }

    clusterData.nodegroups.forEach((nodeGroup) => {
      nodeGroup.nodegroupstate = parseState(nodeGroup.nodegroupstate, nodeGroup.nodegroupstatus)
    })

    clusterData.vips.forEach((vip) => {
      vip.vipstate = parseState(vip.vipstate, vip.vipstatus)
    })

    clustersData.push(clusterData)
  }
  return clustersData
}

const createSecurityRulesArray = (rows: any): SecurityRulesTypes[] => {
  const securityRules: SecurityRulesTypes[] = []

  const responseJson = rows

  for (let i = 0; i < responseJson.length; i++) {
    const vipname = responseJson[i].vipname
    const destinationip = responseJson[i].destinationip
    const state = responseJson[i].state
    const sourceip = responseJson[i].sourceip
    const protocol = responseJson[i].protocol
    const port = responseJson[i].port
    const vipid = responseJson[i].vipid
    const viptype = responseJson[i].viptype

    const securityRule: SecurityRulesTypes = {
      vipname,
      destinationip,
      port,
      protocol,
      sourceip,
      state,
      vipid,
      viptype
    }
    securityRules.push(securityRule)
  }
  return securityRules
}

const buildComputeData = (productDetail: any): Product => {
  const metadata = {
    ...productDetail.metadata
  }
  const rates = productDetail.rates
  let accountType = ''
  let unit = ''
  let rateValue = ''
  let usageExpr = ''

  if (rates.length > 0) {
    const rate = rates[0]

    if (rate) {
      accountType = rate.accountType
      unit = rate.unit
      rateValue = rate.rate
      usageExpr = rate.usageExpr
    }
  }

  const product: Product = {
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
    imageSource: ''
  }

  return product
}

const buildStorageData = (response: any): Storage => {
  const metadata = {
    ...response.metadata
  }

  const rates = response.rates
  let accountType = ''
  let unit = ''
  let rateValue = ''
  let usageExpr = ''

  if (rates.length > 0) {
    const rate = rates[0]

    if (rate) {
      accountType = rate.accountType
      unit = rate.unit
      rateValue = rate.rate
      usageExpr = rate.usageExpr
    }
  }

  const storage: Storage = {
    name: response.name,
    id: response.id,
    created: moment(response.created).format('MM/DD/YYYY h:mm a'),
    vendorId: response.vendorId,
    familyId: response.familyId,
    description: response.description,
    storageCategories: metadata.instanceCategories,
    access: metadata.access,
    billingEnable: metadata.billingEnable,
    category: metadata.instanceCategories,
    disableForAccountTypes: metadata.disableForAccountTypes,
    minimumSize: Number(metadata['volume.size.min']),
    maximumSize: Number(metadata['volume.size.max']),
    displayName: metadata.displayName,
    familyDisplayDescription: metadata['family.displayDescription'],
    familyDisplayName: metadata['family.displayName'],
    highlight: metadata.highlight,
    information: metadata.information,
    recommendedUseCase: metadata.recommendedUseCase,
    region: metadata.region,
    releaseStatus: metadata.releaseStatus,
    service: metadata.service,
    eccn: response.eccn,
    pcq: response.pcq,
    matchExpr: response.matchExpr,
    accountType,
    unit,
    rate: rateValue ? Number(rateValue) : 0,
    usageExpr,
    status: response.status,
    unitSize: metadata['volume.size.unit'],
    usageUnit: metadata['usage.unit']
  }

  return storage
}

const createNodegroupArray = async (rows: any): Promise<NodeGroup[]> => {
  const clustersNodegroups: NodeGroup[] = []
  const responseJson = rows

  for (let i = 0; i < responseJson.length; i++) {
    const annotations = responseJson[i].annotations
    const clusteruuid = responseJson[i].clusteruuid
    const count = responseJson[i].count
    const createddate = responseJson[i].createddate
    const description = responseJson[i].description
    const imiid = responseJson[i].imiid
    const instancetypeid = responseJson[i].instancetypeid
    const name = responseJson[i].name
    const networkinterfacename = responseJson[i].networkinterfacename
    const nodegroupstate = parseState(responseJson[i].nodegroupstate, responseJson[i].nodegroupstatus)
    const nodegroupstatus = responseJson[i].nodegroupstatus
    const nodegroupuuid = responseJson[i].nodegroupuuid
    const nodes = responseJson[i].nodes
    const sshkeyname = responseJson[i].sshkeyname
    const tags = responseJson[i].tags
    const upgradeavailable = responseJson[i].upgradeavailable
    const upgradeimiid = responseJson[i].upgradeimiid
    const upgradestrategy = responseJson[i].upgradestrategy
    const userdataurl = responseJson[i].userdataurl ? responseJson[i].userdataurl : ''
    const vnets = responseJson[i].vnets
    const nodeGroupType = responseJson[i].nodegrouptype
    const nodeGroupData: NodeGroup = {
      annotations,
      clusteruuid,
      count,
      createddate,
      description,
      imiid,
      instancetypeid,
      name,
      networkinterfacename,
      nodegroupstate,
      nodegroupstatus,
      nodegroupuuid,
      nodes,
      sshkeyname,
      tags,
      upgradeavailable,
      upgradeimiid,
      upgradestrategy,
      userdataurl,
      vnets,
      instanceTypeDetails: null,
      nodeGroupType
    }

    nodeGroupData.nodes.forEach((node) => {
      node.state = parseState(node.state, node)
    })

    const products = useHardwareStore.getState().products
    if (products.length === 0) {
      await useHardwareStore.getState().setProducts()
    }
    const instanceType = useSuperComputerStore
      .getState()
      .scProducts.find((x) => x.name === nodeGroupData.instancetypeid)
    nodeGroupData.instanceTypeDetails = instanceType ?? null

    clustersNodegroups.push(nodeGroupData)
  }
  return clustersNodegroups
}

const createInstanceTypesJson = (rows: any): InstanceTypes[] => {
  const clustersInstanceTypes: InstanceTypes[] = []
  const responseJson = rows

  for (let i = 0; i < responseJson.length; i++) {
    const cpu = responseJson[i].cpu
    const instancetypename = responseJson[i].instancetypename
    const memory = responseJson[i].memory
    const storage = responseJson[i].storage

    const instanceTypesData: InstanceTypes = {
      cpu,
      instancetypename,
      memory,
      storage
    }
    clustersInstanceTypes.push(instanceTypesData)
  }
  return clustersInstanceTypes
}

const createRuntimeJson = (rows: any): runtimes[] => {
  const clustersRuntimes: runtimes[] = []
  const responseJson = rows
  for (let i = 0; i < responseJson.length; i++) {
    const k8sversions = responseJson[i].k8sversionname
    const runtimename = responseJson[i].runtimename

    for (let indexVersion = 0; indexVersion < k8sversions.length; indexVersion++) {
      const clusterData: runtimes = {
        k8sversionname: k8sversions[indexVersion],
        runtimename
      }
      clustersRuntimes.push(clusterData)
    }
  }
  return clustersRuntimes
}

const filterValue = (object: any, property: any): any => {
  return object[property] ? object[property] : ''
}

const allowedStates = ['Pending', 'Deleting', 'DeletePending', 'Updating', 'Reconciling']

const needToRefreshNodeGroupNodes = (nodeGroup: NodeGroup): boolean => {
  if (!nodeGroup.nodes || nodeGroup.nodes.length === 0) {
    return false
  }
  return nodeGroup.nodes.some((x) => allowedStates.includes(x.state))
}

const needToRefreshNodeGroup = (cluster: SuperCluster): boolean => {
  if (!cluster.nodegroups || cluster.nodegroups.length === 0) {
    return false
  }
  return cluster.nodegroups.some((x) => allowedStates.includes(x.nodegroupstate) || needToRefreshNodeGroupNodes(x))
}

const needToRefreshLoadBalancer = (cluster: SuperCluster): boolean => {
  if (!cluster.vips || cluster.vips.length === 0) {
    return false
  }
  return cluster.vips.some((x) => allowedStates.includes(x.vipstate))
}

const needToRefreshStorage = (cluster: SuperCluster): boolean => {
  if (!cluster.storages || cluster.storages.length === 0) {
    return false
  }
  return cluster.storages.some((x) => allowedStates.includes(x.state))
}

const needToRefreshSecurityRules = (cluster: SuperCluster): boolean => {
  if (!cluster.securityRules || cluster.securityRules.length === 0) {
    return false
  }
  return cluster.securityRules.some((x) => allowedStates.includes(x.state))
}

const needToRefreshClusters = (): boolean => {
  const clusters = useSuperComputerStore.getState().clusters
  const shouldRefreshClusters = useSuperComputerStore.getState().shouldRefreshClusters
  if (shouldRefreshClusters && clusters) {
    return clusters.some(
      (c: SuperCluster) =>
        allowedStates.includes(c.clusterstate) ||
        needToRefreshNodeGroup(c) ||
        needToRefreshLoadBalancer(c) ||
        needToRefreshStorage(c) ||
        needToRefreshSecurityRules(c)
    )
  }
  return false
}

const needToRefreshClusterDetail = (): boolean => {
  const clusterDetail = useSuperComputerStore.getState().clusterDetail
  const shouldRefreshClusterDetail = useSuperComputerStore.getState().shouldRefreshClusterDetail
  if (shouldRefreshClusterDetail && clusterDetail) {
    return (
      allowedStates.includes(clusterDetail.clusterstate) ||
      needToRefreshNodeGroup(clusterDetail) ||
      needToRefreshLoadBalancer(clusterDetail) ||
      needToRefreshStorage(clusterDetail)
    )
  }
  return false
}

const sortObjectsByKey = (arr: any, key: string, direction: string): any => {
  if (direction === 'asc') {
    return arr.sort((a: any, b: any) => parseFloat(a[key]) - parseFloat(b[key]))
  } else if (direction === 'desc') {
    return arr.sort((a: any, b: any) => parseFloat(b[key]) - parseFloat(a[key]))
  } else {
    throw new Error('Invalid direction. Use "asc" or "desc".')
  }
}

setInterval(() => {
  if (needToRefreshClusters()) {
    void (async () => {
      await useSuperComputerStore.getState().setClusters(true)
    })()
  }
  if (needToRefreshClusterDetail()) {
    void (async () => {
      const idCluster = useSuperComputerStore.getState().clusterDetail?.name
      if (idCluster) {
        await useSuperComputerStore.getState().setClusterDetail(idCluster, true)
      }
    })()
  }
}, 10000)

const useSuperComputerStore = create<SuperComputerStore>()((set, get) => ({
  ...initialState,
  setClusters: async (isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }
    let superClusters: SuperCluster[] = []
    const response = await SuperComputerService.getClusters()
    const clusterInfo = response.data.clusters
    const clusterResourceLimit = response.data?.resourcelimits || null

    if (clusterInfo) {
      superClusters = await createClustersJson(clusterInfo)

      set({ clusters: superClusters })
    }
    set({ loading: false, clusterResourceLimit })
  },
  setProducts: async () => {
    try {
      set({ loading: true })

      // Iks Supported Types
      let instanceTypesResponse: InstanceTypes[] = []
      const instanceTypes = await SuperComputerService.getInstanceTypes()
      if (instanceTypes.data.instancetypes) {
        instanceTypesResponse = createInstanceTypesJson(instanceTypes.data.instancetypes)
      }

      // Product Catalog items
      const response = await PublicService.getSuperComputerCatalog()
      const productDetails = [...response.data.products]

      if (productDetails.length > 0) {
        const coreComputeDetail = productDetails.filter(
          (product) =>
            product.metadata.recommendedUseCase === superComputerProductCatalogTypes.coreCompute &&
            instanceTypesResponse.find((instance) => instance.instancetypename === product.name)
        )

        const aiDetail = productDetails.filter(
          (product) =>
            product.metadata.recommendedUseCase === superComputerProductCatalogTypes.aiCompute &&
            instanceTypesResponse.find((instance) => instance.instancetypename === product.name)
        )

        const storageDetail = productDetails.filter(
          (product) => product.metadata.recommendedUseCase === superComputerProductCatalogTypes.fileStorage
        )

        const controlPlaneDetail = productDetails.filter(
          (product) => product.metadata.instanceType === superComputerProductCatalogTypes.controlPlane
        )

        const coreItems: Product[] = []
        const scProducts: Product[] = []
        const aiItems: Product[] = []
        const stoageItems: Storage[] = []
        const controlPlaneItems: Product[] = []
        productDetails.forEach((item) => {
          scProducts.push(buildComputeData(item))
        })

        coreComputeDetail.forEach((item) => {
          coreItems.push(buildComputeData(item))
        })
        aiDetail.forEach((item) => {
          aiItems.push(buildComputeData(item))
        })
        storageDetail.forEach((item) => {
          stoageItems.push(buildStorageData(item))
        })
        controlPlaneDetail.forEach((item) => {
          controlPlaneItems.push(buildComputeData(item))
        })

        const aifamilies = aiItems.reduce((acc: Product[], current: Product) => {
          if (!acc.find((item: Product) => item.name === current.name)) {
            acc.push(current)
          }
          return acc
        }, [])
        aifamilies.sort((p1, p2) => Number(p1.nodesCount) - Number(p2.nodesCount))

        const coreComputefamilies = coreItems.reduce((acc: Product[], current: Product) => {
          if (!acc.find((item: Product) => item.familyDisplayName === current.familyDisplayName)) {
            acc.push(current)
          }
          return acc
        }, [])

        if (aiItems.length === 0 || controlPlaneItems.length === 0 || stoageItems.length === 0) {
          set({ isWhitelisted: false })
        } else {
          set({ isWhitelisted: true })
        }
        if (coreItems.length > 0) {
          set({ isGeneralComputeAvailable: true })
        } else {
          set({ isGeneralComputeAvailable: false })
        }
        set({ loading: false })
        set({ scProducts })
        set({ aiFamilies: aifamilies })
        set({ coreComputeFamilies: coreComputefamilies })
        set({ coreComputeProducts: coreItems })
        set({ aiProducts: aiItems })
        set({ fileStorage: stoageItems })
        set({ scControlPlane: controlPlaneItems })
        set({ kaasInstanceTypes: instanceTypesResponse })
      } else {
        set({ isWhitelisted: false })
      }
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setClusterDetail: async (clusterName: string, isBackGround: boolean) => {
    try {
      if (!isBackGround) {
        set({ loadingDetail: true })
      }
      let clusters = get().clusters
      if (!clusters || clusters.length === 0 || isBackGround) {
        await get().setClusters(isBackGround)
      }
      clusters = get().clusters
      if (clusters !== null) {
        const clusterDetail = clusters.find((item) => item.name === clusterName)
        if (clusterDetail) {
          // Update nodes
          const nodeGroupResponse = await SuperComputerService.getAllManagedNodeGroupData(clusterDetail.uuid)

          if (nodeGroupResponse.data.nodegroups) {
            clusterDetail.nodegroups = await createNodegroupArray(nodeGroupResponse.data.nodegroups)
          }
          set({ clusterDetail })
        } else {
          set({ clusterDetail: null })
        }
      } else {
        set({ clusterDetail: null })
      }
      set({ loadingDetail: false })
    } catch (error) {
      set({ loadingDetail: false })
    }
  },
  setNodeTabNumber: (tabNumber: number | null) => {
    set({ nodeTabNumber: tabNumber })
  },
  setShouldRefreshClusters: (value: boolean) => {
    set({ shouldRefreshClusters: value })
  },
  setShouldRefreshClusterDetail: (value: boolean) => {
    set({ shouldRefreshClusterDetail: value })
  },
  setDebounceDetailRefresh: (value: boolean) => {
    set({ debounceDetailRefresh: value })
  },
  setKaasInstanceTypes: async () => {
    let instanceTypesResponse: InstanceTypes[] = []
    const instanceTypes = await SuperComputerService.getInstanceTypes()
    if (instanceTypes.data.instancetypes) {
      instanceTypesResponse = createInstanceTypesJson(instanceTypes.data.instancetypes)
      set({ kaasInstanceTypes: instanceTypesResponse })
    }
  },
  setClustersRuntimes: async () => {
    let runtimesResponse: runtimes[] = []
    const runtimes = await SuperComputerService.getRuntime()
    if (runtimes.data.runtimes) {
      runtimesResponse = sortObjectsByKey(createRuntimeJson(runtimes.data.runtimes), 'k8sversionname', 'desc')
      set({ clustersRuntimes: runtimesResponse })
    }
  },
  setEditSecurityRule: (rule: SecurityRulesTypes) => {
    set({ editSecurityRule: rule })
  },
  reset: () => {
    set(initialState)
  }
}))

export default useSuperComputerStore
