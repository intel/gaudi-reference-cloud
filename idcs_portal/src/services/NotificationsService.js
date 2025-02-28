// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import idcConfig from '../config/configurator'
import useUserStore from '../store/userStore/UserStore'
import { AxiosInstance } from '../utils/AxiosInstance'

class NotificationsService {
  getNotifications() {
    const cloudAccountNumber = useUserStore.getState().user?.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/billing/events/all?cloudAccountId=${cloudAccountNumber}`
    return AxiosInstance.get(route)
  }
}

export default new NotificationsService()
