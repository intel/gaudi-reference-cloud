// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import useUserStore from '../../store/userStore/UserStore'
import { AppRolesEnum } from '../Enums'

/**
 * Returns true if user has the specified role
 * @param {string} role App role to validate
 * @returns {boolean}
 */
const checkAppRole = (role) => {
  const { isStandardUser, isPremiumUser, isEnterpriseUser, isEnterprisePendingUser, isIntelUser } =
    useUserStore.getState()
  switch (role) {
    case AppRolesEnum.Standard:
      return isStandardUser()
    case AppRolesEnum.Premium:
      return isPremiumUser()
    case AppRolesEnum.Enterprise:
      return isEnterpriseUser()
    case AppRolesEnum.EnterprisePending:
      return isEnterprisePendingUser()
    case AppRolesEnum.Intel:
      return isIntelUser()
    default:
      return false
  }
}

/**
 * Returns true if user has at least one of the specified roles
 * @param {Array<string>} allAllowedRoles App roles to validate
 * @returns {boolean}
 */
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
