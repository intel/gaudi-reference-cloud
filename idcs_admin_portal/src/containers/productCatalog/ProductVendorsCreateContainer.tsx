// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState } from 'react'
import ProductVendorsCreate from '../../components/productCatalog/ProductVendorsCreate'
import { useNavigate } from 'react-router-dom'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  showFormRequiredFields
} from '../../utility/updateFormHelper/UpdateFormHelper'
import useToastStore from '../../store/toastStore/ToastStore'
import PublicService from '../../services/PublicService'
import { toastMessageEnum } from '../../utility/Enums'

const ProductVendorsCreateContainer = (): JSX.Element => {
  // Navigation
  const navigate = useNavigate()

  // initial state
  const initialState = {
    desciption: 'Create New Vendor',
    form: {
      name: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Vendor Name',
        placeholder: 'Vendor Name',
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
      description: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Vendor Description',
        placeholder: 'Vendor Description',
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
      organizationName: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Vendor Organization Name',
        placeholder: 'Vendor Organization Name',
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
      }
    },
    isValidForm: false,
    navigationTop: [
      {
        label: '⟵ Back to home',
        buttonVariant: 'link',
        function: () => {
          onCancel('/')
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
          onCancel('/products/vendors')
        }
      }
    ]
  }

  const [state, setState] = useState(initialState)
  const [showModal, setShowModal] = useState(false)

  // Global Store
  const showError = useToastStore((state) => state.showError)

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
    const description = getFormValue('description', updateForm)
    const organizationName = getFormValue('organizationName', updateForm)

    const payload = {
      name,
      description,
      organizationName
    }

    try {
      setShowModal(true)
      await PublicService.createVendorCatalog(payload)
      setShowModal(false)
      navigate('/products/vendors')
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

    const inputValue = event.target.value
    const updatedForm = UpdateFormHelper(inputValue, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  return <ProductVendorsCreate state={state} showModal={showModal} onSubmit={onSubmit} onChangeInput={onChangeInput} />
}

export default ProductVendorsCreateContainer
