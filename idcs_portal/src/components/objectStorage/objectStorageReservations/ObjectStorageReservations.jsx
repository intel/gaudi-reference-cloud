// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import GridPagination from '../../../utils/gridPagination/gridPagination'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import { NavLink } from 'react-router-dom'
import AfterAction from '../../../utils/modals/afterAction/AfterAction'
import SearchBox from '../../../utils/searchBox/SearchBox'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'

const ObjectStorageReservations = (props) => {
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
  const afterActionShowModal = props.afterActionShowModal
  const afterActionModalContent = props.afterActionModalContent
  const afterActionClickModal = props.setAfterActionShowModal
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
      <AfterAction
        showModal={afterActionShowModal}
        modalContent={afterActionModalContent}
        onClickModal={afterActionClickModal}
      />
      <div className={'filter'}>
        <NavLink
          to="/buckets/reserve"
          className={`btn btn-primary ${!myreservations || myreservations.length === 0 ? 'd-none' : ''}`}
          intc-id={'btn-navigate-Create-bucket'}
        >
          Create bucket
        </NavLink>
        <NavLink
          to="/buckets/users"
          className="btn btn-outline-primary me-auto"
          intc-id={'btn-navigate-Manage-users-and-permissions'}
        >
          Manage principals and permissions
        </NavLink>
        <div className={`d-flex justify-content-end ${!myreservations || myreservations.length === 0 ? 'd-none' : ''}`}>
          <SearchBox
            intc-id="searchBuckets"
            value={filterText}
            onChange={setFilter}
            placeholder="Search buckets..."
            aria-label="Type to search bucket.."
          />
        </div>
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} emptyGrid={emptyGrid} loading={loading} fixedFirstColumn />
      </div>
    </>
  )
}

export default ObjectStorageReservations
