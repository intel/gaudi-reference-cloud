// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import ClusterService from '../../services/ClusterService'
import useHardwareStore from '../hardwareStore/HardwareStore'
import type Product from '../models/Product/Product'
import PublicService from '../../services/PublicService'
import moment from 'moment'
import { iksClusterTypes } from '../../utils/Enums'

interface Annotation {
  key: string
  value: string
}

interface Network {
  clustercidr: string
  clusterdns: string
  enableloadbalancer: boolean
  region: string
  servicecidr: string
}

interface Node {
  createddate: string
  dnsname: string
  imi: string
  ipaddress: string
  name: string
  state: string
  status: string
}

interface UpgradeStrategy {
  drainnodes: boolean
  maxunavailablepercentage: number
}

interface Vnet {
  availabilityzonename: string
  networkinterfacevnetname: string
}

interface ProvisioningLog {
  logentry: string
  loglevel: string
  logobject: string
  timestamp: string
}

interface Tag {
  key: string
  value: string
}

interface VipMember {
  ipaddresses: string[]
}

interface Vip {
  description: string
  dnsalias: string[]
  members: VipMember[]
  name: string
  poolport: number
  port: number
  vipIp: string
  vipid: number
  vipstate: string
  vipstatus: string
  viptype: string
}

interface NodeGroup {
  annotations: Annotation[]
  clusteruuid: string
  count: number
  createddate: string
  description: string
  imiid: string
  instancetypeid: string
  name: string
  networkinterfacename: string
  nodegroupstate: string
  nodegroupstatus: string
  nodegroupuuid: string
  nodes: Node[]
  sshkeyname: string[]
  tags: Tag[]
  upgradeavailable: boolean
  upgradeimiid: string[]
  upgradestrategy: UpgradeStrategy
  userdataurl: string
  vnets: Vnet[]
  instanceTypeDetails: Product | null
}

interface Cluster {
  annotations: Annotation[]
  clusterstate: string
  clusterstatus: string
  createddate: string
  description: string
  k8sversion: string
  name: string
  network: Network
  nodegroups: NodeGroup[]
  provisioningLog: ProvisioningLog[]
  tags: Tag[]
  upgradeavailable: boolean
  upgradek8sversionavailable: string[]
  uuid: string
  vips: Vip[]
  kubeconfig: string
  storages: Storage[]
  clustertype: string
}

interface Storage {
  storageprovider: string
  size: string
  state: string
  reason: string
  message: string
}

// Zustand state interface
export interface runtimes {
  runtimename: string
  k8sversionname: string[]
  // hosts: any
}
export interface InstanceTypes {
  cpu: number
  memory: number
  storage: number
  instancetypename: string[]
}

export interface SecurityRulesTypes {
  destinationip: string
  port: string
  protocol: string[]
  sourceip: string[]
  state: string
  vipid: number
  vipname: string
  viptype: string
}

export interface ClusterResourceLimit {
  maxclusterpercloudaccount: number
  maxnodegroupspercluster: number
  maxvipspercluster: number
}

export interface ClusterStore {
  clusterProducts: Product[] | []
  clustersData: Cluster[] | null
  clusterResourceLimit: ClusterResourceLimit | null
  setClusterProducts: () => Promise<void>
  setClustersData: (isBackGround: boolean) => Promise<void>
  currentSelectedCluster: string | null
  setCurrentSelectedCluster: (clusterName: string | null) => void
  clustersRuntimes: runtimes[] | []
  setClustersRuntimes: () => Promise<void>
  clusterNodegroups: NodeGroup[] | []
  setClusterNodegroups: (clusteruuid: string) => Promise<void>
  securityRuleUuid: string | null
  clusterSecurityRules: SecurityRulesTypes[] | null
  shouldRefreshSecurityRules: boolean
  setShouldRefreshSecurityRules: (value: boolean) => void
  setClusterSecurityRules: (clusteruuid: string) => Promise<void>
  editSecurityRule: SecurityRulesTypes | null
  setEditSecurityRule: (rule: SecurityRulesTypes) => void
  refreshRate: boolean
  setRefreshRate: (value: boolean) => void
  kaasInstanceTypes: InstanceTypes[] | null
  setKaasInstanceTypes: () => Promise<void>
  loading: boolean
  shouldRefreshClusters: boolean
  setShouldRefreshClusters: (value: boolean) => void
  reset: () => void
}

