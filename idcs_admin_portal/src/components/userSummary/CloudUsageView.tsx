// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Button, Form } from 'react-bootstrap'
import GridPagination from '../../utility/gridPagination/gridPagination'
import LabelValuePair from '../../utility/labelValuePair/LabelValuePair'
import SearchBox from '../../utility/searchBox/SearchBox'
import Spinner from '../../utility/modals/spinner/Spinner'

const CloudUsageView = (props: any): JSX.Element => {
  // props
  const state = props.state
  const form = state.form
  const columns = props.columns
  const myUsages = props.filteredUsages
  const totalAmount = props.totalAmount
  const hasUsages = props.hasUsages
  const currentDateTime = props.currentDateTime

  // props function
  const onChangeInput = props.onChangeInput

  // variables
  const configInput = { ...form.filterUsages }

  const emptyGridByFilter = {
    title: 'No usages found',
    subTitle: 'The applied filter criteria did not match any items.',
    action: {
      type: 'function' as any,
      href: () => {
        refreshFilters()
      },
      label: 'Clear filters'
    }
  }

  const refreshFilters = (): void => {
    onChangeInput(null, 'clearFilters')
  }

  const getLoading = (): JSX.Element => {
    return <Spinner />
  }

  const getNoUsage = (): JSX.Element => {
    return (
      <div className="section text-center align-items-center">
        <h1>No usages found</h1>
        <p>Your account currently has no usage to report for this month.</p>
      </div>
    )
  }

  const getGrid = (): JSX.Element => {
    return (
      <div className="section">
        <GridPagination
          data={myUsages}
          emptyGrid={emptyGridByFilter}
          columns={columns}
          loading={props.loading}
          hidePaginationControl
        />
      </div>
    )
  }

  const getFilter = (): JSX.Element => {
    return (
      <>
        <div className="d-flex flex-wrap gap-s6">
          <Form.Group className="d-flex-customInput w-auto">
            <SearchBox
              intc-id="FilterInput"
              aria-label={configInput.label}
              value={configInput.value}
              placeholder={configInput.placeholder}
              onChange={(event) => onChangeInput(event, 'filterUsages')}
            />
          </Form.Group>
        </div>
        <div className="d-flex flex-column gap-s6">
          <Button
            intc-id="clearFilterButton"
            data-wap_ref="clearFilterButton"
            aria-label="Clear filters"
            variant="outline-primary"
            onClick={() => {
              refreshFilters()
            }}
          >
            Clear filters
          </Button>
        </div>
      </>
    )
  }

  const getTitle = (): JSX.Element => {
    return (
      <>
        <div className="section">
          <h4 className="h4">Current month usage</h4>
          <span>As of {currentDateTime} </span>
          <div className="d-flex flex-xs-column flex-md-row w-100 gap-xs-s6 gap-md-s8">
            <LabelValuePair label="Estimated cost:" className="gap-s3 flex-fill" labelClassName="h4">
              <span className="lead" intc-id="totalAmountLabel">
                ${totalAmount}
              </span>
            </LabelValuePair>
            {getFilter()}
          </div>
        </div>
      </>
    )
  }

  const getContent = (): JSX.Element => {
    return !hasUsages ? (
      getNoUsage()
    ) : (
      <>
        {getTitle()} {getGrid()}
      </>
    )
  }

  return props.loading ? getLoading() : getContent()
}

export default CloudUsageView
