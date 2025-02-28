// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState } from 'react'
import InvitationCode from '../../components/profile/accounts/InvitationCode'
import { UpdateFormHelper } from '../../utils/updateFormHelper/UpdateFormHelper'
import useToastStore from '../../store/toastStore/ToastStore'

const InvitationCodeContainer = ({ account, confirmInvite, showModal, setShowModal }) => {
  const formInitial = {
    form: {
      invitationCode: {
        type: 'text',
        label: 'Invitation code',
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
        helperMessage: `The code sent to your email by account owner ${account.email}`
      }
    },
    actions: {
      onConfirm: () => {
        confirmInviteCode()
      },
      onCancel: () => {
        setShowModal(false)
      }
    }
  }

  const showSuccess = useToastStore((state) => state.showSuccess)
  const showError = useToastStore((state) => state.showError)

  const [formState, setFormState] = useState(formInitial)

  const onChange = (event, formInputName) => {
    const updatedState = {
      ...formState
    }
    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)
    setFormState(updatedForm)
  }

  const confirmInviteCode = async () => {
    const inviteCode = formState.form.invitationCode.value
    if (inviteCode !== '') {
      try {
        await confirmInvite(account.cloudAccountId, inviteCode)
        showSuccess('Invitation Confirmed!')
        setShowModal(false)
      } catch (error) {
        if (error.response) {
          showError(error.response.data.message)
        } else {
          showError(error.message)
        }
        setShowModal(false)
      }
    }
  }

  return <InvitationCode isModalOpen={showModal} form={formState} account={account} onChange={onChange} />
}

export default InvitationCodeContainer
