// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Button } from 'react-bootstrap'
import { NavLink } from 'react-router-dom'
import SearchBox from '../../utility/searchBox/SearchBox'
import GridPagination from '../../utility/gridPagination/gridPagination'
import OnConfirmModal from '../../utility/modals/onConfirmModal/OnConfirmModal'
import OnSubmitModal from '../../utility/modals/onSubmitModal/OnSubmitModal'

const RegionManagement = (props: any): JSX.Element => {
  const regionColumns = props.regionColumns
  const filterText = props.filterText
  const emptyGrid = props.emptyGrid
  const setFilter = props.setFilter
  const onCancel = props.onCancel
  const regions = props.regions
  const loading = props.loading
  const confirmModalData = props.confirmModalData
  const showLoader = props.showLoader
  const defaultRegion = props.defaultRegion

  const onSubmit = props.onSubmit

  const getFilteredData = (): any[] => {
    if (!filterText || !regions) return regions

    const input = filterText.toLowerCase()
    return regions.filter((item: any) => {
      return Object.keys(item).some((key) => {
        const value = item[key]
        return value?.toString().toLowerCase().includes(input)
      })
    })
  }

  const gridItems = getFilteredData()

  return (
    <>
      <div className="section">
        <Button variant="link" className="p-s0" onClick={() => onCancel()}>
          ⟵ Back to Home
        </Button>
      </div>
      <OnSubmitModal showModal={showLoader.isShow} message={showLoader.message}></OnSubmitModal>
      <OnConfirmModal confirmModalData={confirmModalData} onSubmit={onSubmit}></OnConfirmModal>
      <div className="section">
        <h2 className="h4">Region Management</h2>
        <h3 className="h5">Current default region: {defaultRegion?.name ?? 'Not defined'}</h3>
      </div>
      <div className="section">
        <div className="filter flex-wrap p-0">
          <NavLink to="/regionmanagement/regions/create" className="btn btn-primary" intc-id="btn-RegionCreate">
            Create new region
          </NavLink>
          <SearchBox
            intc-id="filterAccounts"
            value={filterText}
            onChange={setFilter}
            placeholder="Filter regions..."
            aria-label="Type to filter cloud regions..."
          />
        </div>
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={regionColumns} loading={loading} emptyGrid={emptyGrid} />
      </div>
    </>
  )
}

export default RegionManagement
