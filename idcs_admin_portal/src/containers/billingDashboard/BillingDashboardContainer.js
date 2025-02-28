import React, { Component } from 'react'
import BillingDashboardMainPage from '../../pages/billingDashboard/BillingDashboardMainPage'

class BillingDashboardContainer extends Component {
  // A container handle the function, api call and redux for each module

  render () {
    return (
            <BillingDashboardMainPage />
    )
  }
}

export default BillingDashboardContainer
