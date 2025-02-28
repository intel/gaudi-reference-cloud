import { create } from 'zustand'
import { abortGetRequests, resetAbortController } from '../../utility/axios/AxiosInstance'
import idcConfig, { updateCurrentRegion } from '../../config/configurator'
import RegionService from '../../services/RegionService'

interface AppStore {
  firstLoadedPage: string | null
  firstLoadComplete: boolean
  changingRegion: boolean
  setFirstLoadComplete: (firstLoadComplete: boolean) => void
  setFirstLoadedPage: (firstLoadComplete: string) => void
  changeRegion: (changeRegion: string) => void
  resetStores: () => void
}

const useAppStore = create<AppStore>()((set, get) => ({
  firstLoadedPage: null,
  firstLoadComplete: false,
  changingRegion: false,
  setFirstLoadComplete: (firstLoadComplete) => {
    set({ firstLoadComplete })
  },
  setFirstLoadedPage: (firstLoadedPage) => {
    set({ firstLoadedPage })
  },
  changeRegion: (region) => {
    const urlParams = new URLSearchParams(window.location.search)
    urlParams.set('region', region)
    window.history.replaceState({}, '', '?' + urlParams.toString())
    RegionService.saveLastChangedRegion(region)
    set({ changingRegion: true })
    abortGetRequests()
    setTimeout(() => {
      resetAbortController()
      updateCurrentRegion(idcConfig)
      set({ changingRegion: false })
      get().resetStores()
    }, 500)
  },
  resetStores: () => {

  }
}))

export default useAppStore
