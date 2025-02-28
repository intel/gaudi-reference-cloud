// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import useAppStore from '../../../store/appStore/AppStore'
import RecentlyVisitedDashboard from '../../../components/homePage/secondUseDashboard/RecentlyVisitedDashboard'
import { useNavigate } from 'react-router-dom'

const RecentlyVisitedDashboardContainer = (): JSX.Element => {
  const visitedSites = useAppStore((state) => state.visitedSites)

  const navigate = useNavigate()

  const onRedirectTo = (url: string): void => {
    navigate({ pathname: url })
  }

  return <RecentlyVisitedDashboard visitedSites={visitedSites} onRedirectTo={onRedirectTo} />
}
export default RecentlyVisitedDashboardContainer
