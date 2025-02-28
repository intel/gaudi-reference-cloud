import { AxiosInstance } from '../utility/axios/AxiosInstance'
import idcConfig from '../config/configurator'

const apiURLGlobal = idcConfig.REACT_APP_API_GLOBAL_SERVICE

class PublicService {
  // Method to retrieve all allocations.

  getProductCatalog() {
    const route = `${apiURLGlobal}/products/admin`
    return AxiosInstance.post(route, {})
  }

  getFamilies() {
    const route = `${apiURLGlobal}/families`
    return AxiosInstance.get(route)
  }

  getVendorCatalog() {
    const route = `${apiURLGlobal}/vendors`
    return AxiosInstance.get(route)
  }

  getCatalogFamilies() {
    const route = `${apiURLGlobal}/families`
    return AxiosInstance.get(route)
  }

  deleteCatalogFamily(name) {
    const route = `${apiURLGlobal}/families/${name}`
    return AxiosInstance.delete(route, { data: {} })
  }

  createCatalogFamily(payload) {
    const route = `${apiURLGlobal}/families`
    return AxiosInstance.post(route, payload)
  }

  putCatalogFamily(name, payload) {
    const route = `${apiURLGlobal}/families/${name}`
    return AxiosInstance.put(route, payload)
  }

  deleteVendorCatalog(name) {
    const route = `${apiURLGlobal}/vendors/${name}`
    return AxiosInstance.delete(route, { data: {} })
  }

  createVendorCatalog(payload) {
    const route = `${apiURLGlobal}/vendors`
    return AxiosInstance.post(route, payload)
  }

  putVendorCatalog(name, payload) {
    const route = `${apiURLGlobal}/vendors/${name}`
    return AxiosInstance.put(route, payload)
  }

  getProducts() {
    const route = `${apiURLGlobal}/products/details`
    return AxiosInstance.get(route)
  }

  getProductByFamilyName(familyName) {
    const route = `${apiURLGlobal}/products/details?product_family_name=${familyName}`
    return AxiosInstance.get(route)
  }

  getProductByName(productName) {
    const route = `${apiURLGlobal}/products/details?name=${productName}`
    return AxiosInstance.get(route)
  }

  getRegions() {
    const route = `${apiURLGlobal}/regions/admin`
    return AxiosInstance.post(route, {})
  }

  getProductServices() {
    const route = `${apiURLGlobal}/intelcloudserviceregistration`
    return AxiosInstance.get(route)
  }

  postProduct(payload) {
    const route = `${apiURLGlobal}/products/add`
    return AxiosInstance.post(route, payload)
  }

  putProduct(payload, productId) {
    const route = `${apiURLGlobal}/products/update/${productId}`
    return AxiosInstance.put(route, payload)
  }
}

export default new PublicService()
