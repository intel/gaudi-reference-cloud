// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { NavLink } from 'react-router-dom'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import SearchBox from '../../../utils/searchBox/SearchBox'
const ClusterMyReservations = (props) => {
  // *****
  // props
  // *****

  const columns = props.columns
  const myreservations = props.myreservations
  const showActionModal = props.showActionModal
  const actionOnModal = props.actionOnModal
  const actionModalContent = props.actionModalContent
  const emptyGrid = props.emptyGrid
  const filterText = props.filterText
  const loading = props.loading
  const clusterLimit = props.clusterLimit
  const gridFeedBack = `${myreservations.length} of ${clusterLimit} clusters requested`
  const setFilter = props.setFilter
  // *****
  // variables
  // *****

  let gridItems = []

  if (filterText !== '' && myreservations) {
    const input = filterText.toLowerCase()
    gridItems = myreservations.filter((item) =>
      item['cluster-name'].value
        ? item['cluster-name'].value.toLowerCase().includes(input)
        : item['cluster-name'].toLowerCase().includes(input)
    )
    if (gridItems.length === 0) {
      gridItems = myreservations.filter((item) => item.k8sversion.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = myreservations.filter((item) => item.status.value.clusterstate.toLowerCase().includes(input))
    }
  } else {
    gridItems = myreservations
  }

  // *****
  // functions
  // *****

  // variables

  return (
    <>
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={actionOnModal}
        showModalActionConfirmation={showActionModal}
      />
      <div className={`filter ${!myreservations || myreservations.length === 0 ? 'd-none' : ''}`}>
        <NavLink
          to="/cluster/reserve"
          className="btn btn-primary"
          intc-id="btn-iksLaunch-cluster"
          data-wap_ref="btn-iksLaunch-cluster"
        >
          Launch Cluster
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
      <div className="section">
        <GridPagination
          feedBack={gridFeedBack}
          data={gridItems}
          columns={columns}
          emptyGrid={emptyGrid}
          loading={loading}
          fixedFirstColumn
        />
      </div>
    </>
  )
}

export default ClusterMyReservations
