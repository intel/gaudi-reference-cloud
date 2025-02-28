// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import useCloudAccountStore from '../cloudAccountStore/CloudAccountStore'
import useBucketStore from '../bucketStore/BucketStore'
import useBucketUsersPermissionsStore from '../bucketUsersPermissionsStore/BucketUsersPermissionsStore'
import useClusterStore from '../clusterStore/ClusterStore'
import useHardwareStore from '../hardwareStore/HardwareStore'
import useImageStore from '../imageStore/ImageStore'
import useProductStore from '../productStore/ProductStore'
import useSoftwareStore from '../SoftwareStore/SoftwareStore'
import useStorageStore from '../storageStore/StorageStore'
import useTrainingStore from '../trainingStore/TrainingStore'
import useCloudCreditsStore from '../billingStore/CloudCreditsStore'
import useInvoicesStore from '../billingStore/InvoicesStore'
import usePaymentMethodStore from '../billingStore/PaymentMethodStore'
import useUsagesReport from '../billingStore/UsagesStore'
import useSuperComputerStore from '../superComputer/SuperComputerStore'
import useLoadBalancerStore from '../loadBalancerStore/LoadBalancerStore'
import PublicService from '../../services/PublicService'
import type ConsoleUI from '../models/consoleUI/consoleUI'
import { abortGetRequests, resetAbortController } from '../../utils/AxiosInstance'
import idcConfig, { updateCurrentRegion } from '../../config/configurator'

export interface VisitedSite {
  title: string
  path: string
}

interface AppStore {
  firstLoadedPage: string | null
  firstLoadComplete: boolean
  changinRegion: boolean
  showLearningBar: boolean
  learningArticlesAvailable: boolean
  showSideNavBar: boolean
  consoleUIs: ConsoleUI[] | null
  paymentServices: ConsoleUI[] | null
  breadcrumCustomTitles: Map<string, string>
  visitedSites: VisitedSite[]
  setVisitedSites: (sites: VisitedSite[]) => void
  changeRegion: (changeRegion: string) => void
  setFirstLoadComplete: (firstLoadComplete: boolean) => void
  setFirstLoadedPage: (firstLoadComplete: string) => void
  setShowLearningBar: (showLearningBar: boolean, savePreference: boolean) => void
  toggleSideBars: (isMdBreakpoint: boolean) => void
  setLearningArticlesAvailable: (learningArticlesAvailable: boolean) => void
  setShowSideNavBar: (showSideNavBar: boolean, savePreference: boolean) => void
  setConsoleUIs: () => Promise<void>
  setPaymentServices: () => Promise<void>
  addBreadcrumCustomTitle: (path: string, title: string) => void
  resetStores: () => void
}

