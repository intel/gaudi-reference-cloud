// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import { type BucketUsersPermission } from '../bucketUsersPermissionsStore/BucketUsersPermissionsStore'
import BucketService from '../../services/BucketService'

export interface BucketReservation {
  cloudAccountId: string
  name: string
  description: string
  resourceId: string
  accessPolicy: string
  availabilityZone: string
  versioned: string
  storage: string
  size: string
  status: string
  message: string
  creationTimestamp: string
  lifecycleRulePolicies: BucketLifecycleRule[] | []
  userAccessPolicies: BucketUserAccess[] | []
  accessEndpoint: string
  subnet: string
}

export interface BucketUserAccess {
  userId: string
  name: string
  spec: BucketUsersPermission
}

export interface BucketLifecycleRule {
  resourceId: string
  ruleName: string
  prefix: string
  deleteMarker: boolean
  expireDays: number
  noncurrentExpireDays: number
  status: string
}

export interface BucketUserReservation {
  cloudAccountId: string
  userId: string
  name: string
  status: string
  creationTimestamp: string
  updateTimestamp: string
  password: string
  spec: string[]
}

export interface Bucket {
  loading: boolean
  objectStorages: BucketReservation[] | []
  shouldRefreshObjectStorages: boolean
  bucketUsers: BucketUserReservation[] | []
  shouldRefreshBucketUsers: boolean
  currentSelectedBucket: BucketReservation | null
  currentSelectedBucketUser: string | null
  setObjectStorages: (isBackground: boolean) => Promise<void>
  setShouldRefreshObjectStorages: (value: boolean) => void
  setBucketUsers: (isBackground: boolean) => Promise<void>
  setShouldRefreshBucketUsers: (value: boolean) => void
  setCurrentSelectedBucket: (bucket: BucketReservation | null) => void
  setCurrentSelectedBucketUser: (userName: string | null) => void
  bucketActiveTab: number
  setBucketActiveTab: (tabNumber: number) => void
  buildCurrentSelectedBucket: (bucket: any) => void
  setBucketsOptionsList: () => Promise<any>
  setLifecycleRulesOptionsList: () => Promise<any>
  setBucketUsersOptionsList: () => Promise<any>
  reset: () => void
}

const initialState = {
  objectStorages: [],
  loading: false,
  shouldRefreshObjectStorages: false,
  bucketUsers: [],
  shouldRefreshBucketUsers: false,
  currentSelectedBucket: null,
  currentSelectedBucketUser: null,
  bucketActiveTab: 0
}

