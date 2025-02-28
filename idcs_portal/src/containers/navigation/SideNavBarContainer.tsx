// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useMemo, useState, useEffect } from 'react'
import SideNavBar from '../../components/header/SideNavBar'
import { getAllowedNavigations } from './Navigation'
import useUserStore from '../../store/userStore/UserStore'
import useAppStore from '../../store/appStore/AppStore'
import { useMediaQuery } from 'react-responsive'

const SideNavBarContainer: React.FC = (): JSX.Element => {
  const isOwnCloudAccount = useUserStore((state) => state.isOwnCloudAccount)
  const setShowSideNavBar = useAppStore((state) => state.setShowSideNavBar)
  const showSideNavBar = useAppStore((state) => state.showSideNavBar)
  const [currentPath, setCurrentPath] = useState(window.location.pathname)
  const isSmScreen = useMediaQuery({
    query: '(max-width: 767px)'
  })

  const allowedNavigations = useMemo(() => {
    const navigationArr = getAllowedNavigations(isOwnCloudAccount)
    return navigationArr
  }, [isOwnCloudAccount])

  useEffect(() => {
    if (isSmScreen) {
      setShowSideNavBar(false, false)
    } else {
      const showSideNav = localStorage.getItem('showSideNav')
      setShowSideNavBar(showSideNav === 'true' || showSideNav === null, true)
    }
  }, [])

  useEffect(() => {
    if (isSmScreen && currentPath !== window.location.pathname) {
      setShowSideNavBar(false, false)
    }
    setCurrentPath(window.location.pathname)
  }, [window.location.pathname])

  return <SideNavBar navigations={allowedNavigations} showSideNavBar={showSideNavBar} currentPath={currentPath} />
}

export default SideNavBarContainer
