import { Modal, Button } from 'react-bootstrap'

const OnConfirmModal = (props) => {
  const confirmModalData = props.confirmModalData
  const onSubmit = props.onSubmit
  const data = confirmModalData.data

  const buildData = (data) => {
    return data.map((item, i) => {
      return (
        <div className="row" key={i}>
          <div className="col-12 col-lg-4 border fw-bold">{item.col}</div>
          <div className="col-12 col-lg-8 border">{item.value}</div>
        </div>
      )
    })
  }

  return (
    <Modal show={confirmModalData.isShow} onHide={confirmModalData.onClose} size="lg">
      <Modal.Header closeButton>
        <Modal.Title>{confirmModalData.title}</Modal.Title>
      </Modal.Header>
      <Modal.Body className="mx-1">{data.length > 0 ? buildData(data) : null}</Modal.Body>
      <Modal.Footer className="justify-content-start">
        <Button variant="primary" onClick={onSubmit}>
          Confirm
        </Button>
        <Button variant="link" onClick={confirmModalData.onClose ? confirmModalData.onClose : onSubmit}>
          Cancel
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default OnConfirmModal
