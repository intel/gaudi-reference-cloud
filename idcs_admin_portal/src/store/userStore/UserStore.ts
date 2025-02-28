import { create } from 'zustand'
import { AzureRolesEnum } from '../../utility/Enums'

export interface User {
  displayName: string
  email: string
  roles: string[] | []
  idToken: string
}

interface UserStore {
  user: User | null
  setUser: (idTokenClaims: any, idToken: string) => void
  getUserRoles: () => string[]
  isGlobalAdminUser: () => boolean
  isSuperUser: () => boolean
  isSREAdminUser: () => boolean
  isIKSAdminUser: () => boolean
  isComputeAdminUser: () => boolean
  isProductAdminUser: () => boolean
  isStorageAdminUser: () => boolean
  isQuotaAdminUser: () => boolean
  isSlurmAdminUser: () => boolean
  isBannerAdmin: () => boolean
  isCatalogAdminUser: () => boolean
  isRegionAdminUser: () => boolean
  isNodePoolsAdminUser: () => boolean

  isLogoutInProgress: boolean
  setIsLogoutInProgress: (isInProgress: boolean) => void

  loading: boolean
  isOwnCloudAccount: boolean
}

const useUserStore = create<UserStore>()((set, get) => ({
  user: null,
  loading: false,
  isOwnCloudAccount: true,
  setUser: (idTokenClaims: any, idToken: string) => {
    const newUser: User = {
      displayName: idTokenClaims.name || '',
      email: idTokenClaims.preferred_username || '',
      idToken: idToken || '',
      roles: idTokenClaims.roles || []
    }
    set(() => ({ user: newUser }))
  },
  getUserRoles: () => {
    const user: User | null = get().user
    return user !== null ? user.roles : []
  },
  isGlobalAdminUser: () => {
    const userRoles = get().getUserRoles()
    return userRoles.length > 0 ? userRoles.some((r) => r === AzureRolesEnum.GlobalAdmin.toString()) : false
  },
  isSuperUser: () => {
    const userRoles = get().getUserRoles()
    return userRoles.length > 0 ? userRoles.some((r) => r === AzureRolesEnum.SuperAdmin.toString()) : false
  },
  isSREAdminUser: () => {
    const userRoles = get().getUserRoles()
    return userRoles.length > 0 ? userRoles.some((r) => r === AzureRolesEnum.SREAdmin.toString()) : false
  },
  isIKSAdminUser: () => {
    const userRoles = get().getUserRoles()
    return userRoles.length > 0 ? userRoles.some((r) => r === AzureRolesEnum.IKSAdmin.toString()) : false
  },
  isComputeAdminUser: () => {
    const userRoles = get().getUserRoles()
    return userRoles.length > 0 ? userRoles.some((r) => r === AzureRolesEnum.ComputeAdmin.toString()) : false
  },
  isProductAdminUser: () => {
    const userRoles = get().getUserRoles()
    return userRoles.length > 0 ? userRoles.some((r) => r === AzureRolesEnum.ProductAdmin.toString()) : false
  },
  isStorageAdminUser: () => {
    const userRoles = get().getUserRoles()
    return userRoles.length > 0 ? userRoles.some((r) => r === AzureRolesEnum.StorageAdmin.toString()) : false
  },
  isQuotaAdminUser: () => {
    const userRoles = get().getUserRoles()
    return userRoles.length > 0 ? userRoles.some((r) => r === AzureRolesEnum.QuotaAdmin.toString()) : false
  },
  isCatalogAdminUser: () => {
    const userRoles = get().getUserRoles()
    return userRoles.length > 0 ? userRoles.some((r) => r === AzureRolesEnum.CatalogAdmin.toString()) : false
  },
  isSlurmAdminUser: () => {
    const userRoles = get().getUserRoles()
    return userRoles.length > 0 ? userRoles.some((r) => r === AzureRolesEnum.SlurmAdmin.toString()) : false
  },
  isBannerAdmin: () => {
    const userRoles = get().getUserRoles()
    return userRoles.length > 0 ? userRoles.some((r) => r === AzureRolesEnum.BannerAdmin.toString()) : false
  },
  isRegionAdminUser: () => {
    const userRoles = get().getUserRoles()
    return userRoles.length > 0 ? userRoles.some((r) => r === AzureRolesEnum.RegionAdmin.toString()) : false
  },
  isNodePoolsAdminUser: () => {
    const userRoles = get().getUserRoles()
    return userRoles.length > 0 ? userRoles.some((r) => r === AzureRolesEnum.NodePoolAdmin.toString()) : false
  },
  isLogoutInProgress: false,
  setIsLogoutInProgress: (isInProgress: boolean) => {
    set(() => ({ isLogoutInProgress: isInProgress }))
  }
}))

export default useUserStore
