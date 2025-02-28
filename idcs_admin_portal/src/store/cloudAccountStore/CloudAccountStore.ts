import { create } from 'zustand'
import CloudAccountApproveListService from '../../services/CloudAccountApproveListService'

interface CloudAccountsData {
  account: string
  providername: string
  status: boolean
  enableStorage: boolean
  maxclusterilb_override: number
  maxclusterng_override: number
  maxclusters_override: number
  maxclustervm_override: number
  maxnodegroupvm_override: number
}

interface CloudAccountPayload {
  account: string
  status: boolean
  enableStorage: boolean
  maxclusterilb_override: number
  maxclusterng_override: number
  maxclusters_override: number
  maxclustervm_override: number
  maxnodegroupvm_override: number
  iksadminkey?: string
}

interface ResourceLimits {
  maxclusterpercloudaccount: number
  maxnodegroupspercluster: number
  maxvipspercluster: number
  maxnodespernodegroup: number
  maxclustervm: number
}

interface CloudAccountStore {
  loading: boolean
  stopLoading: () => void
  cloudAccountsData: CloudAccountsData[] | []
  resourceLimits: ResourceLimits | null
  getCloudAccountsData: (isBackGround: boolean) => Promise<void>
  createCloudAccountsData: (payload: CloudAccountPayload, isBackGround: boolean) => Promise<any>
  updateCloudAccountsData: (payload: CloudAccountPayload, isBackGround: boolean) => Promise<any>
}

const useCloudAccountStore = create<CloudAccountStore>()((set, get) => ({
  loading: false,
  stopLoading: () => {
    set({ loading: false })
  },
  cloudAccountsData: [],
  resourceLimits: null,
  getCloudAccountsData: async (isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    let socketResponse: CloudAccountsData[] = []

    const socket = await CloudAccountApproveListService.getCloudAccountsApproveList()
    const { approveListResponse, existingresourcelimits } = socket.data ?? {}

    if (approveListResponse) {
      socketResponse = createApproveListJson(approveListResponse)
      set({ cloudAccountsData: socketResponse })
    }

    if (existingresourcelimits) {
      set({ resourceLimits: { ...existingresourcelimits } })
    }
  },
  createCloudAccountsData: async (payload: CloudAccountPayload, isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    const socket = await CloudAccountApproveListService.createCloudAccountsApproveList(payload)
    const updatedApproveListData = socket.data

    if (payload.account === updatedApproveListData?.account) {
      return await Promise.resolve(updatedApproveListData)
    }

    return await Promise.reject(new Error('Payload Data does not match with Response data'))
  },
  updateCloudAccountsData: async (payload: CloudAccountPayload, isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    const socket = await CloudAccountApproveListService.updateCloudAccountsApproveList(payload)
    const updatedApproveListData = socket.data

    if (payload.account === updatedApproveListData?.account) {
      return await Promise.resolve(updatedApproveListData)
    }

    return await Promise.reject(new Error('Payload Data does not match with Response Data'))
  }
}))

const createApproveListJson = (rows: any): CloudAccountsData[] => {
  const approvalListResponse: CloudAccountsData[] = []
  const socketResponseJson = rows

  for (let i = 0; i < socketResponseJson.length; i++) {
    const account = socketResponseJson[i].account
    const providername = socketResponseJson[i].providername
    const status = socketResponseJson[i].status
    const enableStorage = socketResponseJson[i].enableStorage
    const maxclusterilbOverride = socketResponseJson[i].maxclusterilb_override
    const maxclusterngOverride = socketResponseJson[i].maxclusterng_override
    const maxclustersOverride = socketResponseJson[i].maxclusters_override
    const maxclustervmOverride = socketResponseJson[i].maxclustervm_override
    const maxnodegroupvmOverride = socketResponseJson[i].maxnodegroupvm_override

    const approvalListData: CloudAccountsData = {
      account,
      providername,
      status,
      enableStorage,
      maxclusterilb_override: maxclusterilbOverride,
      maxclusterng_override: maxclusterngOverride,
      maxclusters_override: maxclustersOverride,
      maxclustervm_override: maxclustervmOverride,
      maxnodegroupvm_override: maxnodegroupvmOverride
    }

    approvalListResponse.push(approvalListData)
  }
  return approvalListResponse
}

export default useCloudAccountStore
