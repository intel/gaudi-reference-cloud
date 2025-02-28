// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import SpinnerIcon from '../../spinner/SpinnerIcon'
import CustomAlerts from '../../customAlerts/CustomAlerts'

export interface KeyInUseModalProps {
  modalContent: any
  showModal: boolean
  onClickModalConfirmation: (status: boolean) => void
  loading: boolean
}

const KeyInUseModal: React.FC<KeyInUseModalProps> = (props): JSX.Element => {
  const modalContent = props.modalContent
  const label = modalContent.label
  const feedback = modalContent.feedback
  const body = modalContent.body
  const buttonLabel = modalContent.buttonLabel
  const buttonVariant = modalContent.buttonVariant
  const hasPendingRequest = modalContent.hasPendingRequest
  const disableAction = modalContent.disableAction || false

  const getID = (label: string, type: string): string => {
    const id = `btn-confirm-${label}-${type}`
    return id.replace(' ', '')
  }

  return (
    <Modal
      show={props.showModal}
      onHide={() => {
        props.onClickModalConfirmation(false)
      }}
      backdrop="static"
      keyboard={false}
      intc-id="generalConfirmModal"
      aria-label="Confirmation modal"
    >
      <Modal.Header closeButton>
        <Modal.Title>{label}</Modal.Title>
      </Modal.Header>

      <Modal.Body>
        <CustomAlerts showAlert={true} alertType="warning" message="Cannot delete this key" showIcon={true} />
        <div className="section">
          {feedback}
          <br />
          {body}
        </div>
      </Modal.Body>
      <Modal.Footer>
        <Button
          intc-id={getID(label, 'cancel')}
          data-wap_ref={getID(label, 'cancel')}
          aria-label="Cancel"
          variant="outline-primary"
          onClick={() => {
            props.onClickModalConfirmation(false)
          }}
        >
          {hasPendingRequest ? 'Close' : 'Cancel'}
        </Button>
        {!hasPendingRequest && (
          <Button
            intc-id={getID(label, buttonLabel)}
            data-wap_ref={getID(label, buttonLabel)}
            aria-label={buttonLabel}
            variant={buttonVariant || 'primary'}
            onClick={() => {
              props.onClickModalConfirmation(true)
            }}
            disabled={disableAction || props.loading}
          >
            {props.loading && <SpinnerIcon />}
            {buttonLabel}
          </Button>
        )}
      </Modal.Footer>
    </Modal>
  )
}

export default KeyInUseModal
