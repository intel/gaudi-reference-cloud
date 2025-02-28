import { AxiosInstance } from '../utility/axios/AxiosInstance'
import idcConfig from '../config/configurator'

const apiURLGlobal = idcConfig.REACT_APP_API_GLOBAL_SERVICE

class CouponService {
  // Method to retrieve all coupons by cloud account.
  getCoupons() {
    const route = `${apiURLGlobal}/cloudcredits/coupons`
    return AxiosInstance.get(route)
  }

  // Method to disable the specific coupon.
  disableCoupon(payload) {
    const route = `${apiURLGlobal}/cloudcredits/coupons/disable`
    return AxiosInstance.post(route, payload)
  }
}

export default new CouponService()
