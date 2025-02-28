// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Modal } from 'react-bootstrap'
import SearchBox from '../../searchBox/SearchBox'
import SelectedAccountCard from '../../selectedAccountCard/SelectedAccountCard'

interface SearchCloudAccountModalProps {
  showModal: boolean
  selectedCloudAccount: any
  cloudAccount: string
  cloudAccountError: string
  showLoader?: any
  setShowModal: (status: boolean) => void
  handleSearchInputChange: (e: any) => void
  onClickSearchButton: (e: any) => Promise<void>
}

const SearchCloudAccountModal: React.FC<SearchCloudAccountModalProps> = (props): JSX.Element => {
  // props variables
  const selectedCloudAccount = props.selectedCloudAccount
  const cloudAccount = props.cloudAccount
  const cloudAccountError = props.cloudAccountError
  const showLoader = props.showLoader

  // props function
  const setShowModal = props.setShowModal
  const handleSearchInputChange = props.handleSearchInputChange
  const onClickSearchButton: any = props.onClickSearchButton

  return (
    <Modal
      show={props.showModal}
      onHide={() => {
        setShowModal(false)
      }}
      backdrop="static"
    >
      <Modal.Header closeButton closeLabel="Close request extention modal">
        <Modal.Title>Search Cloud Account</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="section">
          <SearchBox
            intc-id="searchAccounts"
            value={cloudAccount || ''}
            onChange={handleSearchInputChange}
            onClickSearchButton={onClickSearchButton}
            placeholder="Enter cloud account email or ID"
            aria-label="Enter cloud account email or ID"
          />
          {cloudAccountError && (
            <div className="text-danger text-center w-100" intc-id="cloudAccountSearchError">
              {cloudAccountError}
            </div>
          )}
          {showLoader?.isShow && (
            <div className="text-center w-100" intc-id="cloudAccountSearchLoadingMessage">
              {showLoader?.message}
            </div>
          )}
          <SelectedAccountCard selectedCloudAccount={selectedCloudAccount} />
        </div>
      </Modal.Body>
    </Modal>
  )
}

export default SearchCloudAccountModal
