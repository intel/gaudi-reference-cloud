// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { AxiosInstance } from '../utils/AxiosInstance'
import useUserStore from '../store/userStore/UserStore'
import idcConfig from '../config/configurator'
import useVendorStore from '../store/vendorStore/VendorStore'
import catalogDescriptions from '../config/catalogDescriptions.json'

class PublicService {
  async getHardwareCatalog() {
    const user = useUserStore.getState().user
    const computeVendorFamily = await useVendorStore.getState().getComputeVendorFamily()

    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: computeVendorFamily.id,
        region: idcConfig.REACT_APP_SELECTED_REGION
      }
    }

    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`

    return AxiosInstance.post(route, payload)
  }

  async getTrainingCatalog() {
    const user = useUserStore.getState().user
    const trainingVendorFamily = await useVendorStore.getState().getTrainingVendorFamily()

    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: trainingVendorFamily.id,
        region: idcConfig.REACT_APP_SELECTED_REGION
      }
    }

    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`

    return AxiosInstance.post(route, payload)
  }

  async getLabsCatalog() {
    const user = useUserStore.getState().user
    const labsVendorFamily = await useVendorStore.getState().getLabsVendorFamily()

    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: labsVendorFamily.id,
        region: idcConfig.REACT_APP_SELECTED_REGION
      }
    }

    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`

    return AxiosInstance.post(route, payload)
  }

  async getTrainingDetail(id) {
    const user = useUserStore.getState().user
    const trainingVendorFamily = await useVendorStore.getState().getTrainingVendorFamily()
    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: trainingVendorFamily.id,
        region: idcConfig.REACT_APP_SELECTED_REGION,
        id
      }
    }

    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`

    return AxiosInstance.post(route, payload)
  }

  async getSoftwareCatalog() {
    const user = useUserStore.getState().user
    const softwareVendorFamily = await useVendorStore.getState().getSoftwareVendorFamily()
    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: softwareVendorFamily.id,
        region: idcConfig.REACT_APP_SELECTED_REGION
      }
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`
    return AxiosInstance.post(route, payload)
  }

  async getDpaiCatalog() {
    const user = useUserStore.getState().user
    const dpaiVendorFamily = await useVendorStore.getState().getDpaiVendorFamily()
    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: dpaiVendorFamily.id,
        region: idcConfig.REACT_APP_SELECTED_REGION
      }
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`
    return AxiosInstance.post(route, payload)
  }

  async getMaaSCatalog() {
    const user = useUserStore.getState().user
    const maaSVendorFamily = await useVendorStore.getState().getMaaSVendorFamily()
    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: maaSVendorFamily.id,
        region: idcConfig.REACT_APP_SELECTED_REGION
      }
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`
    return AxiosInstance.post(route, payload)
  }

  async getSoftwareDetail(id) {
    const user = useUserStore.getState().user
    const softwareVendorFamily = await useVendorStore.getState().getSoftwareVendorFamily()
    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: softwareVendorFamily.id,
        id,
        region: idcConfig.REACT_APP_SELECTED_REGION
      }
    }

    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`

    return AxiosInstance.post(route, payload)
  }

  async getDpaiDetail(id) {
    const user = useUserStore.getState().user
    const softwareVendorFamily = await useVendorStore.getState().getDpaiVendorFamily()
    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: softwareVendorFamily.id,
        id,
        region: idcConfig.REACT_APP_SELECTED_REGION
      }
    }

    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`

    return AxiosInstance.post(route, payload)
  }

  async getMassDetail(id) {
    const user = useUserStore.getState().user
    const softwareVendorFamily = await useVendorStore.getState().getMaaSVendorFamily()
    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: softwareVendorFamily.id,
        id,
        region: idcConfig.REACT_APP_SELECTED_REGION
      }
    }

    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`

    return AxiosInstance.post(route, payload)
  }

  async getStorageCatalog() {
    const user = useUserStore.getState().user
    const storageVendorFamily = await useVendorStore.getState().getStorageVendorFamily()

    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: storageVendorFamily.id,
        region: idcConfig.REACT_APP_SELECTED_REGION
      }
    }

    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`

    return AxiosInstance.post(route, payload)
  }

  async getSuperComputerCatalog() {
    const user = useUserStore.getState().user
    const superComuterVendorFamily = await useVendorStore.getState().getSuperComputerVendorFamily()
    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: superComuterVendorFamily.id,
        region: idcConfig.REACT_APP_SELECTED_REGION
      }
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`
    return AxiosInstance.post(route, payload)
  }

  async getKubernetesCatalog() {
    const user = useUserStore.getState().user
    const kubernetesVendorFamily = await useVendorStore.getState().getKubernetesVendorFamily()
    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: kubernetesVendorFamily.id,
        region: idcConfig.REACT_APP_SELECTED_REGION
      }
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`
    return AxiosInstance.post(route, payload)
  }

  async getGuiCatalog() {
    const user = useUserStore.getState().user
    const guiVendorFamily = await useVendorStore.getState().getGuiVendorFamily()
    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: guiVendorFamily.id
      }
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`
    return AxiosInstance.post(route, payload)
  }

  async getPaymentServicesCatalog() {
    const user = useUserStore.getState().user
    const paymentVendorFamily = await useVendorStore.getState().getPaymentVendorFamily()
    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: paymentVendorFamily.id
      }
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`
    return AxiosInstance.post(route, payload)
  }

  async getNetworkCatalog() {
    const user = useUserStore.getState().user
    const networkVendorFamily = await useVendorStore.getState().getNetworkVendorFamily()

    const payload = {
      cloudaccountId: user.cloudAccountNumber,
      productFilter: {
        accountType: user.cloudAccountType,
        familyId: networkVendorFamily.id,
        region: idcConfig.REACT_APP_SELECTED_REGION
      }
    }

    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/api/products`

    return AxiosInstance.post(route, payload)
  }

  getInstanceTypes() {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/instancetypes`
    return AxiosInstance.get(route)
  }

  getMachineImages() {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/machineimages`
    return AxiosInstance.get(route)
  }

  getVendorCatalog() {
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/vendors`

    return AxiosInstance.get(route)
  }

  getMachineImage(imageName) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/machineimages/${imageName}`
    return AxiosInstance.get(route)
  }

  getCatalogDescription(vendor, exactMatch) {
    const vendorMatches = catalogDescriptions[vendor] ?? []
    const match = vendorMatches.find((x) => x.match === exactMatch)
    return match ? match.description : ' '
  }

  getCatalogShortDescription(vendor, exactMatch) {
    const vendorMatches = catalogDescriptions[vendor] ?? []
    const match = vendorMatches.find((x) => x.match === exactMatch)
    return match ? match.shortDescription : ''
  }

  getCatalogOrder(vendor, exactMatch) {
    const vendorMatches = catalogDescriptions[vendor] ?? []
    const match = vendorMatches.find((x) => x.match === exactMatch)
    return match ? match.order : 0
  }
}

export default new PublicService()
