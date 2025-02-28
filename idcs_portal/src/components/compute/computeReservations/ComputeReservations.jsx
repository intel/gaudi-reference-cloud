// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import GridPagination from '../../../utils/gridPagination/gridPagination'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import { NavLink } from 'react-router-dom'
import SearchBox from '../../../utils/searchBox/SearchBox'
import ExtendReservationModal from '../../../utils/modals/extendReservation/ExtendReservationModal'

const ComputeReservations = (props) => {
  // props
  const columns = props.columns
  const myreservations = props.myreservations
  const showActionModal = props.showActionModal
  const setShowActionModal = props.setShowActionModal
  const actionModalContent = props.actionModalContent
  const emptyGrid = props.emptyGrid
  const loading = props.loading
  const filterText = props.filterText
  const setFilter = props.setFilter
  const launchPagePath = props.launchPagePath
  const customLaunchButtonLabel = props?.customLaunchButtonLabel ? props?.customLaunchButtonLabel : 'Launch instance'
  const showExtendReservationModal = props?.showExtendReservationModal
  const extendModalState = props?.extendModalState
  const extendModalOnChange = props?.extendModalOnChange
  const extendModalOnSubmit = props?.extendModalOnSubmit
  const extendModalOnHide = props?.extendModalOnHide

  let gridItems = []

  if (filterText !== '') {
    const input = filterText.toLowerCase()
    gridItems = myreservations.filter((item) =>
      item.name.value ? item.name.value.toLowerCase().includes(input) : item.name.toLowerCase().includes(input)
    )
    if (gridItems.length === 0) {
      gridItems = myreservations.filter((item) => item.IpDefault.toLowerCase().includes(input))
    }
    if (gridItems.length === 0) {
      gridItems = myreservations.filter((item) =>
        item.displayName.value.instanceTypeDetails.name.toLowerCase().includes(input)
      )
    }
    if (gridItems.length === 0) {
      gridItems = myreservations.filter((item) => item.status.value.status.toLowerCase().includes(input))
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
      {showExtendReservationModal && (
        <ExtendReservationModal
          showModal={showExtendReservationModal}
          state={extendModalState}
          onChange={extendModalOnChange}
          onSubmit={extendModalOnSubmit}
          onHide={extendModalOnHide}
        />
      )}
      <div className={`filter ${!myreservations || myreservations.length === 0 ? 'd-none' : ''}`}>
        <NavLink to={launchPagePath} className="btn btn-primary" intc-id="btn-navigate Launch Instance">
          {customLaunchButtonLabel}
        </NavLink>
        <div className="d-flex justify-content-end">
          <SearchBox
            intc-id="searchInstances"
            value={filterText}
            onChange={setFilter}
            placeholder="Search instances..."
            aria-label="Type to search instance.."
          />
        </div>
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} emptyGrid={emptyGrid} loading={loading} fixedFirstColumn />
      </div>
    </>
  )
}

export default ComputeReservations
