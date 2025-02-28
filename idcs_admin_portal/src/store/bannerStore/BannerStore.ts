import { create } from 'zustand'
import BannerService from '../../services/BannerService'
const ALL = 'all'

export interface BannerLink {
  label: string
  href: string
  openInNewTab: boolean
}

interface BannerDefinition {
  id: number
  type: string
  title: string
  status: string
  message: string
  userTypes: string[]
  routes: string[]
  regions: string[]
  expirationDatetime: string
  link?: BannerLink
  isMaintenance?: string
  updatedTimestamp?: number
}

interface BannerStore {
  bannerList: BannerDefinition[]
  loading: boolean
  removeBanner: (id: number) => Promise<void>
  updateBanner: (payload: BannerDefinition) => Promise<void>
  addBanner: (payload: BannerDefinition) => Promise<void>
  setBannerList: () => Promise<void>
}

const useBannerStore = create<BannerStore>()((set, get) => ({
  bannerList: [],
  loading: false,
  setBannerList: async () => {
    try {
      set({ loading: true })
      const newData = await BannerService.getBanners()
      set({ bannerList: buildBannersResponse(newData) })
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  removeBanner: async (id: number) => {
    const bannerList = get().bannerList
    try {
      set({ loading: true })
      await BannerService.removeBanner(id, bannerList)
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  updateBanner: async (payload: BannerDefinition) => {
    const bannerList = get().bannerList
    try {
      set({ loading: true })
      validateMaintenanceBanners(bannerList, payload)
      await BannerService.updateBanner(payload, bannerList)
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  addBanner: async (payload: BannerDefinition) => {
    const bannerList = get().bannerList
    try {
      set({ loading: true })
      validateMaintenanceBanners(bannerList, payload)
      await BannerService.postBanner(payload, bannerList)
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  }
}))

const buildBannersResponse = (data: any): BannerDefinition[] => {
  return data
    ?.map((result: any) => {
      const item = result
      return {
        id: item?.id ? item.id : null,
        type: item?.type ? item.type : '',
        title: item?.title ? item.title : '',
        message: item?.message ? item.message : '',
        userTypes: item?.userTypes ? item.userTypes : [],
        routes: item?.routes ? item?.routes : [],
        regions: item?.regions ? item?.regions : [],
        status: item?.status === 'active' ? item?.status : 'inactive',
        expirationDatetime: item?.expirationDatetime ? new Date(item.expirationDatetime).toISOString() : '',
        link: item?.link,
        isMaintenance: item?.isMaintenance ?? 'False',
        updatedTimestamp: item?.updatedTimestamp
      }
    })
    .sort((a: BannerDefinition, b: BannerDefinition) => b.id - a.id)
}

const validateMaintenanceBanners = (bannerList: BannerDefinition[], payload: BannerDefinition): void => {
  if (
    !(
      payload.status === 'inactive' ||
      payload.isMaintenance === 'False' ||
      new Date(payload.expirationDatetime).getTime() < Date.now()
    )
  ) {
    const currentMaintenanceBanners = bannerList.filter(
      (banner) =>
        banner.id !== payload.id &&
        banner.status === 'active' &&
        banner.isMaintenance === 'True' &&
        (!banner.expirationDatetime || new Date(banner.expirationDatetime).getTime() > Date.now())
    )

    const isRouteDuplicated =
      payload.routes.includes(ALL) ||
      currentMaintenanceBanners.some(
        (banner) => banner?.routes.includes(ALL) || banner?.routes.some((route) => payload.routes.includes(route))
      )

    const isRegionDuplicated = currentMaintenanceBanners.some(banner => banner?.regions.some(region => payload.regions.includes(region)))

    if (currentMaintenanceBanners.length > 0 && (isRouteDuplicated && isRegionDuplicated)) {
      const errorMessage: string =
        'Only one active maintenance banner is permitted for each route per region at a time. Verify that there are no other active maintenance banners before proceeding.'
      throw Error(errorMessage)
    }
  }
}

export default useBannerStore
