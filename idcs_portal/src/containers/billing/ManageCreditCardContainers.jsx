// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useNavigate } from 'react-router'
import ManageCreditCard from '../../components/billing/manageCreditCard/ManageCreditCard'
import useToastStore from '../../store/toastStore/ToastStore'

const ManageCreditCardContainers = () => {
  // States
  const titles = {
    pageTitle: 'Add a credit card'
  }

  const cancelButtonOptions = {
    label: 'Cancel',
    onClick: () => goToPaymentMethods()
  }

  const formActions = {
    afterSuccess: () => goToPaymentMethods(),
    afterError: toggleError
  }

  const showError = useToastStore((state) => state.showError)

  // Navigation
  const navigate = useNavigate()

  // Functions
  function goToPaymentMethods() {
    navigate({ pathname: '/billing/managePaymentMethods' })
  }

  function toggleError(errorMessage) {
    showError(errorMessage)
  }

  return <ManageCreditCard titles={titles} cancelButtonOptions={cancelButtonOptions} formActions={formActions} />
}

export default ManageCreditCardContainers
