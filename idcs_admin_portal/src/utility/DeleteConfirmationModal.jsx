import React from 'react'
import { Modal, Button } from 'react-bootstrap'

const DeleteConfirmationModal = ({
  showModal,
  hideModal,
  confirmModal,
  message
}) => {
  return (
    <Modal show={showModal} onHide={hideModal} intc-id="deleteConfirmModal">
      <Modal.Header closeButton>
        <Modal.Title>Delete Public key?</Modal.Title>
      </Modal.Header>
      <Modal.Body className="p-4 pt-2 pb-2">{message}</Modal.Body>
      <Modal.Footer>
        <Button variant="outline-primary" onClick={hideModal} intc-id="btnDeleteConfirmModalCancel">
          Cancel
        </Button>
        <Button variant="danger" onClick={() => confirmModal()} intc-id="btnDeleteConfirmModalDelete">
          Delete
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default DeleteConfirmationModal
