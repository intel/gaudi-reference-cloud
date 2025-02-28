// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockAxios } from '../../../setupTests'
import idcConfig from '../../../config/configurator'

const mockVendors = () => {
  return {
    vendors: [
      {
        name: 'idc',
        id: '4015bb99-0522-4387-b47e-c821596dc735',
        created: '2023-07-21T22:06:06Z',
        description: idcConfig.REACT_APP_CONSOLE_LONG_NAME,
        families: [
          {
            name: 'compute',
            id: '61befbee-0607-47c5-b140-c6b4ea10b9da',
            created: '2023-07-21T22:06:06Z',
            description: 'Compute as a Service: Bare Metal and Virtual Machine'
          },
          {
            name: 'network',
            id: 'c1c4767c-d718-41d4-9076-02f270b68f5b',
            created: '2023-07-21T22:06:06Z',
            description: 'Network as a Service'
          },
          {
            name: 'storage',
            id: 'fabe738e-1edd-4d07-b1b4-5d9eadc9f28d',
            created: '2023-07-21T22:06:06Z',
            description: 'Storage as a Service'
          },
          {
            name: 'training',
            id: '07af0540-2fda-11ee-be56-0242ac120002',
            created: '2023-07-21T22:06:06Z',
            description: 'Training as a Service'
          },
          {
            name: 'software',
            id: '368e03ac-7f44-11ee-b962-0242ac12000',
            created: '2023-07-21T22:06:06Z',
            description: 'Software as a Service'
          },
          {
            name: 'userinterface',
            id: 'c643e188-f6d8-493f-b1e1-9582f328acaa',
            created: '2023-07-21T22:06:06Z',
            description: 'User Interface'
          },
          {
            name: 'supercomputer',
            id: '1a700ca4-0e26-4465-abdf-d10934be3803',
            created: '2023-08-31T22:41:24Z',
            description: 'Super Computing as a Service'
          },
          {
            name: 'paymentservices',
            id: '411e9d35-88e6-4783-bc0c-aeca3b11913b',
            created: '2023-08-31T22:41:24Z',
            description: 'Payment Services'
          }
        ]
      }
    ]
  }
}

export const mockVendorsApi = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/vendors'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockVendors()
      })
    )
}
