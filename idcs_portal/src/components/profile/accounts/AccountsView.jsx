// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import GridPagination from '../../../utils/gridPagination/gridPagination'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import ToastContainer from '../../../utils/toast/ToastContainer'
import SearchBox from '../../../utils/searchBox/SearchBox'
import './AccountsView.scss'

const AccountsView = (props) => {
  const loading = props.loading
  const listCloudAccounts = props.listCloudAccounts
  const listOwnedAccounts = props.listOwnedAccounts
  const columns = props.columns
  const setFilter = props.setFilter
  const filterText = props.filterText
  const actionModalContent = props.actionModalContent
  const onClickModalConfirmation = props.onClickModalConfirmation
  const showModalActionConfirmation = props.showModalActionConfirmation

  let gridItems = []

  if (filterText !== '' && listCloudAccounts.length > 0) {
    const input = filterText.toLowerCase()
    gridItems = listCloudAccounts.filter((item) =>
      item.cloudAccountId.value
        ? item.cloudAccountId.value.toLowerCase().includes(input)
        : item.cloudAccountId.toLowerCase().includes(input)
    )

    if (gridItems.length === 0) {
      gridItems = listCloudAccounts.filter((item) => item.email.toLowerCase().includes(input))
    }
  } else {
    gridItems = listCloudAccounts
  }

  return (
    <>
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={onClickModalConfirmation}
        showModalActionConfirmation={showModalActionConfirmation}
      />
      <ToastContainer />
      <div className="accountSelection">
        <div className="section ">
          <h1 intc-id="selectAccount">Select a cloud account</h1>
        </div>
        <div className="section">
          <h2 className="h5">My account</h2>
          <GridPagination data={listOwnedAccounts} columns={columns} loading={loading} hidePaginationControl={true} />
        </div>
        <div className="section">
          <h2 className="h5">Membership accounts</h2>
          <div className="d-flex justify-content-end">
            <SearchBox
              intc-id="searchaccounts"
              value={filterText}
              onChange={setFilter}
              placeholder="Search accounts..."
              aria-label="Search accounts"
            />
          </div>
          <GridPagination data={gridItems} columns={columns} loading={loading} hidePaginationControl={true} />
        </div>
      </div>
    </>
  )
}

export default AccountsView
