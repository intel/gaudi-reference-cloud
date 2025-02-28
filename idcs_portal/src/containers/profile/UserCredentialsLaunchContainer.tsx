// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState } from 'react'
import { useNavigate } from 'react-router'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  showFormRequiredFields
} from '../../utils/updateFormHelper/UpdateFormHelper'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'
import UserCredentialsLaunch from '../../components/profile/userCredentials/UserCredentialsLaunch'
import CloudAccountService from '../../services/CloudAccountService'

const UserCredentialsLaunchContainer = (): JSX.Element => {
  // *****
  // Global state
  // *****
  const showError = useToastStore((state) => state.showError)

  // *****
  // local state
  // *****
  const initialState = {
    mainTitle: 'Generate Secret',
    form: {
      credentialName: {
        type: 'text', // options = 'text ,'textArea'
        label: 'Secret name:',
        placeholder: 'Secret name',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        validationRules: {
          isRequired: true,
          onlyAlphaNumLower: true,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: (
          <div className="valid-feedback" intc-id={'CredentialsNameValidMessage'}>
            Name must be 63 characters or less, and can include letters, numbers, and ‘-‘ only.
            <br />
            It should start and end with an alphanumeric character.
          </div>
        )
      }
    },
    isValidForm: false,
    servicePayload: {
      appClientName: ''
    },
    showTokenAsServiceModal: false,
    navigationBottom: [
      {
        buttonAction: 'Submit',
        buttonLabel: 'Generate Secret',
        buttonVariant: 'primary'
      },
      {
        buttonAction: 'Cancel',
        buttonLabel: 'Cancel',
        buttonVariant: 'link',
        buttonFunction: () => {
          onCancel()
        }
      }
    ]
  }

  const accessTokenModalInitial = {
    show: false,
    isTokenReady: false,
    command: <>Your personal client secret that grants you access to our API endpoints.</>,
    token: ''
  }

  const [state, setState] = useState(initialState)
  const [accessTokenModal, setAccessTokenModal] = useState(accessTokenModalInitial)

  // *****
  // Navigation
  // *****
  const navigate = useNavigate()

  // *****
  // Functions
  // *****
  const onChangeInput = (event: any, formInputName: string): void => {
    const value = event.target.value

    const updatedState = { ...state }

    const updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  const onCancel = (): void => {
    // Navigates back to the page when this method triggers.
    navigate('/profile/credentials')
  }

  const showRequiredFields = (): void => {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(stateCopy.form)
    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    stateCopy.form = updatedForm
    setState(stateCopy)
  }

  const onSubmit = async (): Promise<void> => {
    try {
      const isValidForm = state.isValidForm
      if (!isValidForm) {
        showRequiredFields()
        return
      }
      const credentialName = getFormValue('credentialName', state.form)
      const servicePayloadCopy = state.servicePayload
      servicePayloadCopy.appClientName = credentialName
      onAccessTokenModal(true, false)
      const response = await CloudAccountService.postUserCredentials(servicePayloadCopy)
      const data = { ...response.data }
      const { clientSecret } = data
      onAccessTokenModal(true, true, clientSecret)
    } catch (error: any) {
      const message = String(error.message)
      if (error.response) {
        const errData = error.response.data
        const errCode = errData.code
        const errMessage = errData.message
        if (errCode === 11) {
          showError('Client Secret limit reached.', false)
        } else {
          showError(errMessage, false)
        }
      } else {
        showError(message, false)
      }
      onAccessTokenModal(false, false, '')
    }
  }

  const onAccessTokenModal = (show: boolean, isTokenReady: boolean, token?: string): void => {
    setAccessTokenModal({ ...accessTokenModal, show, token: token ?? '', isTokenReady })
  }

  const onCloseModal = (): void => {
    setAccessTokenModal(accessTokenModal)
    navigate({
      pathname: '/profile/credentials'
    })
  }

  return (
    <UserCredentialsLaunch
      title={state.mainTitle}
      form={state.form}
      onChangeInput={onChangeInput}
      navigationBottom={state.navigationBottom}
      onSubmit={onSubmit}
      accessTokenModal={accessTokenModal}
      onCloseModal={onCloseModal}
    />
  )
}

export default UserCredentialsLaunchContainer
