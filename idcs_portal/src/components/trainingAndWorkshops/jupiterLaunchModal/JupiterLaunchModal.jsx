// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Modal from 'react-bootstrap/Modal'
import Spinner from '../../../utils/spinner/Spinner'
import { ReactComponent as ExternalLink } from '../../../assets/images/ExternalLink.svg'
import Button from 'react-bootstrap/Button'

const JupiterLaunchModal = (props) => {
  const show = props.show
  const onClose = props.onClose
  const jupyterRes = props.jupyterRes
  const type = props.type ? props.type : 'notebook'

  const onCloseTag = () => {
    window.open(jupyterRes, '_blank', 'noopener,noreferrer')
    onClose()
  }

  return (
    <Modal
      aria-label="Launch Notebook Modal"
      show={show}
      onHide={(e) => onClose(e, false)}
      backdrop={true}
      keyboard={false}
      size={'mg'}
      centered
    >
      <Modal.Header closeButton>
        <Modal.Title>
          <h2 className="h5"> Launching {type}</h2>
        </Modal.Title>
      </Modal.Header>
      <Modal.Body intc-id="enroll-training">
        {!jupyterRes && <Spinner />}
        {jupyterRes && (
          <div className="section justify-content-center align-items-center">
            To ensure optimum security, you&apos;ll be asked to sign in before the notebook loads.
          </div>
        )}
      </Modal.Body>
      <Modal.Footer className="justify-content-center">
        <div className="d-flex justify-content-center">
          <Button
            intc-id="btn_open_training"
            data-wap_ref="btn_open_training"
            variant="primary"
            onClick={() => onCloseTag()}
            aria-label="Open Notebook"
            disabled={!jupyterRes}
          >
            Launch
            <ExternalLink alt="Launch" />
          </Button>
        </div>
      </Modal.Footer>
    </Modal>
  )
}

export default JupiterLaunchModal
