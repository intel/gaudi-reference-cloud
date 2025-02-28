// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { AxiosInstance } from '../utils/AxiosInstance'
import useUserStore from '../store/userStore/UserStore'
import idcConfig from '../config/configurator'

class UsagesService {
  getUsages() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/billing/usages?cloudAccountId=${cloudAccountNumber}`
    return AxiosInstance.get(route)
  }
}

export default new UsagesService()
