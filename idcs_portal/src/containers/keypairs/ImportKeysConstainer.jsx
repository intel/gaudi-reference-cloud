// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import KeyPairsForm from '../../components/keypairs/KeyPairsForm'
import { useNavigate } from 'react-router'
import useToastStore from '../../store/toastStore/ToastStore'

const ImportKeysContainer = () => {
  const navigate = useNavigate()
  const showSuccess = useToastStore((state) => state.showSuccess)

  const callAfterSuccess = () => {
    showSuccess('Key added successfully.')
    navigate({
      pathname: '/security/publickeys'
    })
  }

  return <KeyPairsForm callAfterSuccess={callAfterSuccess} keysPagePath="/security/publickeys/"></KeyPairsForm>
}

export default ImportKeysContainer
