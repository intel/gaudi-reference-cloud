// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

export default interface ImageOS {
  name: string
  displayName: string
  description: string
  imageSource: string
  instanceCategories: string
  instanceTypes: string
  md5sum: string
  sha256sum: string
  sha512sum: string
  architecture: string
  family: string
  imageCategories: string
  components: ComponentsOs[]
}

export interface ComponentsOs {
  name: string
  type: string
  version: string
  description: string
  infoUrl: string
  imageSource: string
}
