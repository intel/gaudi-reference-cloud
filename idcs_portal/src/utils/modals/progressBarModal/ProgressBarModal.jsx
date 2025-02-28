// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import ProgressBar from 'react-bootstrap/ProgressBar'

const ProgressBarModal = (props) => {
  const show = props.show
  const onClose = props.onClose
  const size = props.size
  const backdrop = props.backdrop
  const title = props.title
  const progressValue = props.progressValue
  const steps = [...props.steps]
  const actions = [...props.actions]
  return (
    <Modal
      show={show}
      onHide={() => onClose(false)}
      backdrop={backdrop}
      keyboard={false}
      size={size}
      centered
      aria-label="Progress bar modal"
    >
      <Modal.Header className="modal-header-training">
        <Modal.Title>{title}</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="d-flex flex-column bd-highlight">
          <div className="p-1 bd-highlight">
            <ProgressBar now={progressValue} label={`${progressValue}%`} />
          </div>
          <div className="p-0 bd-highlight ">
            <div className="d-flex justify-content-center">
              <div className="d-flex flex-column bd-highlight">
                {steps.map((step, index) => (
                  <div className="p-0 bd-highlight" key={index}>
                    <span>{step.text}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </Modal.Body>
      <Modal.Footer className="modal-footer-training">
        <div className="d-flex flex-column bd-highlight">
          <div className="p-1 bd-highlight">
            {actions.map((action, index) => (
              <button
                key={index}
                intc-id={`btn-launch-training-${index}`}
                className="btn btn-primary"
                disabled={action.disabled}
              >
                {action.startIcon ? action.startIcon : ''}
                {action.label}
                {action.endIcon ? action.endIcon : ''}
              </button>
            ))}
          </div>
        </div>
      </Modal.Footer>
    </Modal>
  )
}

export default ProgressBarModal
