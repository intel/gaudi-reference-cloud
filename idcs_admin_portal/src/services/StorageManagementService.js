import { AxiosInstance } from '../utility/axios/AxiosInstance'
import idcConfig from '../config/configurator'

class StorageManagementService {
  getUsages() {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/storage/admin/stats`
    return AxiosInstance.get(route)
  }

  postUsages(payload) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/storage/admin/storageQuota`
    return AxiosInstance.post(route, payload)
  }

  putUsages(payload) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/storage/admin/storageQuota`
    return AxiosInstance.put(route, payload)
  }

  getQuotaById(cloudAccountId) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/storage/admin/storageQuota/${cloudAccountId}`
    return AxiosInstance.get(route)
  }

  getQuota() {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/storage/admin/storageQuota`
    return AxiosInstance.get(route)
  }

  getServices() {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/quota/registrations`
    return AxiosInstance.get(route)
  }

  getServiceById(id) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/quota/service/${id}/serviceResources`
    return AxiosInstance.get(route)
  }

  getServiceQuotaById(id) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/quota/service/${id}`
    return AxiosInstance.get(route)
  }

  postService(payload) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/quota/service/register`
    return AxiosInstance.post(route, payload)
  }

  postServiceQuota(id, payload) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/quota/service/${id}/create`
    return AxiosInstance.post(route, payload)
  }

  putService(id, payload) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/quota/service/${id}/update`
    return AxiosInstance.put(route, payload)
  }

  getServiceQuota(id, resourceName) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/quota/service/${id}/resource/${resourceName}`
    return AxiosInstance.get(route)
  }

  putServiceQuota(id, resourceName, payload) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/quota/service/${id}/resource/${resourceName}/update`
    return AxiosInstance.put(route, payload)
  }

  deleteServiceQuota(id, resourceType, ruleId) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/quota/service/${id}/resource/${resourceType}/ruleId/${ruleId}`
    return AxiosInstance.delete(route, { data: {} })
  }

  deleteService(id) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/quota/service/${id}/delete`
    return AxiosInstance.delete(route, { data: {} })
  }
}

export default new StorageManagementService()
