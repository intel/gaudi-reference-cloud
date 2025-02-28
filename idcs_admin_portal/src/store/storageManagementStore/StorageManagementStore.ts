// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import { type fileSystemUsage, type bucketUsage, type storageQuota, type storageDefaultQuota, type serviceQuota, type serviceQuotaResource, type serviceResourceItem, type serviceResource } from '../models/storageManagement/StorageManagement'
import StorageManagementService from '../../services/StorageManagementService'

interface StorageManagementStore {
  services: serviceResource[] | null
  serviceResourceItem: serviceResource | null
  serviceQuotaResource: serviceQuota | null
  serviceQuota: serviceQuota | null
  fileSystemUsages: fileSystemUsage[] | null
  bucketUsages: bucketUsage[] | null
  editStorageQuota: storageQuota | null
  storageQuotas: storageQuota[] | null
  storageDefaultQuotas: storageDefaultQuota[] | null
  loading: boolean
  getStorageUsages: (isBackground: boolean) => Promise<void>
  getQuotaAssigments: (isBackground: boolean) => Promise<void>
  setEditStorageQuota: (storageQuota: any) => void
  getServices: (isBackground: boolean) => Promise<void>
  getServicesById: (idService: string) => Promise<void>
  getServiceQuotas: (idService: string) => Promise<void>
  getServiceQuota: (idService: string, resourceName: string) => Promise<void>
  resetServiceQuotaResource: () => void
}

const buildFileSystemUsages = (rows: any): fileSystemUsage[] => {
  const usagesResponse: fileSystemUsage[] = []
  rows.forEach((row: any) => {
    usagesResponse.push({
      region: row.region,
      cloudAccountId: row.cloudAccountId,
      accountType: row.accountType,
      email: row.email,
      orgId: row.orgId,
      numFilesystems: row.numFilesystems,
      totalProvisioned: row.totalProvisioned,
      clusterScheduled: row.clusterScheduled,
      hasIksVolumes: row.hasIksVolumes
    })
  })

  return usagesResponse
}

const buildBucketUsages = (rows: any): bucketUsage[] => {
  const usagesResponse: bucketUsage[] = []
  rows.forEach((row: any) => {
    usagesResponse.push({
      region: row.region,
      cloudAccountId: row.cloudAccountId,
      accountType: row.accountType,
      email: row.email,
      clusterScheduled: row.clusterScheduled,
      buckets: row.buckets,
      bucketSize: row.bucketSize,
      usedCapacity: row.usedCapacity
    })
  })

  return usagesResponse
}

const buildStorageQuota = (rows: any): storageQuota[] => {
  const usagesResponse: storageQuota[] = []
  rows.forEach((row: any) => {
    usagesResponse.push({
      cloudAccountId: row.cloudAccountId,
      accountType: row.cloudAccountType,
      bucketsQuota: row.bucketsQuota,
      filesizeQuotaInTB: row.filesizeQuotaInTB,
      filevolumesQuota: row.filevolumesQuota,
      reason: row.reason
    })
  })

  return usagesResponse
}

const buildStorageDefaultQuota = (rows: any): storageDefaultQuota[] => {
  const defaultResponse: storageDefaultQuota[] = []
  rows.forEach((row: any) => {
    defaultResponse.push({
      cloudAccountType: row.cloudAccountType,
      bucketsQuota: row.bucketsQuota,
      filesizeQuotaInTB: row.filesizeQuotaInTB,
      filevolumesQuota: row.filevolumesQuota
    })
  })

  return defaultResponse
}

const buildServiceArray = (rows: any): serviceResource[] => {
  const response: serviceResource[] = []
  rows.forEach((row: any) => {
    const item = buildServiceResource(row)
    response.push(item)
  })
  return response
}

const buildServiceResource = (row: any): serviceResource => {
  const serviceResourceItems: serviceResourceItem[] = []

  for (const index in row.serviceResources) {
    const item = { ...row.serviceResources[index] }
    const serviceResource: serviceResourceItem = {
      maxLimit: item.maxLimit,
      name: item.name,
      quotaUnit: item.quotaUnit
    }
    serviceResourceItems.push(serviceResource)
  }

  const response: serviceResource = {
    serviceId: row.serviceId,
    serviceName: row.serviceName,
    serviceResources: serviceResourceItems
  }
  return response
}

