import Modal from 'react-bootstrap/Modal'

const OnSubmitModal = (props) => {
  const showModal = props.showModal
  const message = Object.prototype.hasOwnProperty.call(props, 'message') ? props.message : 'Working on your request'
  return (
    <Modal show={showModal} backdrop="static" keyboard={false} intc-id="OnSubmitModal">
      <Modal.Header></Modal.Header>
      <Modal.Body>
        <div className="modal-body row justify-content-center">
          <div className="col-12 row">
            <div className="spinner-border text-primary center"></div>
            <span className="text-center col-12 pl-0 pt-2">
              <strong>{message}</strong>
            </span>
          </div>
        </div>
      </Modal.Body>
      <Modal.Footer></Modal.Footer>
    </Modal>
  )
}

export default OnSubmitModal
