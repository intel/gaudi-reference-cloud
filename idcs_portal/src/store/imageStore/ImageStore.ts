// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import PublicService from '../../services/PublicService'
import { type ComponentsOs } from '../models/imageOs/ImageOs'
import type ImageOS from '../models/imageOs/ImageOs'

export interface ImageStore {
  ImagesOs: ImageOS[] | []
  loading: boolean
  setImagesOs: () => Promise<void>
  image: ImageOS | null
  setImage: (imageName: string) => Promise<void>
  reset: () => void
}

const initialState = {
  ImagesOs: [],
  loading: false,
  image: null
}

const useImageStore = create<ImageStore>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setImagesOs: async () => {
    set({ loading: true })

    const response = await PublicService.getMachineImages()

    const imagesOss: ImageOS[] = []

    const imageItems = { ...response.data.items }

    for (const index in imageItems) {
      const machineItem = { ...imageItems[index] }
      imagesOss.push(buildMachineImage(machineItem))
    }

    set({ ImagesOs: imagesOss })
    set({ loading: false })
  },
  setImage: async (imageName: string) => {
    set({ loading: true })

    const response = await PublicService.getMachineImage(imageName)

    const machineItem = { ...response.data }
    const image: ImageOS = buildMachineImage(machineItem)

    set({ image })
    set({ loading: false })
  }
}))

const buildMachineImage = (machineItem: any): ImageOS => {
  const components = { ...machineItem.spec.components }
  const machineComponents: ComponentsOs[] = []

  for (const indexComponent in components) {
    const component = { ...components[indexComponent] }

    const componentItem: ComponentsOs = {
      name: component.name,
      type: component.type,
      version: component.version,
      description: component.description,
      infoUrl: component.infoUrl,
      imageSource: component.imageUrl
    }
    machineComponents.push(componentItem)
  }

  const spec = { ...machineItem.spec }

  const labels = { ...spec.labels }

  return {
    name: machineItem.metadata.name,
    displayName: spec.displayName,
    description: spec.description,
    imageSource: spec.icon,
    instanceCategories: spec.instanceCategories,
    instanceTypes: spec.instanceTypes,
    md5sum: spec.md5sum,
    sha256sum: spec.sha256sum,
    sha512sum: spec.sha512sum,
    architecture: labels.architecture,
    family: labels.family,
    imageCategories: spec.imageCategories,
    components: machineComponents
  }
}

export default useImageStore
