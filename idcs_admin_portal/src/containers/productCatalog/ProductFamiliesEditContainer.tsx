// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import ProductFamiliesForm from '../../components/productCatalog/ProductFamiliesForm'
import { useNavigate, useParams } from 'react-router-dom'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  setFormValue,
  showFormRequiredFields,
  setSelectOptions
} from '../../utility/updateFormHelper/UpdateFormHelper'
import useToastStore from '../../store/toastStore/ToastStore'
import PublicService from '../../services/PublicService'
import useVendorStore from '../../store/vendorStore/VendorStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { toastMessageEnum } from '../../utility/Enums'
import useFamilyStore from '../../store/familyStore/FamilyStore'

const ProductFamiliesEditContainer = (): JSX.Element => {
  // Navigation and params
  const navigate = useNavigate()
  const { param: name } = useParams()

  const throwError = useErrorBoundary()

  // initial state
  const initialState = {
    desciption: 'Edit Family',
    form: {
      name: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Family Name',
        placeholder: 'Family Name',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: true,
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
          onCancel('/products/families')
        }
      }
    ]
  }
  // Global Store
  const showError = useToastStore((state) => state.showError)
  const vendors = useVendorStore((state) => state.vendors)
  const getVendor = useVendorStore((state) => state.getVendor)
  const loadingVendors = useVendorStore((state) => state.loading)
  const families = useFamilyStore((state) => state.families)
  const getFamilies = useFamilyStore((state) => state.getFamilies)
  const loadingFamilies = useFamilyStore((state) => state.loading)

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
        const calls = []
        if (!vendors) calls.push(getVendor())
        if (!families) calls.push(getFamilies())
        await Promise.all(calls)
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
  }, [vendors, families, isPageReady])

  const updateDetails = (): void => {
    const family = families?.find((e: any) => e.name === name)
    if (!family && isPageReady) navigate('/products/families')
    const updatedState = { ...state }
    updatedState.form = setFormValue('name', family?.name, updatedState.form)
    updatedState.form = setFormValue('description', family?.description, updatedState.form)
    updatedState.form = setFormValue('vendorName', family?.vendor, updatedState.form)
    const options: any = []
    vendors?.forEach((item) => {
        options.push({
            name: item.name,
            value: item.name
        })
    })
    updatedState.form = setSelectOptions('vendorName', options, updatedState.form)
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
    const vendorName = getFormValue('vendorName', updateForm)

    const payload = {
      description,
      vendorName
    }

    try {
      setShowModal(true)
      await PublicService.putCatalogFamily(name, payload)
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

  const showRequiredFields = (): void => {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(stateCopy.form)

    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    stateCopy.form = updatedForm
    setState(stateCopy)
  }

  return <ProductFamiliesForm state={state} showModal={showModal} onSubmit={onSubmit} onChangeInput={onChangeInput} loading={loadingVendors || loadingFamilies}/>
}

export default ProductFamiliesEditContainer
