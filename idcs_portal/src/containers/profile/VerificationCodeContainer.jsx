// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import useUserStore from '../../store/userStore/UserStore'
import { useState, useRef } from 'react'
import { UpdateFormHelper, isValidForm, getFormValue } from '../../utils/updateFormHelper/UpdateFormHelper'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import CloudAccountService from '../../services/CloudAccountService'
import VerificationCodeView from '../../components/profile/accountSettings/VerificationCodeView'

const VerificationCodeContainer = ({
  showVerificationModal,
  emailAddress,
  successVerification,
  cancelVerification
}) => {
  const verificationFormIntial = {
    verificationCode: {
      type: 'text',
      label: 'Verification code:',
      placeholder: '',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      validationRules: {
        isRequired: true,
        number: true
      },
      validationMessage: '',
      helperMessage: 'Please provide the verification code sent to your email'
    }
  }

  const validFormIntial = false

  const resendCodeMessageIntial = ''

  const formActions = {
    reSendOtp: () => {
      reSendOtp()
    },
    verifyOtp: () => {
      verifyOtp()
    }
  }

  const blockedUserTime = 60000
  const blockedUserCountDown = blockedUserTime / 1000

  const [isFormValid, setIsFormValid] = useState(validFormIntial)
  const [isModalOpen, setIsModalOpen] = useState(showVerificationModal)
  const [verificationForm, setVerificationForm] = useState(verificationFormIntial)
  const [resendCodeMessage, setResendCodeMessage] = useState(resendCodeMessageIntial)

  const [isUserBlocked, setIsUserBlocked] = useState(false)
  const [blockedTimer, setBlockedTimer] = useState(blockedUserCountDown - 1)

  // Global State
  const user = useUserStore((state) => state.user)

  const otpRef = useRef(null)
  const interval = useRef(null)

  const throwError = useErrorBoundary()

  const verifyOtp = async () => {
    try {
      const otp = getFormValue('verificationCode', otpRef.current)
      const { data } = await CloudAccountService.verifyOtp(emailAddress, otp)

      if (data.validated === true) {
        setIsModalOpen(false)
        successVerification()
      } else {
        let message = ''

        if (data.otpState === 'OTP_STATE_EXPIRED') {
          message = 'Verification code expired'
        } else {
          if (data.retryLeft === 0) {
            message = 'The maximum verification limit has been reached. Kindly retry after 1 minute.'
            handleTimeOut(true)
            setTimeout(() => {
              handleTimeOut(false)
            }, blockedUserTime)
          } else {
            message = `Incorrect verification code. ${data.retryLeft} attempt(s) left.`
          }
        }

        const verificationFormCopy = { ...verificationForm }
        verificationFormCopy.verificationCode.validationMessage = message
        verificationFormCopy.verificationCode.isValid = false
        setVerificationForm({ ...verificationFormCopy })

        setIsFormValid(false)
      }
    } catch (error) {
      throwError(error)
    }
  }

  const reSendOtp = async () => {
    try {
      await CloudAccountService.resendOtp(emailAddress)
      setResendCodeMessage('verification code resent.')
    } catch (error) {
      throwError(error)
    }
  }

  function onChangeInput(event) {
    const value = event.target.value

    validateInput(value)
  }

  const validateInput = (value) => {
    const updatedForm = UpdateFormHelper(value, 'verificationCode', verificationForm)
    let isValid = isValidForm(updatedForm)

    if (isUserBlocked) {
      isValid = false
    }

    setIsFormValid(isValid)
    setVerificationForm(updatedForm)
    otpRef.current = updatedForm
  }

  const handleTimeOut = (action) => {
    if (action) {
      setIsUserBlocked(true)
      interval.current = setInterval(() => {
        setBlockedTimer((blockedTimer) => blockedTimer - 1)
      }, 1000)
    } else {
      clearInterval(interval.current)
      setBlockedTimer(blockedUserCountDown - 1)
      setIsUserBlocked(false)
      validateInput('')
    }
  }

  return (
    <VerificationCodeView
      cloudAccountId={user.cloudAccountNumber}
      onChangeInput={onChangeInput}
      verificationCodeField={verificationForm.verificationCode}
      isFormValid={isFormValid}
      isModalOpen={isModalOpen}
      formActions={formActions}
      resendCodeMessage={resendCodeMessage}
      setResendCodeMessage={setResendCodeMessage}
      cancelVerification={cancelVerification}
      userEmail={user.email}
      isUserBlocked={isUserBlocked}
      blockedTimer={blockedTimer}
    />
  )
}

export default VerificationCodeContainer
