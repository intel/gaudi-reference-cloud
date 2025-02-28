// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useState, useEffect } from 'react'
import { useLocation, useSearchParams } from 'react-router-dom'
import CouponCode from '../../../components/billing/couponCode/CouponCode'
import {
  UpdateFormHelper,
  getFormValue,
  isValidForm,
  showFormRequiredFields
} from '../../../utils/updateFormHelper/UpdateFormHelper'
import useUserStore from '../../../store/userStore/UserStore'
import CloudCreditsService from '../../../services/CloudCreditsService'
import useToastStore from '../../../store/toastStore/ToastStore'
import CloudAccountService from '../../../services/CloudAccountService'
import PaymentMethodService from '../../../services/PaymentMethodService'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'
import { toastMessageEnum } from '../../../utils/Enums'

const CouponCodeContainers = (props) => {
  // Props Variables
  const cancelButtonOptions = props.cancelButtonOptions
  const formActions = props.formActions

  // //Global state
  const cloudAccountNumber = useUserStore((state) => state.user.cloudAccountNumber)
  const showSuccess = useToastStore((state) => state.showSuccess)
  const showErrorToast = useToastStore((state) => state.showError)
  const enroll = useUserStore((state) => state.enroll)
  const isStandardUser = useUserStore((state) => state.isStandardUser)
  const isPremiumUser = useUserStore((state) => state.isPremiumUser)
  const { pathname } = useLocation()
  const [searchParam] = useSearchParams()
  const servicePayload = {
    cloudAccountId: null,
    code: null
  }

  // local state
  const initialState = {
    form: {
      couponCode: {
        sectionGroup: 'couponCode',
        type: 'text', // options = 'text ,'textArea'
        label: 'Coupon Code:',
        placeholder: 'Coupon Code',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        isRequired: true,
        maxLength: 25,
        validationRules: {
          checkMaxLength: true,
          isRequired: true
        },
        validationMessage: ''
      },
      isValidForm: false
    },
    timeoutMiliseconds: 1000,
    showSpinner: false
  }

  const errorModal = {
    showErrorModal: false
  }

  const [state, setState] = useState(initialState)
  const [showError, setShowError] = useState(errorModal)

  useEffect(() => {
    handleCouponParam()
  }, [])

  function handleCouponParam() {
    const coupon = searchParam.get('coupon')
    if (coupon) {
      const updatedState = {
        ...state
      }
      const updatedForm = UpdateFormHelper(coupon, 'couponCode', updatedState.form)

      updatedForm.isValidForm = isValidForm(updatedForm)

      updatedState.form = updatedForm

      setState(updatedState)
    }
  }

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedForm.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  function showRequiredFields() {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(stateCopy.form)
    // Create toast
    showErrorToast(toastMessageEnum.formValidationError, false)
    stateCopy.form = updatedForm
    setState(stateCopy)
  }

  async function submitForm() {
    const updatedState = { ...state }
    const isValidForm = state.form.isValidForm
    if (!isValidForm) {
      showRequiredFields()
      return
    }
    updatedState.showSpinner = true

    const payload = { ...servicePayload }
    payload.cloudAccountId = cloudAccountNumber
    payload.code = getFormValue('couponCode', state.form).trim()

    try {
      setState(updatedState)
      if (pathname === '/upgradeaccount' && isStandardUser()) {
        await CloudAccountService.upgradeCloudAccountByCoupon(payload.code)
        await enroll(false, false)
      } else {
        await CloudCreditsService.postCredit(payload)
      }

      await creditMigrate()
      showSuccess('Coupon redeemed successfully')
      formActions.afterSuccess()
    } catch (error) {
      if (error.config.url.includes('/creditmigrate')) {
        setShowError({
          showErrorModal: true,
          titleMessage: 'Credit Migration fails',
          description: 'The coupon was redeemed successfully, but credit migration failed.',
          message: 'Please contact support.',
          hideRetryMessage: true,
          actionButton: {
            variant: 'secondary'
          }
        })
      } else {
        if (error.response) {
          formActions.afterError(error.response.data.message)
        } else {
          formActions.afterError(error.message)
        }
      }
    } finally {
      setState((prev) => ({ ...prev, showSpinner: false }))
    }
  }

  const onClickCloseErrorModal = () => {
    formActions.afterSuccess()
  }

  async function creditMigrate() {
    if (pathname === '/upgradeaccount' || isPremiumUser()) {
      await PaymentMethodService.creditMigrate()
    }
  }

  return (
    <>
      <CouponCode
        state={state}
        onChangeInput={onChangeInput}
        submitForm={submitForm}
        cancelButtonOptions={cancelButtonOptions}
      />
      <ErrorModal
        showModal={showError.showErrorModal}
        titleMessage={showError?.titleMessage}
        message={showError?.message}
        description={showError?.description}
        hideRetryMessage={showError?.hideRetryMessage}
        actionButton={showError?.actionButton}
        onClickCloseErrorModal={onClickCloseErrorModal}
      />
    </>
  )
}

export default CouponCodeContainers
