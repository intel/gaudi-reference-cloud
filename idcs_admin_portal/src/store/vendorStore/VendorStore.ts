// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import PublicService from '../../services/PublicService'
import { IDCVendorFamilies } from '../../utility/Enums'

interface VendorFamily {
  name: string
  id: string
  description: string
}

interface Vendor {
  name: string
  id: string
  organizationName: string
  description: string
  families: VendorFamily[]
}

interface VendorStore {
  vendors: Vendor[] | null
  loading: boolean
  getVendor: () => Promise<Vendor>
  getIDCVendor: () => Promise<Vendor>
  getComputeVendorFamily: () => Promise<VendorFamily | undefined>
  getNetworkVendorFamily: () => Promise<VendorFamily | undefined>
  getStorageVendorFamily: () => Promise<VendorFamily | undefined>
  getTrainingVendorFamily: () => Promise<VendorFamily | undefined>
  getSoftwareVendorFamily: () => Promise<VendorFamily | undefined>
  getSuperComputerVendorFamily: () => Promise<VendorFamily | undefined>
  getKubernetesVendorFamily: () => Promise<VendorFamily | undefined>
  getGuiVendorFamily: () => Promise<VendorFamily | undefined>
}

const buildVendorResponse = (data: any): Vendor[] => {
  const response: Vendor[] = []

  data.forEach((item: any) => {
    response.push({
      name: item.name,
      id: item.id,
      description: item.description,
      organizationName: item.organization_name,
      families: item.families
        })
  })
  return response
}

const useVendorStore = create<VendorStore>()((set, get) => ({
  vendors: null,
  loading: false,
  getVendor: async () => {
    set({ loading: true })
      const response = await PublicService.getVendorCatalog()
      let newVendors: any = buildVendorResponse(response.data.vendors)
      newVendors = newVendors.sort((a: Vendor, b: Vendor) => a.name.localeCompare(b.name, undefined, { numeric: true }))
    set({ vendors: [...newVendors], loading: false })
    return newVendors
  },
  getIDCVendor: async () => {
    const vendors = get().vendors
    const IdcVendorName = 'idc'
    if (vendors === null) {
      const newVendors: any = await get().getVendor()
      return newVendors.find((x: Vendor) => x.name === IdcVendorName)
    } else {
      return vendors.find((x) => x.name === IdcVendorName)
    }
  },
  getComputeVendorFamily: async () => {
    const idcVendor = await get().getIDCVendor()
    return idcVendor.families.find((x) => x.name === IDCVendorFamilies.Compute)
  },
  getNetworkVendorFamily: async () => {
    const idcVendor = await get().getIDCVendor()
    return idcVendor.families.find((x) => x.name === IDCVendorFamilies.Network)
  },
  getStorageVendorFamily: async () => {
    const idcVendor = await get().getIDCVendor()
    return idcVendor.families.find((x) => x.name === IDCVendorFamilies.Storage)
  },
  getTrainingVendorFamily: async () => {
    const idcVendor = await get().getIDCVendor()
    return idcVendor.families.find((x) => x.name === IDCVendorFamilies.Training)
  },
  getSoftwareVendorFamily: async () => {
    const idcVendor = await get().getIDCVendor()
    return idcVendor.families.find((x) => x.name === IDCVendorFamilies.Software)
  },
  getSuperComputerVendorFamily: async () => {
    const idcVendor = await get().getIDCVendor()
    return idcVendor.families.find((x) => x.name === IDCVendorFamilies.SuperComputer)
  },
  getKubernetesVendorFamily: async () => {
    const idcVendor = await get().getIDCVendor()
    return idcVendor.families.find((x) => x.name === IDCVendorFamilies.Kubernetes)
  },
  getGuiVendorFamily: async () => {
    const idcVendor = await get().getIDCVendor()
    return idcVendor.families.find((x) => x.name === IDCVendorFamilies.UserInterface)
  }
}))

export default useVendorStore
