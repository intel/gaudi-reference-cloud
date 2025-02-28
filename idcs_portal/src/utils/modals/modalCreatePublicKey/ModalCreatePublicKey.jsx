// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import KeyPairsForm from '../../../components/keypairs/KeyPairsForm'
import ToastContainer from '../../toast/ToastContainer'
import useToastStore from '../../../store/toastStore/ToastStore'

const ModalCreatePublicKey = (props) => {
  const showSuccess = useToastStore((state) => state.showSuccess)

  // local variable

  const callAfterSuccess = (name) => {
    showSuccess('Key added successfully.')
    props.closeCreatePublicKeyModal()
    props.afterPubliKeyCreate(name)
  }

  return (
    <Modal
      show={props.showModalActionConfirmation}
      onHide={() => props.closeCreatePublicKeyModal()}
      backdrop="static"
      keyboard={false}
      size="lg"
      aria-label="Upload public key modal"
    >
      <ToastContainer />
      <Modal.Header closeButton>
        <Modal.Title intc-id="upload-key">Upload a public key</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <KeyPairsForm
          isKeyPairFormOpen={props.closeCreatePublicKeyModal}
          callAfterSuccess={callAfterSuccess}
          isModal={props.isModal}
          handleClose={props.closeCreatePublicKeyModal}
        />
      </Modal.Body>
    </Modal>
  )
}

export default ModalCreatePublicKey
