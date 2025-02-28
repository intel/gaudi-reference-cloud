import { create } from 'zustand'
import IKSService from '../../services/IKSService'

export interface K8sversion {
  cpimi: string
  major: string
  minor: string
  name: string
  provider: string
  releasename: string
  runtime: string
  state: string
  workimi: string
}

interface K8sStore {
  loading: boolean
  stopLoading: () => void
  k8sData: K8sversion[] | []
  getK8sData: (isBackGround: boolean) => Promise<void>
}

const useK8SStore = create<K8sStore>()((set, get) => ({
  loading: false,
  stopLoading: () => {
    set({ loading: false })
  },
  k8sData: [],
  getK8sData: async (isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }

    let socketResponse: K8sversion[] = []

    const socket = await IKSService.getK8s()
    const k8sresponse = socket.data.k8sversions ?? []
    if (k8sresponse) {
      socketResponse = createK8sDataJSON(k8sresponse)
      set({ k8sData: socketResponse })
    }
  }
}))

const createK8sDataJSON = (rows: any[]): K8sversion[] => {
  const k8sResponse: K8sversion[] = []

  rows.forEach((row) => {
    k8sResponse.push({
      cpimi: row.cpimi,
      major: row.major,
      minor: row.minor,
      name: row.name,
      provider: row.provider,
      releasename: row.releasename,
      runtime: row.runtime,
      state: row.state,
      workimi: row.workimi
    })
  })

  return k8sResponse
}

export default useK8SStore
