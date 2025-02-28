// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useLayoutEffect } from 'react'
import './App.scss'
import Routing from './containers/routing/Routing'
import { type IPublicClientApplication } from '@azure/msal-browser'
import { BrowserRouter } from 'react-router-dom'
import { MsalProvider } from '@azure/msal-react'
import NoAccessError from './pages/error/AccessDenied'
import AccessControlWrapper from './utils/accessControlWrapper/AccessControlWrapper'
import AuthWrapper from './utils/authWrapper/AuthWrapper'
import { AppRolesEnum } from './utils/Enums'
import 'bootstrap/dist/js/bootstrap.bundle.min.js'
import DarkModeContainer from './utils/darkMode/DarkModeContainer'
import useAppStore from './store/appStore/AppStore'
import AuthenticationSpinner from './utils/authenticationSpinner/AuthenticationSpinner'
import { useMediaQuery } from 'react-responsive'
import idcConfig from './config/configurator'

interface AppProps {
  msalInstance: IPublicClientApplication
  children?: React.ReactNode
}

// For production mode, devtools is disabled
const App = ({ msalInstance }: AppProps): React.ReactElement => {
  const firstLoadComplete = useAppStore((state) => state.firstLoadComplete)
  const showLearningBar = useAppStore((state) => state.showLearningBar)
  const showSideNavBar = useAppStore((state) => state.showSideNavBar)
  const setShowSideNavBar = useAppStore((state) => state.setShowSideNavBar)
  const setShowLearningBar = useAppStore((state) => state.setShowLearningBar)
  const toggleSideBars = useAppStore((state) => state.toggleSideBars)
  const changeRegion = useAppStore((state) => state.changeRegion)

  const content = <Routing />

  const isMdScreen = useMediaQuery({
    query: '(max-width: 991px)'
  })

  const isSmScreen = useMediaQuery({
    query: '(max-width: 767px)'
  })

  useEffect(() => {
    loadGlobalSearchParams()
  }, [])

  useEffect(() => {
    if (isSmScreen) {
      setShowSideNavBar(false, false)
      setShowLearningBar(false, false)
    } else {
      const showSideNav = localStorage.getItem('showSideNav')
      setShowSideNavBar(showSideNav === 'true' || showSideNav === null, true)
      const showLearningBar = localStorage.getItem('showLearningBar')
      setShowLearningBar(showLearningBar === 'true', true)
    }
  }, [])

  useEffect(() => {
    if (isMdScreen && showLearningBar && showSideNavBar) {
      setShowSideNavBar(false, false)
    }
  }, [showLearningBar])

  useEffect(() => {
    toggleSideBars(isMdScreen)
  }, [showSideNavBar])

  useLayoutEffect(() => {
    try {
      const rootElement = document.getElementById('root')
      const resizeObserver = new ResizeObserver((event) => {
        setTimeout(() => {
          const isMd = event[0]?.contentRect?.width <= 992
          toggleSideBars(isMd)
        }, 10)
      })
      resizeObserver.observe(rootElement as Element)
      return () => {
        resizeObserver?.disconnect()
      }
    } catch (error) {
      // No Support for resizeObserver
    }
  }, [])

  function loadGlobalSearchParams(): void {
    const params = new URLSearchParams(window.location.search)
    const region = params.get('region')

    if (region && idcConfig.REACT_APP_DEFAULT_REGIONS.includes(region)) {
      if (region.toLowerCase() !== idcConfig.REACT_APP_SELECTED_REGION.toLowerCase()) changeRegion(region)
    } else changeRegion(idcConfig.REACT_APP_SELECTED_REGION)
  }

  if (!firstLoadComplete) {
    return (
      <>
        <DarkModeContainer />
        <AuthenticationSpinner />
      </>
    )
  }

  return (
    <MsalProvider instance={msalInstance}>
      <DarkModeContainer />
      <AuthWrapper>
        <AccessControlWrapper
          allowedRoles={[
            AppRolesEnum.Enterprise,
            AppRolesEnum.EnterprisePending,
            AppRolesEnum.Intel,
            AppRolesEnum.Premium,
            AppRolesEnum.Standard
          ]}
          renderNoAccess={() => <NoAccessError />}
        >
          <BrowserRouter>{content}</BrowserRouter>
        </AccessControlWrapper>
      </AuthWrapper>
    </MsalProvider>
  )
}

export default App
