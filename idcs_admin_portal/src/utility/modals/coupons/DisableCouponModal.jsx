import { Modal, Button } from 'react-bootstrap'

const DisableCouponModal = (props) => {
  const handleClose = () => props.triggerHideDisableCouponModal(false)
  const handleDisableCoupon = () => props.handleDisableCoupon(props.coupon)

  return (
    <Modal show={props.showModal} onHide={handleClose}>
      <Modal.Header closeButton>
        <Modal.Title>Disable coupon code</Modal.Title>
      </Modal.Header>
      <Modal.Body className="p-4 pt-2 pb-2">
        Are you sure you want to disable code &apos;{props.coupon}&apos;?
      </Modal.Body>
      <Modal.Footer>
      <Button variant="outline-primary" onClick={handleClose}>
          Cancel
        </Button>
        <Button variant="primary" onClick={handleDisableCoupon}>
          Disable
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default DisableCouponModal
