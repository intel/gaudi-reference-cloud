// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { NavLink } from 'react-router-dom'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import './CloudCreditsView.scss'
import LabelValuePair from '../../../utils/labelValuePair/LabelValuePair'
import SearchBox from '../../../utils/searchBox/SearchBox'
import Spinner from '../../../utils/spinner/Spinner'

const CloudCreditsView = (props) => {
  // props
  const state = props.state
  const form = state.form
  const columns = props.columns
  const myCredits = props.filteredCredits
  const myRemainingCredits = props.myRemainingCredits
  const myUsedCredits = props.myUsedCredits
  const currentDateTime = props.currentDateTime
  const hasCredits = props.hasCredits

  // props function
  const onChangeInput = props.onChangeInput
  // variables
  const configInput = { ...form.filterCredits }

  // sort myCredits by date
  myCredits.sort(function (a, b) {
    return new Date(b.createdAt) - new Date(a.createdAt)
  })

  const emptyGridByFilter = {
    title: 'No history found',
    subTitle: 'The applied filter criteria did not match any items.',
    action: {
      type: 'function',
      href: () => onChangeInput({ target: { value: '' } }, 'filterCredits'),
      label: 'Clear filters'
    }
  }

  return (
    <>
      {props.loading ? (
        <Spinner />
      ) : (
        <>
          {!hasCredits ? (
            <div className="section text-center align-items-center">
              <h1>No cloud credits</h1>
              <p>
                Your account currently has no cloud credits.
                <br />
                Redeem a coupon to get cloud credits.
              </p>
              <NavLink
                to={'/billing/credits/managecouponcode'}
                aria-label="Redeem Coupon"
                intc-id="redeemCouponButton"
                className="btn btn-primary"
              >
                Redeem coupon
              </NavLink>
            </div>
          ) : (
            <>
              <div className="section">
                <h2 intc-id="cloudCreditsTitle">Cloud Credits</h2>
                <span>As of {currentDateTime} </span>
              </div>
              <div className="section">
                <h3>Balance</h3>
                <div className="d-flex flex-xs-column flex-md-row creditBalanceToolbar gap-xs-s6 gap-md-s8">
                  <LabelValuePair label="Estimated remaining credits:" className="gap-s3" labelClassName="h5">
                    <span className="lead" intc-id="remainingCreditsLabel">
                      {myRemainingCredits}
                    </span>
                  </LabelValuePair>
                  <LabelValuePair label="Estimated used credits:" className="gap-s3" labelClassName="h5">
                    <span className="lead" intc-id="usedCreditsLabel">
                      {myUsedCredits}
                    </span>
                  </LabelValuePair>
                  {
                    <div className="d-flex ms-md-auto">
                      <NavLink
                        to={'/billing/credits/managecouponcode'}
                        aria-label="Redeem Coupon"
                        intc-id="redeemCouponButton"
                        className="btn btn-primary mb-auto"
                      >
                        Redeem coupon
                      </NavLink>
                    </div>
                  }
                </div>
                <span className="valid-feedback">
                  *Any overage charges will be automatically deducted from your most recent redeemed coupon.
                </span>
              </div>
              <div className="section">
                <h3>Cloud credit history</h3>
                <div className="d-flex ms-md-auto">
                  <SearchBox
                    intc-id="FiltercreditsInput"
                    aria-label="Filter credits"
                    value={configInput.value}
                    placeholder="Type to filter credits..."
                    onChange={(event) => onChangeInput(event, 'filterCredits')}
                  />
                </div>
              </div>
              <div className="section">
                <GridPagination data={myCredits} emptyGrid={emptyGridByFilter} columns={columns} loading={false} />
              </div>
            </>
          )}
        </>
      )}
    </>
  )
}

export default CloudCreditsView
