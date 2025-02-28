// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

export interface Storage {
  name: string
  id: string
  created: string
  vendorId: string
  familyId: string
  description: string
  storageCategories: string
  access: string
  billingEnable: string
  category: string
  disableForAccountTypes: string
  minimumSize: number
  maximumSize: number
  displayName: string
  familyDisplayDescription: string
  familyDisplayName: string
  highlight: string
  information: string
  recommendedUseCase: string
  region: string
  releaseStatus: string
  service: string
  eccn: string
  pcq: string
  matchExpr: string
  accountType: string
  unit: string
  rate: number
  usageExpr: string
  status: string
  unitSize: string
  usageUnit: string
}

export interface ObjectStorage {
  name: string
  id: string
  created: string
  vendorId: string
  familyId: string
  description: string
  access: string
  billingEnable: string
  disableForAccountTypes: string
  displayName: string
  familyDisplayDescription: string
  familyDisplayName: string
  information: string
  instanceCategories: string
  instanceType: string
  region: string
  releaseStatus: string
  service: string
  eccn: string
  pcq: string
  matchExpr: string
  accountType: string
  unit: string
  rate: number
  usageExpr: string
  unitSize: string
  usageUnit: string
}
