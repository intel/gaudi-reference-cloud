export interface fileSystemUsage {
  orgId: string
  region: string
  cloudAccountId: string
  accountType: string
  email: string
  numFilesystems: string
  totalProvisioned: string
  clusterScheduled: string
  hasIksVolumes: string
}

export interface bucketUsage {
  accountType: string
  bucketSize: string
  buckets: string
  cloudAccountId: string
  clusterScheduled: string
  email: string
  region: string
  usedCapacity: string
}

export interface storageQuota {
  accountType: string
  cloudAccountId: string
  reason: string
  filesizeQuotaInTB: string
  filevolumesQuota: string
  bucketsQuota: string
}

export interface storageDefaultQuota {
  cloudAccountType: string
  filesizeQuotaInTB: string
  filevolumesQuota: string
  bucketsQuota: string
}

export interface serviceQuota {
  serviceId: string
  serviceName: string
  serviceQuotaResources: serviceQuotaResource[] | []
}

export interface serviceResource {
  serviceId: string
  serviceName: string
  serviceResources: serviceResourceItem[] | []
}

export interface serviceResourceItem {
  name: string
  quotaUnit: string
  maxLimit: string
}

export interface serviceQuotaResource {
  ruleId: string
  resourceType: string
  quotaUnit: string
  maxLimit: string
  scopeType: string
  scopeValue: string
  reason: string
}
