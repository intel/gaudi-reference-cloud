// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React from 'react'
import GridPagination from '../../../utility/gridPagination/gridPagination'
import { Button } from 'react-bootstrap'
import SearchBox from '../../../utility/searchBox/SearchBox'
import { NavLink } from 'react-router-dom'

const StorageDetailView = (props: any): JSX.Element => {
  // *****
  // Variables
  // *****
  const storageUsages = props.storageUsages
  const columns = props.columns
  const emptyGrid = props.emptyGrid
  const loading = props.loading
  const onCancel = props.onCancel
  const filterText = props.filterText
  const setFilter = props.setFilter
  const moduleName = props.moduleName
  const setBucketFilter = props.setBucketFilter
  const bucketFilterText = props.bucketFilterText
  const bucketUsages = props.bucketUsages
  const bucketColumns = props.bucketColumns
  const bucketEmptyGrid = props.bucketEmptyGrid
  const setDefaultQuotaFilter = props.setDefaultQuotaFilter
  const defaultQuotaFilterText = props.defaultQuotaFilterText
  const defaultQuotaEmptyGrid = props.defaultQuotaEmptyGrid
  const defaultQuotaColumns = props.defaultQuotaColumns
  const defaultQuotaItems = props.defaultQuotaItems

  let gridItems = storageUsages
  if (filterText !== '' && storageUsages) {
    const input = filterText.toLowerCase()

    gridItems = storageUsages.filter((item: any) => item.cloudAccountId.toString().startsWith(input))
  }

  let gridBucketItems = bucketUsages
  if (bucketFilterText !== '' && bucketUsages) {
    const input = bucketFilterText.toLowerCase()

    gridBucketItems = bucketUsages.filter((item: any) => item.cloudAccountId.toString().startsWith(input))
  }

  let gridDefaultItems = defaultQuotaItems
  if (defaultQuotaFilterText !== '' && defaultQuotaItems) {
    const input = defaultQuotaFilterText.toLowerCase()

    gridDefaultItems = defaultQuotaItems.filter((item: any) => item.accountType.toLowerCase().toString().startsWith(input))
  }

  return (
    <>
      <div className="section">
        <Button variant="link" className="p-s0" onClick={() => onCancel()}>
          ⟵ Back to Home
        </Button>
      </div>
      <div className="section">
        <div className="filter flex-wrap p-0">
          {
            moduleName === 'Usages'
              ? <h2 className='h4'>File System Usages</h2>
              : <NavLink to="/storagemanagement/managequota"
                className="btn btn-primary" intc-id="btn-ManageStorageQuota">
                Manage Storage Quota
              </NavLink>
          }
          <SearchBox
            intc-id="filterAccounts"
            value={filterText}
            onChange={setFilter}
            placeholder="Filter accounts..."
            aria-label="Type to filter cloud accounts..."
          />
        </div>
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} loading={loading} emptyGrid={emptyGrid} />
      </div>
      {
        moduleName === 'Usages'
          ? <>
            <div className='section'>
              <div className="filter flex-wrap p-0">
                <h2 className='h4'>Bucket Usages</h2>
                <SearchBox
                  intc-id="filterAccounts"
                  value={bucketFilterText}
                  onChange={setBucketFilter}
                  placeholder="Filter accounts..."
                  aria-label="Type to filter cloud accounts..."
                />
              </div>
            </div>
            <div className="section">
              <GridPagination data={gridBucketItems} columns={bucketColumns} loading={loading} emptyGrid={bucketEmptyGrid} />
            </div>
          </>
          : null
      }
      {
        moduleName === 'Quotas'
          ? <>
            <div className='section'>
              <div className="filter flex-wrap p-0">
                <h2 className='h4'>Default Quota</h2>
                <SearchBox
                  intc-id="filterAccountsTypes"
                  value={defaultQuotaFilterText}
                  onChange={setDefaultQuotaFilter}
                  placeholder="Filter account types..."
                  aria-label="Type to filter cloud account types..."
                />
              </div>
            </div>
            <div className="section">
              <GridPagination data={gridDefaultItems} columns={defaultQuotaColumns} loading={loading} emptyGrid={defaultQuotaEmptyGrid} />
            </div>
          </>
          : null
      }
    </>
  )
}

export default StorageDetailView