const initialState = {
  loading: false,
  clusterProducts: [],
  clustersData: [],
  clusterResourceLimit: null,
  currentSelectedCluster: null,
  clustersRuntimes: [],
  clusterNodegroups: [],
  clusterSecurityRules: [],
  refreshRate: false,
  kaasInstanceTypes: [],
  shouldRefreshClusters: false,
  shouldRefreshSecurityRules: false,
  securityRuleUuid: null,
  editSecurityRule: null
}

const useClusterStore = create<ClusterStore>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setClusterProducts: async () => {
    set({ loading: true })
    try {
      const response = await PublicService.getKubernetesCatalog()
      const productDetails = { ...response.data.products }

      const clusterProducts: Product[] = []

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

        clusterProducts.push(product)
      }
      set({ loading: false })
      set({ clusterProducts })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setCurrentSelectedCluster: (clusterName: string | null) => {
    set({ currentSelectedCluster: clusterName })
  },
  setClustersData: async (isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }
    let socketResponse: Cluster[] = []
    const socket = await ClusterService.getAllClustersDataStatus()

    const clusterInfo = socket.data.clusters
    const clusterResourceLimit = socket.data?.resourcelimits || null

    const allowedClusterTypes = ['', iksClusterTypes.iksCluster]

    if (clusterInfo) {
      socketResponse = createClustersJson(clusterInfo).filter((cluster) =>
        allowedClusterTypes.includes(cluster.clustertype)
      )
      set({ clustersData: socketResponse })
    }
    set({ loading: false, clusterResourceLimit })
  },
  setClustersRuntimes: async () => {
    let runtimesResponse: runtimes[] = []
    const runtimes = await ClusterService.getRuntime()
    if (runtimes.data.runtimes) {
      runtimesResponse = createRuntimeJson(runtimes.data.runtimes)
      set({ clustersRuntimes: runtimesResponse })
    }
  },
  setClusterNodegroups: async (clusteruuid: string) => {
    let nodegroupResponse: NodeGroup[] = []
    const nodegroups = await ClusterService.getAllManagedNodeGroupData(clusteruuid)
    if (nodegroups.data.nodegroups) {
      nodegroupResponse = await createNodegroupArray(nodegroups.data.nodegroups)
      set({ clusterNodegroups: nodegroupResponse })
    }
  },
  setRefreshRate: (value: boolean) => {
    set({ refreshRate: value })
  },
  setKaasInstanceTypes: async () => {
    let instanceTypesResponse: InstanceTypes[] = []
    const instanceTypes = await ClusterService.getInstanceTypes()
    if (instanceTypes.data.instancetypes) {
      instanceTypesResponse = createInstanceTypesJson(instanceTypes.data.instancetypes)
      set({ kaasInstanceTypes: instanceTypesResponse })
    }
  },
  setShouldRefreshClusters: (value: boolean) => {
    set({ shouldRefreshClusters: value })
  },
  setShouldRefreshSecurityRules: (value: boolean) => {
    set({ shouldRefreshSecurityRules: value })
  },
  setClusterSecurityRules: async (clusteruuid: string) => {
    let SecurityRuleItems: SecurityRulesTypes[] = []
    const securityResponse = await ClusterService.getSecurityRules(clusteruuid)
    if (securityResponse.data.getfirewallresponse.length > 0) {
      SecurityRuleItems = createSecurityRulesArray(securityResponse.data.getfirewallresponse)
      set({ clusterSecurityRules: SecurityRuleItems })
      set({ securityRuleUuid: clusteruuid })
    }
  },
  setEditSecurityRule: (rule: SecurityRulesTypes) => {
    set({ editSecurityRule: rule })
  }
}))

