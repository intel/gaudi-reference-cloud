// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useNavigate } from 'react-router'
import UpgradeAccount from '../../components/billing/upgradeAccount/UpgradeAccount'
import useToastStore from '../../store/toastStore/ToastStore'

const UpgradeAccountContainers = () => {
  // States
  const titles = {
    pageTitle: 'Upgrade your account',
    pageDesc: 'Please add a payment method to upgrade your account.',
    mainTitle: 'Add payment method',
    mainDesc: 'Provide a payment method for billing purposes. Choose one of the following:'
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
    navigate(-1)
  }

  function toggleError(errorMessage) {
    showError(errorMessage, true)
  }

  return <UpgradeAccount titles={titles} cancelButtonOptions={cancelButtonOptions} formActions={formActions} />
}

export default UpgradeAccountContainers
