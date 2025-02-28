// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import QuotaManagementServiceCreate from '../../components/storageManagement/quotaManagementService/QuotaManagementServiceCreate'
import { useNavigate, useParams } from 'react-router'
import { UpdateFormHelper, isValidForm, showFormRequiredFields, getFormValue, setFormValue } from '../../utility/updateFormHelper/UpdateFormHelper'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utility/Enums'
import idcConfig from '../../config/configurator'
import useStorageManagementStore from '../../store/storageManagementStore/StorageManagementStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import StorageManagementService from '../../services/StorageManagementService'

const QuotaManagementServiceEditContainer = (): JSX.Element => {
  // *****
  // Params
  // *****
  const { param } = useParams()

  // *****
  // Global state
  // *****
  const showError = useToastStore((state) => state.showError)
  const serviceResourceItem = useStorageManagementStore((state) => state.serviceResourceItem)
  const getServicesById = useStorageManagementStore((state) => state.getServicesById)
  const throwError = useErrorBoundary()

  // *****
  // Hooks
  // *****
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        if (param) {
          await getServicesById(param)
        }
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    updateFormValues()
  }, [serviceResourceItem])

  // *****
  // local state
  // *****
  const navigate = useNavigate()
  const [isPageReady, setIsPageReady] = useState(false)

  const serviceResourceInput = {
    name: {
      type: 'text', // options = 'text ,'textArea'
      label: 'Resource name',
      placeholder: 'Resource name: ',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 50,
      validationRules: {
        isRequired: true,
        onlyAlphaNumLower: true,
        checkMaxLength: true
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: (
        <>
          Name must be 63 characters or less
        </>
      )
    },
    maxLimit: {
      type: 'integer', // options = 'text ,'textArea'
      label: 'Max limit:',
      placeholder: '',
      value: '',
      isValid: false,
      isTouched: false,
      isReadOnly: false,
      maxLength: 6,
      validationRules: {
        isRequired: true,
        checkMaxValue: 10000,
        checkMinValue: 0
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: <>
        Enter a value between 0 and 10,000
      </>
    },
    quotaUnit: {
      type: 'dropdown', // options = 'text ,'textArea'
      label: 'Quota unit:',
      placeholder: 'Please select',
      value: '',
      isValid: false,
      isTouched: false,
      isReadOnly: false,
      validationRules: {
        isRequired: true
      },
      options: [
        {
          name: 'Count',
          value: 'COUNT'
        },
        {
          name: 'Request by second',
          value: 'REQ_SEC'
        },
        {
          name: 'Request by minute',
          value: 'REQ_MIN'
        }, {
          name: 'Request by hour',
          value: 'REQ_HOUR'
        }
      ],
      validationMessage: '',
      helperMessage: <>
        Choose the desired unit for your resource
      </>
    }
  }

  const initialState = {
    mainTitle: 'Edit service',
    form: {
      serviceName: {
        type: 'text', // options = 'text ,'textArea'
        label: 'Service name:',
        placeholder: 'Service name',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: true, // Input create as read only
        maxLength: 63,
        validationRules: {
          isRequired: true,
          onlyAlphaNumLower: true,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: <>Name must be 63 characters or less, and can include letters, numbers, and ‘-‘ only.</>
      },
      serviceResources: {
        label: 'Service Resource:',
        items: [{ ...serviceResourceInput }],
        isValid: false,
        validationRules: {
          isRequired: true
        }
      }
    },
    isValidForm: false,
    servicePayload: {
      region: '',
      serviceName: '',
      serviceResources: [
      ]
    },
    navigationBottom: [
      {
        buttonAction: 'Submit',
        buttonLabel: 'Save',
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

  const resourceItemPayload = {
    name: '',
    quotaUnit: '',
    maxLimit: ''
  }

  const submitModalInitial = {
    show: false,
    message: 'Updating service'
  }

  const emptyViewInitial = {
    show: false
  }

  const [state, setState] = useState(initialState)
  const [submitModal, setSubmitModal] = useState(submitModalInitial)

  // *****
  // functions
  // *****
  const updateFormValues = (): void => {
    if (serviceResourceItem) {
      const stateUpdated = { ...state }
      let formUpdated = state.form
      const nameValue = serviceResourceItem.serviceName
      const serviceResources = serviceResourceItem.serviceResources
      const serviceResourcesUpdated = { ...formUpdated.serviceResources }

      for (let index = 0; index < serviceResources.length; index++) {
        if (serviceResources.length > serviceResourcesUpdated.items.length) {
          serviceResourcesUpdated.items.push({
            name: { ...serviceResourceInput.name },
            maxLimit: { ...serviceResourceInput.maxLimit },
            quotaUnit: { ...serviceResourceInput.quotaUnit }
          })
        }
        serviceResourcesUpdated.items[index].name.value = serviceResources[index].name.trim()
        serviceResourcesUpdated.items[index].name.isValid = true
        serviceResourcesUpdated.items[index].maxLimit.value = serviceResources[index].maxLimit.trim()
        serviceResourcesUpdated.items[index].maxLimit.isValid = true
        serviceResourcesUpdated.items[index].quotaUnit.value = serviceResources[index].quotaUnit.trim()
        serviceResourcesUpdated.items[index].quotaUnit.isValid = true
      }

      serviceResourcesUpdated.isValid = validateRows(serviceResourcesUpdated)
      formUpdated.serviceResources = serviceResourcesUpdated
      formUpdated = setFormValue('serviceName', nameValue, formUpdated)
      stateUpdated.isValidForm = isValidForm(formUpdated)
      stateUpdated.form = formUpdated
      setState(stateUpdated)
      setIsPageReady(true)
    }
  }

  const onChangeInput = (event: any, formInputName: string, idParent: string = '', index: number): void => {
    const value = event.target.value
    const updatedState = {
      ...state
    }

    let updatedForm = updatedState.form

    if (idParent === 'serviceResources') {
      const serviceResources = updatedForm.serviceResources
      const serviceResourcesItems = [...serviceResources.items]
      const serviceResourceItem = serviceResourcesItems[index]
      const updatedserviceResourceItem = UpdateFormHelper(value, formInputName, serviceResourceItem)
      serviceResourcesItems[index] = updatedserviceResourceItem
      updatedForm.serviceResources.items = serviceResourcesItems
      // // Validate rows
      updatedForm.serviceResources.isValid = validateRows(serviceResources)
    } else {
      updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)
    }

    updatedState.form = updatedForm
    updatedState.isValidForm = isValidForm(updatedForm)

    setState(updatedState)
  }

  const validateRows = (serviceResources: any): boolean => {
    let isValidArray = true
    for (const index in serviceResources.items) {
      const computeItem = { ...serviceResources.items[index] }
      const isValidRow = isValidForm(computeItem)
      if (!isValidRow) {
        isValidArray = false
        break
      }
    }
    return isValidArray
  }

  const onClickActionResourceItem = (index: number, action: string): void => {
    const updatedState = {
      ...state
    }
    const form = state.form
    const serviceResourcesUpdated = { ...form.serviceResources }
    const itemsCopy = [...serviceResourcesUpdated.items]
    switch (action) {
      case 'Delete':
        serviceResourcesUpdated.items.splice(index, 1)
        break
      default: {
        const newSourceIp = { ...serviceResourceInput }
        itemsCopy.push(newSourceIp)
        serviceResourcesUpdated.items = itemsCopy
        break
      }
    }
    serviceResourcesUpdated.isValid = validateRows(serviceResourcesUpdated)
    form.serviceResources = serviceResourcesUpdated
    updatedState.isValidForm = isValidForm(form)
    updatedState.form = form
    setState(updatedState)
  }

  const showRequiredFields = (): void => {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(stateCopy.form)
    // Mark array Inputs
    const serviceResources = updatedForm.serviceResources
    const items = serviceResources.items
    const itemsUpdated = []
    for (const index in items) {
      const item = { ...items[index] }
      const updatedItem = showFormRequiredFields(item)
      itemsUpdated.push(updatedItem)
    }
    serviceResources.items = itemsUpdated
    updatedForm.serviceResources = serviceResources

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
      setSubmitModal({ ...submitModal, show: true })
      const payload = { ...state.servicePayload }
      payload.serviceName = getFormValue('serviceName', state.form)
      const serviceResourceArray: any = []
      const serviceResourceItems = { ...state.form.serviceResources.items }
      for (const index in serviceResourceItems) {
        const serviceResource = { ...serviceResourceItems[index] }
        const resourceItemPayloadCopy = { ...resourceItemPayload }
        resourceItemPayloadCopy.maxLimit = serviceResource.maxLimit.value
        resourceItemPayloadCopy.name = serviceResource.name.value
        resourceItemPayloadCopy.quotaUnit = serviceResource.quotaUnit.value
        serviceResourceArray.push(resourceItemPayloadCopy)
      }
      payload.serviceResources = serviceResourceArray
      payload.region = idcConfig.REACT_APP_SELECTED_REGION

      await StorageManagementService.putService(param, payload)
      setSubmitModal({ ...submitModal, show: false })
      onCancel()
    } catch (error: any) {
      const message = String(error.message)
      if (error.response) {
        const errData = error.response.data
        const errMessage = errData.message
        showError(errMessage, false)
      } else {
        showError(message, false)
      }
      setSubmitModal({ ...submitModal, show: false })
    }
  }

  const onCancel = (): void => {
    navigate('/quotamanagement/services')
  }
  return <>
    <QuotaManagementServiceCreate onCancel={onCancel} onSubmit={onSubmit} state={state} onClickActionResourceItem={onClickActionResourceItem} onChangeInput={onChangeInput} serviceResourceLimit={100} submitModal={submitModal} moduleName={'service'} isPageReady={isPageReady} emptyView={emptyViewInitial} />
  </>
}

export default QuotaManagementServiceEditContainer
