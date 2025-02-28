// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import '@testing-library/jest-dom'
import { when } from 'jest-when'
import { mockAxios, mockMsalAccount } from '../../../setupTests'
import { EnrollAccountType, InvitationStateSelection } from '../../../utils/Enums'
import { mockBaseEnrollResponse } from '../authentication/authHelper'

export const mockAdminInvitationList = () => {
  return {
    invites: []
  }
}

export const premiumAdminMock = {
  cloudAccountId: '120657890959', // Other cloud account than the own of current user,
  email: 'premiumuser@domain.com',
  type: EnrollAccountType.premium
}

export const mockGetUserCloudAccountsList = () => {
  const { cloudAccountId } = mockBaseEnrollResponse(EnrollAccountType.standard, true)
  return {
    memberAccount: [
      {
        id: cloudAccountId,
        invitationState: InvitationStateSelection.INVITE_STATE_UNSPECIFIED,
        name: mockMsalAccount.idTokenClaims.email,
        owner: mockMsalAccount.idTokenClaims.email,
        type: EnrollAccountType.premium
      },
      {
        id: premiumAdminMock.cloudAccountId,
        invitationState: InvitationStateSelection.INVITE_STATE_ACCEPTED,
        name: premiumAdminMock.email,
        owner: premiumAdminMock.email,
        type: premiumAdminMock.type
      }
    ]
  }
}

export const mockMultiUserAdminInvitationList = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/cloudaccounts/invitations/read?adminAccountId='))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockAdminInvitationList()
      })
    )
}

export const mockUserCloudAccountsList = () => {
  const { idTokenClaims } = mockMsalAccount
  when(mockAxios.get)
    .calledWith(expect.stringContaining(`/cloudaccounts/name/${idTokenClaims.email}/members`))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockGetUserCloudAccountsList()
      })
    )
}
