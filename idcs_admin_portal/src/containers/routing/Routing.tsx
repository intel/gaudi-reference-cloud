// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useMemo, useState } from 'react'
import { withMsal } from '@azure/msal-react'
import useUserStore from '../../store/userStore/UserStore'
import RoutesMain from '../../components/routes/RoutesMain'
import { routes } from './Routes'
import { isFeatureFlagEnable, isFeatureRegionBlocked } from '../../config/configurator'
import { checkRoles } from '../../utility/wrapper/AccessControlWrapper'
import { type IdcRoute } from './Routes.types'
import { useLocation } from 'react-router-dom'

export const getAllowedRoutes = (isOwnCloudAccount: boolean): IdcRoute[] => {
  const allowedRoutes = routes.filter(
    (x) =>
      (!x.featureFlag || isFeatureFlagEnable(x.featureFlag) || isFeatureRegionBlocked(x.featureFlag)) &&
      (!x.roles || checkRoles(x.roles)) &&
      (!x.memberNotAllowed || isOwnCloudAccount) &&
      (!x.allowedFn || x.allowedFn())
  )
  return allowedRoutes
}

const Routing = (): JSX.Element => {
  const { pathname } = useLocation()
  const user = useUserStore((state) => state.user)
  const isOwnCloudAccount = useUserStore((state) => state.isOwnCloudAccount)
  const [previousPath, setPreviousPath] = useState(window.location.pathname)

  useEffect(() => {
    if (previousPath !== pathname) {
      setPreviousPath(window.location.pathname)
    }
  }, [pathname])

  const allowedRoutes = useMemo(() => {
    const routesArr = getAllowedRoutes(isOwnCloudAccount)
    return routesArr
  }, [])

  return <RoutesMain allowedRoutes={allowedRoutes} userDetails={user} isOwnCloudAccount={isOwnCloudAccount} />
}

export default withMsal(Routing)
