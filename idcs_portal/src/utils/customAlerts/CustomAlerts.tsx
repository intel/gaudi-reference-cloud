// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { BsExclamationTriangle, BsCheckCircle, BsInfoCircle, BsExclamationOctagon } from 'react-icons/bs'
import { Alert } from 'react-bootstrap'
import './CustomAlerts.scss'
import { type BannerLink } from '../../store/bannerStore/BannerStore'

interface CustomAlertsProps {
  showAlert: boolean
  alertType: string
  title?: string
  strongText?: string
  message: string
  link?: BannerLink
  onCloseAlert?: (action: boolean) => void
  showIcon?: boolean
  className?: string
}

const CustomAlerts: React.FC<CustomAlertsProps> = ({
  showAlert,
  alertType,
  title,
  strongText,
  message,
  link,
  onCloseAlert,
  showIcon,
  className
}): JSX.Element => {
  let variant = 'info'
  let icon = null
  const showCloseButton = typeof onCloseAlert === 'function' && onCloseAlert !== undefined

  switch (alertType) {
    case 'error':
      variant = 'danger'
      icon = <BsExclamationOctagon />
      break
    case 'success':
      variant = 'success'
      icon = <BsCheckCircle />
      break
    case 'info':
      variant = 'info'
      icon = <BsInfoCircle />
      break
    case 'secondary':
      variant = 'secondary'
      icon = <BsInfoCircle />
      break
    case 'warning':
      variant = 'warning'
      icon = <BsExclamationTriangle />
      break
    default:
      break
  }

  const body = strongText ? (
    <span className="alert-body">
      <strong>{strongText}</strong>
      <span className="d-flex-inline gap-s4">
        {` ${message}`}
        {link && (
          <>
            &nbsp;
            <a
              className="link"
              aria-label={link.label}
              target={link.openInNewTab ? '_blank' : '_self'}
              rel={link.openInNewTab ? 'noreferrer' : undefined}
              href={link.href}
            >
              {link.label}
            </a>
            .
          </>
        )}
      </span>
    </span>
  ) : (
    <span className="alert-body">
      {' '}
      <span className="d-flex-inline gap-s4">
        {message}
        {link && (
          <>
            &nbsp;
            <a
              className="link"
              aria-label={link.label}
              target={link.openInNewTab ? '_blank' : '_self'}
              rel={link.openInNewTab ? 'noreferrer' : undefined}
              href={link.href}
            >
              {link.label}
            </a>
            .
          </>
        )}
      </span>
    </span>
  )

  return (
    <>
      {showAlert ? (
        <Alert
          variant={variant}
          className={`${className ? `${className} ` : ''}`}
          role="alert"
          onClose={() => {
            if (onCloseAlert) {
              onCloseAlert(false)
            }
          }}
          dismissible={showCloseButton}
        >
          {showIcon ? <div className="d-flex flex-column justify-content-center align-self-center">{icon}</div> : null}
          <div className="alert-body gap-s3">
            {title ? <Alert.Heading className="h6">{title}</Alert.Heading> : null}
            {body}
          </div>
        </Alert>
      ) : null}
    </>
  )
}

export default CustomAlerts
