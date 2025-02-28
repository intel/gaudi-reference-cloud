// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import AuthorizationService from '../../services/AuthorizationService'

interface ResourceAction {
  description: string
  name: string
  type: string
}

interface ResourceDefinition {
  type: string
  description: string
  actions: ResourceAction[] | []
}

interface PermissionDefinition {
  type: string
  resourceId: string
  id: string[]
  actions: string[]
}

export interface RoleDefinition {
  id?: string
  alias: string
  cloudAccountId?: string
  effect: 'allow' | 'deny'
  permissions: PermissionDefinition[]
  users: string[]
}

export interface UserRoleDefinition {
  cloudAccountRoleId?: string
  alias: string
  cloudAccountId?: string
  effect: 'allow' | 'deny'
  permissions: PermissionDefinition[]
}

interface AuthorizationStore {
  loading: boolean
  resources: ResourceDefinition[]
  roles: RoleDefinition[]
  userRoles: UserRoleDefinition[]
  setResources: () => Promise<void>
  setRoles: () => Promise<void>
  setUserRoles: (userEmail: string) => Promise<void>
  reset: () => void
}

const initialState = {
  loading: false,
  resources: [],
  roles: [],
  userRoles: []
}

const useAuthorizationStore = create<AuthorizationStore>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setResources: async () => {
    set({ loading: true })
    try {
      const resourcesResponse: any = await AuthorizationService.getResources()
      const resources: ResourceDefinition[] = []

      for (const index in resourcesResponse.data.resources) {
        const resourceItem = { ...resourcesResponse.data.resources[index] }
        resources.push(resourceItem)
      }
      resources.sort((r1, r2) => (r1.type < r2.type ? 1 : r1.type > r2.type ? -1 : 0))
      set({ loading: false, resources })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setRoles: async () => {
    set({ loading: true })
    try {
      const rolesResponse: any = await AuthorizationService.getCloudAccountRoles()
      const roles: RoleDefinition[] = []
      for (const index in rolesResponse.data.cloudAccountRoles) {
        const roleItem = { ...rolesResponse.data.cloudAccountRoles[index] }
        roles.push(roleItem)
      }
      roles.sort((r1, r2) => r1.alias.localeCompare(r2.alias, undefined, { numeric: true }))
      set({ loading: false, roles })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setUserRoles: async (userEmail) => {
    set({ loading: true })
    try {
      const { data } = await AuthorizationService.getUserRoles(userEmail)

      const userRoles: UserRoleDefinition[] = []

      for (const index in data.cloudAccountRoles) {
        const roleItem = { ...data.cloudAccountRoles[index] }
        userRoles.push(roleItem)
      }
      userRoles.sort((r1, r2) => r1.alias.localeCompare(r2.alias, undefined, { numeric: true }))

      set({ loading: false, userRoles })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  }
}))

export default useAuthorizationStore
