// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Route, Routes, useLocation } from 'react-router-dom'

import Header from '../header/Header'
import FooterMini from '../footer/FooterMini'

import ErrorBoundary from '../../pages/error/ErrorBoundary'
import { ErrorBoundaryLevel } from '../../utils/Enums'
import { type User } from '../../store/userStore/UserStore'
import { type IdcRoute } from '../../containers/routing/Routes.types'
import ToastContainer from '../../utils/toast/ToastContainer'
import useAppStore from '../../store/appStore/AppStore'
import LearningBarContainer from '../../containers/navigation/LearningBarContainer'
import SideNavBarContainer from '../../containers/navigation/SideNavBarContainer'
import { Container } from 'react-bootstrap'
import BreadcrumbContainer from '../../containers/navigation/BreadcrumbContainer'
import { isFeatureRegionBlocked } from '../../config/configurator'
import FeatureNotAvailable from '../../pages/error/FeatureNotAvailable'
import GlobalSearchParams from './GlobalSearchParams'

interface RoutesMainProps {
  userDetails: User | null
  isOwnCloudAccount: boolean
  allowedRoutes: IdcRoute[]
}

const hideMenuRoutes = ['/premium', '/accounts']

const RoutesMain: React.FC<RoutesMainProps> = ({ userDetails, isOwnCloudAccount, allowedRoutes }): JSX.Element => {
  const { pathname } = useLocation()
  const showLearningBar = useAppStore((state) => state.showLearningBar)
  const learningArticlesAvailable = useAppStore((state) => state.learningArticlesAvailable)
  const showSideNavBar = useAppStore((state) => state.showSideNavBar)

  const hideMenu = hideMenuRoutes.some((x) => x === pathname.toLowerCase())

  return (
    <>
      <GlobalSearchParams />
      <Header userDetails={userDetails} isOwnCloudAccount={isOwnCloudAccount} pathname={pathname} hideMenu={hideMenu} />
      <Container
        role="main"
        className={`siteContainer ${!hideMenu && showLearningBar && learningArticlesAvailable ? 'learningBarMargingEnd' : ''} ${!hideMenu && showSideNavBar ? 'sideNavBarMargingStart' : ''}`}
      >
        <ErrorBoundary errorBoundaryLevel={ErrorBoundaryLevel.RouteLevel}>
          <div className="sheet">
            <BreadcrumbContainer />
            <Routes>
              {allowedRoutes.map((x) => (
                <Route
                  key={x.path}
                  path={x.path}
                  Component={
                    (x.href ?? isFeatureRegionBlocked(x.featureFlag ?? ''))
                      ? () => {
                          if (isFeatureRegionBlocked(x.featureFlag ?? '')) {
                            return <FeatureNotAvailable featureFlag={x.featureFlag} />
                          }
                          window.location.href = x.href ?? ''
                          return null
                        }
                      : x.component
                  }
                />
              ))}
            </Routes>
          </div>
          <ToastContainer />
        </ErrorBoundary>
      </Container>
      <LearningBarContainer />
      {!hideMenu && <SideNavBarContainer />}
      <FooterMini />
    </>
  )
}

export default RoutesMain
