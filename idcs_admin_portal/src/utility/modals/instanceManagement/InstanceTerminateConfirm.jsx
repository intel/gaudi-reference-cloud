import { useState } from 'react'
import { Modal, Button } from 'react-bootstrap'
import { AiOutlineWarning } from 'react-icons/ai'

const InstanceTerminateConfirm = (props) => {
  const onInstanceTerminateSubmit = props.onInstanceTerminateSubmit
  const instancesModalData = props.instancesModalData
  const [inputInvalidClass, setInputInvalidClass] = useState(instancesModalData.isValidInput)
  const [isButtonDisabled, setIsButtonDisabled] = useState(true)

  const onChangeInputHandler = (e) => {
    const inputName = e.target.value
    setInputInvalidClass('')
    setIsButtonDisabled(false)
    if (instancesModalData.name !== inputName) {
      setIsButtonDisabled(true)
      setInputInvalidClass('is-invalid')
    }
  }

  return (
    <Modal
      show={instancesModalData?.isShow}
      onHide={instancesModalData?.onClose}
      size="lg"
      backdrop="static"
      keyboard={false}
      intc-id='terminateInstanceModal'
      >
      <Modal.Header closeButton className='bg-warning'>
        <Modal.Title className='h1' intc-id='terminateInstanceModalTitle'><AiOutlineWarning size={45} className='pb-1'/>&nbsp;&nbsp;{instancesModalData?.title}</Modal.Title>
      </Modal.Header>
      <Modal.Body className='mx-1 container'>
        <div className='row mb-3'>
          <strong className='h3'>{instancesModalData.subtitle}</strong>
        </div>
        <div className='row'>
          <div className='col-3 col-lg-3 fw-bold'>Cloud Account:</div>
          <div className='col-3 col-lg-3' intc-id='terminateInstanceAccount'>
            {instancesModalData?.cAccount}
          </div>
        </div>
        <div className='row mb-4'>
          <div className="col-3 col-lg-3 fw-bold">Instance{instancesModalData.type} Name:</div>
          <div className='col-3 col-lg-3' intc-id='terminateInstanceName'>
            {instancesModalData?.name}
          </div>
        </div>

        <div className='row mt-2'>
          <div className="col-12 mb-2">{instancesModalData.message}</div>
          <div className='col-12 col-lg-6 justify-content-center'>
            <input
              type="text"
              className={`form-control border-left-0 ${inputInvalidClass}`}
              intc-id="terminateInstanceNameInput"
              required=""
              placeholder={instancesModalData?.name}
              onChange={onChangeInputHandler}
            />
          </div>
        </div>
      </Modal.Body>
      <Modal.Footer className="justify-content-start">
        <Button variant="link" onClick={instancesModalData.onClose} intc-id='terminateInstancesModalCancelBtn'>
          Cancel
        </Button>
        <Button
          variant="primary"
          onClick={
            (e) => onInstanceTerminateSubmit(
              instancesModalData.name,
              instancesModalData.resourceId,
              instancesModalData.cAccountId
            )
          }
          disabled={isButtonDisabled}
          intc-id='instanceTerminateSubmitBtn'
        >
          Yes, I&apos;m sure
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default InstanceTerminateConfirm
