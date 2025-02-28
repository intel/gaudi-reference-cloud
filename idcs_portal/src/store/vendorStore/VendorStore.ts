// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import PublicService from '../../services/PublicService'
import { IDCVendorFamilies } from '../../utils/Enums'

interface VendorFamily {
  name: string
  id: string
  description: string
}

interface Vendor {
  name: string
  id: string
  families: VendorFamily[]
}

interface VendorStore {
  vendors: Vendor[] | null
  getIDCVendor: () => Promise<Vendor>
  getComputeVendorFamily: () => Promise<VendorFamily | undefined>
  getNetworkVendorFamily: () => Promise<VendorFamily | undefined>
  getStorageVendorFamily: () => Promise<VendorFamily | undefined>
  getTrainingVendorFamily: () => Promise<VendorFamily | undefined>
  getSoftwareVendorFamily: () => Promise<VendorFamily | undefined>
  getSuperComputerVendorFamily: () => Promise<VendorFamily | undefined>
  getKubernetesVendorFamily: () => Promise<VendorFamily | undefined>
  getGuiVendorFamily: () => Promise<VendorFamily | undefined>
  getDpaiVendorFamily: () => Promise<VendorFamily | undefined>
  getMaaSVendorFamily: () => Promise<VendorFamily | undefined>
  getPaymentVendorFamily: () => Promise<VendorFamily | undefined>
  getLabsVendorFamily: () => Promise<VendorFamily | undefined>
}

const useVendorStore = create<VendorStore>()((set, get) => ({
  vendors: null,
  getIDCVendor: async () => {
    const vendors = get().vendors
    const IdcVendorName = 'idc'
    if (vendors === null) {
      const response = await PublicService.getVendorCatalog()
      const newVendors = response.data.vendors
      set({ vendors: [...newVendors] })
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
  },
  getDpaiVendorFamily: async () => {
    const idcVendor = await get().getIDCVendor()
    return idcVendor.families.find((x) => x.name === IDCVendorFamilies.Dpai)
  },
  getMaaSVendorFamily: async () => {
    const idcVendor = await get().getIDCVendor()
    return idcVendor.families.find((x) => x.name === IDCVendorFamilies.MaaS)
  },
  getPaymentVendorFamily: async () => {
    const idcVendor = await get().getIDCVendor()
    return idcVendor.families.find((x) => x.name === IDCVendorFamilies.payment)
  },
  getLabsVendorFamily: async () => {
    const idcVendor = await get().getIDCVendor()
    return idcVendor.families.find((x) => x.name === IDCVendorFamilies.labs)
  }
}))

export default useVendorStore
