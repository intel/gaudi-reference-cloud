// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockAxios } from '../../../setupTests'
import { mockBaseEnrollResponse } from '../authentication/authHelper'

export const mockBaseVnetsStore = () => {
  return {
    items: []
  }
}

export const mockVnets = () => {
  const { cloudAccountId } = mockBaseEnrollResponse()
  when(mockAxios.get)
    .calledWith(expect.stringContaining(`/cloudaccounts/${cloudAccountId}/vnets`))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseVnetsStore()
      })
    )
}
