// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import GridPagination from '../../../utils/gridPagination/gridPagination'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import { NavLink } from 'react-router-dom'
import SearchBox from '../../../utils/searchBox/SearchBox'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'

const ObjectStorageUsersReservations = (props) => {
  // props
  const columns = props.columns
  const myUsers = props.myUsers
  const showActionModal = props.showActionModal
  const actionModalContent = props.actionModalContent
  const emptyGrid = props.emptyGrid
  const loading = props.loading
  const filterText = props.filterText
  const setShowActionModal = props.setShowActionModal
  const setFilter = props.setFilter
  const errorModalContent = props.errorModalContent
  const showErrorModal = props.showErrorModal
  const setShowErrorModal = props.setShowErrorModal
  let gridItems = []

  if (filterText !== '') {
    const input = filterText.toLowerCase()

    gridItems = myUsers.filter((item) =>
      item.name.value ? item.name.value.toLowerCase().includes(input) : item.name.toLowerCase().includes(input)
    )

    if (gridItems.length === 0) {
      gridItems = myUsers.filter((item) => item.status.value.status.toLowerCase().includes(input))
    }
  } else {
    gridItems = myUsers
  }

  return (
    <>
      <ErrorModal
        showModal={showErrorModal}
        titleMessage={errorModalContent.titleMessage}
        description={errorModalContent.description}
        message={errorModalContent.message}
        onClickCloseErrorModal={() => {
          setShowErrorModal(false)
        }}
      />
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={setShowActionModal}
        showModalActionConfirmation={showActionModal}
      />
      <div className="section">
        <h2>Manage Principals and Permissions</h2>
      </div>
      <div className={`filter ${!myUsers || myUsers.length === 0 ? 'd-none' : ''}`}>
        <NavLink to="/buckets/users/reserve" className="btn btn-primary" intc-id={'btn-navigate-Create-users'}>
          Create principal
        </NavLink>
        <NavLink
          to="/buckets"
          className="btn btn-outline-primary me-auto"
          intc-id={'btn-navigate-Manage-users-and-permissions'}
          end
        >
          View buckets
        </NavLink>
        <div className="d-flex justify-content-end">
          <SearchBox
            intc-id="searchUsers"
            value={filterText}
            onChange={setFilter}
            placeholder="Search principals..."
            aria-label="Type to search principals.."
          />
        </div>
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} emptyGrid={emptyGrid} loading={loading} fixedFirstColumn />
      </div>
    </>
  )
}

export default ObjectStorageUsersReservations
