// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import ManagePayment from '../../../components/billing/managePayment/ManagePayment'
import { appFeatureFlags, isFeatureFlagEnable } from '../../../config/configurator'
import { useState } from 'react'

const ManagePaymentContainers = (props) => {
  // Props Variables
  const cancelButtonOptions = props.cancelButtonOptions
  const formActions = props.formActions

  const buttonSelection = {
    creditCardVariant: 'primary',
    couponVariant: 'outline-primary',
    isShow: isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_DIRECTPOST)
      ? 'creditCard'
      : isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_UPGRADE_PREMIUM_COUPON_CODE)
        ? 'couponCode'
        : null
  }

  const [showFormOnButton, setShowFormOnButton] = useState(buttonSelection)

  function actionOnButtonSelection(type) {
    const updateForm = {}

    if (type === 'coupon') {
      updateForm.creditCardVariant = 'outline-primary'
      updateForm.couponVariant = 'primary'
      updateForm.isShow = 'couponCode'
    } else {
      updateForm.creditCardVariant = 'primary'
      updateForm.couponVariant = 'outline-primary'
      updateForm.isShow = 'creditCard'
    }

    setShowFormOnButton(updateForm)
  }

  return (
    <ManagePayment
      showFormOnButton={showFormOnButton}
      actionOnButtonSelection={actionOnButtonSelection}
      cancelButtonOptions={cancelButtonOptions}
      formActions={formActions}
    />
  )
}

export default ManagePaymentContainers
