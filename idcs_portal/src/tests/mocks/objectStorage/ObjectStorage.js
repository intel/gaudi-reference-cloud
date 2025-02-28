// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockAxios } from '../../../setupTests'

export const mockBaseObjectStorageStore = () => {
  return {
    items: [
      {
        metadata: {
          cloudAccountId: '549881761988',
          name: '549881761988-test1',
          resourceId: '3ad9dbc7-8683-4e0b-959e-58ca18df8250',
          resourceVersion: '',
          description: 'test1',
          labels: {},
          creationTimestamp: '2024-03-06T15:02:12.833489337Z',
          deletionTimestamp: null
        },
        spec: {
          availabilityZone: 'us-dev3-1a',
          request: {
            size: '10GB'
          },
          versioned: false,
          accessPolicy: 'UNSPECIFIED'
        },
        status: {
          phase: 'BucketReady',
          message: '',
          policy: {
            lifecycleRules: [],
            userAccessPolicies: [
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  name: 'test1',
                  userId: '0b9c5b0b-6e51-43ff-9c11-7f31cf6e6a33',
                  labels: {},
                  creationTimestamp: '2024-03-07T12:19:14.880209355Z',
                  updateTimestamp: null,
                  deleteTimestamp: null
                },
                spec: [
                  {
                    bucketId: '549881761988-test1',
                    prefix: 'path',
                    permission: ['DeleteBucket'],
                    actions: ['ListBucketMultipartUploads', 'ListMultipartUploadParts', 'GetBucketTagging']
                  }
                ]
              },
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  name: 'test12',
                  userId: '83ec446f-dbee-4d73-a23e-e807dd6154ba',
                  labels: {},
                  creationTimestamp: '2024-03-07T12:20:39.562183690Z',
                  updateTimestamp: null,
                  deleteTimestamp: null
                },
                spec: [
                  {
                    bucketId: '549881761988-test1',
                    prefix: 'path1',
                    permission: ['DeleteBucket', 'WriteBucket'],
                    actions: [
                      'ListBucketMultipartUploads',
                      'ListMultipartUploadParts',
                      'GetBucketTagging',
                      'ListBucket'
                    ]
                  }
                ]
              }
            ]
          }
        }
      },
      {
        metadata: {
          cloudAccountId: '549881761988',
          name: '549881761988-test2',
          resourceId: 'f8ade4d9-1e1d-4801-a972-8afb3a007417',
          resourceVersion: '',
          description: '',
          labels: {},
          creationTimestamp: '2024-03-06T15:30:52.089327979Z',
          deletionTimestamp: null
        },
        spec: {
          availabilityZone: 'us-dev3-1a',
          request: {
            size: '10GB'
          },
          versioned: false,
          accessPolicy: 'UNSPECIFIED'
        },
        status: {
          phase: 'BucketReady',
          message: '',
          policy: {
            lifecycleRules: [
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  ruleName: 'rule1',
                  resourceId: '8e4c86c2-3b2f-4584-8617-d752940e4ab0',
                  bucketId: 'f8ade4d9-1e1d-4801-a972-8afb3a007417',
                  creationTimestamp: '2024-03-06T15:31:59.096366588Z',
                  updateTimestamp: null,
                  deletionTimestamp: null
                },
                spec: {
                  prefix: '',
                  expireDays: 0,
                  noncurrentExpireDays: 0,
                  deleteMarker: false
                },
                status: {
                  ruleId: '',
                  phase: 'LFRuleReady'
                }
              },
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  ruleName: 'rule2',
                  resourceId: 'daafe9b2-9321-493f-9cec-105b95a426fa',
                  bucketId: 'f8ade4d9-1e1d-4801-a972-8afb3a007417',
                  creationTimestamp: '2024-03-06T15:54:36.280450347Z',
                  updateTimestamp: '2024-03-06T16:40:01.757731782Z',
                  deletionTimestamp: null
                },
                spec: {
                  prefix: 'test1',
                  expireDays: 331,
                  noncurrentExpireDays: 221,
                  deleteMarker: false
                },
                status: {
                  ruleId: '',
                  phase: 'LFRuleReady'
                }
              },
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  ruleName: 'rule3',
                  resourceId: '903eabb2-d134-4e15-b76d-8fa676765cad',
                  bucketId: 'f8ade4d9-1e1d-4801-a972-8afb3a007417',
                  creationTimestamp: '2024-03-06T17:58:18.820013464Z',
                  updateTimestamp: '2024-03-06T17:59:47.908399336Z',
                  deletionTimestamp: null
                },
                spec: {
                  prefix: 'rr1',
                  expireDays: 31,
                  noncurrentExpireDays: 31,
                  deleteMarker: false
                },
                status: {
                  ruleId: '',
                  phase: 'LFRuleReady'
                }
              },
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  ruleName: 'rule4',
                  resourceId: '8e258baa-0877-4d10-9055-1bb6aba888a8',
                  bucketId: 'f8ade4d9-1e1d-4801-a972-8afb3a007417',
                  creationTimestamp: '2024-03-07T14:46:08.794774968Z',
                  updateTimestamp: '2024-03-07T14:46:58.847590161Z',
                  deletionTimestamp: null
                },
                spec: {
                  prefix: 'test1',
                  expireDays: 2,
                  noncurrentExpireDays: 2,
                  deleteMarker: false
                },
                status: {
                  ruleId: '',
                  phase: 'LFRuleReady'
                }
              }
            ],
            userAccessPolicies: [
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  name: 'test1',
                  userId: '0b9c5b0b-6e51-43ff-9c11-7f31cf6e6a33',
                  labels: {},
                  creationTimestamp: '2024-03-07T12:19:14.880209355Z',
                  updateTimestamp: null,
                  deleteTimestamp: null
                },
                spec: [
                  {
                    bucketId: '549881761988-test2',
                    prefix: 'root',
                    permission: ['ReadBucket', 'WriteBucket'],
                    actions: ['GetBucketLocation', 'GetBucketPolicy', 'ListBucket']
                  }
                ]
              },
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  name: 'test12',
                  userId: '83ec446f-dbee-4d73-a23e-e807dd6154ba',
                  labels: {},
                  creationTimestamp: '2024-03-07T12:20:39.562183690Z',
                  updateTimestamp: null,
                  deleteTimestamp: null
                },
                spec: [
                  {
                    bucketId: '549881761988-test2',
                    prefix: 'root1',
                    permission: ['ReadBucket', 'WriteBucket', 'DeleteBucket'],
                    actions: ['GetBucketLocation', 'GetBucketPolicy', 'ListBucket', 'ListBucketMultipartUploads']
                  }
                ]
              }
            ]
          }
        }
      },
      {
        metadata: {
          cloudAccountId: '549881761988',
          name: '549881761988-test3',
          resourceId: '3ad9dbc7-8683-4e0b-959e-58ca18df8261',
          resourceVersion: '',
          description: 'test3',
          labels: {},
          creationTimestamp: '2024-03-06T15:02:12.833489337Z',
          deletionTimestamp: null
        },
        spec: {
          availabilityZone: 'us-dev3-1a',
          request: {
            size: '10GB'
          },
          versioned: false,
          accessPolicy: 'UNSPECIFIED'
        },
        status: {
          phase: 'BucketTerminating',
          message: '',
          policy: {
            lifecycleRules: [
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  ruleName: 'rule1',
                  resourceId: '8e4c86c2-3b2f-4584-8617-d752940e4ab0',
                  bucketId: 'f8ade4d9-1e1d-4801-a972-8afb3a007417',
                  creationTimestamp: '2024-03-06T15:31:59.096366588Z',
                  updateTimestamp: null,
                  deletionTimestamp: null
                },
                spec: {
                  prefix: '',
                  expireDays: 0,
                  noncurrentExpireDays: 0,
                  deleteMarker: false
                },
                status: {
                  ruleId: '',
                  phase: 'LFRuleReady'
                }
              },
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  ruleName: 'rule2',
                  resourceId: 'daafe9b2-9321-493f-9cec-105b95a426fa',
                  bucketId: 'f8ade4d9-1e1d-4801-a972-8afb3a007417',
                  creationTimestamp: '2024-03-06T15:54:36.280450347Z',
                  updateTimestamp: '2024-03-06T16:40:01.757731782Z',
                  deletionTimestamp: null
                },
                spec: {
                  prefix: 'test1',
                  expireDays: 331,
                  noncurrentExpireDays: 221,
                  deleteMarker: false
                },
                status: {
                  ruleId: '',
                  phase: 'LFRuleReady'
                }
              },
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  ruleName: 'rule3',
                  resourceId: '903eabb2-d134-4e15-b76d-8fa676765cad',
                  bucketId: 'f8ade4d9-1e1d-4801-a972-8afb3a007417',
                  creationTimestamp: '2024-03-06T17:58:18.820013464Z',
                  updateTimestamp: '2024-03-06T17:59:47.908399336Z',
                  deletionTimestamp: null
                },
                spec: {
                  prefix: 'rr1',
                  expireDays: 31,
                  noncurrentExpireDays: 31,
                  deleteMarker: false
                },
                status: {
                  ruleId: '',
                  phase: 'LFRuleReady'
                }
              },
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  ruleName: 'rule4',
                  resourceId: '8e258baa-0877-4d10-9055-1bb6aba888a8',
                  bucketId: 'f8ade4d9-1e1d-4801-a972-8afb3a007417',
                  creationTimestamp: '2024-03-07T14:46:08.794774968Z',
                  updateTimestamp: '2024-03-07T14:46:58.847590161Z',
                  deletionTimestamp: null
                },
                spec: {
                  prefix: 'test1',
                  expireDays: 2,
                  noncurrentExpireDays: 2,
                  deleteMarker: false
                },
                status: {
                  ruleId: '',
                  phase: 'LFRuleReady'
                }
              }
            ],
            userAccessPolicies: [
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  name: 'test1',
                  userId: '0b9c5b0b-6e51-43ff-9c11-7f31cf6e6a33',
                  labels: {},
                  creationTimestamp: '2024-03-07T12:19:14.880209355Z',
                  updateTimestamp: null,
                  deleteTimestamp: null
                },
                spec: [
                  {
                    bucketId: '549881761988-test2',
                    prefix: 'root',
                    permission: ['ReadBucket', 'WriteBucket'],
                    actions: ['GetBucketLocation', 'GetBucketPolicy', 'ListBucket']
                  }
                ]
              },
              {
                metadata: {
                  cloudAccountId: '549881761988',
                  name: 'test12',
                  userId: '83ec446f-dbee-4d73-a23e-e807dd6154ba',
                  labels: {},
                  creationTimestamp: '2024-03-07T12:20:39.562183690Z',
                  updateTimestamp: null,
                  deleteTimestamp: null
                },
                spec: [
                  {
                    bucketId: '549881761988-test2',
                    prefix: 'root1',
                    permission: ['ReadBucket', 'WriteBucket', 'DeleteBucket'],
                    actions: ['GetBucketLocation', 'GetBucketPolicy', 'ListBucket', 'ListBucketMultipartUploads']
                  }
                ]
              }
            ]
          }
        }
      }
    ]
  }
}

