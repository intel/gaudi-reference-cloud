// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import moment from 'moment'
import { useNavigate } from 'react-router'
import ManagePaymentMethods from '../../components/billing/managePaymentMethods/ManagePaymentMethods'
import CloudCreditsService from '../../services/CloudCreditsService'
import usePaymentMethodStore from '../../store/billingStore/PaymentMethodStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { formatNumber } from '../../utils/numberFormatHelper/NumberFormatHelper'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import useToastStore from '../../store/toastStore/ToastStore'

const ManagePaymentMethodsContainer = (props) => {
  const title = 'Manage Payment Methods'
  const helperTitle = 'Choose between adding a credit card or a cloud credit'

  const cloudInitial = {
    title: 'Cloud Credits',
    type: 'cloud',
    subTitle: 'Cloud credits outstanding balance',
    balance: '',
    expDt: '',
    helperMessage: 'Your account has no cloud credits',
    actions: [
      {
        buttonLabel: 'View Credit Details',
        function: () => setAction('View')
      },
      {
        buttonLabel: 'Redeem Coupon',
        function: () => setAction('Redeem')
      }
    ]
  }

  const cardInitial = {
    title: 'Credit Card',
    type: 'card',
    subTitle: 'Default payment card',
    cardHolderName: '',
    cardType: '',
    cardNbr: '',
    expDt: '',
    brandtype: '',
    helperMessage: 'Your account has no credit card on file',
    actions: isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_DIRECTPOST)
      ? [
          {
            buttonLabel: 'Add card',
            function: setAction
          }
        ]
      : []
  }

  // Navigation
  const navigate = useNavigate()

  // Local state
  const [cardCredits, setCardCredits] = useState(cardInitial)
  const [cloudCredits, setCloudCredits] = useState(cloudInitial)
  const [loading, setLoading] = useState(false)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // Global State
  const creditCardResponse = usePaymentMethodStore((state) => state.creditCardResponse)
  const setCreditCardResponse = usePaymentMethodStore((state) => state.setCreditCardResponse)

  const throwError = useErrorBoundary()

  // Hook
  useEffect(() => {
    const fetch = async () => {
      try {
        await getCreditDetails()
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    if (creditCardResponse === null) {
      return
    }
    if (creditCardResponse.success) {
      showSuccess(creditCardResponse.message)
    } else {
      showError(creditCardResponse.message, true)
    }
    setCreditCardResponse(null)
  }, [creditCardResponse])

  function setAction(action) {
    switch (action) {
      case 'Redeem':
        navigate({
          pathname: '/billing/credits/managecouponcode'
        })
        break
      case 'View':
        navigate({
          pathname: '/billing/credits'
        })
        break
      default:
        navigate({
          pathname: '/billing/managePaymentMethods/managecreditcard'
        })
        break
    }
  }

  async function getCreditDetails() {
    setLoading(true)
    const response = await CloudCreditsService.getResumeCredits()
    const optionsResponse = await CloudCreditsService.getCreditOptions()
    const data = response?.data
    if (data) {
      const cloudUpdated = { ...cloudInitial }
      cloudUpdated.balance = `$ ${formatNumber(data.totalRemainingAmount, 2)}`
      cloudUpdated.expDt = moment(data.expirationDate).format('MM/DD/YYYY')
      setCloudCredits(cloudUpdated)
    }
    const dataOpts = { ...optionsResponse?.data }
    const cardUpdated = { ...cardCredits }
    const paymentType = dataOpts.paymentType
    if (paymentType !== 'PAYMENT_UNSPECIFIED') {
      const data = dataOpts
      cardUpdated.cardHolderName = data?.firstName + ' ' + data?.lastName
      cardUpdated.cardNbr = data?.creditCard?.suffix
      cardUpdated.expDt = data?.creditCard?.expiration
      cardUpdated.brandtype = data?.creditCard?.type
    }
    setLoading(false)
    setCardCredits(cardUpdated)
  }

  return (
    <ManagePaymentMethods
      title={title}
      helperTitle={helperTitle}
      cloudCredits={cloudCredits}
      cardCredits={cardCredits}
      loading={loading}
    />
  )
}

export default ManagePaymentMethodsContainer
