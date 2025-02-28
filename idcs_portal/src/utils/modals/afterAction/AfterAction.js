// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'

const AfterAction = (props) => {
  const modalContent = props.modalContent
  const label = modalContent.label
  const feedback = modalContent.feedback
  const buttonLabel = modalContent.buttonLabel
  const buttonVariant = modalContent.buttonVariant

  return (
    <Modal
      show={props.showModal}
      onHide={() => props.onClickModal(false)}
      backdrop="static"
      keyboard={false}
      size="md"
      aria-label="After action modal"
    >
      <Modal.Header closeButton>
        <Modal.Title>{label}</Modal.Title>
      </Modal.Header>

      <Modal.Body>
        <div className="text-left">{feedback}</div>
      </Modal.Body>
      <Modal.Footer>
        <Button
          intc-id={`btn-confirm-${label}-${buttonLabel}`}
          data-wap_ref={`btn-confirm-${label}-${buttonLabel}`}
          variant={buttonVariant || 'primary'}
          onClick={() => props.onClickModal(false)}
        >
          {buttonLabel}
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default AfterAction
