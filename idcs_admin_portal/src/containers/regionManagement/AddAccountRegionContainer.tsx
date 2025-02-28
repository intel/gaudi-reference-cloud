// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  setFormValue,
  setSelectOptions
} from '../../utility/updateFormHelper/UpdateFormHelper'
import CloudAccountService from '../../services/CloudAccountService'
import RegionManagementService from '../../services/RegionManagementService'
import useToastStore from '../../store/toastStore/ToastStore'
import useRegionStore from '../../store/regionStore/RegionStore'
import AddAccountRegion from '../../components/regionManagement/AddAccountRegion'
import useUserStore from '../../store/userStore/UserStore'

const AddAccountRegionContainer = (): JSX.Element => {
  // Initial state for form and validation
  const initialState = {
    mainSubtitle: 'Specify the needed Information',
    form: {
      region: {
        sectionGroup: 'region',
        type: 'select', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Controlled Region:',
        placeholder: 'Please select',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: ''
      },
      cloudAccount: {
        sectionGroup: 'cloudAccount',
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Cloud Account:',
        placeholder: 'Enter Cloud Account Email or ID',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '' // Errror message to display to the user
      }
    },
    isValidForm: false,
    timeoutMiliseconds: 1000
  }

  // Initial state for loader
  const initialLoaderData = {
    isShow: false,
    message: ''
  }

  // Error Boundry
  const throwError = useErrorBoundary()

  // Navigation
  const navigate = useNavigate()

  const [state, setState] = useState(initialState)
  const [showLoader, setShowLoader] = useState<any>(initialLoaderData)
  const [selectedCloudAccount, setSelectedCloudAccount] = useState<any>('')

  // Global States)
  const showError = useToastStore((state) => state.showError)
  const regions = useRegionStore((state) => state.regions)
  const setRegions = useRegionStore((state) => state.setRegions)
  const user: any = useUserStore((state) => state.user)

  const backButtonLabel = '⟵ Back to Region Whitelist Accounts'

  // Hooks
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        await setRegions({ type: 'controlled' })
      } catch (error) {
        throwError(error)
      }
    }

    fetch().catch(() => {})
  }, [])

  useEffect(() => {
    setForm()
  }, [regions])

  // functions
  const getRegionOptions = (): any[] => {
    return regions.map((region) => {
      return {
        name: region.name,
        value: region.name
      }
    })
  }
  function setForm(): void {
    const stateUpdated = {
      ...state
    }

    // Load families
    const selectableRegions = getRegionOptions()
    if (selectableRegions.length > 0) {
      stateUpdated.form = setSelectOptions('region', selectableRegions, stateUpdated.form)
      stateUpdated.form = setFormValue('region', selectableRegions[0].value, stateUpdated.form)
    }

    setState(stateUpdated)
  }

  function onChangeInput(event: any, formInputName: string): void {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)
    updatedState.form = updatedForm

    if (formInputName === 'cloudAccount') {
      setSelectedCloudAccount(null)
      updatedState.isValidForm = false
    }

    if (!selectedCloudAccount) {
      updatedState.isValidForm = false
    }

    setState(updatedState)
  }

  async function onSearchCloudAccount(): Promise<void> {
    const cloudAccount = getFormValue('cloudAccount', state.form)
    setCloudAccountError('')
    setSelectedCloudAccount(null)
    if (cloudAccount !== '') {
      try {
        let data: any
        if (cloudAccount.includes('@') || /[a-zA-Z]/.test(cloudAccount)) {
          data = await CloudAccountService.getCloudAccountDetailsByName(cloudAccount)
        } else {
          data = await CloudAccountService.getCloudAccountDetailsById(cloudAccount)
        }

        setSelectedCloudAccount(data?.data)
        const updatedState = { ...state }
        updatedState.form.cloudAccount.validationMessage = ''
        updatedState.isValidForm = isValidForm(updatedState.form)
        setState(updatedState)
      } catch (e: any) {
        const code = e.response.data?.code
        const errorMsg = e.response.data?.message
        const message =
          code && [3, 5].includes(code)
            ? String(errorMsg.charAt(0).toUpperCase()) + String(errorMsg.slice(1))
            : 'Cloud Account ID is not found'
        setCloudAccountError(message)
        setSelectedCloudAccount(null)
      }
    } else {
      setCloudAccountError('Cloud Account Number is required')
    }
  }

  function setCloudAccountError(errorMessage: string): void {
    const updatedState = { ...state }
    const updatedForm = { ...updatedState.form }
    const updatedFormElement = { ...updatedForm.cloudAccount }
    updatedFormElement.isValid = false
    updatedFormElement.validationMessage = errorMessage
    updatedForm.cloudAccount = updatedFormElement
    updatedState.form = updatedForm
    updatedState.isValidForm = false
    setState(updatedState)
  }

  function onCancel(): void {
    navigate('/regionmanagement/whitelist')
  }

  async function onSubmit(): Promise<void> {
    setShowLoader({ isShow: true, message: 'Working on your request' })
    await submitForm()
  }

  async function submitForm(): Promise<void> {
    const region = getFormValue('region', state.form)
    try {
      const payload = {
        cloudaccountId: selectedCloudAccount.id,
        regionName: region,
        adminName: user.email,
        created: new Date()
      }

      await postAcl(payload)
    } catch (error: any) {
      let message = ''
      if (error.response) {
        if (error.response.data.message !== '') {
          message = error.response.data.message
        } else {
          message = error.message
        }
      } else {
        message = error.message
      }
      setShowLoader({ isShow: false })
      showError(message, false)
    }
  }

  async function postAcl(payload: any): Promise<void> {
    await RegionManagementService.postAcl(payload)

    setTimeout(() => {
      navigate('/regionmanagement/whitelist')
    }, state.timeoutMiliseconds)
  }

  return (
    <AddAccountRegion
      state={state}
      showLoader={showLoader}
      backButtonLabel={backButtonLabel}
      onCancel={onCancel}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      selectedCloudAccount={selectedCloudAccount}
      onSearchCloudAccount={onSearchCloudAccount}
    />
  )
}

export default AddAccountRegionContainer
