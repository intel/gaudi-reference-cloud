// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useState } from 'react'
import ActionConfirmation from '../../utils/modals/actionConfirmation/ActionConfirmation'
import useToastStore from '../../store/toastStore/ToastStore'
import InvitationCode from '../../components/profile/accounts/InvitationCode'
import { UpdateFormHelper } from '../../utils/updateFormHelper/UpdateFormHelper'
import AccountsAction from '../../components/profile/accounts/AccountsAction'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import CloudAccountService from '../../services/CloudAccountService'

const AccountsActionContainer = ({ account, confirmInvite, rejectInvite }) => {
  const modalContent = {
    label: 'Decline Cloud Account ' + account.cloudAccountId,
    buttonLabel: 'Decline',
    uuid: '',
    name: '',
    resourceId: '',
    question: `Do you want to decline the invite for Cloud Account ${account.cloudAccountId} ?`,
    feedback: '',
    resourceType: ''
  }

  const formInitial = {
    form: {
      invitationCode: {
        type: 'text',
        label: 'Invitation code:',
        placeholder: 'Enter code from your email invitation',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        validationRules: {
          isRequired: true,
          number: true
        },
        validationMessage: '',
        helperMessage: ''
      }
    },
    actions: {
      onConfirm: () => {
        confirmInviteCode()
      },
      onCancel: () => {
        setShowModalInviteCodeConfirmation(false)
      }
    }
  }

  const [actionModalContent] = useState(modalContent)
  const [formState, setFormState] = useState(formInitial)
  const [showModalInviteCodeConfirmation, setShowModalInviteCodeConfirmation] = useState(false)
  const [showModalActionConfirmation, setShowModalActionConfirmation] = useState(false)
  const [resendInvite, setResendInvite] = useState(false)
  const showSuccess = useToastStore((state) => state.showSuccess)
  const showError = useToastStore((state) => state.showError)
  const throwError = useErrorBoundary()

  const onClickReject = () => {
    setShowModalActionConfirmation(true)
  }

  const onClickModalConfirmation = (status) => {
    if (status) {
      rejectInvitation()
    } else {
      setShowModalActionConfirmation(false)
    }
  }

  const rejectInvitation = async () => {
    try {
      await rejectInvite(account.cloudAccountId, account.invitationState)
      showSuccess('Invite Declined!')
      setShowModalActionConfirmation(false)
    } catch (error) {
      if (error.response && error.response.data && error.response.data.message) {
        showError(error.response.data.message)
      } else {
        throwError(error.message)
      }
      setShowModalActionConfirmation(false)
    }
  }

  const confirmInviteCode = async () => {
    const inviteCode = formState.form.invitationCode.value
    if (inviteCode !== '') {
      try {
        await confirmInvite(account.cloudAccountId, inviteCode)
        showSuccess('Invitation Confirmed!')
        setShowModalInviteCodeConfirmation(false)
      } catch (error) {
        if (error.response && error.response.data && error.response.data.code === 13) {
          const updatedState = {
            ...formState
          }
          updatedState.form.invitationCode.isValid = false
          updatedState.form.invitationCode.validationMessage = 'Incorrect code'
          setFormState(updatedState)
        } else {
          throwError(error)
          setShowModalInviteCodeConfirmation(false)
        }
      }
    }
  }

  const onChangeInput = (event, formInputName) => {
    const updatedState = {
      ...formState
    }
    updatedState.form = UpdateFormHelper(event.target.value, formInputName, updatedState.form)
    setFormState(updatedState)
  }

  const onClickAcceptInvite = async () => {
    // Call api to send code
    try {
      setShowModalInviteCodeConfirmation(true)
      await CloudAccountService.memberNotification(account.cloudAccountId)
    } catch (error) {
      const updatedState = {
        ...formState
      }
      updatedState.form.invitationCode.isValid = false
      updatedState.form.invitationCode.isTouched = true
      updatedState.form.invitationCode.validationMessage = 'There was an error sending the invitation code'
      setFormState(updatedState)
    }
  }

  const sendCodeToMemberEmail = async () => {
    // Call api to resend code
    try {
      const updatedState = {
        ...formState
      }
      updatedState.form.invitationCode.isTouched = false
      updatedState.form.invitationCode.validationMessage = ''
      setFormState(updatedState)
      setResendInvite(true)
      await CloudAccountService.memberNotification(account.cloudAccountId)
      setResendInvite(false)
    } catch (error) {
      const updatedState = {
        ...formState
      }
      updatedState.form.invitationCode.isValid = false
      updatedState.form.invitationCode.isTouched = true
      updatedState.form.invitationCode.validationMessage = 'There was an error sending the invitation code'
      setFormState(updatedState)
      setResendInvite(false)
    }
  }

  return (
    <>
      <ActionConfirmation
        actionModalContent={actionModalContent}
        onClickModalConfirmation={onClickModalConfirmation}
        showModalActionConfirmation={showModalActionConfirmation}
      />
      <InvitationCode
        isModalOpen={showModalInviteCodeConfirmation}
        formState={formState}
        account={account}
        resendInvite={resendInvite}
        onChange={onChangeInput}
        sendCodeToMemberEmail={sendCodeToMemberEmail}
      />
      <AccountsAction onClickAccept={onClickAcceptInvite} onClickReject={onClickReject} />
    </>
  )
}

export default AccountsActionContainer
