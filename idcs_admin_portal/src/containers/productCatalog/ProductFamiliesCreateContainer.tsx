// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import ProductFamiliesForm from '../../components/productCatalog/ProductFamiliesForm'
import { useNavigate } from 'react-router-dom'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  showFormRequiredFields,
  setSelectOptions,
  setFormValue
} from '../../utility/updateFormHelper/UpdateFormHelper'
import useToastStore from '../../store/toastStore/ToastStore'
import PublicService from '../../services/PublicService'
import { toastMessageEnum } from '../../utility/Enums'
import useVendorStore from '../../store/vendorStore/VendorStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'

const ProductFamiliesCreateContainer = (): JSX.Element => {
  // Navigation
  const navigate = useNavigate()

  // initial state
  const initialState = {
    desciption: 'Create New Family',
    form: {
      name: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Family Name',
        placeholder: 'Family Name',
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
        label: 'Family Description',
        placeholder: 'Family Description',
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
      vendorName: {
        type: 'dropdown', // options = 'text ,'textArea'
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
          checkMaxLength: true
        },
        options: [],
        validationMessage: '',
        helperMessage: ''
      }
    },
    isValidForm: false,
    navigationTop: [
      {
        label: '⟵ Back to Home',
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
          onCancel('/products/families')
        }
      }
    ]
  }

  const [state, setState] = useState(initialState)
  const [showModal, setShowModal] = useState(false)

  // Global Store
  const showError = useToastStore((state) => state.showError)
  const vendors = useVendorStore((state) => state.vendors)
  const getVendor = useVendorStore((state) => state.getVendor)
  const loading = useVendorStore((state) => state.loading)

  // Error Boundry
  const throwError = useErrorBoundary()

  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        if (!vendors) await getVendor()
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    setForm()
  }, [vendors])

  // functions
  const setForm = (): void => {
    const stateCopy = { ...state }
    const options: any = []
    vendors?.forEach((item) => {
        options.push({
          name: item.name,
          value: item.name
        })
      })
    let form = setSelectOptions('vendorName', options, stateCopy.form)
    if (options.length > 1) {
      form = setFormValue('vendorName', options[0].value, form)
    }
    stateCopy.form = form
    setState(stateCopy)
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
    const vendorName = getFormValue('vendorName', updateForm)

    const payload = {
      name,
      description,
      vendorName
    }

    try {
      setShowModal(true)
      await PublicService.createCatalogFamily(payload)
      setShowModal(false)
      navigate('/products/families')
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

  return (
    <ProductFamiliesForm
      state={state}
      showModal={showModal}
      loading={loading}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
    />
  )
}

export default ProductFamiliesCreateContainer
