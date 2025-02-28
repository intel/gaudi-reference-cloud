// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import moment from 'moment'
import PublicService from '../../services/PublicService'
import type { Storage, ObjectStorage } from '../models/Storage/Storage'
import { StorageServicesEnum } from '../../utils/Enums'

export interface StorageStore {
  storages: Storage[] | [] | null
  setStorages: () => Promise<void>
  loading: boolean
  objectStorages: ObjectStorage[] | [] | null
  setObjectStorages: () => Promise<void>
  reset: () => void
}

const initialState = {
  storages: null,
  loading: false,
  objectStorages: null
}

const useStorageStore = create<StorageStore>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setStorages: async () => {
    try {
      set({ loading: true })
      const response = await PublicService.getStorageCatalog()
      const fileStorages = response.data.products
        ? response.data.products.filter((x: any) => x.metadata.service === StorageServicesEnum.fileStorage)
        : []
      const storages: Storage[] = []
      fileStorages.forEach((detail: any) => {
        const storage = buildData(detail)
        storages.push(storage)
      })
      set({ loading: false })
      set({ storages })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setObjectStorages: async () => {
    try {
      set({ loading: true })
      const response = await PublicService.getStorageCatalog()
      const objectStorages = response.data.products
        ? response.data.products.filter((x: any) => x.metadata.service === StorageServicesEnum.objectStorage)
        : []
      const storages: ObjectStorage[] = []
      objectStorages.forEach((detail: any) => {
        const storage = buildObjectStorage(detail)
        storages.push(storage)
      })
      set({ loading: false })
      set({ objectStorages: storages })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  }
}))

const buildData = (response: any): Storage => {
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

  const storage: Storage = {
    name: response.name,
    id: response.id,
    created: moment(response.created).format('MM/DD/YYYY h:mm a'),
    vendorId: response.vendorId,
    familyId: response.familyId,
    description: response.description,
    storageCategories: metadata.instanceCategories,
    access: metadata.access,
    billingEnable: metadata.billingEnable,
    category: metadata.category,
    disableForAccountTypes: metadata.disableForAccountTypes,
    minimumSize: Number(metadata['volume.size.min']),
    maximumSize: Number(metadata['volume.size.max']),
    displayName: metadata.displayName,
    familyDisplayDescription: metadata['family.displayDescription'],
    familyDisplayName: metadata['family.displayName'],
    highlight: metadata.highlight,
    information: metadata.information,
    recommendedUseCase: metadata.recommendedUseCase,
    region: metadata.region,
    releaseStatus: metadata.releaseStatus,
    service: metadata.service,
    eccn: response.eccn,
    pcq: response.pcq,
    matchExpr: response.matchExpr,
    accountType,
    unit,
    rate: rateValue ? Number(rateValue) : 0,
    usageExpr,
    status: metadata.status,
    unitSize: metadata['volume.size.unit'],
    usageUnit: metadata['usage.unit']
  }

  return storage
}

const buildObjectStorage = (response: any): ObjectStorage => {
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

  const storage: ObjectStorage = {
    name: response.name,
    id: response.id,
    created: moment(response.created).format('MM/DD/YYYY h:mm a'),
    vendorId: response.vendorId,
    familyId: response.familyId,
    description: response.description,
    access: metadata.access,
    billingEnable: metadata.billingEnable,
    disableForAccountTypes: metadata.disableForAccountTypes,
    displayName: metadata.displayName,
    familyDisplayDescription: metadata['family.displayDescription'],
    familyDisplayName: metadata['family.displayName'],
    information: metadata.information,
    instanceCategories: metadata.instanceCategories,
    instanceType: metadata.instanceType,
    region: metadata.region,
    releaseStatus: metadata.releaseStatus,
    service: metadata.service,
    eccn: response.eccn,
    pcq: response.pcq,
    matchExpr: response.matchExpr,
    accountType,
    unit,
    rate: rateValue ? Number(rateValue) : 0,
    usageExpr,
    unitSize: metadata['volume.size.unit'],
    usageUnit: metadata['usage.unit']
  }

  return storage
}

export default useStorageStore
