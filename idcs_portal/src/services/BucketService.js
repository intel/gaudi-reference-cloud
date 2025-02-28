// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { AxiosInstance } from '../utils/AxiosInstance'
import useUserStore from '../store/userStore/UserStore'
import idcConfig from '../config/configurator'

class BucketService {
  // Method to retrieve all allocations by cloud account

  getObjectBucketsByCloudAccount() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/objects/buckets`
    return AxiosInstance.get(route)
  }

  deleteObjectBucketByCloudAccount(resourceId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/objects/buckets/id/${resourceId}`
    return AxiosInstance.delete(route, { data: {} })
  }

  postObjectStorageReservation(payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/objects/buckets`
    return AxiosInstance.post(route, payload)
  }

  getBucketUsersByCloudAccount() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/objects/users`
    return AxiosInstance.get(route)
  }

  postBucketUsers(payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/objects/users`
    return AxiosInstance.post(route, payload)
  }

  updateBucketUser(payload, userName) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/objects/users/name/${userName}/policy`
    return AxiosInstance.put(route, payload)
  }

  updateBucketUserCredentialsByCloudAccount(userName) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/objects/users/name/${userName}/credentials`
    return AxiosInstance.put(route, { data: {} })
  }

  deleteBucketUserByCloudAccount(userName) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/objects/users/name/${userName}`
    return AxiosInstance.delete(route, { data: {} })
  }

  postObjectBucketRule(payload, bucketId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/objects/buckets/id/${bucketId}/lifecyclerule`
    return AxiosInstance.post(route, payload)
  }

  deleteObjectBucketRule(resourceId, bucketId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/objects/buckets/id/${bucketId}/lifecyclerule/id/${resourceId}`
    return AxiosInstance.delete(route, { data: {} })
  }

  updateObjectBucketRule(payload, resourceId, bucketId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/objects/buckets/id/${bucketId}/lifecyclerule/id/${resourceId}`
    return AxiosInstance.put(route, payload)
  }
}

export default new BucketService()
