// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'

export interface SpinnerBackdropProps {
  showSpinner: boolean
}

const SpinnerBackdrop: React.FC<SpinnerBackdropProps> = ({ showSpinner }): JSX.Element => {
  return (
    <Modal
      show={showSpinner}
      backdrop="static"
      size="lg"
      aria-labelledby="contained-modal-title-vcenter"
      centered
      contentClassName="border-0 shadow-none background-none"
      aria-label="Loading backdrop modal"
    >
      <Modal.Body className="p-3">
        <div className="p-3 col-12 row text-center">
          <div className="spinner-border main-spinner text-primary center"></div>
        </div>
      </Modal.Body>
    </Modal>
  )
}

export default SpinnerBackdrop