export const mockObjectStorage = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/objects/buckets'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseObjectStorageStore()
      })
    )
}

export const mockEmptyObjectStorage = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/objects/buckets'))
    .mockImplementation(() =>
      Promise.resolve({
        data: {
          items: []
        }
      })
    )
}

export const mockBaseObjectStorageUsersStore = () => {
  return {
    users: [
      {
        metadata: {
          cloudAccountId: '549881761988',
          name: 'p2',
          userId: '451ac515-010e-4cf3-b6d8-c880ec71f902',
          labels: {},
          creationTimestamp: '2024-02-27T21:05:23.437271457Z',
          updateTimestamp: null,
          deleteTimestamp: null
        },
        spec: [
          {
            bucketId: '549881761988-test2-pulkit',
            prefix: 'p2',
            permission: ['ReadBucket', 'WriteBucket'],
            actions: ['GetBucketLocation', 'GetBucketPolicy']
          }
        ],
        status: {
          phase: 'ObjectUserReady',
          principal: {
            cluster: {
              clusterId: '918b5026-d516-48c8-bfd3-5998547265b2',
              accessEndpoint: 'https://s3w-pdx05-2.us-staging-1.cloud.intel.com:9000/'
            },
            credentials: {
              accessKey: 'j1hhQ3oOnkSPyQ==',
              secretKey: ''
            }
          }
        }
      },
      {
        metadata: {
          cloudAccountId: '549881761988',
          name: 'latest',
          userId: '9cfe0e2e-8f8a-4ea0-85d1-37174f583b25',
          labels: {},
          creationTimestamp: '2024-02-27T21:13:50.041109876Z',
          updateTimestamp: null,
          deleteTimestamp: null
        },
        spec: [
          {
            bucketId: '549881761988-test3-pulkit',
            prefix: '1',
            permission: ['ReadBucket'],
            actions: ['GetBucketLocation']
          },
          {
            bucketId: '549881761988-test2-pulkit',
            prefix: '2',
            permission: ['WriteBucket'],
            actions: ['GetBucketPolicy']
          }
        ],
        status: {
          phase: 'ObjectUserTerminating',
          principal: {
            cluster: {
              clusterId: '918b5026-d516-48c8-bfd3-5998547265b2',
              accessEndpoint: 'https://s3w-pdx05-2.us-staging-1.cloud.intel.com:9000/'
            },
            credentials: {
              accessKey: '8OPeLbgSL7QDAA==',
              secretKey: ''
            }
          }
        }
      },
      {
        metadata: {
          cloudAccountId: '549881761988',
          name: 'test',
          userId: '97cba53c-a371-4342-9294-f66cae782bce',
          labels: {},
          creationTimestamp: '2024-02-29T15:10:17.898686387Z',
          updateTimestamp: null,
          deleteTimestamp: null
        },
        spec: [
          {
            bucketId: '549881761988-test2',
            prefix: '',
            permission: ['ReadBucket', 'WriteBucket', 'DeleteBucket'],
            actions: [
              'GetBucketLocation',
              'GetBucketPolicy',
              'ListBucket',
              'ListBucketMultipartUploads',
              'ListMultipartUploadParts',
              'GetBucketTagging'
            ]
          }
        ],
        status: {
          phase: 'ObjectUserReady',
          principal: {
            cluster: {
              clusterId: '918b5026-d516-48c8-bfd3-5998547265b2',
              accessEndpoint: 'https://s3w-pdx05-2.us-staging-1.cloud.intel.com:9000/'
            },
            credentials: {
              accessKey: 'ElGYCpJPdx2RXw==',
              secretKey: ''
            }
          }
        }
      },
      {
        metadata: {
          cloudAccountId: '549881761988',
          name: 'test1',
          userId: '6ead942a-4d45-4c66-966d-53dab03a955f',
          labels: {},
          creationTimestamp: '2024-02-29T15:37:42.088529798Z',
          updateTimestamp: null,
          deleteTimestamp: null
        },
        spec: [
          {
            bucketId: '549881761988-test2',
            prefix: '',
            permission: [],
            actions: [
              'GetBucketLocation',
              'GetBucketPolicy',
              'ListBucket',
              'ListBucketMultipartUploads',
              'ListMultipartUploadParts',
              'GetBucketTagging'
            ]
          }
        ],
        status: {
          phase: 'ObjectUserReady',
          principal: {
            cluster: {
              clusterId: '918b5026-d516-48c8-bfd3-5998547265b2',
              accessEndpoint: 'https://s3w-pdx05-2.us-staging-1.cloud.intel.com:9000/'
            },
            credentials: {
              accessKey: 'sDPCLu+CGE9XQA==',
              secretKey: ''
            }
          }
        }
      }
    ]
  }
}

