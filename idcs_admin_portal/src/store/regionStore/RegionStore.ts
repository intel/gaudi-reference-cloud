import { create } from 'zustand'
import moment from 'moment'
import type Region from '../models/Region/Region'
import RegionManagementService from '../../services/RegionManagementService'
import type { AccountWhitelist } from '../models/Region/Region'
export interface RegionStore {
  regions: Region[] | []
  accountWhitelist: AccountWhitelist[] | []
  loading: boolean
  setAccountWhitelist: () => Promise<void>
  setRegions: (payload?: any) => Promise<void>
}
const dateFormat = 'MM/DD/YYYY hh:mm a'

const useRegionStore = create<RegionStore>()((set) => ({
  regions: [],
  accountWhitelist: [],
  loading: false,

  setRegions: async (payload = null) => {
    try {
      set({ loading: true })
      const { data } = await RegionManagementService.getRegions(payload)

      const parsedRegions = data.regions.map((region: any) => ({
        ...region,
        created_at: moment(region.created_at).format(dateFormat),
        updated_at: moment(region.updated_at).format(dateFormat)
      }))

      set({
        regions: parsedRegions,
        loading: false
      })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },

  setAccountWhitelist: async () => {
    try {
      set({ loading: true })
      const { data } = await RegionManagementService.getAccountWhitelist()
      const accountWhitelist = data.acl.map((e: any) => ({ ...e, created: moment(e.created).format(dateFormat) }))
      set({ accountWhitelist, loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  }
}))

export default useRegionStore
