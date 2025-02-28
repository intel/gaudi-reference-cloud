// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import useUserStore from '../store/userStore/UserStore'
import useErrorBoundary from './useErrorBoundary'
import useLogout from './useLogout'

const useSecurity = () => {
  const checkMemberAccess = useUserStore((state) => state.checkMemberAccess)
  const throwError = useErrorBoundary()
  const { logoutHandler } = useLogout()

  const checkMembersAccess = async () => {
    try {
      const passMemberValidation = await checkMemberAccess()
      if (!passMemberValidation) {
        logoutHandler('/')
      }
    } catch (error) {
      throwError(error)
    }
  }

  return {
    checkMembersAccess
  }
}

export default useSecurity
