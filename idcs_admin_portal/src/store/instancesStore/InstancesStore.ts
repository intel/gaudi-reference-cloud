import { create } from 'zustand'
import CloudAccountService from '../../services/CloudAccountService'

export interface Instances {
  cloudAccountId: string
  name: string
  instanceGroup: string
  resourceId: string
  resourceVersion: string
  availabilityZone: string
  instanceType: string
  status: string
  creationTimestamp: string
}

interface InstancesStore {
  instances: Instances[] | [] | any
  loading: boolean
  cloudAccountId: string
  setInstances: (cloudAccountId: string) => Promise<void>
  shouldRefreshInstances: boolean
  setShouldRefreshInstances: (value: boolean) => void
}

const useInstancesStore = create<InstancesStore>()((set) => ({
  instances: [],
  loading: false,
  cloudAccountId: '',
  shouldRefreshInstances: false,
  setInstances: async (cloudAccountId: string) => {
    if (cloudAccountId === '') {
      set({ instances: [] })
      set({ loading: false })
      return
    }

    set({ loading: true })
    // Calling getInstance service to retrive the instances.
    CloudAccountService.getInstancesByCloudAccount(cloudAccountId).then((res) => {
      if (res.data) {
        set({ instances: res.data })
        set({ cloudAccountId })
        set({ loading: false })
      } else {
        set({ instances: [] })
        set({ loading: true })
      }
    }).catch(() => {
      set({ instances: [] })
      set({ loading: false })
    })
  },
  setShouldRefreshInstances: (value: boolean) => {
    set({ shouldRefreshInstances: value })
  }
}))

const needToRefreshInstances = (): boolean => {
  // Fetch instances.
  const instances = useInstancesStore.getState().instances
  // Fetch shouldRefreshInstances value.
  const shouldRefreshInstances = useInstancesStore.getState().shouldRefreshInstances
  // Calls when both operations are TRUE.
  return shouldRefreshInstances && checkInstanceStateMatches(instances)
}

const checkInstanceStateMatches = (instancesObject: { items: any }): any => {
  // Initializing the states to trigger the calls.
  const allowedStates = ['Provisioning', 'Stopping', 'Terminating']
  return instancesObject?.items?.some((item: { status: { phase: string } }) => allowedStates.includes(item.status.phase))
}

// Setup time interval for 10 seconds to fetch the latest instance group data.
setInterval(() => {
  if (needToRefreshInstances()) {
    void (async () => {
      // Collecting Cloud Account Id.
      const cloudAccountId = useInstancesStore.getState().cloudAccountId
      // Calls setInstanceGroups to fetch latest data.
      await useInstancesStore.getState().setInstances(cloudAccountId)
    })()
  }
}, 10000)

export default useInstancesStore
