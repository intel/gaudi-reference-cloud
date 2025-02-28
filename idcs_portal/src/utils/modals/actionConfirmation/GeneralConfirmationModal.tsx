// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import { type ActionConfirmationProps } from './ActionConfirmation'
import SpinnerIcon from '../../spinner/SpinnerIcon'

export interface GeneralConfirmationModalProps extends ActionConfirmationProps {
  loading: boolean
  toggleLoading: (show: boolean) => void
}

const GeneralConfirmationModal: React.FC<GeneralConfirmationModalProps> = (props): JSX.Element => {
  const actionModalContent = props.actionModalContent
  const label = actionModalContent.label
  const question = actionModalContent.question
  const feedback = actionModalContent.feedback
  const buttonLabel = actionModalContent.buttonLabel
  const buttonVariant = actionModalContent.buttonVariant
  const disableAction = actionModalContent.disableAction || false

  const getID = (label: string, type: string): string => {
    const id = `btn-confirm-${label}-${type}`
    return id.replace(' ', '')
  }

  return (
    <Modal
      show={props.showModalActionConfirmation}
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
        <div className="section">
          {question}
          <br />
          {feedback}
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
          Cancel
        </Button>
        <Button
          intc-id={getID(label, buttonLabel)}
          data-wap_ref={getID(label, buttonLabel)}
          aria-label={buttonLabel}
          variant={buttonVariant || 'primary'}
          onClick={() => {
            props.toggleLoading(true)
            props.onClickModalConfirmation(true)
          }}
          disabled={disableAction || props.loading}
        >
          {props.loading && <SpinnerIcon />}
          {buttonLabel}
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default GeneralConfirmationModal
