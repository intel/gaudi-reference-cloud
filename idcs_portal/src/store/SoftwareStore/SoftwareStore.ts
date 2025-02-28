// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import PublicService from '../../services/PublicService'
import type Software from '../models/Software/Software'
import type MediaFile from '../models/MediaFile/MedialFile'
import moment from 'moment'

export interface SoftwareStore {
  softwareList: Software[] | []
  softwareDetail: Software | null
  loading: boolean
  getSoftware: (softwareId: string) => Promise<void>
  setSoftwareList: (background: boolean) => Promise<void>
  reset: () => void
}

const initialState = {
  softwareList: [],
  softwareDetail: null,
  loading: false
}

const useSoftwareStore = create<SoftwareStore>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setSoftwareList: async (background) => {
    if (!background) {
      set({ loading: true })
    }
    try {
      const allCatalogs = [
        PublicService.getSoftwareCatalog(),
        PublicService.getDpaiCatalog(),
        PublicService.getMaaSCatalog()
      ]

      const responses = await Promise.allSettled(allCatalogs)

      const successfulResponses = responses
        .filter((response): response is PromiseFulfilledResult<any> => response.status === 'fulfilled')
        .map((response) => response.value)

      const softwareDetail = successfulResponses.reduce((acc, response) => acc.concat(response.data.products), [])

      const softwareList: Software[] = []
      for (const detail of softwareDetail) {
        const software = buildData(detail)
        softwareList.push(software)
      }
      set({ loading: false })
      set({ softwareList })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  getSoftware: async (softwareId: string) => {
    set({ loading: true })
    try {
      const allCatalogDetails = [
        PublicService.getSoftwareDetail(softwareId),
        PublicService.getDpaiDetail(softwareId),
        PublicService.getMassDetail(softwareId)
      ]

      const responses = await Promise.allSettled(allCatalogDetails)

      const successfulResponses = responses
        .filter((response): response is PromiseFulfilledResult<any> => response.status === 'fulfilled')
        .map((response) => response.value)

      const responseDetail = successfulResponses.reduce((acc, response) => acc.concat(response.data.products), [])

      if (responseDetail.length > 0) {
        const training = buildData(responseDetail[0])
        set({ softwareDetail: training })
      }
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  }
}))

const buildMediaArray = (metadata: any): MediaFile[] => {
  const mediaArray: MediaFile[] = []
  for (let i = 1; i <= 3; i++) {
    if (metadata[`media${i}.URL`]) {
      mediaArray.push({
        src: metadata[`media${i}.URL`],
        type: metadata[`media${i}.type`],
        poster: metadata[`media${i}.posterURL`]
      })
    }
  }
  return mediaArray
}

const buildData = (response: any): Software => {
  const metadata = {
    ...response.metadata
  }
  const rates = response.rates
  let accountType = ''
  let unit = ''
  let rateValue = ''
  let usageExpr = ''

  if (rates.length > 0) {
    const rate = rates[0]

    if (rate) {
      accountType = rate.accountType
      unit = rate.unit
      rateValue = rate.rate
      usageExpr = rate.usageExpr
    }
  }
  const software: Software = {
    name: response.name,
    id: response.id,
    created: moment(response.created).format('MM/DD/YYYY h:mm a'),
    vendorId: response.vendorId,
    familyId: response.familyId,
    description: response.description,
    documentation: metadata.Documentation,
    access: metadata.access,
    billingEnable: metadata.billingEnable,
    category: metadata.category,
    components: metadata.components,
    downloadURL: metadata['detail.downloadURL'],
    features: metadata['detail.features'],
    demoURL: metadata['detail.demoURL'],
    helpURL: metadata['detail.helpURL'],
    licenseURL: metadata['detail.licenseURL'],
    jupyterlab: metadata['detail.jupyterlab'],
    audience: metadata['detail.objectives.audience'],
    overview: metadata['detail.overview'],
    productURL: metadata['detail.productURL'],
    shortDesc: metadata.shortDesc,
    useCases: metadata['detail.useCases']?.split(','),
    displayCatalogDesc: metadata.displayCatalogDesc,
    displayInHomepage: metadata.displayInHomepage === 'true',
    displayName: metadata.displayName,
    familyDisplayDescription: metadata['family.displayDescription'],
    familyDisplayName: metadata['family.displayName'],
    region: metadata.region,
    service: metadata.service,
    eccn: response.eccn,
    pcq: response.pcq,
    matchExpr: response.matchExpr,
    accountType,
    unit,
    usageUnit: metadata['usage.unit'],
    rate: rateValue,
    usageExpr,
    imageSource: metadata.displayPicture,
    homepageDisplayGroup: metadata.homepageDisplayGroup,
    platforms: metadata.platforms,
    status: response.status,
    launchImage: metadata['launch.image'],
    launchDownloadUrl: metadata['launch.downloadUrl'],
    launchLink: metadata['launch.link'],
    launch: metadata.launch,
    mediaArray: buildMediaArray(metadata),
    model: metadata.hfModelName
  }

  return software
}

export default useSoftwareStore
