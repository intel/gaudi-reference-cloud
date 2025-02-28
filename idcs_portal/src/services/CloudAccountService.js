// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { AxiosInstance } from '../utils/AxiosInstance'
import useUserStore from '../store/userStore/UserStore'
import idcConfig from '../config/configurator'

class CloudAccountService {
  // Method to retrieve all allocations by cloud account

  getExpiryTraining() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_SLURM_SERVICE}/cloudaccounts/${cloudAccountNumber}/trainings/expiry`
    return AxiosInstance.post(route)
  }

  postEnrollTraining(payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_SLURM_SERVICE}/cloudaccounts/${cloudAccountNumber}/trainings`
    return AxiosInstance.post(route, payload)
  }

  getExpiry() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_SLURM_SERVICE}/cloudaccounts/${cloudAccountNumber}/trainings/expiry`
    return AxiosInstance.get(route)
  }

  getEnrollTraining(trainingId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_SLURM_SERVICE}/cloudaccounts/${cloudAccountNumber}/trainings/${trainingId}/users`
    return AxiosInstance.get(route)
  }

  getInstancesByCloudAccount() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/instances`
    return AxiosInstance.get(route)
  }

  getInstanceGroupInstances(instanceGroup) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/instances?metadata.instanceGroup=${instanceGroup}`
    return AxiosInstance.get(route)
  }

  getSshByCloud() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/sshpublickeys`
    return AxiosInstance.get(route)
  }

  postSshByCloud(payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/sshpublickeys`
    return AxiosInstance.post(route, payload)
  }

  deleteSshByCloud(name) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/sshpublickeys/name/${name}`
    return AxiosInstance.delete(route, { data: {} })
  }

  getMyVnets() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/vnets`
    return AxiosInstance.get(route)
  }

  enableCloudMonitorForBM() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/cloudmonitor/enable`
    return AxiosInstance.post(route, { data: {} })
  }

  postVnets(payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/vnets`
    return AxiosInstance.post(route, payload)
  }

  postComputeReservation(payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/instances`
    return AxiosInstance.post(route, payload)
  }

  postComputeGroupReservation(payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/instancegroups`
    return AxiosInstance.post(route, payload)
  }

  putComputeReservation(resourceId, payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/instances/id/${resourceId}`
    return AxiosInstance.put(route, payload)
  }

  deleteComputeReservation(resourceId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/instances/id/${resourceId}`
    return AxiosInstance.delete(route, { data: {} })
  }

  stopComputeReservation(machineId) {
    const payload = {
      spec: {
        runStrategy: 'Halted'
      }
    }
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/instances/id/${machineId}`
    return AxiosInstance.put(route, payload)
  }

  startComputeReservation(machineId) {
    const payload = {
      spec: {
        runStrategy: 'Always'
      }
    }
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/instances/id/${machineId}`
    return AxiosInstance.put(route, payload)
  }

  postStorageReservation(payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/filesystems`
    return AxiosInstance.post(route, payload)
  }

  putStorageReservation(resourceId, payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/filesystems/id/${resourceId}`
    return AxiosInstance.put(route, payload)
  }

  enrollCloudAccountDetails(shouldEnrollPremium, acceptTermAndConditions) {
    const payload = {
      premium: Boolean(shouldEnrollPremium),
      termsStatus: acceptTermAndConditions
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/enroll`
    return AxiosInstance.post(route, payload)
  }

  getInstanceGroupsByCloudAccount() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/instancegroups`
    return AxiosInstance.get(route)
  }

  putComputeGroupReservation(resourceName, payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/instancegroups/name/${resourceName}`
    return AxiosInstance.put(route, payload)
  }

  deleteInstanceGroupByName(resourceName) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/instancegroups/name/${resourceName}`
    return AxiosInstance.delete(route, { data: {} })
  }

  getStoragesByCloudAccount() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/filesystems?metadata.filterType=ComputeGeneral`
    return AxiosInstance.get(route)
  }

  deleteStorageByCloudAccount(resourceId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/filesystems/id/${resourceId}`
    return AxiosInstance.delete(route, { data: {} })
  }

  getStorageCredentialsByCloudAccount(fileName) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/filesystems/name/${fileName}/user`
    return AxiosInstance.get(route)
  }

  upgradeCloudAccountByCoupon(couponCode) {
    const payload = {
      cloudAccountId: useUserStore.getState().user.cloudAccountNumber,
      cloudAccountUpgradeToType: 'ACCOUNT_TYPE_PREMIUM',
      code: couponCode
    }

    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/upgrade`
    return AxiosInstance.post(route, payload)
  }

  upgradeCloudAccountByCreditCard() {
    const payload = {
      cloudAccountId: useUserStore.getState().user.cloudAccountNumber,
      cloudAccountUpgradeToType: 'ACCOUNT_TYPE_PREMIUM'
    }

    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/upgradecc`
    return AxiosInstance.post(route, payload)
  }

  getUserCloudAccountsList(email) {
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/name/${email}/members?onlyActive=1`
    return AxiosInstance.get(route)
  }

  rejectUserCloudAccount(adminCloudAccountId, invitationState) {
    const payload = {
      adminAccountId: adminCloudAccountId,
      invitationState,
      memberEmail: useUserStore.getState().user.email
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/invitations/member/reject`
    return AxiosInstance.post(route, payload)
  }

  checkInviteCodeForUserCloudAccount(adminCloudAccountId, inviteCode) {
    const payload = {
      adminCloudAccountId,
      inviteCode,
      memberEmail: useUserStore.getState().user.email
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/validateinvitecode`
    return AxiosInstance.post(route, payload)
  }

  createOtp(memberEmail) {
    const payload = {
      cloudAccountId: useUserStore.getState().user.cloudAccountNumber,
      memberEmail
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/otp/create`
    return AxiosInstance.post(route, payload)
  }

  verifyOtp(memberEmail, otpCode) {
    const payload = {
      cloudAccountId: useUserStore.getState().user.cloudAccountNumber,
      memberEmail,
      otpCode
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/otp/verify`
    return AxiosInstance.post(route, payload)
  }

  resendOtp(memberEmail) {
    const payload = {
      cloudAccountId: useUserStore.getState().user.cloudAccountNumber,
      memberEmail
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/otp/resend`
    return AxiosInstance.post(route, payload)
  }

  multiUserAdminInvitationList() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route =
      `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/invitations/read?adminAccountId=` + cloudAccountNumber

    return AxiosInstance.get(route)
  }

  createInvite(invitePayload) {
    const payload = {
      cloudAccountId: useUserStore.getState().user.cloudAccountNumber,
      invites: invitePayload
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/invitations/create`
    return AxiosInstance.post(route, payload)
  }

  resendInvite(memberEmail) {
    const payload = {
      adminAccountId: useUserStore.getState().user.cloudAccountNumber,
      memberEmail
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/invitations/resend`
    return AxiosInstance.post(route, payload)
  }

  memberNotification(adminAccountId) {
    const payload = {
      adminAccountId,
      memberEmail: useUserStore.getState().user.email
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/sendinvitecode`
    return AxiosInstance.post(route, payload)
  }

  removePendingInvite(memberEmail, invitationState) {
    const payload = {
      adminAccountId: useUserStore.getState().user.cloudAccountNumber,
      invitationState,
      memberEmail
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/invitations/revoke`
    return AxiosInstance.post(route, payload)
  }

  removeAcceptedInvite(memberEmail, invitationState) {
    const payload = {
      adminAccountId: useUserStore.getState().user.cloudAccountNumber,
      invitationState,
      memberEmail
    }
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/invitations/remove`
    return AxiosInstance.post(route, payload)
  }

  getMetricsQueryData(resourceId, payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/cloudmonitor/resources/${resourceId}/query`
    return AxiosInstance.post(route, payload)
  }

  getUserCredentials() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/user/credentials/${cloudAccountNumber}/list`
    return AxiosInstance.get(route)
  }

  postUserCredentials(payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/user/credentials/${cloudAccountNumber}/create`
    return AxiosInstance.post(route, payload)
  }

  getUserCredentialAccessToken(clientId, clientSecret) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/user/credentials/${cloudAccountNumber}/token?clientId=${clientId}&clientSecret=${clientSecret}`
    return AxiosInstance.get(route)
  }

  deleteCredentialUser(clientId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/user/credentials/${cloudAccountNumber}/delete?clientId=${clientId}`
    return AxiosInstance.delete(route, { data: {} })
  }
}

export default new CloudAccountService()
