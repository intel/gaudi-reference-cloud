// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { AxiosInstance } from '../utils/AxiosInstance'
import useUserStore from '../store/userStore/UserStore'
import idcConfig from '../config/configurator'

class AuthorizationService {
  getResources() {
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/authorization/resources`
    return AxiosInstance.get(route)
  }

  getCloudAccountRoles() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/authorization/cloudaccounts/${cloudAccountNumber}/roles`
    return AxiosInstance.get(route)
  }

  createCloudAccountRoles(payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/authorization/cloudaccounts/${cloudAccountNumber}/roles`
    return AxiosInstance.post(route, payload)
  }

  updateCloudAccountRoles(roleId, payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/authorization/cloudaccounts/${cloudAccountNumber}/roles/${roleId}`
    return AxiosInstance.put(route, payload)
  }

  getUserRoles(userId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/authorization/cloudaccounts/${cloudAccountNumber}/users/${userId}`
    return AxiosInstance.get(route)
  }

  addUserToRole(roleId, payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/authorization/cloudaccounts/${cloudAccountNumber}/roles/${roleId}/users`
    return AxiosInstance.post(route, payload)
  }

  removeUserFromRole(roleId, userId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/authorization/cloudaccounts/${cloudAccountNumber}/roles/${roleId}/users?userId=${userId}`
    return AxiosInstance.delete(route, { data: {} })
  }

  deleteRole(roleId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/authorization/cloudaccounts/${cloudAccountNumber}/roles/${roleId}`
    return AxiosInstance.delete(route, { data: {} })
  }

  addRolesToUser(userId, payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/authorization/cloudaccounts/${cloudAccountNumber}/users/${userId}/roles`
    return AxiosInstance.post(route, payload)
  }
}

export default new AuthorizationService()
