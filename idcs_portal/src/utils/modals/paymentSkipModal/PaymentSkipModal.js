// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import { Link } from 'react-router-dom'
import idcConfig from '../../../config/configurator'

const PaymentSkipModal = (props) => {
  // props
  const showSkipModal = props.showSkipModal
  const skipAction = props.skipAction

  return (
    <Modal
      show={showSkipModal}
      backdrop="static"
      size="lg"
      aria-labelledby="contained-modal-title-vcenter"
      onHide={() => skipAction(false)}
      centered
      aria-label="Payment skip modal"
    >
      <Modal.Header>
        <Modal.Title>Skip Payment Method?</Modal.Title>
      </Modal.Header>
      <Modal.Body className="p-3">
        Are you sure you want to skip the payment method? <br />
        {`Adding a payment method unlocks all ${idcConfig.REACT_APP_CONSOLE_LONG_NAME} features and products.`}
        <br />
      </Modal.Body>
      <Modal.Footer className="justify-content-sm-start p-3">
        <Link to="/">
          <Button variant="danger" className="btn" intc-id="btn-premium-skip" data-wap_ref="btn-premium-skip">
            Skip
          </Button>
        </Link>
        <Button
          onClick={() => skipAction(false)}
          variant="outline-primary"
          className="btn"
          intc-id="btnModalGotoPayment"
        >
          Add payment method
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default PaymentSkipModal
