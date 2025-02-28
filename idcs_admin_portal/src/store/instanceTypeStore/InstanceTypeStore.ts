import { create } from 'zustand'
import IKSService from '../../services/IKSService'
import type { IMISData } from '../imiStore/IMIStore'

interface InstanceTypesData {
  instancetypename: string
  memory: number
  cpu: number
  nodeprovidername: string
  storage: number
  status: string
  displayname: string
  imioverride: boolean
  description: string
  category: string
  family: string
  allowManualInsert: boolean
  imiResponse: IMISData[]
  instacetypeimik8scompatibilityresponse: IMISData[]
}

interface ComputeResponseData {
  instancetypename: string
  memory: number
  cpu: number
  nodeprovidername: string
  storage: number
  status: string
  displayname: string
  imioverride: boolean
  description: string
  category: string
  family: string
  iksDB: boolean
}

interface InstanceTypeData {
  iksInstanceType: InstanceTypesData
  computeInstanceType: ComputeResponseData
}

interface InstanceTypeInfoData {
  computeResponse: ComputeResponseData[]
  nodeprovidername: string []
  states: string[]
}

interface InstanceTypesPayload {
  memory: number
  cpu: number
  nodeprovidername: string
  storage: number
  status: string
  displayname: string
  imioverride: boolean
  description: string
  category: string
  family: string
  iksDB: boolean
  allowManualInsert: boolean
  iksadminkey: string
}

interface InstanceTypesDeletePayload {
  iksadminkey: string
}

interface InstanceTypeK8sPayload {
  category: string
  family: string
  iksadminkey: string
  nodeprovidername: string
  instacetypeimik8scompatibilityresponse: Array<{
    artifact: string
    category: string
    cposimageinstances: string
    family: string
    name: string
    os: string
    provider: string
    runtime: string
    type: string
    upstreamreleasename: string
  }>
}

interface InstanceTypeStore {
  loading: boolean
  stopLoading: () => void
  instanceTypesData: InstanceTypesData[] | []
  getInstanceTypesData: (isBackGround: boolean) => Promise<void>
  instanceTypeData: InstanceTypeData | null
  clearInstanceTypeData: () => void
  getInstanceTypeDataByID: (id: string, isBackGround: boolean) => Promise<void>
  createInstanceTypeData: (payload: InstanceTypesData, isBackGround: boolean) => Promise<any>
  updateInstanceTypeData: (id: string, payload: InstanceTypesPayload, isBackGround: boolean) => Promise<any>
  deleteInstanceTypeData: (id: string, payload: InstanceTypesDeletePayload, isBackGround: boolean) => Promise<any>
  instanceTypesInfoData: InstanceTypeInfoData | null
  getInstanceTypesInfoData: (isBackGround: boolean) => Promise<void>
  updateInstanceTypeK8sData: (id: string, payload: InstanceTypeK8sPayload, isBackGround: boolean) => Promise<any>
}

const useInstanceTypeStore = create<InstanceTypeStore>()((set, get) => ({
  loading: false,
  stopLoading: () => {
    set({ loading: false })
  },
  instanceTypesData: [],
  getInstanceTypesData: async (isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    let socketResponse: InstanceTypesData[] = []

    const socket = await IKSService.getInstanceTypes()
    const { instanceTypeResponse } = socket.data ?? []
    if (instanceTypeResponse) {
      socketResponse = createInstanceTypeDataJSON(instanceTypeResponse)
      set({ instanceTypesData: socketResponse })
    }
  },
  instanceTypeData: null,
  clearInstanceTypeData: () => {
    set({ instanceTypeData: null })
  },
  getInstanceTypeDataByID: async (id: string, isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    const socket = await IKSService.getInstanceTypes(id)
    const { iksInstanceType, computeInstanceType } = socket.data

    if (iksInstanceType) {
      set({
        instanceTypeData: {
          iksInstanceType,
          computeInstanceType
        }
      })
    }
  },
  createInstanceTypeData: async (payload: InstanceTypesData, isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    const socket = await IKSService.createInstanceTypes(payload)
    const createdInstanceTypeData = socket.data

    return await Promise.resolve(createdInstanceTypeData)
  },
  updateInstanceTypeData: async (id: string, payload: InstanceTypesPayload, isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    const socket = await IKSService.updateInstanceTypes(id, payload)
    const updatedInstanceTypeData = socket.data

    return await Promise.resolve(updatedInstanceTypeData)
  },
  deleteInstanceTypeData: async (id: string, payload: InstanceTypesDeletePayload, isBackGround) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    const socket = await IKSService.deleteInstanceTypes(id, payload)
    const deletedInstanceTypeData = socket.data

    return await Promise.resolve(deletedInstanceTypeData)
  },
  instanceTypesInfoData: null,
  getInstanceTypesInfoData: async (isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    const socket = await IKSService.getInstanceTypesInfo()
    const infoData: InstanceTypeInfoData = socket.data

    if (infoData) {
      set({ instanceTypesInfoData: infoData })
    }
  },
  updateInstanceTypeK8sData: async (id: string, payload: InstanceTypeK8sPayload, isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    const socket = await IKSService.updateInstanceTypesK8s(id, payload)
    const updatedInstanceTypeK8sData = socket.data

    return await Promise.resolve(updatedInstanceTypeK8sData)
  }
}))

const createInstanceTypeDataJSON = (rows: any[]): InstanceTypesData[] => {
  const instanceTypeResponse: InstanceTypesData[] = []

  rows.forEach((row) => {
    instanceTypeResponse.push({
      instancetypename: row.instancetypename,
      memory: row.memory,
      cpu: row.cpu,
      nodeprovidername: row.nodeprovidername,
      storage: row.storage,
      status: row.status,
      displayname: row.displayname,
      imioverride: row.imioverride,
      description: row.description,
      category: row.category,
      family: row.family,
      allowManualInsert: row.allowManualInsert,
      imiResponse: row.imiResponse ?? [],
      instacetypeimik8scompatibilityresponse: row.instacetypeimik8scompatibilityresponse ?? []
    })
  })

  return instanceTypeResponse
}

export default useInstanceTypeStore
