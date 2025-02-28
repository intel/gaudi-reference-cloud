import React from 'react'
import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'

const DeleteModal = (props: any): JSX.Element => {
  const modalContent = props.modalContent
  const show = modalContent.show
  const title = modalContent.title
  const question = modalContent.question
  const feedback = modalContent.feedback
  const buttonLabel = modalContent.buttonLabel
  const onClickModalConfirmation = props.onClickModalConfirmation

  return (
    <Modal
      show={show}
      onHide={() => onClickModalConfirmation(false)}
      backdrop="static"
      keyboard={false}
      size="lg"
    >
      <Modal.Header closeButton>
        <Modal.Title className="text-break">
          {title}
        </Modal.Title>
      </Modal.Header>

      <Modal.Body>
        <div className="d-flex flex-column">
          {question}<br />
          {feedback}
        </div>
      </Modal.Body>
      <Modal.Footer>
        <Button
          variant="outline-primary"
          onClick={() => onClickModalConfirmation(false)}
        >
          Cancel
        </Button>
        <Button
          variant='danger'
          onClick={() => onClickModalConfirmation(true)}
        >
          {buttonLabel}
        </Button>
      </Modal.Footer>
    </Modal>
  )
}
export default DeleteModal
