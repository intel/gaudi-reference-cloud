// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import { Link } from 'react-router-dom'
import { capitalizeString } from '../../stringFormatHelper/StringFormatHelper'
import idcConfig from '../../../config/configurator'
import useUserStore from '../../../store/userStore/UserStore'

interface EmptyCatalogModalInterface {
  show: boolean
  product: string
  goBackPath: string
  size?: 'sm' | 'lg' | 'xl'
  extraExplanation?: string
  extraActions?: any
}

const EmptyCatalogModal = ({
  show,
  size = 'lg',
  product,
  goBackPath,
  extraExplanation = '',
  extraActions = null
}: EmptyCatalogModalInterface): JSX.Element => {
  const { isPremiumUser, isEnterpriseUser } = useUserStore((state) => state)
  const supportLink = isEnterpriseUser()
    ? idcConfig.REACT_APP_SUBMIT_TICKET_ENTERPRISE
    : isPremiumUser()
      ? idcConfig.REACT_APP_SUBMIT_TICKET_PREMIUM
      : idcConfig.REACT_APP_SUBMIT_TICKET

  return (
    <Modal
      show={show}
      backdrop="static"
      keyboard={false}
      size={size}
      centered
      aria-label="Exclusive Feature Access"
      intc-id="Empty-Catalog"
      data-wap_ref="Empty-Catalog"
    >
      <Modal.Header>
        <Modal.Title>Exclusive Feature Access</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="text-start">
          <p>
            {capitalizeString(product)} is reserved for purposely selected users. To get access, submit a ticket and
            share how {product.toLowerCase()} will enhance your AI projects.
          </p>
          {extraExplanation && <p>{extraExplanation}</p>}
        </div>
      </Modal.Body>
      <Modal.Footer className={extraActions ? 'justify-content-between' : 'justify-content-end'}>
        {extraActions}
        <div>
          <Link to={goBackPath} className="btn btn-outline-primary">
            Go Back
          </Link>
          <a
            href={supportLink}
            target="_blank"
            rel="noreferrer"
            role="button"
            aria-label="Request access"
            intc-id="btn-emptyCatalog-requestAccess"
            className="btn btn-primary ms-3"
          >
            Request access
          </a>
        </div>
      </Modal.Footer>
    </Modal>
  )
}

export default EmptyCatalogModal