const buildResourceQuota = (item: any): serviceQuota => {
  const serviceResourceItems: serviceQuotaResource[] = []
  for (const index in item.serviceQuotaAllResources) {
    const itemQuota = { ...item.serviceQuotaAllResources[index] }
    const serviceQuotaItem = buildQuotaItem(itemQuota)
    serviceResourceItems.push(serviceQuotaItem)
  }

  const resourceQuota: serviceQuota = {
    serviceId: item.serviceId,
    serviceName: item.serviceName,
    serviceQuotaResources: serviceResourceItems
  }

  return resourceQuota
}

const buildResourceQuotaResource = (item: any): serviceQuota => {
  const serviceResourceItems: serviceQuotaResource[] = []
  for (const index in item.serviceQuotaResource) {
    const itemQuota = { ...item.serviceQuotaResource[index] }
    const serviceQuotaItem = buildQuotaItem(itemQuota)
    serviceResourceItems.push(serviceQuotaItem)
  }

  const resourceQuota: serviceQuota = {
    serviceId: item.serviceId,
    serviceName: item.serviceName,
    serviceQuotaResources: serviceResourceItems
  }

  return resourceQuota
}

const buildQuotaItem = (item: any): serviceQuotaResource => {
  const serviceQuotaItem: serviceQuotaResource = {
    ruleId: item.ruleId,
    resourceType: item.resourceType,
    reason: item.reason,
    maxLimit: item.quotaConfig.limits,
    quotaUnit: item.quotaConfig.quotaUnit,
    scopeType: item.scope.scopeType,
    scopeValue: item.scope.scopeValue
  }
  return serviceQuotaItem
}

const useStorageManagementStore = create<StorageManagementStore>()((set, get) => ({
  services: null,
  serviceResourceItem: null,
  serviceQuotaResource: null,
  serviceQuota: null,
  fileSystemUsages: null,
  bucketUsages: null,
  storageQuotas: null,
  storageDefaultQuotas: null,
  editStorageQuota: null,
  loading: false,
  getStorageUsages: async (isBackGround: boolean) => {
    try {
      if (!isBackGround) {
        set({ loading: true })
      }
      const { data } = await StorageManagementService.getUsages()
      const fileSystemUsages = buildFileSystemUsages(data.filesystemUsages)
      const bucketUsages = buildBucketUsages(data.bucketUsages)
      set({ bucketUsages })
      set({ fileSystemUsages })
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  getQuotaAssigments: async (isBackGround: boolean) => {
    try {
      if (!isBackGround) {
        set({ loading: true })
      }
      const { data } = await StorageManagementService.getQuota()
      const storageQuotas = buildStorageQuota(data.storageQuotaByAccount)
      const storageDefaultQuotas = buildStorageDefaultQuota(data.defaultQuotaSection)
      set({ storageQuotas })
      set({ storageDefaultQuotas })
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setEditStorageQuota: (storageQuota: any) => {
    if (storageQuota) {
      const storageQuotaItem: storageQuota = {
        accountType: storageQuota.accountType,
        cloudAccountId: storageQuota.cloudAccountId,
        bucketsQuota: storageQuota.bucketsQuota,
        filesizeQuotaInTB: storageQuota.filesizeQuotaInTB,
        filevolumesQuota: storageQuota.filevolumesQuota,
        reason: storageQuota.reason
      }
      set({ editStorageQuota: storageQuotaItem })
    } else {
      set({ editStorageQuota: null })
    }
  },
  getServices: async (isBackGround: boolean) => {
    try {
      if (isBackGround) {
        set({ loading: true })
      }
      const { data } = await StorageManagementService.getServices()
      const services = buildServiceArray(data.services)
      set({ services })
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  getServicesById: async (idService: string) => {
    try {
      set({ loading: true })
      const { data } = await StorageManagementService.getServiceById(idService)
      const service = buildServiceResource(data)
      set({ serviceResourceItem: service })
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  getServiceQuotas: async (idService: string) => {
    try {
      set({ loading: true })
      const { data } = await StorageManagementService.getServiceQuotaById(idService)
      const serviceQuota = buildResourceQuota(data)
      set({ serviceQuota })
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  getServiceQuota: async (idService: string, resourceName: string) => {
    try {
      set({ loading: true })
      const { data } = await StorageManagementService.getServiceQuota(idService, resourceName)
      const serviceQuotaResource = buildResourceQuotaResource(data)
      set({ serviceQuotaResource })
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  resetServiceQuotaResource: () => {
    set({ serviceQuotaResource: null })
  }
}))

export default useStorageManagementStore
