// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import PublicService from '../../services/PublicService'

export interface Family {
  name: string
  id: string
  description: string
  vendor: string
}

interface FamilyStore {
  loading: boolean
  families: Family[] | null
  getFamilies: () => Promise<void>
}

const buildFamilyJson = (items: any): Family[] => {
  const response: Family[] = []

  items.forEach((item: any) => {
    response.push({
      id: item.id,
      name: item.name,
      description: item.description,
      vendor: item.vendor_name
    })
  })

  return response
}

const useFamilyStore = create<FamilyStore>()((set, get) => ({
  families: null,
  loading: false,
  getFamilies: async () => {
    set({ loading: true })
    const response = await PublicService.getCatalogFamilies()
    const familiesResponse = response.data.product_families
    const families = buildFamilyJson(familiesResponse).sort((a: any, b: any) => b.name - a.name)
    set({ families })
    set({ loading: false })
  }
}))

export default useFamilyStore
