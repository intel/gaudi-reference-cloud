// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import { NavLink } from 'react-router-dom'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'
import idcConfig from '../../../config/configurator'
import SearchBox from '../../../utils/searchBox/SearchBox'
import { ReactComponent as ExternalLink } from '../../../assets/images/ExternalLink.svg'

const AccountRoles = (props: any): JSX.Element => {
  // ****
  // Variables
  // ****
  const userRoles = props.userRoles
  const loading = props.loading
  const columns = props.columns
  const emptyGrid = props.emptyGrid
  const launchPagePath = props.launchPagePath
  const actionModalContent = props.actionModalContent
  const showActionModal = props.showActionModal
  const filterText = props.filterText
  const actionOnModal = props.actionOnModal
  const setFilter = props.setFilter
  const isOwnCloudAccount = props.isOwnCloudAccount

  // variables
  let gridItems = []

  if (filterText !== '') {
    const input = filterText.toLowerCase()

    gridItems = userRoles.filter((item: any) =>
      item.roles.value ? item.roles.value.toLowerCase().includes(input) : item.roles.toLowerCase().includes(input)
    )
  } else {
    gridItems = userRoles
  }

  return (
    <>
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={actionOnModal}
        showModalActionConfirmation={showActionModal}
      />
      <div className="section">
        <h2>My Roles</h2>
        {/* TODO: Replace the official link for the user roles guide when available */}
        <span className="d-flex flex-row align-items-center">
          The following are the roles for this user.&nbsp;
          <a href={idcConfig.REACT_APP_MULTIUSER_GUIDE} target="_blank" rel="noreferrer" className="link">
            Learn about user roles
            <ExternalLink />
          </a>
        </span>
      </div>
      <div className={`filter ${!userRoles || userRoles.length === 0 ? 'd-none' : ''}`}>
        {isOwnCloudAccount && (
          <NavLink
            className="btn btn-primary"
            intc-id="btn-navigate-create-new-role"
            data-wap_ref="btn-navigate-create-new-role"
            to={launchPagePath}
          >
            Create new role
          </NavLink>
        )}
        <div className={`${!isOwnCloudAccount ? 'd-flex justify-content-end w-100' : ''}`}>
          <SearchBox
            intc-id="searchRoles"
            value={filterText}
            onChange={setFilter}
            placeholder="Search roles..."
            aria-label="Type to search role.."
          />
        </div>
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={columns} loading={loading} emptyGrid={emptyGrid} />
      </div>
    </>
  )
}

export default AccountRoles
