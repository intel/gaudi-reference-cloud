// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import idcConfig from '../../config/configurator'
import BucketService from '../../services/BucketService'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  showFormRequiredFields
} from '../../utils/updateFormHelper/UpdateFormHelper'
import ObjectStorageLaunch from '../../components/objectStorage/objectStorageLaunch/ObjectStorageLaunch'
import useUserStore from '../../store/userStore/UserStore'
import useStorageStore from '../../store/storageStore/StorageStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { formatCurrency, formatNumber } from '../../utils/numberFormatHelper/NumberFormatHelper'
import useBucketStore from '../../store/bucketStore/BucketStore'
import {
  friendlyErrorMessages,
  isErrorInsufficientCapacity,
  isErrorInAuthorization,
  isErrorInsufficientCredits
} from '../../utils/apiError/apiError'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'

const getCustomMessage = (messageType) => {
  let message = null

  switch (messageType) {
    case 'name':
      message = (
        <div className="valid-feedback" intc-id={'StorageBucketNameValidMessage'}>
          Max length 50 characters. Letters, numbers and ‘- ‘ accepted.
          <br />
          Name should start and end with an alphanumeric character.
        </div>
      )
      break
    case 'description':
      message = (
        <div className="valid-feedback" intc-id={'StorageBucketDescriptionValidMessage'}>
          Provide a description for this bucket.
        </div>
      )
      break
    default:
      break
  }

  return message
}

const ObjectStorageLaunchContainer = () => {
  // local state
  const cloudAccountNumber = useUserStore((state) => state.user.cloudAccountNumber)
  const showError = useToastStore((state) => state.showError)
  const getPrependMessage = (messageType) => {
    let message = null

    switch (messageType) {
      case 'name':
        message = <div intc-id={'StorageBucketNamePrependMessage'}>{cloudAccountNumber}-</div>
        break
      default:
        break
    }

    return message
  }

  const initialState = {
    mainTitle: 'Create storage bucket',
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
        maxLength: 50,
        validationRules: {
          isRequired: true,
          onlyAlphaNumLower: true,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: getCustomMessage('name'),
        prepend: getPrependMessage('name')
      },
      description: {
        sectionGroup: 'configuration',
        type: 'textArea', // options = 'text ,'textArea'
        label: 'Description:',
        placeholder: 'Description',
        value: '', // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 100,
        validationRules: {
          isRequired: false,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: getCustomMessage('description')
      },
      versioned: {
        sectionGroup: 'configuration',
        type: 'checkbox', // options = 'text ,'textArea'
        label: '',
        placeholder: '',
        value: false, // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        isChecked: false,
        validationRules: {
          isRequired: false
        },
        options: [
          {
            name: 'Enable versioning',
            value: false,
            defaultChecked: false
          }
        ],
        validationMessage: '', // Errror message to display to the user
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
        versioned: false,
        instanceType: ''
      }
    },
    isValidForm: false,
    showReservationModal: false,
    showErrorModal: false,
    errorMessage: '',
    errorHideRetryMessage: null,
    errorDescription: null,
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
  const [costPerHour, setCostPerHour] = useState(0)
  const [showUpgradeNeededModal, setShowUpgradeNeededModal] = useState(false)
  const [emptyCatalogModal, setEmptyCatalogModal] = useState(false)
  const [storageUsageUnit, setStorageUsageUnit] = useState('')

  // Store
  const buildCurrentSelectedBucket = useBucketStore((state) => state.buildCurrentSelectedBucket)
  const objectStorages = useStorageStore((state) => state.objectStorages)
  const setObjectStorages = useStorageStore((state) => state.setObjectStorages)

  const throwError = useErrorBoundary()

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        await setObjectStorages()
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    if (objectStorages === null) {
      return
    }
    if (objectStorages.length > 0) {
      setStorageOptions()
      setEmptyCatalogModal(false)
      getPricing()
    } else {
      setEmptyCatalogModal(true)
    }
  }, [objectStorages])

  // Navigation
  const navigate = useNavigate()

  // Functions

  function onCancel() {
    // Navigates back to the page when this method triggers.
    navigate('/buckets')
  }

  function setStorageOptions() {
    const objectStorage = { ...objectStorages[0] }

    const usageUnit = objectStorage.usageUnit
    setStorageUsageUnit(usageUnit)
  }

  function getPricing() {
    const rate = objectStorages[0].rate
    setCostPerHour(formatCurrency(formatNumber(rate, 2)))
  }

  function onChangeInput(event, formInputName) {
    const value = event.target.value
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    if (formInputName === 'versioned') {
      updatedForm[formInputName].isChecked = event.target.checked
      updatedForm[formInputName].value = event.target.checked
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
      payloadCopy.metadata.description = getFormValue('description', stateCopy.form)
      payloadCopy.spec.versioned = stateCopy.form.versioned.isChecked
      payloadCopy.spec.instanceType = objectStorages[0].name
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
    const { data } = await BucketService.postObjectStorageReservation(servicePayload)
    buildCurrentSelectedBucket(data)
    setTimeout(() => {
      const stateUpdated = { ...state }

      stateUpdated.showReservationModal = false

      setState(stateUpdated)

      navigate({
        pathname: '/buckets'
      })
    }, state.timeoutMiliseconds)
  }
  return (
    <ObjectStorageLaunch
      state={state}
      onClickCloseErrorModal={onClickCloseErrorModal}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      costPerHour={costPerHour}
      showUpgradeNeededModal={showUpgradeNeededModal}
      setShowUpgradeNeededModal={setShowUpgradeNeededModal}
      emptyCatalogModal={emptyCatalogModal}
      storageUsageUnit={storageUsageUnit}
    />
  )
}

export default ObjectStorageLaunchContainer
