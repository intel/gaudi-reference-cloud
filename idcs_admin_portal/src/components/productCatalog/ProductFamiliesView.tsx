// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Button } from 'react-bootstrap'
import SearchBox from '../../utility/searchBox/SearchBox'
import GridPagination from '../../utility/gridPagination/gridPagination'
import DeleteModal from '../../utility/modals/deleteModal/DeleteModal'

const ProductFamiliesView = (props: any): JSX.Element => {
  const columns = props.columns
  const rows = props.rows
  const emptyGrid = props.emptyGrid
  const filterText = props.filterText
  const loading = props.loading
  const onRedirect = props.onRedirect
  const setFilter = props.setFilter
  const backToHome = props.backToHome
  const deleteModal = props.deleteModal
  const onCloseDeleteModal = props.onCloseDeleteModal

  function getFilteredData(): any[] {
    let filteredData = []

    if (filterText !== '') {
      for (const index in rows) {
        const quota = { ...rows[index] }
        if (getFilterCheck(filterText, quota)) {
          filteredData.push(quota)
        }
      }
    } else {
      filteredData = [...rows]
    }

    return filteredData
  }

  function getFilterCheck(filterValue: string, data: any): boolean {
    filterValue = filterValue.toLowerCase()

    return data.name.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data.description.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data.vendor.toString().toLowerCase().indexOf(filterValue) > -1
  }

  const gridItems = getFilteredData()

  return (
    <>
      <DeleteModal modalContent={deleteModal} onClickModalConfirmation={onCloseDeleteModal} />
      <div className="section">
        <Button variant="link" className='p-s0' onClick={() => backToHome()}>
          ⟵ Back to Home
        </Button>
      </div>
      <div className="filter flex-wrap">
        <Button
        onClick={() => onRedirect('families/create')}
        variant='primary'
        intc-id={'btn-navigate-FamilyCreate'}>
          Create New Family
        </Button>
        <SearchBox
          intc-id="searchFamilies"
          value={filterText}
          onChange={setFilter}
          placeholder="Search families..."
          aria-label="Type to search families.."
        />
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} loading={loading} emptyGrid={emptyGrid} />
      </div>
    </>
  )
}

export default ProductFamiliesView
