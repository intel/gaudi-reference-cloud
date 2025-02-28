// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Button, ButtonGroup } from 'react-bootstrap'
import Modal from 'react-bootstrap/Modal'
import { type CostEstimateProps } from './CostEstimate.types'

interface CostEstimateModalProps extends CostEstimateProps {
  showModal: boolean
  onHide: () => void
}

const CostEstimateModal: React.FC<CostEstimateModalProps> = (props): JSX.Element => {
  const showModal = props.showModal
  const title = props.title
  const description = props.description
  const costArray = props.costArray
  const onHide = props.onHide

  return (
    <Modal
      show={showModal}
      onHide={() => {
        onHide()
      }}
      backdrop="static"
      keyboard={false}
      centered
      aria-label="Cost estimate modal"
    >
      <Modal.Header closeButton closeLabel="Close cost estimate modal">
        <Modal.Title>{title}</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="d-flex flex-column gap-s4">
          {description}
          {costArray.map((item, index) => (
            <div key={index} className="d-flex justify-content-between w-100">
              <span className="fw-semibold">{item.label}</span>
              <span>{item.value}</span>
            </div>
          ))}
        </div>
      </Modal.Body>
      <Modal.Footer>
        <ButtonGroup>
          <Button
            intc-id="btn-close-cost-estimate-modal"
            data-wap_ref="btn-close-cost-estimate-modal"
            aria-label="Close cost estimate modal"
            variant="primary"
            onClick={() => {
              onHide()
            }}
          >
            Ok
          </Button>
        </ButtonGroup>
      </Modal.Footer>
    </Modal>
  )
}

export default CostEstimateModal
