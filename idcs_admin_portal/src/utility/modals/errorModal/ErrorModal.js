import React from 'react'
import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import { Link } from 'react-router-dom'
import { MdDangerous } from 'react-icons/md'
import idcConfig from '../../../config/configurator'
import useUserStore from '../../../store/userStore/UserStore'

const ErrorModal = (props) => {
  const { isPremiumUser, isEnterpriseUser, isIntelUser } = useUserStore(
    (state) => state
  )

  const canContactSupport =
    isPremiumUser || isEnterpriseUser || isIntelUser

  const defaultRetryMessage = canContactSupport
    ? 'Please try again or contact support if the issue continues.'
    : 'Please try again or check the community for hints.'

  // props
  const showModal = props.showModal
  const message = props.message
  const titleMessage = props.titleMessage || 'Could not launch your instance'
  const description =
    props.description ||
    'There was an error while processing your compute instance.'
  const retryMessage = props.retryMessage || defaultRetryMessage
  const hideRetryMessage = props.hideRetryMessage
  const onClickCloseErrorModal = props.onClickCloseErrorModal
  const actionButton = props.actionButton

  const supportLink = () => {
    if (canContactSupport) {
      return (
        <Link
          to={{ pathname: idcConfig.REACT_APP_SUPPORT_PAGE }}
          target="_blank"
        >
          <Button intc-id="error-modal-contact-support-link" aria-label="Go to Contact Support page" variant="link" className="text-decoration-none">
            Contact Support
          </Button>
        </Link>
      )
    } else {
      return (
        <Link to={{ pathname: idcConfig.REACT_APP_COMMUNITY }} target="_blank">
          <Button intc-id="error-modal-community-link" aria-label="Go to Contact Support page" variant="link" className="text-decoration-none">
            Community
          </Button>
        </Link>
      )
    }
  }

  return (
    <Modal
      show={showModal}
      backdrop="static"
      size="lg"
      aria-labelledby="contained-modal-title-vcenter"
      centered
    >
      <Modal.Body>
        <br />
        <div className="text-center">
          <MdDangerous color="red" size="3em" />
          <h5>{titleMessage}</h5>
          <p>
            {description}
            {message
              ? (
              <>
                <br />{message}
              </>
                )
              : null}
            <br />
            {hideRetryMessage ? null : retryMessage}
          </p>
        </div>
      </Modal.Body>
      <Modal.Footer>
        <div className="mx-auto">
          {supportLink()}
          <div className="space"></div>
          <div className="space"></div>
          <Button
            onClick={onClickCloseErrorModal}
            className={actionButton?.class || ''}
            variant={actionButton?.variant || 'primary'}
          >
            {actionButton?.label || 'Go back'}
          </Button>
        </div>
      </Modal.Footer>
    </Modal>
  )
}

export default ErrorModal
