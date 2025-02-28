// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import GridPagination from '../../../utils/gridPagination/gridPagination'
import ActionConfirmation from '../../../utils/modals/actionConfirmation/ActionConfirmation'

const ClusterSecurityRules = (props: any): JSX.Element => {
  const securityRules = props.securityRules
  const columns = props.columns
  const loading = props.loading
  const emptyGrid = props.emptyGrid
  const deleteModal = props.deleteModal
  const onClickActionModal = props.onClickActionModal

  return (
    <>
      <ActionConfirmation
        actionModalContent={deleteModal}
        onClickModalConfirmation={onClickActionModal}
        showModalActionConfirmation={deleteModal.show}
      />
      <div className="section">
        <GridPagination data={securityRules} columns={columns} loading={loading} emptyGrid={emptyGrid} />
      </div>
    </>
  )
}

export default ClusterSecurityRules
