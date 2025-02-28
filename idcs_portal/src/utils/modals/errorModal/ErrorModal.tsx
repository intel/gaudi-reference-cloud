// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import { BsArrowRightShort, BsExclamationTriangle } from 'react-icons/bs'
import idcConfig from '../../../config/configurator'
import useUserStore from '../../../store/userStore/UserStore'
import { specificErrorMessageEnum } from '../../Enums'
import { Link } from 'react-router-dom'
import { ButtonGroup } from 'react-bootstrap'
import { friendlyErrorMessages } from '../../apiError/apiError'
import { ReactComponent as ExternalLink } from '../../../assets/images/ExternalLink.svg'
import './ErrorModal.scss'
import { capitalizeString } from '../../stringFormatHelper/StringFormatHelper'

interface ErrorModalProps {
  showModal: boolean
  message: string
  description: string
  retryMessage?: string
  actionButton?: any
  hideRetryMessage?: boolean
  titleMessage: string
  onClickCloseErrorModal: () => void
}

const ErrorModal: React.FC<ErrorModalProps> = ({
  message: initMessage,
  description: initDescription,
  retryMessage: initRetryMessage,
  titleMessage: initTitleMessage,
  showModal,
  actionButton,
  hideRetryMessage = false,
  onClickCloseErrorModal
}): JSX.Element => {
  const { isPremiumUser, isEnterpriseUser, isIntelUser } = useUserStore((state) => state)

  const canContactSupport = isPremiumUser() || isEnterpriseUser() || isIntelUser()

  const isAuthError = specificErrorMessageEnum.authErrorMessage.some((errMsg) =>
    initMessage?.toString().includes(errMsg)
  )

  const backendTryAgainMsg = 'Please try again later.'

  const getDefaultRetryMessage = (): string => {
    const tryMsgIncluded = message?.includes(backendTryAgainMsg)
    if (isAuthError) {
      return 'Contact your administrator for assistance.'
    } else if (canContactSupport) {
      return `${tryMsgIncluded ? '' : 'Please try again'} or contact support if the issue continues.`
    } else {
      return `${tryMsgIncluded ? '' : 'Please try again'} or check the community for hints.`
    }
  }

  const getAuthErrorDescription = (): string => {
    return friendlyErrorMessages.unathorizedAction
  }

  // props
  const message = isAuthError ? '' : initMessage
  const titleMessage = isAuthError ? 'Permission denied' : initTitleMessage || 'Could not launch your instance'
  const description = isAuthError
    ? getAuthErrorDescription()
    : initDescription || 'There was an error while processing your compute instance.'
  const retryMessage = initRetryMessage ?? getDefaultRetryMessage()

  const formatMessage = (message: any): JSX.Element => {
    if (!message) return <></>
    let formattedMessage: string = message.trim()
    formattedMessage = formattedMessage.charAt(0).toUpperCase() + formattedMessage.slice(1)

    if (!formattedMessage.endsWith('.')) formattedMessage += '.'
    if (formattedMessage.includes(backendTryAgainMsg)) formattedMessage = formattedMessage.slice(0, -1)
    return <div className={`d-inline ${isAuthError ? 'fw-bold' : ''}`}>{formattedMessage} </div>
  }

  const supportLink = (): JSX.Element => {
    if (isAuthError) {
      return (
        <Link
          intc-id="error-modal-my-roles-link"
          aria-label="Go to My roles page"
          className="link"
          to="/profile/accountsettings"
        >
          Go to my roles
        </Link>
      )
    } else if (canContactSupport) {
      return (
        <a
          intc-id="error-modal-contact-support-link"
          aria-label="Go to Contact Support page"
          className="link"
          href={idcConfig.REACT_APP_SUPPORT_PAGE}
          rel="noreferrer"
          target="_blank"
        >
          Contact Support
          <ExternalLink />
        </a>
      )
    } else {
      return (
        <a
          intc-id="error-modal-community-link"
          aria-label="Go to Contact Support page"
          className="link"
          href={idcConfig.REACT_APP_COMMUNITY}
          rel="noreferrer"
          target="_blank"
        >
          Community
          <BsArrowRightShort />
        </a>
      )
    }
  }

  return (
    <Modal
      show={showModal}
      backdrop="static"
      aria-labelledby="contained-modal-title-vcenter"
      centered
      aria-label="Error modal"
      className="errorModal m-xs-0"
      fullscreen="sm-down"
    >
      <Modal.Header closeButton onHide={onClickCloseErrorModal}>
        <Modal.Title className="d-flex flex-row gap-s4 align-items-center align-self-stretch">
          <BsExclamationTriangle />
          <h2 className="h5 d-sm-inline-block d-none">{titleMessage}</h2>
          <h2 className="h6 d-xs-inline-block d-sm-none">{titleMessage}</h2>
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="section p-0">
          <span className="h6">{capitalizeString(description)}</span>
          <span>
            {formatMessage(message)}
            {hideRetryMessage ? null : retryMessage}
          </span>
        </div>
      </Modal.Body>
      <Modal.Footer>
        <ButtonGroup className="m-0">
          {supportLink()}
          <Button
            onClick={onClickCloseErrorModal}
            className={actionButton?.class || ''}
            variant={actionButton?.variant || 'primary'}
          >
            {actionButton?.label || 'Go back'}
          </Button>
        </ButtonGroup>
      </Modal.Footer>
    </Modal>
  )
}

export default ErrorModal
