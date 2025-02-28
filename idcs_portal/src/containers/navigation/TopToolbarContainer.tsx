// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useMemo, useState } from 'react'
import { getAllowedNavigations } from './Navigation'
import TopToolbar from '../../components/header/TopToolbar'
import useUserStore from '../../store/userStore/UserStore'
import { type IdcNavigation } from './Navigation.types'

const TopToolbarContainer: React.FC = () => {
  const isOwnCloudAccount = useUserStore((state) => state.isOwnCloudAccount)
  const [currentPath, setCurrentPath] = useState(window.location.pathname)

  const isRouteActive = (pathname: string, children?: IdcNavigation[]): boolean => {
    const paramWildcard = '/:param'
    return (
      currentPath.toLowerCase() === pathname.toLowerCase() ||
      (!children?.some((x) => currentPath.toLowerCase() === x.path) &&
        currentPath.toLowerCase().startsWith(`${pathname.toLowerCase()}/`)) ||
      (!children?.some((x) => currentPath.toLowerCase() === x.path) &&
        pathname.includes(paramWildcard) &&
        currentPath.toLowerCase().startsWith(`${pathname.split(paramWildcard)[0].toLowerCase()}/`))
    )
  }

  const navItemParent = useMemo(() => {
    const navigationArr = getAllowedNavigations(isOwnCloudAccount)
    const parentNavigation = navigationArr.find(
      (x) =>
        x.path?.toLowerCase() === currentPath.toLowerCase() ||
        x.children?.some((y) => isRouteActive(y.path, x.children))
    )
    return parentNavigation
  }, [isOwnCloudAccount, currentPath])

  useEffect(() => {
    setCurrentPath(window.location.pathname)
  }, [window.location.pathname])

  return <TopToolbar idcNavigation={navItemParent} currentPath={currentPath} />
}

export default TopToolbarContainer
