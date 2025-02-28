// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockBaseEnrollResponse } from '../authentication/authHelper'
import { mockAxios } from '../../../setupTests'

export const mockBaseCloudCreditsStore = () => {
  return {
    lastUpdated: 'Mon, 12 Jun 2023 19:49:57 GMT',
    remainingCredits: 300,
    usedCredits: 300,
    loading: true,
    cloudCredits: [mockBaseCloudCredits(), mockBaseCloudCredits(), mockBaseCloudCredits()]
  }
}

export const mockBaseCloudCredits = () => {
  return {
    creditType: 'TEST12',
    createdAt: 'May 1, 2023',
    ExpiryDate: 'May 2, 2023',
    total: 200,
    totalUsed: 100,
    totalRemaining: 100
  }
}

export const mockCloudCredits = () => {
  const { cloudAccountId } = mockBaseEnrollResponse()
  when(mockAxios.get)
    .calledWith(
      expect.stringContaining(`/cloudcredits/credit?cloudAccountId=${cloudAccountId}&history=true`),
      expect.anything()
    )
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseCloudCreditsStore()
      })
    )
}
