// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import QuotaManagementServiceCreate from '../../components/storageManagement/quotaManagementService/QuotaManagementServiceCreate'
import { useNavigate, useParams } from 'react-router'
import { UpdateFormHelper, isValidForm, showFormRequiredFields, getFormValue, setFormValue } from '../../utility/updateFormHelper/UpdateFormHelper'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utility/Enums'
import StorageManagementService from '../../services/StorageManagementService'
import useStorageManagementStore from '../../store/storageManagementStore/StorageManagementStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'

const QuotaManagementServiceEditQuotaContainer = (): JSX.Element => {
  // *****
  // Params
  // *****
  const { param, resourceName, ruleId } = useParams()
  // *****
  // Global state
  // *****
  const showError = useToastStore((state) => state.showError)
  const serviceQuotaResource = useStorageManagementStore((state) => state.serviceQuotaResource)
  const getServiceQuota = useStorageManagementStore((state) => state.getServiceQuota)
  const resetServiceQuotaResource = useStorageManagementStore((state) => state.resetServiceQuotaResource)
  const throwError = useErrorBoundary()

  // *****
  // local state
  // *****
  const navigate = useNavigate()

  const initialState = {
    mainTitle: `Edit ${resourceName} quota`,
    form: {
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
        isReadOnly: true,
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
      }
    },
    isValidForm: false,
    servicePayload: {
      ruleId: '',
      quotaConfig: {
        limits: '',
        quotaUnit: ''
      },
      reason: ''
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

  const submitModalInitial = {
    show: false,
    message: 'Saving changes'
  }

  const emptyViewInitial = {
    show: false,
    title: 'Rule Id not found',
    action: {
      type: 'function',
      href: () => { onCancel() },
      label: 'Back to quotas'
    }
  }

  const [state, setState] = useState(initialState)
  const [submitModal, setSubmitModal] = useState(submitModalInitial)
  const [isPageReady, setIsPageReady] = useState(false)
  const [emptyView, setEmptyView] = useState(emptyViewInitial)
  // *****
  // Hooks
  // *****[]
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        if (param && resourceName) {
          await getServiceQuota(param, resourceName)
        }
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
    return () => { resetServiceQuotaResource() }
  }, [])

  useEffect(() => {
      updateFormValues()
  }, [serviceQuotaResource])

  // *****
  // functions
  // *****
  const updateFormValues = (): void => {
    if (serviceQuotaResource) {
      const serviceQuotas = serviceQuotaResource?.serviceQuotaResources
      const serviceQuota = serviceQuotas?.find((item) => item.ruleId === ruleId)
      if (serviceQuota) {
        const stateUpdated = { ...state }
        let formUpdated = state.form
        formUpdated = setFormValue('quotaUnit', serviceQuota?.quotaUnit, formUpdated)
        formUpdated = setFormValue('maxLimit', serviceQuota?.maxLimit, formUpdated)
        formUpdated = setFormValue('reason', serviceQuota?.reason, formUpdated)
        stateUpdated.isValidForm = isValidForm(formUpdated)
        stateUpdated.form = formUpdated
        setState(stateUpdated)
      } else {
        setEmptyView({ ...emptyView, show: true })
      }
      setIsPageReady(true)
    }
  }

  const onChangeInput = (event: any, formInputName: string): void => {
    const value = event.target.value
    const updatedState = {
      ...state
    }

    let updatedForm = updatedState.form

    updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)

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

  const onSubmit = async (): Promise<void> => {
    try {
      const isValidForm = state.isValidForm
      if (!isValidForm) {
        showRequiredFields()
        return
      }
      setSubmitModal({ ...submitModal, show: true })
      const payloadCopy = { ...state.servicePayload }
      const ruleIdValue = ruleId?.toString() ?? ''
      if (ruleIdValue === '') {
        throw new Error('Invalid rule id')
      }
      payloadCopy.ruleId = ruleIdValue
      payloadCopy.quotaConfig.limits = getFormValue('maxLimit', state.form)
      payloadCopy.quotaConfig.quotaUnit = getFormValue('quotaUnit', state.form)
      payloadCopy.reason = getFormValue('reason', state.form)
      await StorageManagementService.putServiceQuota(param, resourceName, payloadCopy)
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
  return <QuotaManagementServiceCreate onCancel={onCancel} onSubmit={onSubmit} state={state} onChangeInput={onChangeInput} serviceResourceLimit={10} submitModal={submitModal} moduleName={'quotas'} isPageReady={isPageReady} emptyView={emptyView} />
}

export default QuotaManagementServiceEditQuotaContainer
