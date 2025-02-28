import { React } from 'react'
import { Modal, Button } from 'react-bootstrap'
import { Link } from 'react-router-dom'
import { BsCopy } from 'react-icons/bs'
import { useCopy } from '../../../hooks/useCopy'
import ToastContainer from '../../toast/ToastContainer'
import idcConfig from '../../../config/configurator'

const OnCreateCoupon = (props) => {
  const modal = props.modal
  const show = modal.show
  const onClose = modal.onClose
  const redirecTo = modal.redirecTo
  const body = modal.message
  const { copyToClipboard } = useCopy()

  return (
    <Modal show={show} intc-id='OnCreateCouponModal'>
      <ToastContainer />
      <Modal.Header>
        <Modal.Title>Your Coupon Code is:</Modal.Title>
      </Modal.Header>
      <Modal.Body className="d-flex flex-row justify-content-between align-items-center">
            <h3 intc-id='CouponValue' className='h6'>{body}</h3>
          <Button variant="outline-primary" onClick={() => copyToClipboard(body)}>
            <BsCopy />
            Copy
          </Button>
          <Button className='me-1' variant="outline-primary" onClick={() => copyToClipboard(`${idcConfig.REACT_APP_GUI_DOMAIN}/billing/credits/managecouponcode?coupon=${body}`)}>
            <BsCopy />
            Copy shareable link
          </Button>
      </Modal.Body>
      <Modal.Footer>
      <Button variant="outline-primary" intc-id='CouponModalCreateMoreBtn' onClick={onClose}>
          Create more
        </Button>
        <Link to={redirecTo}>
          <Button variant="primary" intc-id='CouponModalOkBtn'>Ok</Button>
        </Link>
      </Modal.Footer>
    </Modal>
  )
}

export default OnCreateCoupon
