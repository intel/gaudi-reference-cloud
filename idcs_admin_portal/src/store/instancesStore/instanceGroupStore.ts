import { create } from 'zustand'
import CloudAccountService from '../../services/CloudAccountService'

export interface InstanceGroups {
  cloudAccountId: string
  name: string
  availabilityZone: string
  instanceType: string
  status: string
}

interface InstanceGroupsStore {
  instanceGroups: InstanceGroups[] | [] | any
  loading: boolean
  cloudAccountId: string
  setInstanceGroups: (cloudAccountId: string) => Promise<void>
  shouldRefreshInstanceGroups: boolean
  setShouldRefreshInstanceGroups: (value: boolean) => void
}

const useInstanceGroupsStore = create<InstanceGroupsStore>()((set) => ({
  instanceGroups: [],
  loading: false,
  cloudAccountId: '',
  shouldRefreshInstanceGroups: false,
  setInstanceGroups: async (cloudAccountId: string) => {
    if (cloudAccountId === '') {
      set({ instanceGroups: [] })
      set({ loading: false })
      return
    }

    set({ loading: true })

    CloudAccountService.getInstanceGroupsById(cloudAccountId).then((res) => {
      if (res.data) {
        set({ instanceGroups: res.data })
        set({ cloudAccountId })
        set({ loading: false })
      } else {
        set({ instanceGroups: [] })
        set({ loading: true })
      }
    }).catch(() => {
      set({ instanceGroups: [] })
      set({ loading: false })
    })
  },
  setShouldRefreshInstanceGroups: (value: boolean) => {
    set({ shouldRefreshInstanceGroups: value })
  }
}))

const needToRefreshInstanceGroups = (): boolean => {
  // Fetch Instance Groups.
  const shouldRefreshInstanceGroups = useInstanceGroupsStore.getState().shouldRefreshInstanceGroups
  // Fetch shouldRefreshInstanceGroups value.
  const instanceGroups = useInstanceGroupsStore.getState().instanceGroups
  return shouldRefreshInstanceGroups && !checkInstanceCountsMatch(instanceGroups)
}

// Custom method to check whether the instanceCount and readyCount properties are same or not.
const checkInstanceCountsMatch = (instanceGroupsObject: { items: any }): any => {
  // Returns TRUE when 'instanceCount' and 'readyCount' properties are equal. Otherwise, returns FALSE.
  return instanceGroupsObject?.items?.every((item: { spec: { instanceCount: any }, status: { readyCount: any } }) => item.spec.instanceCount === item.status.readyCount)
}

// Setup time interval for 10 seconds to fetch the latest instance group data.
setInterval(() => {
  if (needToRefreshInstanceGroups()) {
    void (async () => {
      // Collecting Cloud Account Id.
      const cloudAccountId = useInstanceGroupsStore.getState().cloudAccountId
      // Calls setInstanceGroups to fetch latest data.
      await useInstanceGroupsStore.getState().setInstanceGroups(cloudAccountId)
    })()
  }
}, 10000)

export default useInstanceGroupsStore
