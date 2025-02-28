// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useNavigate } from 'react-router'

import ManageCouponCode from '../../components/billing/manageCouponCode/ManageCouponCode'
import useToastStore from '../../store/toastStore/ToastStore'

const ManageCouponCodeContainers = () => {
  // States
  const titles = {
    pageTitle: 'Redeem coupon'
  }

  const cancelButtonOptions = {
    label: 'Cancel',
    onClick: () => goBack()
  }

  const formActions = {
    afterSuccess: () => goBack(),
    afterError: toggleError
  }

  const showError = useToastStore((state) => state.showError)

  // Navigation
  const navigate = useNavigate()

  // Functions
  function goBack() {
    navigate({ pathname: '/billing/credits' })
  }

  function toggleError(errorMessage) {
    showError(errorMessage, true)
  }

  return <ManageCouponCode titles={titles} cancelButtonOptions={cancelButtonOptions} formActions={formActions} />
}

export default ManageCouponCodeContainers
