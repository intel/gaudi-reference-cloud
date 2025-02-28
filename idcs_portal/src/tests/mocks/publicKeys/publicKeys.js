// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockAxios } from '../../../setupTests'
import { mockBaseEnrollResponse } from '../authentication/authHelper'

export const mockBasePublicKeysStore = () => {
  return {
    items: [
      {
        metadata: {
          cloudAccountId: '549881761988',
          name: 'test1',
          resourceId: 'ca82c23a-6c30-4c6b-a0b2-3f52f90685b5',
          labels: {},
          creationTimestamp: '2023-08-07T04:56:06.959653Z',
          allowDelete: true
        },
        spec: {
          sshPublicKey: 'ssh-ed25519 SSH KEY',
          ownerEmail: 'test@intel.com'
        }
      },
      {
        metadata: {
          cloudAccountId: '549881761988',
          name: 'test2',
          resourceId: '1750adaf-5b4f-4cc4-ba16-b35fb6176677',
          labels: {},
          creationTimestamp: '2023-08-07T07:39:41.204077Z',
          allowDelete: true
        },
        spec: {
          sshPublicKey: 'ssh-ed25519 SSH KEY'
        }
      },
      {
        metadata: {
          cloudAccountId: '549881761988',
          name: 'test3',
          resourceId: '5c230f45-337b-4f95-b788-9e6606d5960f',
          labels: {},
          creationTimestamp: '2023-08-07T10:57:53.853822Z',
          allowDelete: true
        },
        spec: {
          sshPublicKey: 'ssh-ed25519 SSH KEY'
        }
      }
    ]
  }
}

export const mockCreatePublicKeys = () => {
  const { cloudAccountId } = mockBaseEnrollResponse()
  when(mockAxios.post)
    .calledWith(expect.stringContaining(`/cloudaccounts/${cloudAccountId}/sshpublickeys`), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        status: 200
      })
    )
}

export const mockPublicKeys = () => {
  const { cloudAccountId } = mockBaseEnrollResponse()
  when(mockAxios.get)
    .calledWith(expect.stringContaining(`/cloudaccounts/${cloudAccountId}/sshpublickeys`))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBasePublicKeysStore()
      })
    )
}

export const mockEmptyPublicKeys = () => {
  const { cloudAccountId } = mockBaseEnrollResponse()
  when(mockAxios.get)
    .calledWith(expect.stringContaining(`/cloudaccounts/${cloudAccountId}/sshpublickeys`))
    .mockImplementation(() =>
      Promise.resolve({
        data: []
      })
    )
}

export const mockDeletePublicKeys = () => {
  const { cloudAccountId } = mockBaseEnrollResponse()
  when(mockAxios.delete)
    .calledWith(expect.stringContaining(`/cloudaccounts/${cloudAccountId}/sshpublickeys`), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: {}
      })
    )
}

export const mockFailedCreatePublicKeys = (error) => {
  const { cloudAccountId } = mockBaseEnrollResponse()
  when(mockAxios.post)
    .calledWith(expect.stringContaining(`/cloudaccounts/${cloudAccountId}/sshpublickeys`), expect.anything())
    .mockRejectedValue(new Error(error))
}
