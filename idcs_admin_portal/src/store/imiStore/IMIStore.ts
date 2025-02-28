import { create } from 'zustand'
import IKSService from '../../services/IKSService'

interface IMISComponents {
  artifact: string
  name: string
  version: string
}

interface InstanceTypes {
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

export interface IMISData {
  artifact: string
  components: IMISComponents[]
  name: string
  os: string
  provider: string
  runtime: string
  state: string
  type: string
  upstreamreleasename: string
  family: string
  category: string
  iscompatabilityactiveimi: boolean
  instacetypeimik8scompatibilityresponse: InstanceTypes[]
  instanceTypeResponse: InstanceTypes[]
  cposimageinstances: string[]
  isk8sActive: boolean
}

type IMICreatePayload = Omit<IMISData, 'instanceTypeResponse' | 'instacetypeimik8scompatibilityresponse'>

interface IMISPayload {
  artifact: string
  components: IMISComponents[]
  os: string
  provider: string
  runtime: string
  state: string
  type: string
  upstreamreleasename: string
  family: string
  category: string
  iksadminkey: string
}

interface IMISDeletePayload {
  iksadminkey: string
}

interface IMISInfoData {
  runtime: string[]
  osimage: string[]
  provider: string[]
  state: string[]
}

interface IMIK8sPayload {
  upstreamreleasename: string
  provider: string
  runtime: string
  os: string
  family: string
  category: string
  instancetypes?: string[]
}

interface IMIStore {
  loading: boolean
  stopLoading: () => void
  imisData: IMISData[] | []
  getIMISData: (isBackGround: boolean) => Promise<void>
  imiData: IMISData | null
  clearIMIData: () => void
  getIMIDataByID: (id: string, isBackGround: boolean) => Promise<void>
  createIMIData: (payload: IMICreatePayload, isBackGround: boolean) => Promise<any>
  updateIMIData: (id: string, payload: IMISPayload, isBackGround: boolean) => Promise<any>
  deleteIMIData: (id: string, payload: IMISDeletePayload, isBackground: boolean) => Promise<any>
  imiInfoData: IMISInfoData | null
  getIMIInfoData: (isBackGround: boolean) => Promise<any>
  updateIMIK8sData: (id: string, payload: IMIK8sPayload, isBackGround: boolean) => Promise<any>
}

const useIMIStore = create<IMIStore>()((set, get) => ({
  loading: false,
  stopLoading: () => {
    set({ loading: false })
  },
  imisData: [],
  getIMISData: async (isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    let socketResponse: IMISData[] = []

    const socket = await IKSService.getIMIS()
    const { imiresponse } = socket.data ?? []
    if (imiresponse) {
      socketResponse = createIMIDataJSON(imiresponse)
      set({ imisData: socketResponse })
    }
  },
  imiData: null,
  clearIMIData: () => {
    set({ imiData: null })
  },
  getIMIDataByID: async (id: string, isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    // let socketResponse: IMISData

    const socket = await IKSService.getIMIS(id)
    const response = socket.data

    if (response) {
      // socketResponse = createIMIDataJSON(response)
      set({ imiData: response })
    }
  },
  createIMIData: async (payload: IMICreatePayload, isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    const socket = await IKSService.createIMIS(payload)
    const createdIMIData = socket.data

    return await Promise.resolve(createdIMIData)
  },
  updateIMIData: async (id: string, payload: IMISPayload, isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    const socket = await IKSService.updateIMIS(id, payload)
    const updatedIMIData = socket.data

    return await Promise.resolve(updatedIMIData)
  },
  deleteIMIData: async (id: string, payload: IMISDeletePayload, isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    const socket = await IKSService.deleteIMIS(id, payload)
    const deletedIMIData = socket.data

    return await Promise.resolve(deletedIMIData)
  },
  imiInfoData: null,
  getIMIInfoData: async (isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    const socket = await IKSService.getIMISInfo()
    const imiInfo = socket.data
    if (imiInfo) {
      mapToStringArray(imiInfo)

      set({ imiInfoData: imiInfo })
    }
  },
  updateIMIK8sData: async (id: string, payload: IMIK8sPayload, isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    const socket = await IKSService.updateIMISK8s(id, payload)
    const updatedIMIData = socket.data

    return await Promise.resolve(updatedIMIData)
  }
}))

const createIMIDataJSON = (rows: any[]): IMISData[] => {
  const imiResponse: IMISData[] = []

  rows.forEach((row) => {
    const instacetypeimik8scompatibilityresponse: InstanceTypes[] = []

    if (Array.isArray(row.instacetypeimik8scompatibilityresponse) && row.instacetypeimik8scompatibilityresponse.length > 0) {
      row.instacetypeimik8scompatibilityresponse.forEach((res: any) => {
        instacetypeimik8scompatibilityresponse.push({
          instancetypename: res.instancetypename,
          category: res.category,
          cpu: res.cpu,
          description: res.description,
          displayname: res.displayname,
          family: res.family,
          iksDB: res.iksDB,
          imioverride: res.imioverride,
          memory: res.memory,
          nodeprovidername: res.nodeprovidername,
          status: res.status,
          storage: res.storage
        })
      })
    }

    const instanceTypeResponse: InstanceTypes[] = []

    if (Array.isArray(row.instanceTypeResponse) && row.instanceTypeResponse.length > 0) {
      row.instanceTypeResponse.forEach((res: any) => {
        instanceTypeResponse.push({
          instancetypename: res.instancetypename,
          category: res.category,
          cpu: res.cpu,
          description: res.description,
          displayname: res.displayname,
          family: res.family,
          iksDB: res.iksDB,
          imioverride: res.imioverride,
          memory: res.memory,
          nodeprovidername: res.nodeprovidername,
          status: res.status,
          storage: res.storage
        })
      })
    }

    imiResponse.push({
      artifact: row.artifact,
      components: row.components,
      name: row.name,
      os: row.os,
      provider: row.provider,
      runtime: row.runtime,
      state: row.state,
      type: row.type,
      upstreamreleasename: row.upstreamreleasename,
      family: row.family ?? '',
      category: row.category ?? '',
      iscompatabilityactiveimi: row.iscompatabilityactiveimi,
      instacetypeimik8scompatibilityresponse,
      instanceTypeResponse,
      cposimageinstances: row.cposimageinstances,
      isk8sActive: row.isk8sActive
    })
  })

  return imiResponse
}

const mapToStringArray = (input: any): null => {
  for (const key in input) {
    input[key] = input[key].map((obj: any) => obj[key])
  }
  return null
}

export default useIMIStore
