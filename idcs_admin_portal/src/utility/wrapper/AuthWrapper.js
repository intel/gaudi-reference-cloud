import { useMsal, useMsalAuthentication } from '@azure/msal-react'
import { EventType, InteractionStatus, InteractionType } from '@azure/msal-browser'
import AuthenticationSpinner from '../authenticationSpinner/AuthenticationSpinner'
import { useEffect, useState } from 'react'
import { LoginRequest } from '../../AuthConfig'
import SomethingWentWrong from '../../pages/error/SomethingWentWrong'
import useUserStore from '../../store/userStore/UserStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useAppStore from '../../store/appStore/AppStore'

const AuthWrapper = ({ children }) => {
  const { accounts, inProgress, instance } = useMsal()
  const user = useUserStore((state) => state.user)
  const setUser = useUserStore((state) => state.setUser)
  const isLogoutInProgress = useUserStore((state) => state.isLogoutInProgress)
  const setIsLogoutInProgress = useUserStore((state) => state.setIsLogoutInProgress)
  const changingRegion = useAppStore((state) => state.changingRegion)
  const setFirstLoadedPage = useAppStore((state) => state.setFirstLoadedPage)

  const { login, error } = useMsalAuthentication(InteractionType.Silent, LoginRequest)
  const [retryLogin, setRetryLogin] = useState(false)

  const throwError = useErrorBoundary()

  const removeMsalKeysFromLocalStorage = () => {
    const allKeys = []
    for (let i = 0; i < localStorage.length; i++) {
      allKeys.push(localStorage.key(i))
    }
    const nonMsalKeys = [] // Add local storage keys in lowercase that you wnat to keep after logout
    const msalKeys = allKeys.filter((x) => x && !nonMsalKeys.some((x) => x === x.toLowerCase()))
    msalKeys.forEach((key) => {
      localStorage.removeItem(key)
    })
  }

  const initUser = (idTokenClaims, idToken) => {
    setUser(idTokenClaims, idToken)
  }

  useEffect(() => {
    const logincallback = async (event) => {
      if (
        (event.eventType === EventType.LOGIN_SUCCESS ||
          event.eventType === EventType.ACQUIRE_TOKEN_SUCCESS ||
          event.eventType === EventType.SSO_SILENT_SUCCESS) &&
        event.payload.account
      ) {
        const { account } = event.payload
        instance.setActiveAccount(account)

        if (!user || user.idToken !== account.idToken) {
          initUser(account.idTokenClaims, account.idToken)
        }
      } else if (event.eventType === EventType.INITIALIZE_END) {
        const shouldCleanLocalStorage = Boolean(accounts.length > 0 && !instance.getActiveAccount())
        if (shouldCleanLocalStorage) {
          removeMsalKeysFromLocalStorage()
          window.location.reload()
        }
      } else if (event.eventType === EventType.ACCOUNT_REMOVED) {
        throwError('idc logout window')
      }
    }
    instance.enableAccountStorageEvents()
    instance.addEventCallback(logincallback)
    setFirstLoadedPage(window.location.pathname)
    return () => {
      instance.disableAccountStorageEvents()
      instance.removeEventCallback(logincallback)
    }
  }, [])

  useEffect(() => {
    if (retryLogin && !isLogoutInProgress) {
      login(InteractionType.Redirect, LoginRequest)
      setRetryLogin(false)
    }

    if (isLogoutInProgress) {
      setIsLogoutInProgress(false)
    }
  }, [retryLogin])

  const needsInteraction =
    inProgress === InteractionStatus.None &&
    error &&
    (error.errorCode === 'interaction_required' ||
      error.errorCode === 'monitor_window_timeout' ||
      error.errorCode === 'state_not_found' ||
      error.errorMessage.indexOf('AADB2C90077') !== -1 || // User does not have an existing session error
      error.errorMessage.indexOf('AADSTS50058') !== -1) // cookies blocked
  if (needsInteraction && !retryLogin) {
    setRetryLogin(true)
  }

  if (inProgress === InteractionStatus.None && error && !needsInteraction) {
    return <SomethingWentWrong error={error.errorMessage} />
  }

  const isAuthenticationInProgress = !accounts || accounts.length === 0 || !user || changingRegion

  if (isAuthenticationInProgress) {
    return <AuthenticationSpinner />
  }

  return children
}

export default AuthWrapper
