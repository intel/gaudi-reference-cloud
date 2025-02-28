// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import idcConfig from '../../../config/configurator'

const TermsAndConditionsModal = ({ show, onClickOption }) => {
  return (
    <Modal
      show={show}
      onHide={() => onClickOption(false)}
      centered
      backdrop="static"
      keyboard={false}
      size="lg"
      aria-label="Terms and conditions modal"
    >
      <Modal.Header closeButton>
        <Modal.Title>{`${idcConfig.REACT_APP_CONSOLE_LONG_NAME} User Terms and Conditions`}</Modal.Title>
      </Modal.Header>

      <Modal.Body>
        <div className="text-left">
          {`Thank you for your interest in ${idcConfig.REACT_APP_CONSOLE_LONG_NAME} Services articulated in the Order Form (the
          'Services'). ${idcConfig.REACT_APP_COMPANY_SHORT_NAME}’s offer and delivery of the Services to You are subject to the terms
          and conditions articulated in this Order Form, `}
          <a target="_blank" href={idcConfig.REACT_APP_SERVICE_AGREEMENT_URL} rel="noreferrer">
            {`${idcConfig.REACT_APP_COMPANY_SHORT_NAME}’s ${idcConfig.REACT_APP_CONSOLE_SHORT_NAME} Service Agreement`}
          </a>
          , and&nbsp;
          <a target="_blank" href={idcConfig.REACT_APP_SOFTWARE_AGREEMENT_URL} rel="noreferrer">
            {`${idcConfig.REACT_APP_COMPANY_SHORT_NAME}’s Standard Commercial Software and Services Terms and Conditions`}
          </a>{' '}
          (collectively, the &apos;Legal Terms and Conditions&apos;).
          <br /> <br /> By clicking &apos;I Accept&apos;, below, You acknowledge that You have reviewed, understand, and
          accept the Legal Terms and Conditions as a prerequisite to and condition of Your access and use of the
          Services. Please do not access or use the Services unless and until you agree to the Legal Terms and
          Conditions.
        </div>
      </Modal.Body>
      <Modal.Footer>
        <Button
          intc-id={'btn-TermsAndConditions-cancel'}
          data-wap_ref={'btn-TermsAndConditions-cancel'}
          variant="outline-primary"
          onClick={() => onClickOption(false)}
        >
          Cancel
        </Button>
        <Button
          intc-id={'btn-TermsAndConditions-Accept'}
          data-wap_ref={'btn-TermsAndConditions-Accept'}
          variant="primary"
          onClick={() => onClickOption(true)}
        >
          I Accept
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default TermsAndConditionsModal
