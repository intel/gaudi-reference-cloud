// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { NavLink } from 'react-router-dom'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import SearchBox from '../../../utils/searchBox/SearchBox'

const SuperComputerReservations = ({
  myreservations,
  emptyGrid,
  columns,
  loading,
  filterText,
  setFilter,
  actionModal,
  actionOnModal
}) => {
  let gridItems = []

  if (filterText !== '' && myreservations) {
    const input = filterText.toLowerCase()
    gridItems = myreservations.filter((item) =>
      item['cluster-name'].value
        ? item['cluster-name'].value.toLowerCase().includes(input)
        : item['cluster-name'].toLowerCase().includes(input)
    )
    if (gridItems.length === 0) {
      gridItems = myreservations.filter((item) => item.storage.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = myreservations.filter((item) => item.status.value.clusterstate.toLowerCase().includes(input))
    }
  } else {
    gridItems = myreservations
  }

  return (
    <>
      <ActionConfirmation
        actionModalContent={actionModal}
        onClickModalConfirmation={actionOnModal}
        showModalActionConfirmation={actionModal.show}
      />
      <div className={`filter ${!myreservations || myreservations.length === 0 ? 'd-none' : ''}`}>
        <NavLink to="/supercomputer/launch" className="btn btn-primary" intc-id="launch-SCCluster">
          Launch Supercomputing Cluster
        </NavLink>
        <div className="d-flex justify-content-end">
          <SearchBox
            intc-id="searchCluster"
            value={filterText}
            onChange={setFilter}
            placeholder="Search clusters..."
            aria-label="Type to search cluster.."
          />
        </div>
      </div>
      <div className="section" intc-id="Super-Compute-Title">
        <GridPagination data={gridItems} columns={columns} loading={loading} emptyGrid={emptyGrid} fixedFirstColumn />
      </div>
    </>
  )
}

export default SuperComputerReservations
