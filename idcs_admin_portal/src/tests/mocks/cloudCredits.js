import { when } from 'jest-when'
import { mockAxios } from '../../setupTests'

export const mockBaseCloudCreditsStore = () => {
  return {
    coupons: [mockBaseCloudCredits(), mockBaseCloudCredits(), mockBaseCloudCredits()]
  }
}

export const mockBaseCloudCredits = () => {
  return {
    code: '52XL-RZ73-YLYZ',
    creator: 'anyadminuser',
    created: '2023-06-10T15:49:01Z',
    start: '2023-06-10T15:49:11Z',
    expires: '2023-08-15T02:00:00Z',
    disabled: null,
    amount: 2500,
    numUses: 1,
    numRedeemed: 1,
    redemptions: [mockBaseRedeemptions(), mockBaseRedeemptions()]
  }
}

export const mockBaseRedeemptions = () => {
  return {
    code: 'YB3F-FZZ1-QKE8',
    cloudAccountId: '216823269847',
    redeemed: '2023-06-10T21:45:04Z',
    installed: true
  }
}

export const mockCreateCoupon = () => {
  return {
    code: '52XL-RZ73-YLYZ',
    creator: 'veera.prasad.dagudu@intel.com',
    created: '2023-06-10T15:49:01Z',
    start: '2023-06-10T15:49:11Z',
    expires: '2023-08-15T02:00:00Z',
    disabled: null,
    amount: 2500,
    numUses: 1,
    numRedeemed: 0,
    redemptions: []
  }
}

export const mockCloudCredits = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/cloudcredits/credits'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseCloudCreditsStore()
      })
    )
}

export const mockCreateCouponAPI = () => {
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/cloudcredits/coupons'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockCreateCoupon()
      })
    )
}

export const mockCouponsAPI = () => {
  when(mockAxios.get)
    .calledWith(expect.anything('/cloudcredits/coupons'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseCloudCreditsStore()
      })
    )
}
