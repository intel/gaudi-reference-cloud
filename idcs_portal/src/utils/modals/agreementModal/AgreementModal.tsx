// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'

export interface AgreementModalProps {
  modalContent: any
  showAgreementModal: boolean
  onClickModalConfirmation: (status: boolean) => void
  onDismiss: () => void
}

const AgreementModal: React.FC<AgreementModalProps> = (props): JSX.Element => {
  const modalContent = props.modalContent
  const label = modalContent.label
  const agreementContent = modalContent.agreementContent
  const acceptButtonLabel = modalContent.acceptButtonLabel
  const rejectButtonLabel = modalContent.rejectButtonLabel
  const acceptButtonVariant = modalContent.acceptButtonVariant
  const disableAction = modalContent.disableAction || false

  const getID = (label: string, type: string): string => {
    const id = `btn-confirm-${label}-${type}`
    return id.replace(' ', '')
  }

  return (
    <Modal
      show={props.showAgreementModal}
      onHide={() => {
        props.onDismiss()
      }}
      backdrop="static"
      keyboard={false}
      intc-id="generalConfirmModal"
      aria-label="Service agreement modal"
    >
      <Modal.Header closeButton>
        <Modal.Title>{label}</Modal.Title>
      </Modal.Header>

      <Modal.Body>
        <div className="section">{agreementContent}</div>
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
          {rejectButtonLabel}
        </Button>
        <Button
          intc-id={getID(label, acceptButtonLabel)}
          data-wap_ref={getID(label, acceptButtonLabel)}
          aria-label={acceptButtonLabel}
          variant={acceptButtonVariant || 'primary'}
          onClick={() => {
            props.onClickModalConfirmation(true)
          }}
          disabled={disableAction}
        >
          {acceptButtonLabel}
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default AgreementModal
