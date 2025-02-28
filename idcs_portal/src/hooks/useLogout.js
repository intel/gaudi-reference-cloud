// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { LoginRequest, RedirectLogout, msalInstance } from '../AuthConfig'
import useUserStore from '../store/userStore/UserStore'

const useLogout = () => {
  const setIsLogoutInProgress = useUserStore((state) => state.setIsLogoutInProgress)

  const logoutHandler = async (customRedirect) => {
    try {
      setIsLogoutInProgress(true)

      const account = msalInstance.getActiveAccount()
      if (!account) {
        await msalInstance.ssoSilent(LoginRequest)
        logoutHandler()
        return
      }
      const logoutRequest = {
        account,
        postLogoutRedirectUri: customRedirect || `${RedirectLogout.postLogoutRedirectUri}`,
        idTokenHint: useUserStore.getState().user?.idToken
      }

      await msalInstance.logoutRedirect(logoutRequest)
    } catch (error) {
      // Just redirect because there is no session active nor cookies to restore it
      window.location.href = RedirectLogout.postLogoutRedirectUri
    }
  }

  return {
    logoutHandler
  }
}

export default useLogout
