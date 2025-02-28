// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import GridPagination from '../../utils/gridPagination/gridPagination'
import { NavLink } from 'react-router-dom'
import SearchBox from '../../utils/searchBox/SearchBox'
import ActionConfirmation from '../../utils/modals/actionConfirmation/ActionConfirmation'

const KeyPairsView = (props) => {
  // props variables
  const columns = props.columns
  const myPublicKeys = props.myPublicKeys
  const importPagePath = props.importPagePath
  const filterText = props.filterText
  const emptyGrid = props.emptyGrid
  const loading = props.loading
  const showActionModal = props.showActionModal
  const actionOnModal = props.actionOnModal

  // props functions
  const actionModalContent = props.actionModalContent
  const setFilter = props.setFilter

  return (
    <>
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={actionOnModal}
        showModalActionConfirmation={showActionModal}
      />
      <div className={`filter ${!myPublicKeys || myPublicKeys.length === 0 ? 'd-none' : ''}`}>
        <NavLink
          to={importPagePath}
          className="btn btn-primary"
          intc-id="btn-sshview-UploadKey"
          data-wap_ref="btn-sshview-UploadKey"
        >
          Upload key
        </NavLink>
        <div className="d-flex justify-content-end">
          <SearchBox
            intc-id="searchKeys"
            value={filterText}
            onChange={setFilter}
            placeholder="Search keys..."
            aria-label="Type to search key.."
          />
        </div>
      </div>
      <div className="section">
        <GridPagination
          intc-id="sshPublicKeysTable"
          data={myPublicKeys}
          columns={columns}
          emptyGrid={emptyGrid}
          loading={loading}
          fixedFirstColumn
        />
      </div>
    </>
  )
}

export default KeyPairsView
