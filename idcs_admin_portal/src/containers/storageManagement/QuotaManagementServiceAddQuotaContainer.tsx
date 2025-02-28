// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import QuotaManagementServiceCreate from '../../components/storageManagement/quotaManagementService/QuotaManagementServiceCreate'
import { useNavigate, useParams } from 'react-router'
import { UpdateFormHelper, isValidForm, showFormRequiredFields, getFormValue, setSelectOptions, markErrorOnElement } from '../../utility/updateFormHelper/UpdateFormHelper'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utility/Enums'
import StorageManagementService from '../../services/StorageManagementService'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useStorageManagementStore from '../../store/storageManagementStore/StorageManagementStore'
import CloudAccountService from '../../services/CloudAccountService'

const QuotaManagementServiceAddQuotaContainer = (): JSX.Element => {
  // *****
  // Params
  // *****
  const { param } = useParams()
  // *****
  // Global state
  // *****
  const serviceResourceItem = useStorageManagementStore((state) => state.serviceResourceItem)
  const getServicesById = useStorageManagementStore((state) => state.getServicesById)
  const showError = useToastStore((state) => state.showError)
  // *****
  // local state
  // *****
  const navigate = useNavigate()

  const initialState = {
    mainTitle: 'Add new quota',
    form: {
      resourceName: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Resource:',
        placeholder: 'Please select',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: '',
        helperMessage: <>
          Choose the desired unit for your quota
        </>
      },
      reason: {
        type: 'textArea', // options = 'text ,'textArea'
        label: 'Reason:',
        placeholder: 'Reason',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 200,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: <>Plese provide a reason.</>
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
          Choose the desired unit for your quota
        </>
      },
      scopeType: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Scope:',
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
            name: 'Account type',
            value: 'QUOTA_ACCOUNT_TYPE'
          },
          {
            name: 'Account Id',
            value: 'QUOTA_ACCOUNT_ID'
          }
        ],
        validationMessage: '',
        helperMessage: <>
          Choose the desired scoupe for your quota
        </>
      },
      userType: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'User type:',
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
            name: 'PREMIUM',
            value: 'PREMIUM'
          },
          {
            name: 'STANDARD',
            value: 'STANDARD'
          },
          {
            name: 'INTEL',
            value: 'INTEL'
          }
        ],
        validationMessage: '',
        helperMessage: <>
          Choose the desired scoupe for your quota
        </>,
        hidden: true
      },
      cloudAccount: {
        type: 'text', // options = 'text ,'textArea'
        label: 'Cloud Account:',
        placeholder: 'Cloud Account',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 200,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: <>Please provide a cloud account.</>,
        hidden: true
      }
    },
    isValidForm: false,
    servicePayload: {
      serviceQuotaResource: {
        resourceType: '',
        quotaConfig: {
          limits: '',
          quotaUnit: ''
        },
        scope: {
          scopeType: '',
          scopeValue: ''
        },
        reason: ''
      }
    },
    navigationBottom: [
      {
        buttonAction: 'Submit',
        buttonLabel: 'Create',
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

  const submitModalInitial = {
    show: false,
    message: 'Creating new service'
  }

  const searchModalInitial = {
    show: false,
    message: ''
  }

  const cloudAccountSelectedInitial = {
    show: false,
    name: '',
    id: '',
    type: ''
  }

  const emptyViewInitial = {
    show: false
  }

  const [state, setState] = useState(initialState)
  const [searchModal, setSearchModal] = useState(searchModalInitial)
  const [submitModal, setSubmitModal] = useState(submitModalInitial)
  const [cloudAccountSelected, setCloudAccountSelected] = useState(cloudAccountSelectedInitial)
  const [quotaCloudType, setQuotaCloudType] = useState(false)
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
    updateForm()
  }, [serviceResourceItem])

  // *****
  // functions
  // *****
  const updateForm = (): void => {
    if (serviceResourceItem) {
      const items = [...serviceResourceItem.serviceResources]
      const options: any = []
      const stateUpdate = { ...state }
      items.forEach((item) => {
        options.push({
          name: item.name,
          value: item.name
        })
      })
      const form = setSelectOptions('resourceName', options, stateUpdate.form)
      stateUpdate.form = form
      setState(stateUpdate)
    }
  }

  const onChangeInput = (event: any, formInputName: string): void => {
    const value = event.target.value
    const updatedState = {
      ...state
    }

    let updatedForm = updatedState.form

    updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)

    if (formInputName === 'scopeType') {
      setCloudAccountSelected(cloudAccountSelectedInitial)
      if (value === 'QUOTA_ACCOUNT_ID') {
        updatedForm.cloudAccount.hidden = false
        updatedForm.userType.hidden = true
        updatedForm.userType.value = ''
        setQuotaCloudType(false)
      }
      if (value === 'QUOTA_ACCOUNT_TYPE') {
        updatedForm.userType.hidden = false
        updatedForm.cloudAccount.hidden = true
        updatedForm.cloudAccount.value = ''
        setQuotaCloudType(true)
      }
    }

    updatedState.form = updatedForm
    updatedState.isValidForm = isValidForm(updatedForm)

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

  const onSearchCloudAccount = async (): Promise<void> => {
    const stateCopy = { ...state }
    const cloudAccount = getFormValue('cloudAccount', state.form)
    if (cloudAccount) {
      await getCloudAccountInfo(cloudAccount)
    } else {
      const updatedForm = markErrorOnElement(state.form, 'cloudAccount')
      stateCopy.form = updatedForm
      setState(stateCopy)
    }
  }

  const getCloudAccountInfo = async (cloudAccount: string): Promise<void> => {
    try {
      setSearchModal({ ...searchModal, show: true, message: 'Searching for Cloud Account' })
      const response = await CloudAccountService.getCloudAccountDetailsById(cloudAccount)
      const { data } = response
      const { id, name, type } = data
      setCloudAccountSelected({
        ...cloudAccountSelected,
        show: true,
        id,
        name,
        type
      })
      setSearchModal(searchModalInitial)
      return id
    } catch (error: any) {
      setCloudAccountSelected(cloudAccountSelectedInitial)
      setSearchModal(searchModalInitial)
      const code = error.response.data?.code
      const message = code && [3, 5].includes(code) ? error.response.data?.message : 'Cloud Account: is not found'
      const stateCopy = { ...state }
      const updatedForm = markErrorOnElement(state.form, 'cloudAccount', message)
      stateCopy.form = updatedForm
      setState(stateCopy)
    }
  }

  const onSubmit = async (): Promise<void> => {
    try {
      const isValidForm = state.isValidForm
      if (!isValidForm) {
        showRequiredFields()
        return
      }
      if (!quotaCloudType && !cloudAccountSelected.id) {
        await onSearchCloudAccount()
        return
      }
      setSubmitModal({ ...submitModal, show: true })
      const payloadCopy = { ...state.servicePayload }
      const serviceQuotaResourceCopy = { ...payloadCopy.serviceQuotaResource }
      serviceQuotaResourceCopy.resourceType = getFormValue('resourceName', state.form)
      serviceQuotaResourceCopy.reason = getFormValue('reason', state.form)
      serviceQuotaResourceCopy.quotaConfig.limits = getFormValue('maxLimit', state.form)
      serviceQuotaResourceCopy.quotaConfig.quotaUnit = getFormValue('quotaUnit', state.form)
      serviceQuotaResourceCopy.scope.scopeType = getFormValue('scopeType', state.form)
      serviceQuotaResourceCopy.scope.scopeValue = quotaCloudType ? getFormValue('userType', state.form) : cloudAccountSelected.id
      payloadCopy.serviceQuotaResource = serviceQuotaResourceCopy
      await StorageManagementService.postServiceQuota(param, payloadCopy)
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
    navigate(`/quotamanagement/services/d/${param}/quotas`)
  }
  return <QuotaManagementServiceCreate
    onCancel={onCancel}
    onSubmit={onSubmit}
    state={state}
    onChangeInput={onChangeInput}
    serviceResourceLimit={10}
    submitModal={submitModal}
    moduleName={'quotas'}
    cloudAccountSelected={cloudAccountSelected}
    onSearchCloudAccount={onSearchCloudAccount}
    searchModal={searchModal}
    isPageReady={true}
    emptyView={emptyViewInitial}/>
}

export default QuotaManagementServiceAddQuotaContainer
