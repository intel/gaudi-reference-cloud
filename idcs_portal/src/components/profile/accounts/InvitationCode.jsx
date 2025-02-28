// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import CustomInput from '../../../utils/customInput/CustomInput'
import idcConfig from '../../../config/configurator'
import { BsEnvelope } from 'react-icons/bs'
import { AiOutlineLoading3Quarters } from 'react-icons/ai'

const InvitationCode = ({ isModalOpen, account, formState, onChange, sendCodeToMemberEmail, resendInvite }) => {
  if (!isModalOpen) {
    return null
  }

  const { form, actions } = formState
  const { invitationCode } = form

  return (
    <Modal
      onHide={actions.onCancel}
      show={isModalOpen}
      backdrop="static"
      keyboard={false}
      centered
      size="lg"
      aria-label="Invitation code modal"
    >
      <Modal.Header closeButton>
        <Modal.Title>
          Accept invitation from cloud account: <a href={`mailto: ${account.email}`}>{account.email}</a>
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="py-s6">
          <div className="d-flex flex-column bd-highlight">
            <div className="d-flex py-s3 bd-highlight gap-s4">
              <div className="flex-grow-1 bd-highlight">
                <CustomInput
                  type={invitationCode.type}
                  fieldSize={invitationCode.fieldSize}
                  placeholder={invitationCode.placeholder}
                  isRequired={invitationCode.validationRules.isRequired}
                  label={invitationCode.label}
                  value={invitationCode.value}
                  onChanged={(event) => {
                    onChange(event, 'invitationCode')
                  }}
                  isValid={invitationCode.isValid}
                  isTouched={invitationCode.isTouched}
                  helperMessage={invitationCode.helperMessage}
                  isReadOnly={invitationCode.isReadOnly}
                  validationMessage={invitationCode.validationMessage}
                />
              </div>
              <div className="d-flex flex-column gap-s4 bd-highlight">
                <div className="mt-s5 pt-s2 pb-s1"></div>
                <button
                  onClick={() => sendCodeToMemberEmail()}
                  className="btn btn-primary mt-s4"
                  aria-label="clear search"
                  type="button"
                >
                  {resendInvite ? <AiOutlineLoading3Quarters /> : <BsEnvelope />} Resend
                </button>
              </div>
            </div>
            <div className="py-s3 bd-highlight">
              By clicking &apos;Confirm&apos;, you agree to the{' '}
              <a target="_blank" href={idcConfig.REACT_APP_SERVICE_AGREEMENT_URL} rel="noreferrer">
                {`${idcConfig.REACT_APP_COMPANY_SHORT_NAME}’s ${idcConfig.REACT_APP_CONSOLE_SHORT_NAME} Service Agreement`}
              </a>{' '}
              and&nbsp;
              <a target="_blank" href={idcConfig.REACT_APP_SOFTWARE_AGREEMENT_URL} rel="noreferrer">
                {`${idcConfig.REACT_APP_COMPANY_SHORT_NAME}’s Standard Commercial Software and Services Terms and Conditions`}
              </a>
            </div>
          </div>
        </div>
      </Modal.Body>
      <Modal.Footer>
        <Button
          intc-id={'btn-accessaccount-accept-invite-cancel'}
          data-wap_ref={'btn-accessaccount-accept-invite-cancel'}
          aria-label="Cancel"
          variant="outline-primary"
          onClick={actions.onCancel}
        >
          Cancel
        </Button>
        <Button
          intc-id={'btn-accessaccount-accept-invite-confirm'}
          data-wap_ref={'btn-accessaccount-accept-invite-confirm'}
          aria-label="Confirm"
          variant="primary"
          disabled={!invitationCode.value}
          onClick={actions.onConfirm}
        >
          Confirm
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default InvitationCode
