import React from 'react'
import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'

const ActionConfirmation = (props) => {
  const actionModalContent = props.actionModalContent
  const label = actionModalContent.label
  const question = actionModalContent.question
  const feedback = actionModalContent.feedback
  const buttonLabel = actionModalContent.buttonLabel

  return (

        <Modal
            show={props.showModalActionConfirmation}
            onHide={() => props.onClickModalConfirmation(false)}
            backdrop="static"
            keyboard={false}
            size="md"
        >
            <Modal.Header closeButton>
                <Modal.Title className="text-break">
                    {label}
                </Modal.Title>
            </Modal.Header>

            <Modal.Body>
                <div className="text-left small">
                        {question}<br/>
                        {feedback}
                </div>
            </Modal.Body>
            <Modal.Footer>
                <Button
                    variant="outline-primary"
                    onClick={() => props.onClickModalConfirmation(false)}
                >
                    Cancel
                </Button>
                <Button
                    variant='danger'
                    onClick={() => props.onClickModalConfirmation(true)}
                >
                    {buttonLabel}
                </Button>
            </Modal.Footer>
        </Modal>

  )
}

export default ActionConfirmation