const useBucketStore = create<Bucket>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setObjectStorages: async (isBackground = false) => {
    try {
      if (!isBackground) {
        set({ loading: true })
      }

      const objectStoragesResponse = await BucketService.getObjectBucketsByCloudAccount()
      const objectStorages: BucketReservation[] = []

      for (const index in objectStoragesResponse.data.items) {
        const storageItem = { ...objectStoragesResponse.data.items[index] }

        objectStorages.push(buildBucket(storageItem))
      }
      objectStorages.sort((p1, p2) => (p1.name < p2.name ? 1 : p1.name > p2.name ? -1 : 0))
      set({ loading: false, objectStorages })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setShouldRefreshObjectStorages: (value: boolean) => {
    set({ shouldRefreshObjectStorages: value })
  },
  setBucketUsers: async (isBackground = false) => {
    try {
      if (!isBackground) {
        set({ loading: true })
      }

      const bucketUsersResponse = await BucketService.getBucketUsersByCloudAccount()
      const bucketUsers: BucketUserReservation[] = []

      const getUserStatusPhase = (status: any): string => {
        if (!status.phase) {
          return ''
        }
        const phase: string = status.phase.replace('ObjectUser', '')
        return `${phase.charAt(0)}${phase.substring(1)}`
      }

      for (const index in bucketUsersResponse.data.users) {
        const userItem = { ...bucketUsersResponse.data.users[index] }

        const { metadata, status, spec } = userItem

        const bucketUser: BucketUserReservation = {
          cloudAccountId: metadata.cloudAccountId,
          name: metadata.name,
          userId: metadata.userId,
          status: getUserStatusPhase(status),
          creationTimestamp: metadata.creationTimestamp,
          updateTimestamp: metadata.updateTimestamp,
          password: '',
          spec: buildUserPermissionData(spec)
        }
        bucketUsers.push(bucketUser)
      }
      bucketUsers.sort((p1, p2) => (p1.name < p2.name ? 1 : p1.name > p2.name ? -1 : 0))
      set({ loading: false, bucketUsers })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setShouldRefreshBucketUsers: (value: boolean) => {
    set({ shouldRefreshBucketUsers: value })
  },
  setCurrentSelectedBucket: (bucket: BucketReservation | null) => {
    set({ currentSelectedBucket: bucket })
  },
  setCurrentSelectedBucketUser: (userName: string | null) => {
    set({ currentSelectedBucketUser: userName })
  },
  setBucketActiveTab: (tabNumber: number) => {
    set({ bucketActiveTab: tabNumber })
  },
  buildCurrentSelectedBucket: (bucket: any) => {
    set({ currentSelectedBucket: buildBucket(bucket) })
  },
  setBucketsOptionsList: async () => {
    if (get().objectStorages.length === 0) {
      await get().setObjectStorages(false)
    }

    return buildBucketOptionsList(get().objectStorages)
  },
  setLifecycleRulesOptionsList: async () => {
    if (get().objectStorages.length === 0) {
      await get().setObjectStorages(false)
    }

    return buildLifecycleRulesOptionsList(get().objectStorages)
  },
  setBucketUsersOptionsList: async () => {
    if (get().bucketUsers.length === 0) {
      await get().setBucketUsers(false)
    }

    return buildBucketUsersOptionsList(get().bucketUsers)
  }
}))

const allowedStates = ['Provisioning', 'Stopping', 'Terminating', 'Deleting']

const needToRefreshObjectStorages = (): boolean => {
  const objectStorages = useBucketStore.getState().objectStorages
  const shouldRefreshObjectStorages = useBucketStore.getState().shouldRefreshObjectStorages
  return (
    shouldRefreshObjectStorages &&
    objectStorages?.some((objectStorage: BucketReservation) => allowedStates.includes(objectStorage.status))
  )
}

const needToRefreshBucketUsers = (): boolean => {
  const bucketUsers = useBucketStore.getState().bucketUsers
  const shouldRefreshBucketUsers = useBucketStore.getState().shouldRefreshBucketUsers
  return (
    shouldRefreshBucketUsers &&
    bucketUsers?.some((bucketUser: BucketUserReservation) => allowedStates.includes(bucketUser.status))
  )
}

setInterval(() => {
  if (needToRefreshObjectStorages()) {
    void (async () => {
      await useBucketStore.getState().setObjectStorages(true)
    })()
  }
  if (needToRefreshBucketUsers()) {
    void (async () => {
      await useBucketStore.getState().setBucketUsers(true)
    })()
  }
}, 10000)

const buildBucket = (bucket: any): BucketReservation => {
  const getStorageStatusPhase = (status: any): string => {
    if (!status.phase) {
      return ''
    }
    const phase: string = status.phase.replace('Bucket', '')
    return `${phase.charAt(0)}${phase.substring(1)}`
  }

  const getStorageGBSize = (spec: any): string => {
    return spec.request.size.replace(/\D/g, '')
  }

  const getSubnet = (status: any): string => {
    if (!status.securityGroup?.networkFilterAllow) {
      return ''
    }

    return status.securityGroup.networkFilterAllow.map((x: any) => {
      return String(x.subnet) + '/' + String(x.prefixLength)
    })
  }

  const { metadata, spec, status } = bucket

  return {
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
    userAccessPolicies: status.policy ? buildBucketUsers(status.policy) : [],
    accessEndpoint: status?.cluster && status?.cluster?.accessEndpoint ? status.cluster.accessEndpoint : '',
    subnet: getSubnet(status)
  }
}

const buildLifecycleRule = (bucketPolicy: any): any => {
  const lifecycleRules = bucketPolicy.lifecycleRules
  const getStatus = (status: any): string => {
    if (!status.phase) {
      return ''
    }
    const phase: string = status.phase.replace('LFRule', '')
    return `${phase.charAt(0)}${phase.substring(1)}`
  }
  return lifecycleRules.map((rule: any): BucketLifecycleRule => {
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

const buildBucketUsers = (bucketPolicy: any): any => {
  const userPolicies = bucketPolicy.userAccessPolicies
  return userPolicies.map((user: any): BucketUserAccess => {
    return {
      userId: user.metadata.userId,
      name: user.metadata.name,
      spec: user.spec[0]
    }
  })
}

const buildUserPermissionData = (spec: any): any => {
  const obj: any = {}
  for (const i in spec) {
    const permissionObj = spec[i]
    const bucketId = permissionObj.bucketId
    obj[bucketId] = permissionObj
  }
  return obj
}

const buildBucketOptionsList = (response: any): any[] => {
  return response.map((record: any) => {
    return {
      name: record.name,
      value: record.resourceId
    }
  })
}

const buildBucketUsersOptionsList = (response: any): any[] => {
  return response.map((record: any) => {
    return {
      name: record.name,
      value: record.userId
    }
  })
}

const buildLifecycleRulesOptionsList = (response: any): any[] => {
  const lifecycleRuleArray = []
  for (let i = 0; i < response.length; i++) {
    const bucket = response[i]
    const name = bucket.name
    const ruleLength = bucket.lifecycleRulePolicies.length

    if (ruleLength > 0) {
      for (let j = 0; j < ruleLength; j++) {
        const rule = bucket.lifecycleRulePolicies[j]
        lifecycleRuleArray.push({
          value: rule.resourceId,
          name: `${rule.ruleName} - ${name}`
        })
      }
    }
  }
  return lifecycleRuleArray
}

export default useBucketStore
