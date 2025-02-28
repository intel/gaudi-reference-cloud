import { AxiosInstance } from '../utility/axios/AxiosInstance'
import idcConfig from '../config/configurator'

const apiURLGlobal = idcConfig.REACT_APP_API_GLOBAL_SERVICE

class BillingService {
  getCloudCredits() {
    return AxiosInstance.get('credits')
  }

  submitCloudCredits(payload) {
    return AxiosInstance.post('credits', payload)
  }

  submitCoupons(payload) {
    const route = `${apiURLGlobal}/cloudcredits/coupons`
    return AxiosInstance.post(route, payload)
  }
}

export default new BillingService()
