// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React from 'react'
import { Button } from 'react-bootstrap'
import { NavLink } from 'react-router-dom'
import GridPagination from '../../../utility/gridPagination/gridPagination'
import DeleteModal from '../../../utility/modals/deleteModal/DeleteModal'
import SearchBox from '../../../utility/searchBox/SearchBox'

const QuotaManagementServiceQuotas = (props: any): JSX.Element => {
  const onCancel = props.onCancel
  const emptyGrid = props.emptyGrid
  const loading = props.loading
  const serviceQuotaColumns = props.serviceQuotaColumns
  const serviceQuotaItems = props.serviceQuotaItems
  const serviceDetail = props.serviceDetail
  const serviceName = serviceDetail.serviceName
  const serviceId = serviceDetail.serviceId
  const deleteModal = props.deleteModal
  const onCloseDeleteModal = props.onCloseDeleteModal
  const filterText = props.filterText
  const setFilter = props.setFilter

  function getFilteredData(): any[] {
    let filteredData = []

    if (filterText !== '') {
      for (const index in serviceQuotaItems) {
        const quota = { ...serviceQuotaItems[index] }
        if (getFilterCheck(filterText, quota)) {
          filteredData.push(quota)
        }
      }
    } else {
      filteredData = [...serviceQuotaItems]
    }

    return filteredData
  }

  const gridItems = getFilteredData()

  function getFilterCheck(filterValue: string, data: any): boolean {
    filterValue = filterValue.toLowerCase()

    return data?.resourceType?.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data?.reason?.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data?.quotaUnit?.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data?.maxLimit?.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data?.scopeType?.toString().toLowerCase().indexOf(filterValue) > -1 ||
    data?.scopeValue?.toString().toLowerCase().indexOf(filterValue) > -1
  }

  return <>
    <DeleteModal modalContent={deleteModal} onClickModalConfirmation={onCloseDeleteModal} />
    <div className="section">
      <Button variant="link" className="p-s0" onClick={() => onCancel()}>
        ⟵ Back to Services
      </Button>
    </div>
    <div className="section">
      <h2 className='h4'>Service Quotas for {serviceName}</h2>
    </div>
    <div className="section">
      <div className="filter flex-wrap p-0">
        <NavLink to={`/quotamanagement/services/d/${serviceId}/quotas/add`}
          className="btn btn-primary" intc-id="btn-ManageStorageQuota">
          Create new Quota
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
      <GridPagination data={gridItems} columns={serviceQuotaColumns} loading={loading} emptyGrid={emptyGrid} />
    </div>
  </>
}

export default QuotaManagementServiceQuotas
