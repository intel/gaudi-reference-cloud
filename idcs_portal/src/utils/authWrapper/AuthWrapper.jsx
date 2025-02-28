// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useMsal, useMsalAuthentication } from '@azure/msal-react'
import { EventType, InteractionStatus, InteractionType } from '@azure/msal-browser'
import AuthenticationSpinner from '../authenticationSpinner/AuthenticationSpinner'
import { useEffect, useState } from 'react'
import { LoginRequest } from '../../AuthConfig'
import SomethingWentWrong from '../../pages/error/SomethingWentWrong'
import useUserStore from '../../store/userStore/UserStore'
import AccountCreationError from '../../pages/error/AccountCreationError'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { EnrollActionResponse } from '../Enums'
import idcConfig, { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import AccountsContainer from '../../containers/profile/AccountsContainer'
import useAppStore from '../../store/appStore/AppStore'
import ConsoleAccessDenied from '../../pages/error/ConsoleAccessDenied'
import { Container } from 'react-bootstrap'
import SingleTopNavBar from '../../components/header/SingleTopNavBar'
import FooterMini from '../../components/footer/FooterMini'
import TermsAndConditions from '../../pages/error/TermsAndConditions'
import { ErrorBoundaryLevelWrapper } from '../../pages/error/ErrorBoundary'
import useSecurity from '../../hooks/useSecurity'

const AuthPendingWrapper = ({ children }) => {
  return (
    <>
      <SingleTopNavBar />
      <Container className="siteContainer-no-toolbar container" role="main">
        <div className="sheet">{children}</div>
      </Container>
      <FooterMini />
    </>
  )
}

const AuthWrapper = ({ children }) => {
  const { accounts, inProgress, instance } = useMsal()
  const user = useUserStore((state) => state.user)
  const setUser = useUserStore((state) => state.setUser)
  const enroll = useUserStore((state) => state.enroll)
  const enrollResponse = useUserStore((state) => state.enrollResponse)
  const isLogoutInProgress = useUserStore((state) => state.isLogoutInProgress)
  const consoleUIs = useAppStore((state) => state.consoleUIs)
  const setConsoleUIs = useAppStore((state) => state.setConsoleUIs)
  const paymentServices = useAppStore((state) => state.paymentServices)
  const setPaymentServices = useAppStore((state) => state.setPaymentServices)
  const changinRegion = useAppStore((state) => state.changinRegion)
  const setFirstLoadedPage = useAppStore((state) => state.setFirstLoadedPage)
  const setIsLogoutInProgress = useUserStore((state) => state.setIsLogoutInProgress)
  const { login, error } = useMsalAuthentication(InteractionType.Silent, LoginRequest)
  const [enrollApiCallCount, setEnrollApiCallCount] = useState(0)
  const [retryLogin, setRetryLogin] = useState(false)
  const { checkMembersAccess } = useSecurity()

  const throwError = useErrorBoundary()

  const removeMsalKeysFromLocalStorage = () => {
    const allKeys = []
    for (let i = 0; i < localStorage.length; i++) {
      allKeys.push(localStorage.key(i))
    }
    const msalKeys = allKeys.filter(
      (x) => x && x.toLowerCase().indexOf(idcConfig.REACT_APP_AZURE_B2C_UNIFIED_FLOW.toLowerCase()) !== -1
    )
    msalKeys.forEach((key) => {
      localStorage.removeItem(key)
    })
  }

  // Functions
  const initUser = (idTokenClaims, idToken) => {
    setUser(idTokenClaims, idToken)
  }

  useEffect(() => {
    const logincallback = async (event) => {
      if (
        (event.eventType === EventType.LOGIN_SUCCESS ||
          event.eventType === EventType.SSO_SILENT_SUCCESS ||
          event.eventType === EventType.ACQUIRE_TOKEN_SUCCESS) &&
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
    if (enrollResponse && enrollResponse.action) {
      checkEnrollFlow()
    }
  }, [enrollResponse])

  useEffect(() => {
    async function enrollFn() {
      const shouldCallEnroll = user?.idToken !== ''
      if (shouldCallEnroll) {
        const shouldEnrollPremium = window.location.pathname.toLowerCase() === '/premium'
        try {
          await enroll(shouldEnrollPremium, false)
        } catch (error) {
          throwError(error)
        }
        checkEnrollFlow()
      }
    }
    if (enrollApiCallCount <= 3) {
      enrollFn()
    }
  }, [enrollApiCallCount, user?.idToken])

  useEffect(() => {
    const checkMembersInterval = setInterval(async () => {
      await checkMembersAccess()
    }, 60000)
    return () => {
      clearInterval(checkMembersInterval)
    }
  }, [])

  const checkEnrollFlow = () => {
    if (enrollResponse && enrollResponse.action) {
      const action = enrollResponse.action

      if (action === EnrollActionResponse.ENROLL_ACTION_REGISTER) {
        setEnrollApiCallCount(enrollApiCallCount + 1)
      }

      if (action === EnrollActionResponse.ENROLL_ACTION_RETRY) {
        setEnrollApiCallCount(enrollApiCallCount + 1)
      }
    }
  }

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
      error.errorMessage.indexOf('AADB2C90077') !== -1) // User does not have an existing session error
  if (needsInteraction && !retryLogin) {
    setRetryLogin(true)
  }

  if (inProgress === InteractionStatus.None && error && !needsInteraction) {
    return (
      <ErrorBoundaryLevelWrapper>
        <SomethingWentWrong error={error.errorMessage} />
      </ErrorBoundaryLevelWrapper>
    )
  }

  if (enrollApiCallCount > 3) {
    return (
      <ErrorBoundaryLevelWrapper>
        <AccountCreationError />
      </ErrorBoundaryLevelWrapper>
    )
  }

  if (enrollResponse?.action === EnrollActionResponse.ENROLL_ACTION_TC) {
    return (
      <AuthPendingWrapper>
        <TermsAndConditions />
      </AuthPendingWrapper>
    )
  }

  const isAuthenticationInProgress = !accounts || accounts.length === 0 || !user || !user.authenticated || changinRegion
  if (isAuthenticationInProgress) {
    return (
      <AuthPendingWrapper>
        <AuthenticationSpinner />
      </AuthPendingWrapper>
    )
  }

  if (!user.cloudAccountNumber) {
    return (
      <AuthPendingWrapper>
        <AccountsContainer />
      </AuthPendingWrapper>
    )
  }

  if (consoleUIs === null || paymentServices === null) {
    const fetchUIProducts = async () => {
      const promises = []
      if (consoleUIs === null) {
        promises.push(setConsoleUIs())
      }
      if (paymentServices === null) {
        promises.push(setPaymentServices())
      }
      await Promise.all(promises)
    }
    fetchUIProducts()
    return (
      <AuthPendingWrapper>
        <AuthenticationSpinner />
      </AuthPendingWrapper>
    )
  }

  const isUserWhitelisted =
    !isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_UX_WHITELIST) ||
    consoleUIs.some((x) => x.url === window.location.hostname)

  if (!isUserWhitelisted) {
    return (
      <AuthPendingWrapper>
        <ConsoleAccessDenied />
      </AuthPendingWrapper>
    )
  }

  const isCreditCardEnabled = paymentServices?.find((x) => x.name === 'credit-card') !== undefined
  idcConfig.REACT_APP_FEATURE_DIRECTPOST = isCreditCardEnabled ? 1 : 0

  return children
}

export default AuthWrapper
