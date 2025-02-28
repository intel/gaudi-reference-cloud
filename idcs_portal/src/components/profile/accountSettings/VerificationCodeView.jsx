// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import Modal from 'react-bootstrap/Modal'
import Button from 'react-bootstrap/Button'
import CustomInput from '../../../utils/customInput/CustomInput'

const VerificationCodeView = ({
  cloudAccountId,
  verificationCodeField,
  onChangeInput,
  isFormValid,
  isModalOpen,
  formActions,
  userEmail,
  resendCodeMessage,
  setResendCodeMessage,
  cancelVerification,
  isUserBlocked,
  blockedTimer
}) => {
  const modalTitle = `Grant access to cloud account ID: ${cloudAccountId}`

  if (!isModalOpen) {
    return null
  }

  if (resendCodeMessage !== '') {
    setTimeout(() => {
      setResendCodeMessage('')
    }, 3000)
  }

  return (
    <Modal
      onHide={() => cancelVerification()}
      show={isModalOpen}
      backdrop="static"
      keyboard={false}
      centered
      size="lg"
      aria-label="Verification code modal"
    >
      <Modal.Header closeButton>
        <Modal.Title>{modalTitle}</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="section">
          <span>We sent a code to {userEmail}.</span>
          <CustomInput
            type={verificationCodeField.type}
            fieldSize={verificationCodeField.fieldSize}
            placeholder={verificationCodeField.placeholder}
            isRequired={verificationCodeField.validationRules.isRequired}
            label={
              verificationCodeField.validationRules.isRequired
                ? verificationCodeField.label + ' *'
                : verificationCodeField.label
            }
            value={verificationCodeField.value}
            onChanged={(event) => onChangeInput(event)}
            isValid={verificationCodeField.isValid}
            isTouched={verificationCodeField.isTouched}
            helperMessage={verificationCodeField.helperMessage}
            isReadOnly={verificationCodeField.isReadOnly}
            validationMessage={verificationCodeField.validationMessage}
          />
        </div>
      </Modal.Body>
      <Modal.Footer>
        {resendCodeMessage !== '' ? <span>{resendCodeMessage}</span> : null}
        <Button
          intc-id={'btn-accessmanagement-reSendCode'}
          data-wap_ref={'btn-accessmanagement-reSendCode'}
          variant="outline-primary"
          aria-label="Send new code"
          onClick={() => formActions.reSendOtp()}
        >
          Send new code
        </Button>
        <Button
          intc-id={'btn-accessmanagement-verifyCode'}
          data-wap_ref={'btn-accessmanagement-verifyCode'}
          variant="primary"
          disabled={!isFormValid}
          aria-label="Verify"
          onClick={() => formActions.verifyOtp()}
        >
          Verify {isUserBlocked ? `(${blockedTimer})` : ''}
        </Button>
      </Modal.Footer>
    </Modal>
  )
}

export default VerificationCodeView