const useAppStore = create<AppStore>()((set, get) => ({
  firstLoadedPage: null,
  firstLoadComplete: false,
  changinRegion: false,
  showLearningBar: false,
  learningArticlesAvailable: false,
  showSideNavBar: false,
  consoleUIs: null,
  paymentServices: null,
  breadcrumCustomTitles: new Map(),
  visitedSites: [],
  setVisitedSites: (visitedSites) => {
    set({ visitedSites })
    localStorage.setItem('recentlyVisited', JSON.stringify(visitedSites))
  },
  changeRegion: (region) => {
    const urlParams = new URLSearchParams(window.location.search)
    urlParams.set('region', region)
    window.history.replaceState({}, '', '?' + urlParams.toString())
    set({ changinRegion: true })
    abortGetRequests()
    setTimeout(() => {
      resetAbortController()
      updateCurrentRegion(idcConfig)
      set({ changinRegion: false })
      get().resetStores()
    }, 500)
  },
  setFirstLoadComplete: (firstLoadComplete) => {
    set({ firstLoadComplete })
  },
  setFirstLoadedPage: (firstLoadedPage) => {
    set({ firstLoadedPage })
  },
  setShowLearningBar: (showLearningBar, savePreference) => {
    set({ showLearningBar })
    if (savePreference) {
      localStorage.setItem('showLearningBar', showLearningBar.toString())
    }
    if (!showLearningBar) {
      const showSideNavBar = localStorage.getItem('showSideNav') === 'true'
      set({ showSideNavBar })
    }
  },
  toggleSideBars: (isMdBreakpoint) => {
    const haveLearningBar = get().showLearningBar
    const haveSideBar = get().showSideNavBar
    if (isMdBreakpoint && haveLearningBar && haveSideBar) {
      set({ showLearningBar: false })
    } else if (!isMdBreakpoint) {
      const showLearningBar = localStorage.getItem('showLearningBar') === 'true'
      set({ showLearningBar })
      const showSideNavBar = localStorage.getItem('showSideNav') === 'true'
      set({ showSideNavBar })
    }
  },
  setLearningArticlesAvailable: (learningArticlesAvailable) => {
    set({ learningArticlesAvailable })
  },
  setShowSideNavBar: (showSideNavBar, savePreference) => {
    set({ showSideNavBar })
    if (savePreference) {
      localStorage.setItem('showSideNav', showSideNavBar.toString())
    }
  },
  setConsoleUIs: async () => {
    try {
      const response = await PublicService.getGuiCatalog()
      const responseDetail = { ...response.data.products }
      const consoleUIList: ConsoleUI[] = []
      for (const index in responseDetail) {
        const detail = { ...responseDetail[index] }
        const consoleUI: ConsoleUI = {
          id: detail.id,
          familyId: detail.familyId,
          vendorId: detail.vendorId,
          name: detail.name,
          description: detail.description,
          category: detail.metadata.category,
          service: detail.metadata.service,
          familyDisplayName: detail.metadata['family.displayName'],
          instanceType: detail.metadata.instanceType,
          displayName: detail.metadata.displayName,
          url: detail.metadata.url
        }
        consoleUIList.push(consoleUI)
      }
      set({ consoleUIs: consoleUIList })
    } catch (error) {
      set({ consoleUIs: [] })
    }
  },
  setPaymentServices: async () => {
    try {
      const response = await PublicService.getPaymentServicesCatalog()
      const responseDetail = { ...response.data.products }
      const paymentservicesList: ConsoleUI[] = []
      for (const index in responseDetail) {
        const detail = { ...responseDetail[index] }
        const paymentService: ConsoleUI = {
          id: detail.id,
          familyId: detail.familyId,
          vendorId: detail.vendorId,
          name: detail.name,
          description: detail.description,
          category: detail.metadata.category,
          service: detail.metadata.service,
          familyDisplayName: detail.metadata['family.displayName'],
          instanceType: detail.metadata.instanceType,
          displayName: detail.metadata.displayName,
          url: detail.metadata.url
        }
        paymentservicesList.push(paymentService)
      }
      set({ paymentServices: paymentservicesList })
    } catch (error) {
      set({ paymentServices: [] })
    }
  },
  addBreadcrumCustomTitle: (path, title) => {
    const currentMap = get().breadcrumCustomTitles
    set({ breadcrumCustomTitles: new Map(currentMap).set(path, title) })
  },
  resetStores: () => {
    useCloudAccountStore.getState().reset()
    useBucketStore.getState().reset()
    useBucketUsersPermissionsStore.getState().reset()
    useClusterStore.getState().reset()
    useHardwareStore.getState().reset()
    useImageStore.getState().reset()
    useProductStore.getState().reset()
    useSoftwareStore.getState().reset()
    useStorageStore.getState().reset()
    useTrainingStore.getState().reset()
    useCloudCreditsStore.getState().reset()
    useInvoicesStore.getState().reset()
    usePaymentMethodStore.getState().reset()
    useUsagesReport.getState().reset()
    useSuperComputerStore.getState().reset()
    useLoadBalancerStore.getState().reset()
  }
}))

export default useAppStore
