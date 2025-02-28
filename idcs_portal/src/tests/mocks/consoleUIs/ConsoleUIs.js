// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockAxios } from '../../../setupTests'

export const mockBaseConsoleUIs = () => {
  return {
    products: [
      {
        id: '7cfc5a1f-a92d-49be-8cbb-4f627af78fb9',
        name: 'gui-beta',
        created: '2023-11-10T17:16:48Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: 'c643e188-f6d8-493f-b1e1-9582f328acaa',
        description: 'Intel Beta UI 2.0',
        metadata: {
          billingEnable: 'false',
          category: 'UserInterface',
          service: 'IDC User Interface',
          'family.displayName': '',
          instanceType: 'gui-beta',
          displayName: 'IDC Console Beta',
          url: 'localhost'
        }
      }
    ]
  }
}

export const mockBasePaymentServices = () => {
  return {
    products: [
      {
        id: '7cfc5a1f-a92d-49be-8cbb-4f627af78fb9',
        name: 'credit-card',
        created: '2023-11-10T17:16:48Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: 'c643e188-f6d8-493f-b1e1-9582f328acaa',
        description: 'Intel Beta UI 2.0',
        metadata: {
          billingEnable: 'false',
          category: 'UserInterface',
          service: 'Payment Service',
          'family.displayName': '',
          instanceType: 'credit-card',
          displayName: 'Credit Card'
        }
      }
    ]
  }
}

export const mockConsoleUIList = () => {
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/products'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseConsoleUIs()
      })
    )
}

export const mockPaymentServices = () => {
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/products'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBasePaymentServices()
      })
    )
}
