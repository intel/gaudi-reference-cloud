// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import '@testing-library/jest-dom'
import { when } from 'jest-when'
import { EnrollAccountType, EnrollActionResponse } from '../../../utils/Enums'
import useUserStore from '../../../store/userStore/UserStore'
import { mockAxios, mockMsalAccount } from '../../../setupTests'

const mockMsalUserInit = () => {
  useUserStore.getState().setUser(mockMsalAccount.idTokenClaims, 'myIDToken')
}

export const mockBaseEnrollResponse = (
  accountType = EnrollAccountType.standard,
  isMember = false,
  action = EnrollActionResponse.ENROLL_ACTION_NONE
) => {
  return {
    action,
    cloudAccountId: '120354053959',
    cloudAccountType: accountType,
    enrolled: true,
    haveBillingAccount: true,
    haveCloudAccount: true,
    registered: true,
    isInviteAccounts: false,
    cloudAccountEmail: '',
    isMember
  }
}

export const mockEnterpriseUser = () => {
  mockMsalUserInit()
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/cloudaccounts/enroll'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseEnrollResponse(EnrollAccountType.enterprise)
      })
    )
}

export const mockStandardUser = (isMember = false) => {
  mockMsalUserInit()
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/cloudaccounts/enroll'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseEnrollResponse(EnrollAccountType.standard, isMember)
      })
    )
}

export const mockPremiumUser = () => {
  mockMsalUserInit()
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/cloudaccounts/enroll'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseEnrollResponse(EnrollAccountType.premium)
      })
    )
}

export const mockIntelUser = () => {
  mockMsalUserInit()
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/cloudaccounts/enroll'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseEnrollResponse(EnrollAccountType.intel)
      })
    )
}

export const mockEnterprisePendingUser = () => {
  mockMsalUserInit()
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/cloudaccounts/enroll'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseEnrollResponse(EnrollAccountType.enterprise_pending)
      })
    )
}

export const mockNonAcceptedTCsUser = () => {
  mockMsalUserInit()
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/cloudaccounts/enroll'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseEnrollResponse(EnrollAccountType.intel, false, EnrollActionResponse.ENROLL_ACTION_TC)
      })
    )
}
