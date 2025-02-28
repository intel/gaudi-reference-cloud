// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import GridPagination from '../../utility/gridPagination/gridPagination'
import './CloudCreditsView.scss'
import LabelValuePair from '../../utility/labelValuePair/LabelValuePair'
import SearchBox from '../../utility/searchBox/SearchBox'
import Spinner from '../../utility/modals/spinner/Spinner'

interface CloudCreditsViewProps {
  state: any
  columns: any[]
  filteredCredits: any[]
  myRemainingCredits: string
  myUsedCredits: string
  hasCredits: boolean
  currentDateTime: string
  loading: boolean
  onChangeInput: (event: any, formInputName: string) => void
}

const CloudCreditsView: React.FC<CloudCreditsViewProps> = (props): JSX.Element => {
  // props
  const loading = props.loading
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
  myCredits.sort(function (a: any, b: any): number {
    return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  })

  const emptyGridByFilter = {
    title: 'No history found',
    subTitle: 'The applied filter criteria did not match any items.',
    action: {
      type: 'function' as any,
      href: () => {
        onChangeInput({ target: { value: '' } }, 'filterCredits')
      },
      label: 'Clear filters'
    }
  }

  return (
    <>
      {loading ? (
        <Spinner />
      ) : (
        <>
          {!hasCredits ? (
            <div className="section text-center align-items-center">
              <h1>No cloud credits</h1>
              <p>Selected account currently has no cloud credits.</p>
            </div>
          ) : (
            <>
              <div className="section">
                <h4 intc-id="cloudCreditsTitle" className="h4">
                  Cloud Credits
                </h4>
                <span>As of {currentDateTime} </span>
              </div>
              <div className="section">
                <h4 className="h4">Balance</h4>
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
                </div>
                <span className="valid-feedback">
                  *Any overage charges will be automatically deducted from most recent redeemed coupon.
                </span>
              </div>
              <div className="section">
                <h4 className="h4">Cloud credit history</h4>
                <div className="d-flex ms-md-auto">
                  <SearchBox
                    intc-id="FiltercreditsInput"
                    aria-label="Filter credits"
                    value={configInput.value}
                    placeholder="Type to filter credits..."
                    onChange={(event) => {
                      onChangeInput(event, 'filterCredits')
                    }}
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
