// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { AxiosInstance } from '../utils/AxiosInstance'
import useUserStore from '../store/userStore/UserStore'
import idcConfig from '../config/configurator'

class LoadBalancerService {
  // Method to retrieve all allocations by cloud account

  getLoadBalancers() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/loadbalancers`
    return AxiosInstance.get(route)
  }

  createLoadBalancer(payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/loadbalancers`
    return AxiosInstance.post(route, payload)
  }

  editLoadBalancer(payload, resourceId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/loadbalancers/id/${resourceId}`
    return AxiosInstance.put(route, payload)
  }

  deleteLoadBalancer(resourceId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/loadbalancers/id/${resourceId}`
    return AxiosInstance.delete(route, { data: {} })
  }
}

export default new LoadBalancerService()
