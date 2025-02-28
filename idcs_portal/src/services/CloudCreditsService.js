// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { AxiosInstance } from '../utils/AxiosInstance'
import useUserStore from '../store/userStore/UserStore'
import idcConfig from '../config/configurator'

class CloudCreditsService {
  getCredits() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudcredits/credit?cloudAccountId=${cloudAccountNumber}&history=true`
    return AxiosInstance.get(route)
  }

  getResumeCredits() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudcredits/credit?cloudAccountId=${cloudAccountNumber}`
    return AxiosInstance.get(route)
  }

  getCreditOptions() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/billing/options?cloudAccountId=${cloudAccountNumber}`
    return AxiosInstance.get(route)
  }

  postCredit(payload) {
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudcredits/coupons/redeem`
    return AxiosInstance.post(route, payload)
  }
}

export default new CloudCreditsService()
