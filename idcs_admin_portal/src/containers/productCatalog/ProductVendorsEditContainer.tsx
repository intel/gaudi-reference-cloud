// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import ProductVendorsCreate from '../../components/productCatalog/ProductVendorsCreate'
import { useNavigate, useParams } from 'react-router-dom'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  setFormValue,
  showFormRequiredFields
} from '../../utility/updateFormHelper/UpdateFormHelper'
import useToastStore from '../../store/toastStore/ToastStore'
import PublicService from '../../services/PublicService'
import useVendorStore from '../../store/vendorStore/VendorStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { toastMessageEnum } from '../../utility/Enums'

const ProductVendorsEditContainer = (): JSX.Element => {
  // Navigation and params
  const navigate = useNavigate()
  const { param: name } = useParams()

  const throwError = useErrorBoundary()

  // initial state
  const initialState = {
    desciption: 'Edit Vendor',
    form: {
      name: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Vendor Name',
        placeholder: 'Vendor Name',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: true,
        maxLength: 70,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
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
          checkMaxLength: true
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
        label: 'Save',
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
  // Global Store
  const showError = useToastStore((state) => state.showError)
  const vendors = useVendorStore((state) => state.vendors)
  const getVendor = useVendorStore((state) => state.getVendor)

  // State
  const [state, setState] = useState(initialState)
  const [showModal, setShowModal] = useState(false)
  const [isPageReady, setIsPageReady] = useState(false)

  // *****
  // Hooks
  // *****
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        if (!vendors) await getVendor()
        setIsPageReady(true)
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    updateDetails()
  }, [vendors, isPageReady])

  const updateDetails = (): void => {
    const vendor = vendors?.find((e: any) => e.name === name)
    if (!vendor && isPageReady) navigate('/products/vendors')
    const updatedState = { ...state }
    updatedState.form = setFormValue('name', vendor?.name, updatedState.form)
    updatedState.form = setFormValue('description', vendor?.description, updatedState.form)
    updatedState.form = setFormValue('organizationName', vendor?.organizationName, updatedState.form)
    updatedState.isValidForm = isValidForm(updatedState.form)
    setState(updatedState)
  }

  // functions
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
      description,
      organizationName
    }

    try {
      setShowModal(true)
      await PublicService.putVendorCatalog(name, payload)
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

  const showRequiredFields = (): void => {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(stateCopy.form)

    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    stateCopy.form = updatedForm
    setState(stateCopy)
  }

  return <ProductVendorsCreate state={state} showModal={showModal} onSubmit={onSubmit} onChangeInput={onChangeInput} />
}

export default ProductVendorsEditContainer
