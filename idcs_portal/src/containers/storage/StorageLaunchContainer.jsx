// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import StorageLaunch from '../../components/storage/storageLaunch/StorageLaunch'
import idcConfig, { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import { useNavigate } from 'react-router'
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
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'
import {
  friendlyErrorMessages,
  isErrorInAuthorization,
  isErrorInsufficientCapacity,
  isErrorInsufficientCredits
} from '../../utils/apiError/apiError'

const getCustomMessage = (messageType, minSize, maxSize, unitSize) => {
  let message = null

  switch (messageType) {
    case 'name':
      message = (
        <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'}>
          Max length 63 characters. Letters, numbers and ‘- ‘ accepted.
          <br />
          Name should start and end with an alphanumeric character.
        </div>
      )
      break
    case 'description':
      message = (
        <div className="valid-feedback" intc-id={'InstanceDetailsVolumeMessage'}>
          Provide a description for this volume.{' '}
        </div>
      )
      break
    case 'diskSize':
      message = (
        <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'}>
          {`Size should be between ${minSize} and ${maxSize} ${unitSize}.`}
        </div>
      )
      break
    default:
      break
  }

  return message
}

const StorageLaunchContainer = () => {
  // local state
  const isStorageFlagAvailable = isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_STORAGE)
  const messageBanner =
    'Get ready to learn the fundamentals and start building on Intel-optimized software stacks hosted on Intel compute platforms.'
  const throwError = useErrorBoundary()
  const initialState = {
    mainTitle: 'Create a storage volume',
    configSectionTitle: 'Volume information',
    form: {
      name: {
        sectionGroup: 'configuration',
        type: 'text', // options = 'text ,'textArea'
        label: 'Name:',
        placeholder: 'Name',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        validationRules: {
          isRequired: true,
          onlyAlphaNumLower: true,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: getCustomMessage('name')
      },
      storageSize: {
        sectionGroup: 'configuration',
        type: 'integer', // options = 'text ,'textArea'
        label: 'Storage Size:',
        placeholder: '',
        // inputAppend: '$0.00 / month',
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
      metadata: {
        name: '',
        description: ''
      },
      spec: {
        availabilityZone: idcConfig.REACT_APP_DEFAULT_REGION_AVAILABILITY_ZONE,
        request: {
          storage: '',
          instanceType: ''
        },
        storageClass: isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_STORAGE_VAST)
          ? 'GeneralPurposeStd'
          : 'GeneralPurpose',
        filesystemType: 'ComputeGeneral'
      }
    },
    isValidForm: false,
    showReservationModal: false,
    showErrorModal: false,
    errorMessage: '',
    navigationBottom: [
      {
        buttonLabel: 'Create',
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

  const [state, setState] = useState(initialState)
  const [costPerMonth, setCostPerMonth] = useState('')
  const [emptyCatalogModal, setEmptyCatalogModal] = useState(false)
  const [storageUnitSize, setStorageUnitSize] = useState('')
  const [storageUsageUnit, setStorageUsageUnit] = useState('')
  const [showUpgradeNeededModal, setShowUpgradeNeededModal] = useState(false)

  // Store
  const storages = useStorageStore((state) => state.storages)
  const setStorages = useStorageStore((state) => state.setStorages)
  const showError = useToastStore((state) => state.showError)

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        if (isStorageFlagAvailable) {
          await setStorages()
        }
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    if (storages === null) {
      return
    }
    if (storages.length > 0) {
      setStorageOptions()
      setEmptyCatalogModal(false)
    } else {
      setEmptyCatalogModal(true)
    }
  }, [storages])

  // Navigation
  const navigate = useNavigate()

  // Functions
  function setStorageOptions() {
    const storage = { ...storages[0] }

    const minimumSize = storage.minimumSize
    const maximumSize = storage.maximumSize
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
    setState(stateCopy)
  }

  function onCancel() {
    // Navigates back to the page when this method triggers.
    navigate('/storage')
  }

  function getPricing(spaceInGb) {
    const storage = { ...storages[0] }
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
      payloadCopy.metadata.name = getFormValue('name', stateCopy.form)
      payloadCopy.metadata.description = ''
      const specCopy = { ...payloadCopy.spec }
      specCopy.request.storage = `${getFormValue('storageSize', stateCopy.form)}${storageUnitSize}`
      specCopy.instanceType = storages[0].name
      payloadCopy.spec = specCopy
      stateCopy.showReservationModal = true
      setState(stateCopy)
      await createReservation(payloadCopy)
    } catch (error) {
      const stateUpdated = { ...state }

      stateUpdated.showReservationModal = false
      stateUpdated.errorHideRetryMessage = false
      stateUpdated.errorDescription = null

      if (error.response) {
        const errData = error.response.data
        const errCode = errData.code
        if (isErrorInAuthorization(error)) {
          stateUpdated.showErrorModal = true
          stateUpdated.errorMessage = error.response.data.message
        } else if (isErrorInsufficientCredits(error)) {
          // No Credits
          setShowUpgradeNeededModal(true)
        } else if (errCode === 11) {
          // No Quota
          stateUpdated.showErrorModal = true
          stateUpdated.errorMessage = error.response.data.message
          stateUpdated.errorHideRetryMessage = true
        } else if (isErrorInsufficientCapacity(error.response.data.message)) {
          stateUpdated.showErrorModal = true
          stateUpdated.errorDescription = friendlyErrorMessages.insufficientCapacity
          stateUpdated.errorMessage = ''
          stateUpdated.errorHideRetryMessage = true
        } else {
          stateUpdated.showErrorModal = true
          stateUpdated.errorMessage = error.response.data.message
        }
      } else {
        stateUpdated.showErrorModal = true
        stateUpdated.errorMessage = error.message
      }

      setState(stateUpdated)
    }
  }

  function onClickCloseErrorModal() {
    const stateUpdated = { ...state }
    stateUpdated.showErrorModal = false
    setState(stateUpdated)
  }

  async function createReservation(servicePayload) {
    await CloudAccountService.postStorageReservation(servicePayload)
    setTimeout(() => {
      const stateUpdated = { ...state }

      stateUpdated.showReservationModal = false

      setState(stateUpdated)

      navigate({
        pathname: '/storage'
      })
    }, state.timeoutMiliseconds)
  }

  return (
    <StorageLaunch
      state={state}
      costPerMonth={costPerMonth}
      emptyCatalogModal={emptyCatalogModal}
      isStorageFlagAvailable={isStorageFlagAvailable}
      messageBanner={messageBanner}
      onClickCloseErrorModal={onClickCloseErrorModal}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      storageUsageUnit={storageUsageUnit}
      showUpgradeNeededModal={showUpgradeNeededModal}
      setShowUpgradeNeededModal={setShowUpgradeNeededModal}
    />
  )
}

export default StorageLaunchContainer
