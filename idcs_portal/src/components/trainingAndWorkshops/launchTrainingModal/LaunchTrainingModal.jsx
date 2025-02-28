// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Modal from 'react-bootstrap/Modal'
import CodeLine from '../../../utils/CodeLine'
import Button from 'react-bootstrap/Button'
import CustomAlerts from '../../../utils/customAlerts/CustomAlerts'
import Spinner from '../../../utils/spinner/Spinner'

const LaunchTrainingModal = (props) => {
  const show = props.show
  const onClose = props.onClose
  const loading = props.loading
  const serviceResponse = props.serviceResponse

  let title = null
  let body = null
  let footer = null
  if (loading) {
    body = (
      <div className="d-flex justify-content-center">
        <Spinner />
        <div className="p-2 my-4">
          <h5>Launching notebook</h5>
        </div>
      </div>
    )
  } else {
    title = 'How to connect to learning node'
    body = (
      <div className="d-flex flex-column">
        <div className="bd-highlight mt-2">
          <CustomAlerts
            showAlert={true}
            alertType="warning"
            message={`Your access will be valid until ${serviceResponse.expiryDate}`}
            onCloseAlert={null}
            showIcon={true}
          />
        </div>
        <div className="bd-highlight mt-2">
          <h6>To access your instance with an SSH client:</h6>
          <ol className="ps-3">
            <li>Open an SSH client.</li>
            <li>
              Connect to the batch service using its public DNS:
              <CodeLine codeline={serviceResponse.sshLoginInfo} />
            </li>
          </ol>
          <span>If you need any assistance connecting to your instance, please see our documentation.</span>
        </div>
      </div>
    )
    footer = (
      <div className="d-flex justify-content-end">
        <Button variant="outline-primary" onClick={() => props.onClose(false)}>
          Close
        </Button>
      </div>
    )
  }

  return (
    <Modal
      show={show}
      onHide={() => onClose(false)}
      backdrop={true}
      keyboard={false}
      size={'lg'}
      centered
      aria-label="Launch Learning Modal"
    >
      <Modal.Header closeButton>
        <Modal.Title>{title}</Modal.Title>
      </Modal.Header>
      <Modal.Body intc-id="enroll-training">{body}</Modal.Body>
      <Modal.Footer>{footer}</Modal.Footer>
    </Modal>
  )
}

export default LaunchTrainingModal
