import { AxiosInstance } from '../utility/axios/AxiosInstance'
import idcConfig from '../config/configurator'

const apiURLGlobal = idcConfig.REACT_APP_API_GLOBAL_SERVICE

class CloudAccountService {
  getCloudAccounts() {
    const route = `${apiURLGlobal}/cloudaccounts`
    return AxiosInstance.get(route)
  }

  // Method to retrieve all allocations by cloud account.
  getCloudAccountDetailsById (cloudAccount) {
    const route = `${apiURLGlobal}/cloudaccounts/id/${cloudAccount}`
    return AxiosInstance.get(route)
  }

  // Method to fetch the instances based specific cloudAccountNumber.
  getInstancesByCloudAccount(cloudAccountNumber) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/instances`
    return AxiosInstance.get(route)
  }

  // Method to delete instance group by specific name.
  deleteInstanceGroupByName(resourceName, cloudAccountNumber) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/instancegroups/name/${resourceName}`
    return AxiosInstance.delete(route, { data: {} })
  }

  deleteComputeReservation(resourceId, cloudAccountNumber) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/instances/id/${resourceId}`
    return AxiosInstance.delete(route, { data: {} })
  }

  getCloudAccountDetailsByName (cloudAccount) {
    const route = `${apiURLGlobal}/cloudaccounts/name/${cloudAccount}`
    return AxiosInstance.get(route)
  }

  // Instance Groups services.
  getInstanceGroupsById(cloudAccountNumber) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/instancegroups`
    return AxiosInstance.get(route)
  }

  updateCloudAccount(cloudAccount, payload) {
    const route = `${apiURLGlobal}/cloudaccounts/id/${cloudAccount}`
    return AxiosInstance.patch(route, payload)
  }
}

export default new CloudAccountService()
