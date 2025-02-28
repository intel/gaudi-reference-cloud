// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import type MediaFile from '../MediaFile/MedialFile'

export default interface Software {
  id: string
  familyId: string
  vendorId: string
  created: string
  name: string
  description: string
  documentation: string
  access: string
  billingEnable: string
  category: string
  components: string
  downloadURL: string
  features: string
  demoURL: string
  helpURL: string
  licenseURL: string
  jupyterlab: string
  audience: string
  overview: string
  productURL: string
  shortDesc: string
  useCases: [string]
  displayCatalogDesc: string
  displayInHomepage: boolean
  displayName: string
  imageSource: string
  familyDisplayDescription: string
  familyDisplayName: string
  homepageDisplayGroup: string
  platforms: string
  region: string
  service: string
  eccn: string
  pcq: string
  matchExpr: string
  accountType: string
  unit: string
  usageUnit: string
  rate: string
  usageExpr: string
  status: string
  launchImage: string
  launchDownloadUrl: string
  launchLink: string
  launch: string
  mediaArray: MediaFile[]
  model: string
}
