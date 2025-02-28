import { Modal, Button } from 'react-bootstrap'

const SkuQuotaCreateConfirm = (props) => {
  const confirmModalData = props.confirmModalData
  const onSubmit = props.onSubmit

  return (
    <Modal
      show={confirmModalData.isShow}
      onHide={confirmModalData.onClose}
      size="lg">
      <Modal.Header closeButton>
        <Modal.Title>{confirmModalData.title}</Modal.Title>
      </Modal.Header>
      <Modal.Body className="mx-1">
        <div className="row">
          <div className="col-12 col-lg-4 border fw-bold">Product Family</div>
          <div className="col-12 col-lg-8 border">{confirmModalData.family}</div>
        </div>
        <div className="row">
          <div className="col-12 col-lg-4 border fw-bold">Instance Name</div>
          <div className="col-12 col-lg-8 border">
            {confirmModalData.instanceName}
          </div>
        </div>
        <div className="row">
          <div className="col-12 col-lg-4 border fw-bold">Instance Type</div>
          <div className="col-12 col-lg-8 border">
            {confirmModalData.instanceType}
          </div>
        </div>
        <div className="row">
          <div className="col-12 col-lg-4 border fw-bold">Cloud Account</div>
          <div className="col-12 col-lg-8 border">
            {confirmModalData.cloudAccount}
          </div>
        </div>
      </Modal.Body>
      <Modal.Footer className="justify-content-start">
        <Button variant="primary" onClick={onSubmit}>
          Confirm
        </Button>
        <Button variant="link" onClick={confirmModalData.onClose}>
          Cancel
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default SkuQuotaCreateConfirm
