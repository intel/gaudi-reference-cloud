// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState } from 'react'
import RegionCreate from '../../components/regionManagement/RegionCreate'
import { useNavigate } from 'react-router-dom'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  showFormRequiredFields
} from '../../utility/updateFormHelper/UpdateFormHelper'
import useToastStore from '../../store/toastStore/ToastStore'
import RegionManagementService from '../../services/RegionManagementService'
import { toastMessageEnum } from '../../utility/Enums'
import useUserStore from '../../store/userStore/UserStore'

const RegionCreateContainer = (): JSX.Element => {
  // Navigation
  const navigate = useNavigate()

  // initial state
  const initialState = {
    desciption: 'Create New Region',
    form: {
      name: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Region name',
        placeholder: 'Region name',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        maxLength: 70,
        validationRules: {
          isRequired: true,
          checkMaxLength: true,
          onlyAlphaNumLower: true
        },
        validationMessage: '',
        helperMessage: ''
      },
      friendlyName: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Region friendly name',
        placeholder: 'Region friendly name',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        maxLength: 200,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '',
        helperMessage: ''
      },
      type: {
        type: 'dropdown', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Region type',
        placeholder: 'Please select region type',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [
          { name: 'open', value: 'open' },
          { name: 'controlled', value: 'controlled' }
        ],
        validationMessage: '',
        helperMessage: ''
      },
      subnet: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Region subnet',
        placeholder: 'Region subnet',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        maxLength: 70,
        validationRules: {
          isRequired: true,
          checkMaxLength: true,
          onlyAlphaNumLower: true
        },
        validationMessage: '',
        helperMessage: ''
      },
      availabilityZone: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Region availability zone',
        placeholder: 'Region availability zone',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        maxLength: 70,
        validationRules: {
          isRequired: true,
          checkMaxLength: true,
          onlyAlphaNumLower: true
        },
        validationMessage: '',
        helperMessage: ''
      },
      prefix: {
        type: 'integer', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Region prefix',
        placeholder: '24',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 70,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        }
      },
      defaultRegion: {
        section: 'banner-link',
        type: 'checkbox', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Default region',
        hiddenLabel: true,
        placeholder: '',
        value: false,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        },
        options: [
          {
            name: 'Set as Default Region',
            value: '1'
          }
        ],
        validationMessage: '',
        helperMessage: (
          <div className="valid-feedback" intc-id={'BannerCreationAdminIsMaintenanceValidMessage'}>
            Only &apos;open&apos; regions can be set as the default region, and only one default region is allowed at
            any given time.
          </div>
        )
      },
      apiDns: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'API DNS',
        placeholder: 'API DNS',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        maxLength: 256,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '',
        helperMessage: ''
      }
    },
    isValidForm: false,
    navigationTop: [
      {
        label: '⟵ Back to regions',
        buttonVariant: 'link',
        function: () => {
          onCancel('/regionmanagement/regions')
        }
      }
    ],
    navigationBottom: [
      {
        buttonAction: 'Submit',
        label: 'Create',
        buttonVariant: 'primary'
      },
      {
        buttonAction: 'Cancel',
        label: 'Cancel',
        buttonVariant: 'link',
        function: () => {
          onCancel('/regionmanagement/regions')
        }
      }
    ]
  }

  const [state, setState] = useState(initialState)
  const [showModal, setShowModal] = useState(false)

  // Global Store
  const showError = useToastStore((state) => state.showError)
  const user = useUserStore((state) => state.user)

  // functions
  const showRequiredFields = (): void => {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(stateCopy.form)

    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    stateCopy.form = updatedForm
    setState(stateCopy)
  }

  async function onSubmit(): Promise<void> {
    const isValidForm = state.isValidForm
    if (!isValidForm) {
      showRequiredFields()
      return
    }

    const updatedState = {
      ...state
    }
    const updateForm = { ...updatedState.form }

    const name = getFormValue('name', updateForm)
    const friendlyName = getFormValue('friendlyName', updateForm)
    const type = getFormValue('type', updateForm)
    const subnet = getFormValue('subnet', updateForm)
    const prefix = parseInt(getFormValue('prefix', updateForm))
    const defaultRegion = getFormValue('defaultRegion', updateForm)
    const apiDns = getFormValue('apiDns', updateForm)
    const availabilityZone = getFormValue('availabilityZone', updateForm)
    const adminName = user?.email

    const payload = {
      name,
      apiDns,
      availabilityZone,
      friendlyName,
      isDefault: defaultRegion,
      type,
      subnet,
      prefix,
      adminName
    }

    try {
      setShowModal(true)
      await RegionManagementService.createRegion(payload)
      setShowModal(false)
      navigate('/regionmanagement/regions')
    } catch (error: any) {
      setShowModal(false)
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
      showError(message, false)
    }
  }

  function onCancel(location: string): void {
    navigate(location)
  }

  function onChangeInput(event: any, formInputName: any): void {
    const updatedState = {
      ...state
    }

    let inputValue = event.target.value
    if (formInputName === 'defaultRegion') {
      inputValue = event.target.checked
    }
    const updatedForm = UpdateFormHelper(inputValue, formInputName, updatedState.form)

    if (formInputName === 'type') {
      updatedForm.defaultRegion.value = false
      updatedForm.defaultRegion.isReadOnly = inputValue === 'controlled'
    }

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  return <RegionCreate state={state} showModal={showModal} onSubmit={onSubmit} onChangeInput={onChangeInput} />
}

export default RegionCreateContainer
