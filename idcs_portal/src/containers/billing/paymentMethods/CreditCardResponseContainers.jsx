// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect } from 'react'
import AriaErrorCodes from '../../../utils/metadata/AriaErrorCodes.json'
import { useNavigate } from 'react-router'
import PaymentMethodService from '../../../services/PaymentMethodService'
import useErrorBoundary from '../../../hooks/useErrorBoundary'
import usePaymentMethodStore from '../../../store/billingStore/PaymentMethodStore'
import Spinner from '../../../utils/spinner/Spinner'

const CreditCardResponseContainers = () => {
  const searchParams = new URLSearchParams(document.location.search)

  const setCreditCardResponse = usePaymentMethodStore((state) => state.setCreditCardResponse)

  const throwError = useErrorBoundary()

  // Navigation
  const navigate = useNavigate()

  const errors = searchParams.get('errors')

  // TODO log credit card error
  const logError = (errorKey, errorCode, errorMessage) => {
    console.error(`ARIA error [errorkey: ${errorKey}] [errorCode: ${errorCode}] [errorMessage: ${errorMessage}]`)
  }

  const getErrorMessageFromQueryString = (index) => {
    const errorKeyString = `error_messages[${index}][error_key]`
    const errorCodeString = `error_messages[${index}][error_code]`
    const errorKey = searchParams.get(errorKeyString)
    const errorCode = searchParams.get(errorCodeString)
    try {
      const error = AriaErrorCodes[errorKey]
      logError(errorKey, errorCode, error.message)
      return error
    } catch (error) {
      return {} // Error not found in message
    }
  }

  const extractErrorMessages = () => {
    const messages = []
    for (let i = 0; i < errors; i++) {
      const errorMessage = getErrorMessageFromQueryString(i)
      if (errorMessage) {
        messages.push(errorMessage)
      }
    }
    return messages
  }

  function redirectToPaymentMethods() {
    navigate({
      pathname: '/billing/managePaymentMethods'
    })
  }

  async function postPaymentAPI() {
    try {
      const paymentMethodNo = searchParams.get('payment_method_no')
      await PaymentMethodService.getPostPayment(paymentMethodNo)
      setCreditCardResponse({ success: true, message: 'Credit card added to account' })
    } catch (error) {
      const isApiErrorWithErrorMessage = Boolean(error?.response?.data?.message)
      if (isApiErrorWithErrorMessage) {
        const authorizationPaymentFailed = error?.response?.data?.code === 13
        if (authorizationPaymentFailed) {
          setCreditCardResponse({
            success: false,
            message:
              'We were unable to place a temporary hold on your card. Please check if your card is active and has sufficient funds.'
          })
        } else {
          throw error
        }
      } else {
        throw error
      }
    }
  }

  async function afterDirectPost() {
    try {
      await postPaymentAPI()
      await PaymentMethodService.creditMigrate()
      redirectToPaymentMethods()
    } catch (error) {
      throwError(error)
    }
  }

  useEffect(() => {
    if (errors === '0') {
      afterDirectPost()
    } else {
      const errorMessages = extractErrorMessages()
      const showVerificationCode = errorMessages.some((x) => x.uiResponse === 'verify')
      if (showVerificationCode) {
        setCreditCardResponse({ success: false, message: 'Please verify your card information and try again.' })
      } else {
        setCreditCardResponse({
          success: false,
          message: 'Could not update your payment information. Please try again.'
        })
      }
      redirectToPaymentMethods()
    }
  }, [])

  return (
    <>
      <Spinner />
      <div className="invisible" intc-id="creditCardResponseContainer">
        ...Loading
      </div>
    </>
  )
}

export default CreditCardResponseContainers
