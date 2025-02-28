// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'

export interface BucketUsersPermission {
  bucketId: string | ''
  prefix: string | ''
  permission: string[] | []
  actions: string[] | []
}

export type BucketUsersSpecs = Record<string, BucketUsersPermission>

export interface BucketUsersPermissionsStore {
  bucketsPermissions: BucketUsersSpecs | null
  selectionType: string | null
  setBucketsPermissions: (obj: any) => Promise<void>
  setBucketId: (bucketId: string) => Promise<void>
  setPrefix: (bucketId: string, prefix: string) => Promise<void>
  setPermissions: (bucketId: string, permissions: string[]) => Promise<void>
  setActions: (bucketId: string, actions: string[]) => Promise<void>
  setSelectionType: (value: string) => void
  reset: () => void
}

const initialState = {
  selectionType: null,
  bucketsPermissions: null,
  selectAllPermission: [],
  selectAllAction: []
}

const useBucketUsersPermissionsStore = create<BucketUsersPermissionsStore>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setBucketsPermissions: async (obj) => {
    set({ bucketsPermissions: obj })
  },
  setBucketId: async (bucketId) => {
    const bucketObj = get().bucketsPermissions
    if (bucketObj === null || !bucketObj[bucketId]) {
      const obj: BucketUsersPermission = {
        bucketId,
        prefix: '',
        permission: [],
        actions: []
      }
      const bucketAssign: BucketUsersSpecs = {
        [bucketId]: obj
      }
      const bucketUser = { ...get().bucketsPermissions, ...bucketAssign }
      set({ bucketsPermissions: bucketUser })
    }
  },
  setPrefix: async (bucketId, prefix) => {
    const bucketObj = get().bucketsPermissions
    if (bucketObj) {
      const obj = { ...bucketObj[bucketId] }
      obj.prefix = prefix
      const bucketAssign: BucketUsersSpecs = {
        [bucketId]: obj
      }
      const bucketUser = { ...bucketObj, ...bucketAssign }
      set({ bucketsPermissions: bucketUser })
    }
  },
  setPermissions: async (bucketId, permissions) => {
    const bucketObj = get().bucketsPermissions
    if (bucketObj) {
      const obj = { ...bucketObj[bucketId] }
      obj.permission = permissions
      const bucketAssign: BucketUsersSpecs = {
        [bucketId]: obj
      }
      const bucketUser = { ...bucketObj, ...bucketAssign }
      set({ bucketsPermissions: bucketUser })
    }
  },
  setActions: async (bucketId, actions) => {
    const bucketObj = get().bucketsPermissions
    if (bucketObj) {
      const obj = { ...bucketObj[bucketId] }
      obj.actions = actions
      const bucketAssign: BucketUsersSpecs = {
        [bucketId]: obj
      }
      const bucketUser = { ...bucketObj, ...bucketAssign }
      set({ bucketsPermissions: bucketUser })
    }
  },
  setSelectionType: (selectionType) => {
    set({ selectionType })
  }
}))

export default useBucketUsersPermissionsStore
