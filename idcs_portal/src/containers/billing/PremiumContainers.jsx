// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useState, useEffect } from 'react'

import Premium from '../../components/billing/premium/Premium'
import useUserStore from '../../store/userStore/UserStore'
import { EnrollActionResponse } from '../../utils/Enums'
import idcConfig from '../../config/configurator'
import useToastStore from '../../store/toastStore/ToastStore'

const PremiumContainers = () => {
  // States
  const titles = {
    pageTitle: `Welcome to the ${idcConfig.REACT_APP_CONSOLE_LONG_NAME}`,
    mainTitle: 'Let’s finish setting up your account',
    mainDesc: 'Provide a payment method for billing purposes. Choose one of the following:'
  }

  const isPremiumUser = useUserStore((state) => state.isPremiumUser)
  const enrollResponse = useUserStore((state) => state.enrollResponse)
  const showError = useToastStore((state) => state.showError)

  const cancelButtonOptions = {
    label: 'Skip',
    onClick: (isShow = true) => toggleSkip(isShow)
  }

  const formActions = {
    afterSuccess: () => goHome(),
    afterError: toggleError
  }

  const [isShowSkip, setIsShowSkip] = useState(false)

  // Functions
  function goHome() {
    window.location.href = '/'
  }

  function toggleError(errorMessage) {
    showError(errorMessage, true)
  }

  function toggleSkip(isShow) {
    setIsShowSkip(isShow)
  }

  useEffect(() => {
    const shouldStayOnThePage =
      location.pathname.toLowerCase() === '/premium' &&
      isPremiumUser() &&
      enrollResponse.action === EnrollActionResponse.ENROLL_ACTION_COUPON_OR_CREDIT_CARD
    if (!shouldStayOnThePage) {
      goHome()
    }
  }, [])

  return (
    <Premium
      titles={titles}
      cancelButtonOptions={cancelButtonOptions}
      formActions={formActions}
      isShowSkip={isShowSkip}
    />
  )
}

export default PremiumContainers
