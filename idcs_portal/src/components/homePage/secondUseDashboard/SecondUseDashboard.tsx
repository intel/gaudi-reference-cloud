// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useMemo } from 'react'
import UsageWidgetContainer from '../../../containers/homePage/secondUseDashboard/UsageWidgetContainer'
import ChartsWidgetContainer from '../../../containers/homePage/secondUseDashboard/ChartsWidgetContainer'
import RecentlyVisitedDashboardContainer from '../../../containers/homePage/secondUseDashboard/RecentlyVisitedDashboardContainer'

interface SecondUseDashboardProps {
  showCloudCredits: boolean
  isOwnCloudAccount: boolean
  navigateTo: (url: string) => void
}

const SecondUseDashboard: React.FC<SecondUseDashboardProps> = (props): JSX.Element => {
  const Charts = useMemo(() => <ChartsWidgetContainer />, [])
  const Usages = useMemo(
    () => <UsageWidgetContainer showCloudCredits={props.showCloudCredits} />,
    [props.showCloudCredits]
  )

  return (
    <>
      <div className="col-md-4 col-sm-6 col-xs-12 order-0">
        <RecentlyVisitedDashboardContainer />
      </div>
      <div className={'col-xl-9 col-md-8 d-flex flex-column gap-s8 flex-fill'}>
        {Charts}
        {props.isOwnCloudAccount ? <div className="col-xl-6 d-xl-block d-none">{Usages}</div> : null}
      </div>
      {props.isOwnCloudAccount ? <div className="col-lg-6 col-md-8 d-xl-none d-xs-block">{Usages}</div> : null}
    </>
  )
}

export default SecondUseDashboard
