import idcConfig from '../config/configurator'
import { AxiosInstance } from '../utility/axios/AxiosInstance'

const apiURLGlobal = idcConfig.REACT_APP_API_GLOBAL_SERVICE

class RegionManagementService {
  getRegions(filter) {
    const payload = filter || { data: {} }
    const route = `${apiURLGlobal}/regions/admin`
    return AxiosInstance.post(route, payload)
  }

  createRegion(payload) {
    const route = `${apiURLGlobal}/regions/add`
    return AxiosInstance.post(route, payload)
  }

  updateRegion(name, payload) {
    const route = `${apiURLGlobal}/regions/${name}`
    return AxiosInstance.put(route, payload)
  }

  deleteRegion(name) {
    const route = `${apiURLGlobal}/regions/${name}`
    return AxiosInstance.delete(route, { data: {} })
  }

  getAccountWhitelist() {
    const route = `${apiURLGlobal}/regions/acl`
    return AxiosInstance.get(route)
  }

  postAcl(payload) {
    const route = `${apiURLGlobal}/regions/acl/add`
    return AxiosInstance.post(route, payload)
  }

  deleteAcl(cloudaccountId, regionName) {
    const route = `${apiURLGlobal}/regions/acl/delete?cloudaccountId=${cloudaccountId}&regionName=${regionName}`
    return AxiosInstance.delete(route, { data: {} })
  }
}

export default new RegionManagementService()