export const mockObjectStorageUsers = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/objects/users'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseObjectStorageUsersStore()
      })
    )
}

export const mockEmptyObjectStorageUsers = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/objects/users'))
    .mockImplementation(() =>
      Promise.resolve({
        data: {
          items: []
        }
      })
    )
}

export const mockObjectStorageData = () => {
  const objectStoragesResponse = mockBaseObjectStorageStore()
  const objectStorages = []

  const getStorageStatusPhase = (status) => {
    if (!status.phase) {
      return ''
    }
    const phase = status.phase.replace('Bucket', '')
    return `${phase.charAt(0)}${phase.substring(1)}`
  }

  const getStorageGBSize = (spec) => {
    return spec.request.size.replace(/\D/g, '')
  }

  for (const index in objectStoragesResponse.items) {
    const storageItem = { ...objectStoragesResponse.items[index] }

    const { metadata, spec, status } = storageItem

    const bucketReservation = {
      cloudAccountId: metadata.cloudAccountId,
      name: metadata.name,
      description: metadata.description,
      resourceId: metadata.resourceId,
      accessPolicy: spec.accessPolicy,
      availabilityZone: spec.availabilityZone,
      versioned: spec.versioned,
      storage: spec?.request && spec?.request?.size ? spec.request.size : '',
      size: getStorageGBSize(spec),
      status: getStorageStatusPhase(status),
      message: status.message,
      creationTimestamp: metadata.creationTimestamp,
      lifecycleRulePolicies: status.policy ? buildLifecycleRule(status.policy) : [],
      userAccessPolicies: []
    }
    objectStorages.push(bucketReservation)
  }
  objectStorages.sort((p1, p2) => (p1.name < p2.name ? 1 : p1.name > p2.name ? -1 : 0))
  return objectStorages
}

const buildLifecycleRule = (bucketPolicy) => {
  const lifecycleRules = bucketPolicy.lifecycleRules
  const getStatus = (status) => {
    if (!status.phase) {
      return ''
    }
    const phase = status.phase.replace('LFRule', '')
    return `${phase.charAt(0)}${phase.substring(1)}`
  }
  return lifecycleRules.map((rule) => {
    return {
      resourceId: rule.metadata.resourceId,
      ruleName: rule.metadata.ruleName,
      prefix: rule.spec.prefix,
      deleteMarker: rule.spec.deleteMarker,
      expireDays: rule.spec.expireDays,
      noncurrentExpireDays: rule.spec.noncurrentExpireDays,
      status: getStatus(rule.status)
    }
  })
}
