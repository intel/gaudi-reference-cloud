import { AxiosInstance } from '../utility/axios/AxiosInstance'
import idcConfig from '../config/configurator'

const apiURLGlobal = idcConfig.REACT_APP_API_GLOBAL_SERVICE
class UserManagementService {
  getUsers() {
    const route = `${apiURLGlobal}/cloudaccounts`
    return AxiosInstance.get(route)
  }

  getCreditUsages(cloudAccountNumber) {
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/billing/usages?cloudAccountId=${cloudAccountNumber}`
    return AxiosInstance.get(route)
  }

  getCloudCredits(cloudAccountNumber) {
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudcredits/credit?cloudAccountId=${cloudAccountNumber}&history=true`
    return AxiosInstance.get(route)
  }

  updateUser(id, isBlocked) {
    const payload = {
      delinquent: isBlocked
    }
    const route = `${apiURLGlobal}/cloudaccounts/id/${id}`
    return AxiosInstance.patch(route, payload)
  }
}

export default new UserManagementService()
