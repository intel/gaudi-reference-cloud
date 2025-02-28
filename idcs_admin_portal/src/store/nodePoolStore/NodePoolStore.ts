import { create } from 'zustand'
import NodePoolService from '../../services/NodePoolService'

interface NodeStateDefinition {
  nodeId: string
  instanceTypeStats: NodeStateInstanceTypeStatesDefinition[]
}

interface NodeStateInstanceTypeStatesDefinition {
  instanceType: string
  runningInstances: string
  maxNewInstances: string
}

interface NodeDefinition {
  nodeName: string
  availabilityZone: string
  clusterId: string
  namespace: string
  nodeId: string
  percentageResourcesUsed: string
  region: string
  instanceTypes: string[]
  poolIds: string[]
}

interface PoolDefinition {
  poolId: string
  poolName: string
  poolAccountManagerAgsRole: string
  numberOfNodes: string
}

interface CloudAccountDefinition {
  poolId: string
  cloudAccountId: string
  createAdmin: string
}

interface NodePoolStore {
  nodeStates: NodeStateDefinition[]
  nodeList: NodeDefinition[]
  poolList: PoolDefinition[]
  cloudAccountList: CloudAccountDefinition[]
  loading: boolean
  setNodeStatesList: (nodeId: string | null) => Promise<void>
  setNodeList: (poolId: string | null, isReturn: boolean) => Promise<any>
  setCloudAccountList: (poolId: string) => Promise<void>
  setPoolList: () => Promise<void>
}

const useNodePoolStore = create<NodePoolStore>()((set, get) => ({
  nodeStates: [],
  nodeList: [],
  poolList: [],
  cloudAccountList: [],
  loading: false,
  setNodeStatesList: async (nodeId: string | null = null) => {
    try {
      set({ loading: true })
      const { data } = await NodePoolService.getNodesStates(nodeId)
      const nodeStates = buildNodeSates(data)
      set({ nodeStates })
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setNodeList: async (poolId: string | null = null, isReturn: boolean = false) => {
    try {
      set({ loading: !isReturn })
      const { data } = await NodePoolService.getNodes(poolId)

      const nodes = buildNodes(data)
      set({ nodeList: nodes })
      set({ loading: false })
      if (isReturn) {
        return nodes
      }
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setCloudAccountList: async (poolId: string) => {
    try {
      set({ loading: true })
      const { data } = await NodePoolService.getPoolCloudAccounts(poolId)
      const cloudAccountList = buildCloudAccountList(data)
      set({ cloudAccountList })
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setPoolList: async () => {
    try {
      set({ loading: true })
      const { data } = await NodePoolService.getPools()
      const pools = buildPools(data)
      set({ poolList: pools })
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  }
}))

const buildPools = (data: any): PoolDefinition[] => {
  if (!data?.computeNodePools || data.computeNodePools.length === 0) {
    return []
  }

  return data.computeNodePools.map((item: any) => ({
    poolId: item.poolId,
    poolName: item.poolName,
    poolAccountManagerAgsRole: item.poolAccountManagerAgsRole,
    numberOfNodes: item.numberOfNodes
  }))
}

const buildNodes = (data: any): NodeDefinition[] => {
  if (!data?.computeNodes || data.computeNodes.length === 0) {
    return []
  }

  return data.computeNodes.map((item: any) => ({
    nodeName: item.nodeName,
    availabilityZone: item.availabilityZone,
    clusterId: item.clusterId,
    namespace: item.namespace,
    nodeId: item.nodeId,
    percentageResourcesUsed: item.percentageResourcesUsed,
    region: item.region,
    instanceTypes: item.instanceTypes,
    poolIds: item.poolIds
  }))
}

const buildNodeSates = (data: any): NodeStateDefinition[] => {
  if (!data?.NodeInstanceTypeStats || data.NodeInstanceTypeStats.length === 0) {
    return []
  }

  return data.NodeInstanceTypeStats.map((item: any) => ({
    nodeId: item.nodeId,
    instanceTypeStats: item.InstanceTypeStats
  }))
}

const buildCloudAccountList = (data: any): CloudAccountDefinition[] => {
  if (!data?.CloudAccountsForComputeNodePool || data.CloudAccountsForComputeNodePool.length === 0) {
    return []
  }

  return data.CloudAccountsForComputeNodePool.map((item: any) => ({
    poolId: item.poolId,
    cloudAccountId: item.cloudAccountId,
    createAdmin: item.createAdmin
  }))
}

export default useNodePoolStore
