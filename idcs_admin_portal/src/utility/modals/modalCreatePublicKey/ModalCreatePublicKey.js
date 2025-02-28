import React from 'react'
import Modal from 'react-bootstrap/Modal'
import KeyPairsForm from '../../../components/keypairs/KeyPairsForm'

const ModalCreatePublicKey = (props) => {
  const payload = {

    metadata: {
      name: null
    },
    spec: {
      sshPublicKey: null
    }
  }

  const callAfterSuccess = (name) => {
    props.closeCreatePublicKeyModal()
    props.afterPubliKeyCreate(name)
  }

  return (
        <Modal
            show={props.showModalActionConfirmation}
            onHide={() => props.closeCreatePublicKeyModal()}
            backdrop="static"
            keyboard={false}
            size="xl"
            aria-label="Upload public key modal"
        >

            <Modal.Header closeButton>
                <Modal.Title intc-id='upload-key'>
                    Upload a public key
                </Modal.Title>
            </Modal.Header>
            <Modal.Body>
                <KeyPairsForm
                    payload={payload}
                    isKeyPairFormOpen={props.closeCreatePublicKeyModal}
                    callAfterSuccess={callAfterSuccess}
                    isModal={props.isModal}
                    handleClose={props.closeCreatePublicKeyModal}
                />
                <br/>
            </Modal.Body>
        </Modal>
  )
}

export default ModalCreatePublicKey
