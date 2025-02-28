// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Button } from 'react-bootstrap'
import { NavLink } from 'react-router-dom'
import SearchBox from '../../utility/searchBox/SearchBox'
import GridPagination from '../../utility/gridPagination/gridPagination'
import OnConfirmModal from '../../utility/modals/onConfirmModal/OnConfirmModal'
import OnSubmitModal from '../../utility/modals/onSubmitModal/OnSubmitModal'
import CustomInput from '../../utility/customInput/CustomInput'

const AccountRegion = (props: any): JSX.Element => {
  const regionColumns = props.regionColumns
  const filterText = props.filterText
  const emptyGrid = props.emptyGrid
  const setFilter = props.setFilter
  const onCancel = props.onCancel
  const data = props.data
  const loading = props.loading
  const confirmModalData = props.confirmModalData
  const showLoader = props.showLoader
  const regions = props.regions
  const selectedRegion = props.selectedRegion

  const onSubmit = props.onSubmit
  const setRegionFilter = props.setRegionFilter

  // local variables
  const customStyle = { minWidth: '18rem' }

  function getFilteredData(): any[] {
    let filteredData = []
    if (selectedRegion && filterText) {
      for (const index in data) {
        const item = { ...data[index] }
        if (getFilterCheck(filterText, item) && getFilterCheck(selectedRegion, item)) {
          filteredData.push(item)
        }
      }
    } else if (filterText !== '' || selectedRegion) {
      for (const index in data) {
        const item = { ...data[index] }
        if (getFilterCheck(filterText, item) || getFilterCheck(selectedRegion, item)) {
          filteredData.push(item)
        }
      }
    } else {
      filteredData = [...data]
    }

    return filteredData
  }

  function getFilterCheck(filterValue: any, item: any): boolean {
    if (!filterValue) return false
    const input = filterValue.toLowerCase()
    return Object.keys(item).some((key) => {
      const value = item[key]
      return value?.toString().toLowerCase().includes(input)
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
        <h2 className="h4">Region Whitelist Accounts</h2>
      </div>
      <div className="section">
        <div className="filter flex-wrap p-0">
          <NavLink to="/regionmanagement/whitelist/add" className="btn btn-primary" intc-id="btn-add-cloud-account">
            Add Cloud Account
          </NavLink>
          <div className="d-flex flex-sm-row flex-column gap-s6">
            <div className="d-flex flex-sm-row flex-column gap-s4 align-items-sm-center">
              <span>Region</span>
              <div style={customStyle}>
                <CustomInput
                  type="dropdown"
                  hiddenLabel={true}
                  label="Service Type"
                  placeholder="Service Type"
                  value={selectedRegion}
                  options={regions}
                  onChanged={setRegionFilter}
                />
              </div>
            </div>
            <SearchBox
              intc-id="searchCloudAccounts"
              value={filterText}
              onChange={setFilter}
              placeholder="Search accounts..."
              aria-label="Type to search cloud accounts.."
            />
          </div>
        </div>
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={regionColumns} loading={loading} emptyGrid={emptyGrid} />
      </div>
    </>
  )
}

export default AccountRegion
