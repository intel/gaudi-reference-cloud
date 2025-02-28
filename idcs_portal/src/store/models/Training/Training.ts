// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

export default interface Training {
  id: string
  familyId: string
  vendorId: string
  created: string
  name: string
  description: string
  category: string
  gettingStarted: string
  audience: string
  expectations: string
  overview: string
  prerrequisites: string
  shortDesc: string
  displayCatalogDesc: string
  displayName: string
  familyDisplayDescription: string
  familyDisplayName: string
  launch: string
  region: string
  service: string
  eccn: string
  pcq: string
  matchExpr: string
  accountType: string
  unit: string
  rate: string
  usageExpr: string
  imageSource: string
  displayInHomepage: boolean
  homepageDisplayGroup: string
  featuredSoftware: string | null
}
