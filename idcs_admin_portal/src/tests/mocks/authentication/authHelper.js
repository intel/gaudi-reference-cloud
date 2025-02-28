import '@testing-library/jest-dom'
import { when } from 'jest-when'
import { EnrollAccountType, EnrollActionResponse } from '../../../utility/Enums'
import useUserStore from '../../../store/userStore/UserStore'
import { mockAxios, mockMsalAccount } from '../../../setupTests'

const mockMsalUserInit = () => {
  useUserStore.getState().setUser(mockMsalAccount.idTokenClaims, 'myIDToken')
}

const mockBaseEnrollResponse = (accountType) => {
  return {
    action: EnrollActionResponse.ENROLL_ACTION_NONE,
    cloudAccountId: '120354053959',
    cloudAccountType: accountType,
    enrolled: true,
    haveBillingAccount: true,
    haveCloudAccount: true,
    registered: true
  }
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
