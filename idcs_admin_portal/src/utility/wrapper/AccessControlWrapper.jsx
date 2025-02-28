import useUserStore from '../../store/userStore/UserStore'
import { AppRolesEnum } from '../Enums'

export const checkAppRole = (role) => {
  const {
    isGlobalAdminUser,
    isSREAdminUser,
    isIKSAdminUser,
    isComputeAdminUser,
    isProductAdminUser,
    isSlurmAdminUser,
    isSuperUser,
    isStorageAdminUser,
    isQuotaAdminUser,
    isBannerAdmin,
    isCatalogAdminUser,
    isRegionAdminUser,
    isNodePoolsAdminUser
  } = useUserStore.getState()

  switch (role) {
    case AppRolesEnum.GlobalAdmin:
      return isGlobalAdminUser()
    case AppRolesEnum.SREAdmin:
      return isSREAdminUser()
    case AppRolesEnum.IKSAdmin:
      return isIKSAdminUser()
    case AppRolesEnum.ComputeAdmin:
      return isComputeAdminUser()
    case AppRolesEnum.ProductAdmin:
      return isProductAdminUser()
    case AppRolesEnum.StorageAdmin:
      return isStorageAdminUser()
    case AppRolesEnum.QuotaAdmin:
      return isQuotaAdminUser()
    case AppRolesEnum.SlurmAdmin:
      return isSlurmAdminUser()
    case AppRolesEnum.SuperAdmin:
      return isSuperUser()
    case AppRolesEnum.BannerAdmin:
      return isBannerAdmin()
    case AppRolesEnum.CatalogAdmin:
      return isCatalogAdminUser()
    case AppRolesEnum.RegionAdmin:
      return isRegionAdminUser()
    case AppRolesEnum.NodePoolAdmin:
      return isNodePoolsAdminUser()
    default:
      return false
  }
}

export const checkRoles = (allAllowedRoles) => {
  if (allAllowedRoles.length === 0) {
    return true
  }

  return allAllowedRoles.some((r) => checkAppRole(r))
}

const AccessControlWrapper = ({ allowedRoles = [], children, renderNoAccess = () => <></> }) => {
  const permitted = checkRoles(allowedRoles)

  if (permitted) {
    return children
  }
  return renderNoAccess()
}

export default AccessControlWrapper
