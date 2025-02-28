// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React from 'react'
import { Button } from 'react-bootstrap'
import { NavLink } from 'react-router-dom'
import GridPagination from '../../../utility/gridPagination/gridPagination'
import SearchBox from '../../../utility/searchBox/SearchBox'
import DeleteModal from '../../../utility/modals/deleteModal/DeleteModal'

const QuotaManagementService = (props: any): JSX.Element => {
  const onCancel = props.onCancel
  const emptyGrid = props.emptyGrid
  const loading = props.loading
  const seviceColumns = props.seviceColumns
  const serviceItems = props.serviceItems
  const filterText = props.filterText
  const setFilter = props.setFilter
  const deleteModal = props.deleteModal
  const onCloseDeleteModal = props.onCloseDeleteModal
  let gridItems = serviceItems

  if (filterText !== '' && serviceItems) {
    const input = filterText.toLowerCase()
    gridItems = serviceItems.filter((item: any) => item.serviceName.toString().startsWith(input))
  }

  return <>
  <DeleteModal modalContent={deleteModal} onClickModalConfirmation={onCloseDeleteModal} />
    <div className="section">
      <Button variant="link" className="p-s0" onClick={() => onCancel()}>
        ⟵ Back to Home
      </Button>
    </div>
    <div className="section">
      <h2 className='h4'>Service Management</h2>
    </div>
    <div className="section">
      <div className="filter flex-wrap p-0">
        <NavLink to="/quotamanagement/services/create"
          className="btn btn-primary" intc-id="btn-ManageStorageQuota">
          Create new service
        </NavLink>
        <SearchBox
          intc-id="filterAccounts"
          value={filterText}
          onChange={setFilter}
          placeholder="Filter services..."
          aria-label="Type to filter cloud services..."
        />
      </div>
    </div>
    <div className="section">
      <GridPagination data={gridItems} columns={seviceColumns} loading={loading} emptyGrid={emptyGrid} />
    </div>
  </>
}

export default QuotaManagementService