const parseState = (state: string, status: any): string => {
  const { errorcode } = status
  if (!errorcode) {
    return state
  }
  return 'Error'
}

const createClustersJson = (rows: any): Cluster[] => {
  const clustersData: Cluster[] = []
  const socketResponseJson = rows

  for (let i = 0; i < socketResponseJson.length; i++) {
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
    const vips = socketResponseJson[i].vips
    const kubeconfig = socketResponseJson[i].kubeconfig
    const storages = socketResponseJson[i].storages
    const clustertype = socketResponseJson[i].clustertype ?? ''

    const clusterData: Cluster = {
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
      storages
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
const createRuntimeJson = (rows: any): runtimes[] => {
  const clustersRuntimes: runtimes[] = []
  const responseJson = rows

  for (let i = 0; i < responseJson.length; i++) {
    const k8sversionname = responseJson[i].k8sversionname
    const runtimename = responseJson[i].runtimename

    const clusterData: runtimes = {
      k8sversionname,
      runtimename
    }
    clustersRuntimes.push(clusterData)
  }
  return clustersRuntimes
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
      instanceTypeDetails: null
    }

    nodeGroupData.nodes.forEach((node) => {
      node.state = parseState(node.state, node)
    })

    const products = useHardwareStore.getState().products
    if (products.length === 0) {
      await useHardwareStore.getState().setProducts()
    }
    const instanceType = useHardwareStore.getState().products.find((x) => x.name === nodeGroupData.instancetypeid)
    nodeGroupData.instanceTypeDetails = instanceType ?? null

    clustersNodegroups.push(nodeGroupData)
  }
  return clustersNodegroups
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

const allowedStates = ['Pending', 'Deleting', 'DeletePending', 'Updating', 'Reconciling']

const needToRefreshNodeGroupNodes = (nodeGroup: NodeGroup): boolean => {
  if (!nodeGroup.nodes || nodeGroup.nodes.length === 0) {
    return false
  }
  return nodeGroup.nodes.some((x) => allowedStates.includes(x.state))
}

const needToRefreshNodeGroup = (cluster: Cluster): boolean => {
  if (!cluster.nodegroups || cluster.nodegroups.length === 0) {
    return false
  }
  return cluster.nodegroups.some((x) => allowedStates.includes(x.nodegroupstate) || needToRefreshNodeGroupNodes(x))
}

const needToRefreshLoadBalancer = (cluster: Cluster): boolean => {
  if (!cluster.vips || cluster.vips.length === 0) {
    return false
  }
  return cluster.vips.some((x) => allowedStates.includes(x.vipstate))
}

const needToRefreshClusters = (): boolean => {
  const clusters = useClusterStore.getState().clustersData
  const shouldRefreshClusters = useClusterStore.getState().shouldRefreshClusters
  if (shouldRefreshClusters && clusters) {
    return clusters.some(
      (c: Cluster) =>
        allowedStates.includes(c.clusterstate) || needToRefreshNodeGroup(c) || needToRefreshLoadBalancer(c)
    )
  }

  return false
}

const needToRefreshSecurityRules = (): boolean => {
  const clusterSecurityRules = useClusterStore.getState().clusterSecurityRules
  const shouldRefreshSecurityRules = useClusterStore.getState().shouldRefreshSecurityRules
  if (clusterSecurityRules && shouldRefreshSecurityRules) {
    return clusterSecurityRules.some((c: SecurityRulesTypes) => allowedStates.includes(c.state))
  }
  return false
}

setInterval(() => {
  if (needToRefreshClusters()) {
    void (async () => {
      await useClusterStore.getState().setClustersData(true)
    })()
  }
  if (needToRefreshSecurityRules()) {
    void (async () => {
      const securityRuleUuid = useClusterStore.getState().securityRuleUuid
      if (securityRuleUuid) {
        await useClusterStore.getState().setClusterSecurityRules(securityRuleUuid)
      }
    })()
  }
}, 10000)

const filterValue = (object: any, property: any): any => {
  return object[property] ? object[property] : ''
}

export default useClusterStore
