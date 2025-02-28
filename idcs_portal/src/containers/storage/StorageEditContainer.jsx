// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useNavigate, useParams } from 'react-router'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import CloudAccountService from '../../services/CloudAccountService'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  showFormRequiredFields
} from '../../utils/updateFormHelper/UpdateFormHelper'
import useStorageStore from '../../store/storageStore/StorageStore'
import { formatCurrency, formatNumber } from '../../utils/numberFormatHelper/NumberFormatHelper'
import StorageEdit from '../../components/storage/storageEdit/StorageEdit'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'
import { isErrorInAuthorization } from '../../utils/apiError/apiError'

const getCustomMessage = (messageType, minSize, maxSize, unitSize) => {
  let message = null

  switch (messageType) {
    case 'diskSize':
      message = (
        <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'}>
          {`Size should be between ${minSize} and ${maxSize} ${unitSize}.`}
          <br />
          Size can only be upgraded.
        </div>
      )
      break
    default:
      break
  }

  return message
}

const StorageEditContainer = () => {
  const { param: name } = useParams()
  // local state
  const initialState = {
    mainTitle: `Edit storage volume ${name}`,
    configSectionTitle: 'Volume information',
    form: {
      storageSize: {
        sectionGroup: 'configuration',
        type: 'integer', // options = 'text ,'textArea'
        label: 'Storage Size:',
        placeholder: '',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: '',
        helperMessage: ''
      }
    },
    servicePayload: {
      spec: {
        request: {
          storage: ''
        }
      }
    },
    isValidForm: false,
    showReservationModal: false,
    showErrorModal: false,
    errorMessage: '',
    navigationBottom: [
      {
        buttonLabel: 'Edit',
        buttonVariant: 'primary'
      },
      {
        buttonLabel: 'Cancel',
        buttonVariant: 'link',
        buttonFunction: () => onCancel()
      }
    ],
    timeoutMiliseconds: 4000
  }

  const throwError = useErrorBoundary()
  const [searchParams] = useSearchParams()

  const [state, setState] = useState(initialState)
  const [costPerMonth, setCostPerMonth] = useState('')
  const [emptyCatalogModal, setEmptyCatalogModal] = useState(false)
  const [storageUnitSize, setStorageUnitSize] = useState('')
  const [storageUsageUnit, setStorageUsageUnit] = useState('')
  const [storageToEdit, setStorageToEdit] = useState(null)
  const [isPageReady, setIsPageReady] = useState(false)
  const showError = useToastStore((state) => state.showError)

  // Store
  const storageProducts = useStorageStore((state) => state.storages)
  const setStorageProducts = useStorageStore((state) => state.setStorages)
  const storages = useCloudAccountStore((state) => state.storages)
  const loading = useCloudAccountStore((state) => state.loading)
  const setStorages = useCloudAccountStore((state) => state.setStorages)

  // Navigation
  const navigate = useNavigate()

  const refreshStorageProduct = async () => {
    try {
      await setStorageProducts()
    } catch (error) {
      throwError(error)
    }
  }

  const refreshStorages = async (background) => {
    try {
      await setStorages(background)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  // Hooks
  useEffect(() => {
    if (!storageProducts || storageProducts.length === 0) {
      refreshStorageProduct()
    }
    if (!storages || storages.length === 0) {
      refreshStorages()
    } else {
      setIsPageReady(true)
    }
  }, [])

  useEffect(() => {
    let shouldExit = false
    if (isPageReady) {
      if (storages.length > 0) {
        const storage = storages?.find((item) => item.name === name)
        if (storage !== undefined) {
          setStorageToEdit(storage)
        } else {
          shouldExit = true
        }
      } else {
        shouldExit = true
      }
    }

    if (shouldExit) {
      navigate({
        pathname: '/storage'
      })
    }
  }, [isPageReady])

  useEffect(() => {
    if (storageProducts === null || !storageToEdit) {
      return
    }
    if (storageProducts.length > 0) {
      setStorageOptions()
      setEmptyCatalogModal(false)
    } else {
      setEmptyCatalogModal(true)
    }
  }, [storageProducts, storageToEdit])

  // Functions

  function setStorageOptions() {
    const storage = { ...storageProducts[0] }

    const minSize = Number(storageToEdit.size) + 1
    const minimumSize = minSize > storage.maximumSize ? storage.maximumSize : minSize
    const maximumSize = Number(storage.maximumSize)
    const unitSize = storage.unitSize
    const usageUnit = storage.usageUnit

    setStorageUnitSize(unitSize)
    setStorageUsageUnit(usageUnit)

    const stateCopy = { ...state }
    const formCopy = stateCopy.form

    const element = formCopy.storageSize
    element.maxLength = maximumSize.toString().length
    element.validationRules = {
      ...element.validationRules,
      checkMinValue: minimumSize,
      checkMaxValue: maximumSize
    }
    element.helperMessage = getCustomMessage('diskSize', minimumSize, maximumSize, unitSize)
    element.label = `Storage Size (${unitSize}):`

    stateCopy.form = formCopy

    if (storageToEdit) {
      stateCopy.mainTitle = 'Edit storage volume - ' + storageToEdit.name
    }

    setState(stateCopy)
  }

  function goBack() {
    const backTo = searchParams.get('backTo')
    switch (backTo) {
      case 'detail':
        navigate({
          pathname: `/storage/d/${name}`
        })
        break
      default:
        navigate({
          pathname: '/storage'
        })
        break
    }
  }

  function onCancel() {
    // Navigates back to the page when this method triggers.
    goBack()
  }

  function getPricing(spaceInGb) {
    const storage = { ...storageProducts[0] }
    const costPerMonth = storage.rate * Number(spaceInGb)
    return formatCurrency(formatNumber(costPerMonth, 2))
  }

  function onChangeInput(event, formInputName) {
    const value = event.target.value
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    const storageSizeInput = updatedForm.storageSize

    if (storageSizeInput.isValid) {
      const costPerHour = getPricing(storageSizeInput.value)
      setCostPerMonth(costPerHour)
    } else {
      setCostPerMonth('error')
    }

    updatedState.form = updatedForm

    setState(updatedState)
  }

  function showRequiredFields() {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(stateCopy.form)
    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    stateCopy.form = updatedForm
    setState(stateCopy)
  }

  async function onSubmit(e) {
    try {
      const stateCopy = { ...state }
      const isValidForm = stateCopy.isValidForm
      if (!isValidForm) {
        showRequiredFields()
        return
      }
      const payloadCopy = { ...stateCopy.servicePayload }
      const specCopy = { ...payloadCopy.spec }
      specCopy.request.storage = `${getFormValue('storageSize', stateCopy.form)}${storageUnitSize}`
      payloadCopy.spec = specCopy
      stateCopy.showReservationModal = true
      setState(stateCopy)
      await editStorage(payloadCopy)
    } catch (error) {
      const stateUpdated = { ...state }
      stateUpdated.showReservationModal = false
      if (error.response) {
        // TODO: keep auth check condition for future update, even though it currently produces identical output. This will make it easier to adjust permissions if requirements change.
        if (isErrorInAuthorization(error)) {
          stateUpdated.errorMessage = error.response.data.message
        } else {
          stateUpdated.errorMessage = error.response.data.message
        }
      } else {
        stateUpdated.errorMessage = error.message
      }
      stateUpdated.showErrorModal = true
      setState(stateUpdated)
    }
  }

  function onClickCloseErrorModal() {
    const stateUpdated = { ...state }
    stateUpdated.showErrorModal = false
    setState(stateUpdated)
  }

  async function editStorage(servicePayload) {
    await CloudAccountService.putStorageReservation(storageToEdit.resourceId, servicePayload)
    setTimeout(() => {
      setStorages()
      const stateUpdated = { ...state }
      stateUpdated.showReservationModal = false
      setState(stateUpdated)
      goBack()
    }, state.timeoutMiliseconds)
  }
  return (
    <StorageEdit
      state={state}
      loading={loading || !isPageReady || !storageToEdit}
      costPerMonth={costPerMonth}
      emptyCatalogModal={emptyCatalogModal}
      onClickCloseErrorModal={onClickCloseErrorModal}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      storageUsageUnit={storageUsageUnit}
    />
  )
}

export default StorageEditContainer
