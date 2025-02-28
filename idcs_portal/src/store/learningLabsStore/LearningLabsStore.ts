// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import PublicService from '../../services/PublicService'
import LearningLabsService from '../../services/LearningLabsService'
import type LearningLabsProduct from '../models/LearningLabs/LearningLabsProduct'

export interface LearningLabsStore {
  learningLabsList: LearningLabsProduct[] | []
  loading: boolean
  generatedMessage: string | undefined
  setLearningLabsList: () => Promise<void>
  getMessageFromText: (params: any) => Promise<void>
  reset: () => void
  generatedImage: any
  getImageFromText: (params: any) => Promise<void>
}

const initialState = {
  learningLabsList: [],
  generatedMessage: undefined,
  loading: false,
  generatedImage: undefined
}

const useLearningLabsStore = create<LearningLabsStore>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setLearningLabsList: async () => {
    set({ loading: true })
    try {
      const response = await PublicService.getLabsCatalog()
      const learningLabDetails = { ...response.data.products }
      const learningLabsList: LearningLabsProduct[] = []
      for (const index in learningLabDetails) {
        const detail = { ...learningLabDetails[index] }
        const learningLab = buildData(detail)
        learningLabsList.push(learningLab)
      }
      set({ loading: false, learningLabsList })
    } catch (error) {
      set({ loading: false })
      console.error(error)
    }
  },
  getMessageFromText: async (params) => {
    set({ loading: true })
    try {
      const response = await LearningLabsService.getMessageFromText(params)
      if (response.data) {
        const generatedMessage = response.data.generated_text
        set({ generatedMessage })
      }
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  getImageFromText: async (params: any) => {
    set({ loading: true })
    try {
      const response = await LearningLabsService.getImageFromText(params)
      if (response.data) {
        const imageUrl = URL.createObjectURL(response.data)
        set({ generatedImage: imageUrl })
      }
      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  }
}))

const buildData = (response: any): LearningLabsProduct => {
  const metadata = { ...response.metadata }

  const learningLabs: LearningLabsProduct = {
    id: response.id,
    vendorId: response.vendorId,
    familyId: response.familyId,
    region: metadata.region,
    familyDisplayName: metadata['family.displayName'],
    displayName: metadata.displayName,
    displayCatalogDesc: metadata.displayCatalogDesc,
    launch: metadata.launch
  }

  return learningLabs
}

export default useLearningLabsStore
