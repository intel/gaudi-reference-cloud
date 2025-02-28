// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useState } from 'react'
import Modal from 'react-bootstrap/Modal'
import ComputeInstanceCompare from '../../../components/compute/computeInstanceCompare/ComputeInstanceCompare'
import Button from 'react-bootstrap/Button'

const ModalInstanceCompare = (props) => {
  const [instance, setInstance] = useState('')

  const callAfterSuccess = () => {
    props.afterInstanceSelected(instance)
    props.closeInstanceCompareModal()
  }

  function instanceSelected(instance) {
    setInstance(instance)
  }

  return (
    <Modal
      show={props.showModalActionConfirmation}
      onHide={() => props.closeInstanceCompareModal()}
      backdrop="static"
      keyboard={false}
      size="xl mw-80"
      intc-id="compareInstanceModal"
      aria-label="Compare instance modal"
    >
      <Modal.Header closeButton>
        <Modal.Title>Compare instance types</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <ComputeInstanceCompare products={props.products} afterInstanceSelected={instanceSelected} />
      </Modal.Body>
      <Modal.Footer className="justify-content-sm-start p-3">
        {props.products.length > 0 && (
          <Button variant="primary" onClick={callAfterSuccess} intc-id="btnCompareInstanceSave">
            Select
          </Button>
        )}
        <Button variant="outline-primary" onClick={() => props.closeInstanceCompareModal()}>
          Cancel
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default ModalInstanceCompare
