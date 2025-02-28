// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Button } from 'react-bootstrap'
import { NavLink } from 'react-router-dom'
import GridPagination from '../../utility/gridPagination/gridPagination'
import SearchBox from '../../utility/searchBox/SearchBox'
import DeleteModal from '../../utility/modals/deleteModal/DeleteModal'

const ProductVendors = (props: any): JSX.Element => {
  const onCancel = props.onCancel
  const emptyGrid = props.emptyGrid
  const productColumns = props.productColumns
  const productItems = props.productItems
  const loading = props.loading
  const filterText = props.filterText
  const setFilter = props.setFilter
  const deleteModal = props.deleteModal
  const onCloseDeleteModal = props.onCloseDeleteModal
  let gridItems = productItems

  if (filterText !== '' && productItems) {
    const input = filterText.toLowerCase()
    gridItems = productItems.filter(
      (item: any) =>
        item?.name.toString().toLowerCase().includes(input) ||
        item?.description.toString().toLowerCase().includes(input) ||
        item?.organizationName.toString().toLowerCase().includes(input)
    )
  }

  return (
    <>
      <DeleteModal modalContent={deleteModal} onClickModalConfirmation={onCloseDeleteModal} />
      <div className="section">
        <Button variant="link" className="p-s0" onClick={() => onCancel()}>
          ⟵ Back to Home
        </Button>
      </div>
      <div className="section">
        <h2 className="h4">Vendor Management</h2>
      </div>
      <div className="section">
        <div className="filter flex-wrap p-0">
          <NavLink to="/products/vendors/create" className="btn btn-primary" intc-id="btn-createNewVendor">
            Create new vendor
          </NavLink>
          <SearchBox
            intc-id="filterVendors"
            value={filterText}
            onChange={setFilter}
            placeholder="Filter vendors..."
            aria-label="Type to filter vendors..."
          />
        </div>
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={productColumns} loading={loading} emptyGrid={emptyGrid} />
      </div>
    </>
  )
}

export default ProductVendors
