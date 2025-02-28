// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockAxios } from '../../../setupTests'

export const mockBaseStorageStore = () => {
  return {
    products: [
      {
        name: 'storage-file',
        id: '3bc52387-da79-4947-a562-04aad42e1db7',
        created: '2023-10-16T15:24:03Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: 'fabe738e-1edd-4d07-b1b4-5d9eadc9f28d',
        description: 'High speed Filestorage',
        metadata: {
          StorageCategories: 'FileStorage',
          access: 'open',
          billingEnable: 'true',
          category: 'singlenode',
          disableForAccountTypes: '',
          'disks.sizes': '200,500,1000',
          displayName: 'Storage Service - Filesystem',
          'family.displayDescription': 'Filesystem Storage Service',
          'family.displayName': 'Filesystem Storage Service',
          highlight: '200GB',
          information: 'Filesystem Storage service\n',
          recommendedUseCase: 'Core compute',
          region: 'global',
          releaseStatus: 'Released',
          service: 'Storage Service - File'
        },
        eccn: 'EAR99',
        pcq: '19621',
        matchExpr: "serviceType == 'StorageAsAService' && instanceType == 'vm-spr-sml'",
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.0075',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'error'
      }
    ]
  }
}

export const mockBaseObjectStorageProductStore = () => {
  return {
    products: [
      {
        name: 'storage-object',
        id: '6e9f2eab-76ee-496e-8805-65b5fffca406',
        created: '2024-03-18T15:35:27Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: 'fabe738e-1edd-4d07-b1b4-5d9eadc9f28d',
        description: 'High speed Object storage',
        metadata: {
          access: 'open',
          billingEnable: 'true',
          disableForAccountTypes: '',
          displayName: 'Storage Service - Object',
          'family.displayDescription': 'Object Storage Service',
          'family.displayName': 'Object Storage Service',
          information: 'Storage Service with Object store \n',
          instanceCategories: 'ObjectStorage',
          instanceType: 'storage-object',
          region: 'global',
          releaseStatus: 'Released',
          service: 'Storage Service - Object'
        },
        eccn: 'EAR99',
        pcq: '19621',
        matchExpr: 'serviceType == "StorageAsAService" && instanceType == "storage-object"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.07',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'ready'
      }
    ]
  }
}

export const mockStorage = () => {
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/products'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseStorageStore()
      })
    )
}

export const mockObjectStorageProduct = () => {
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/products'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseObjectStorageProductStore()
      })
    )
}
