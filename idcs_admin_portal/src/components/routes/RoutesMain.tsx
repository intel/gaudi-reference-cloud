// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Route, Routes } from 'react-router-dom'

import Header from '../header/Header'
import FooterMini from '../footer/FooterMini'

import ErrorBoundary from '../../pages/error/ErrorBoundary'
import { ErrorBoundaryLevel } from '../../utility/Enums'
import { type User } from '../../store/userStore/UserStore'
import { type IdcRoute } from '../../containers/routing/Routes.types'
import ToastContainer from '../../utility/toast/ToastContainer'
import { isFeatureRegionBlocked } from '../../config/configurator'
import { Container } from 'react-bootstrap'
import FeatureNotAvailable from '../../pages/error/FeatureNotAvailable'
import GlobalSearchParams from './GlobalSearchParams'

interface RoutesMainProps {
  userDetails: User | null
  isOwnCloudAccount: boolean
  allowedRoutes: IdcRoute[]
}

const RoutesMain: React.FC<RoutesMainProps> = ({ allowedRoutes }): JSX.Element => {
  return (
    <>
      <GlobalSearchParams />
      <Header />
      <Container role="main" className={'siteContainer'}>
        <ErrorBoundary errorBoundaryLevel={ErrorBoundaryLevel.RouteLevel}>
          <div className="sheet">
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
      <FooterMini validateRoute />
    </>
  )
}

export default RoutesMain
