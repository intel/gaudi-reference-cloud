// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import DashboardInfoCard from './DashboardInfoCard'
import './HomePage.scss'
import SecondUseDashboardContainer from '../../containers/homePage/secondUseDashboard/SecondUseDashboardContainer'
import { type initialCardSate, type GetStartedCard } from '../../containers/homePage/Home.types'

interface HomePageProps {
  cardState: initialCardSate
  onRedirectTo: (value: string) => void
}

const HomePage: React.FC<HomePageProps> = (props): JSX.Element => {
  const cardState = props.cardState

  const onRedirectTo: any = props.onRedirectTo

  // variable
  const DASHBOARD_INFO_CARDS = ['learn', 'evaluate', 'deploy']

  return (
    <>
      <div className="section dashboard-container gap-s8">
        <div className="row g-s8">
          {DASHBOARD_INFO_CARDS.map((cardType, index) => {
            let cardInfo: GetStartedCard | undefined

            switch (cardType) {
              case 'learn':
                cardInfo = cardState.learn
                break
              case 'deploy':
                cardInfo = cardState.deploy
                break
              default:
                break
            }

            if (!cardInfo) {
              return <></>
            }
            return (
              <div key={index} className="col-xs-12 col-sm-6 col-md-4">
                <DashboardInfoCard cardInfo={cardInfo} onRedirectTo={onRedirectTo} />
              </div>
            )
          })}
          <SecondUseDashboardContainer />
        </div>
      </div>
    </>
  )
}

export default HomePage
