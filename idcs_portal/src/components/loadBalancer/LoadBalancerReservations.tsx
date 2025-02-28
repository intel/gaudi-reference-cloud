// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import GridPagination from '../../utils/gridPagination/gridPagination'
import ActionConfirmation from '../../utils/modals/actionConfirmation/ActionConfirmation'
import { NavLink } from 'react-router-dom'
import SearchBox from '../../utils/searchBox/SearchBox'

const LoadBalancerReservations = (props: any): JSX.Element => {
  // props

  const myreservations = props.myreservations
  const columns = props.columns
  const showActionModal = props.showActionModal
  const actionModalContent = props.actionModalContent
  const emptyGrid = props.emptyGrid
  const loading = props.loading
  const filterText = props.filterText
  const setFilter = props.setFilter
  const setShowActionModal = props.setShowActionModal

  // variables
  let gridItems = []

  if (filterText !== '') {
    const input = filterText.toLowerCase()

    gridItems = myreservations.filter((item: any) =>
      item.name.value ? item.name.value.toLowerCase().includes(input) : item.name.toLowerCase().includes(input)
    )

    if (gridItems.length === 0) {
      gridItems = myreservations.filter((item: any) => item.vip.toLowerCase().includes(input))
    }

    if (gridItems.length === 0) {
      gridItems = myreservations.filter((item: any) => item.status.value.status.toLowerCase().includes(input))
    }
  } else {
    gridItems = myreservations
  }

  return (
    <>
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={setShowActionModal}
        showModalActionConfirmation={showActionModal}
      />
      <div className={`filter ${!myreservations || myreservations.length === 0 ? 'd-none' : ''}`}>
        <NavLink
          to="/load-balancer/reserve"
          aria-label="Launch Load Balancer"
          className="btn btn-primary"
          intc-id={'btn-navigate-Create-load-balancer'}
        >
          Launch Load Balancer
        </NavLink>
        <div className="d-flex justify-content-end">
          <SearchBox
            intc-id="searchLoadBalancers"
            value={filterText}
            onChange={setFilter}
            placeholder="Search load balancers..."
            aria-label="Type to search load balancers.."
          />
        </div>
      </div>

      <div className="section">
        <GridPagination data={gridItems} columns={columns} emptyGrid={emptyGrid} loading={loading} fixedFirstColumn />
      </div>
    </>
  )
}

export default LoadBalancerReservations
