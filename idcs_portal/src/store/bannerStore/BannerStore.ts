// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface BannerLink {
  label: string
  href: string
  openInNewTab: boolean
}

interface BannerDefinition {
  id: string
  type: string
  title: string
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
  hiddenBannerList: string[]
  addBanner: (
    id: string,
    type: string,
    title: string,
    message: string,
    userTypes: string[],
    routes: string[],
    regions: string[],
    expirationDatetime: string,
    link: BannerLink,
    isMaintenance?: string,
    updatedTimestamp?: number
  ) => void
  hideBanner: (id: string) => void
}

const addBanner = (
  banners: BannerDefinition[],
  id: string,
  type: string,
  title: string,
  message: string,
  userTypes: string[],
  routes: string[],
  regions: string[],
  expirationDatetime: string,
  link: BannerLink,
  isMaintenance?: string,
  updatedTimestamp?: number
): BannerDefinition[] => {
  const newBanner: BannerDefinition = {
    id,
    type,
    title,
    message,
    userTypes,
    routes,
    regions,
    expirationDatetime,
    link,
    isMaintenance,
    updatedTimestamp
  }
  banners.push(newBanner)
  return [...banners]
}

const hideBanner = (hiddenBanners: string[], id: string): string[] => {
  hiddenBanners.push(id)
  return [...hiddenBanners]
}

const useBannerStore = create<BannerStore>()(
  persist(
    (set, get) => ({
      bannerList: [],
      hiddenBannerList: [],
      addBanner: (
        id: string,
        type: string,
        title: string,
        message: string,
        userTypes: string[],
        routes: string[],
        regions: string[],
        expirationDatetime: string,
        link: BannerLink,
        isMaintenance?: string,
        updatedTimestamp?: number
      ) => {
        const bannerList = get().bannerList
        set({
          bannerList: addBanner(
            bannerList,
            id,
            type,
            title,
            message,
            userTypes,
            routes,
            regions,
            expirationDatetime,
            link,
            isMaintenance,
            updatedTimestamp
          )
        })
      },
      hideBanner: (id: string) => {
        const hiddenBannerList = get().hiddenBannerList
        set({ hiddenBannerList: hideBanner(hiddenBannerList, id) })
      }
    }),
    {
      name: 'hiddenBannerList',
      partialize: (state) => ({ hiddenBannerList: state.hiddenBannerList })
    }
  )
)

export default useBannerStore
export type { BannerDefinition }
