// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import GridPagination from '../../../utils/gridPagination/gridPagination'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import { NavLink } from 'react-router-dom'
import SearchBox from '../../../utils/searchBox/SearchBox'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'

const StorageReservations = (props) => {
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
  const errorModalContent = props.errorModalContent
  const showErrorModal = props.showErrorModal
  const setShowErrorModal = props.setShowErrorModal

  let gridItems = []

  if (filterText !== '') {
    const input = filterText.toLowerCase()

    gridItems = myreservations.filter((item) =>
      item.name.value ? item.name.value.toLowerCase().includes(input) : item.name.toLowerCase().includes(input)
    )

    if (gridItems.length === 0) {
      gridItems = myreservations.filter((item) => item.size.toLowerCase().includes(input))
    }

    if (gridItems.length === 0) {
      gridItems = myreservations.filter((item) => item.status.value.status.toLowerCase().includes(input))
    }
  } else {
    gridItems = myreservations
  }

  return (
    <>
      <ErrorModal
        showModal={showErrorModal}
        titleMessage={errorModalContent.titleMessage}
        description={errorModalContent.description}
        message={errorModalContent.message}
        onClickCloseErrorModal={() => setShowErrorModal(false)}
      />
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={setShowActionModal}
        showModalActionConfirmation={showActionModal}
      />
      <div className={`filter ${!myreservations || myreservations.length === 0 ? 'd-none' : ''}`}>
        <NavLink to="/storage/reserve" className="btn btn-primary" intc-id="btn-navigate-create-volume">
          Create volume
        </NavLink>
        <div className="d-flex justify-content-end">
          <SearchBox
            intc-id="searchVolumes"
            value={filterText}
            onChange={setFilter}
            placeholder="Search volumes..."
            aria-label="Type to search volume.."
          />
        </div>
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} emptyGrid={emptyGrid} loading={loading} fixedFirstColumn />
      </div>
    </>
  )
}

export default StorageReservations
