// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React from 'react'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import { Button } from 'react-bootstrap'
import { Link } from 'react-router-dom'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'

const UserCredentialsView = (props: any): JSX.Element => {
  // ****
  // Varaibles
  // ****
  const credentials = props.credentials
  const loading = props.loading
  const columns = props.columns
  const emptyGrid = props.emptyGrid
  const actionModal = props.actionModal
  const showActionModal = props.showActionModal
  const onClickDeleteModal = props.onClickDeleteModal

  return (
    <div className="section" intc-id="Super-Compute-Title">
      <ActionConfirmation
        actionModalContent={actionModal}
        onClickModalConfirmation={onClickDeleteModal}
        showModalActionConfirmation={showActionModal}
      />
      {credentials.length > 0 ? (
        <div className="d-flex flex-row bd-highlight mb-3">
          <div className="bd-highlight">
            <Link to="/profile/credentials/launch">
              <Button>Generate client secret</Button>
            </Link>
          </div>
        </div>
      ) : null}
      <GridPagination data={credentials} columns={columns} loading={loading} emptyGrid={emptyGrid} />
    </div>
  )
}

export default UserCredentialsView
