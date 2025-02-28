import { AxiosInstance } from '../utility/axios/AxiosInstance'
import idcConfig from '../config/configurator'

const apiURLGlobal = idcConfig.REACT_APP_API_GLOBAL_SERVICE
class SkuQuotaService {
  getSkuQuotas() {
    const route = `${apiURLGlobal}/products/acl`
    return AxiosInstance.get(route)
  }

  postAcl(payload) {
    const route = `${apiURLGlobal}/products/acl/add`
    return AxiosInstance.post(route, payload)
  }

  deleteAcl (cloudAccountId, productId, vendorId) {
    const route = `${apiURLGlobal}/products/acl/delete?cloudaccountId=${cloudAccountId}&productId=${productId}&vendorId=${vendorId}`
    return AxiosInstance.delete(route, { data: {} })
  }
}

export default new SkuQuotaService()
