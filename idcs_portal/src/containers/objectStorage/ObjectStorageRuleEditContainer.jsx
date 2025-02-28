// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  setFormValue,
  showFormRequiredFields
} from '../../utils/updateFormHelper/UpdateFormHelper'
import useBucketStore from '../../store/bucketStore/BucketStore'
import BucketService from '../../services/BucketService'
import { formatNumber } from '../../utils/numberFormatHelper/NumberFormatHelper'
import ObjectStorageRuleEdit from '../../components/objectStorage/objectStorageRuleEdit/ObjectStorageRuleEdit'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { isErrorInAuthorization } from '../../utils/apiError/apiError'

const getCustomMessage = (messageType) => {
  let message = null

  switch (messageType) {
    case 'prefix':
      message = (
        <div className="valid-feedback" intc-id={'ObjectStorageRulePrefixValidMessage'}>
          The prefix of objects to apply the rule to. Max length 1024.
        </div>
      )
      break
    default:
      break
  }

  return message
}

const ObjectStorageRuleEditContainer = () => {
  const { param: name, param2: ruleName } = useParams()
  // local state

  const initialState = {
    mainTitle: `Edit Lifecycle Rule - ${ruleName}`,
    form: {
      prefix: {
        sectionGroup: 'configuration',
        type: 'text', // options = 'text ,'textArea'
        label: 'Prefix:',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 1024,
        validationRules: {
          isRequired: false,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: getCustomMessage('prefix'),
        columnSize: 6
      },
      deleteMarker: {
        sectionGroup: 'deleteMarker',
        type: 'radio', // options = 'text ,'textArea'
        label: '',
        placeholder: '',
        value: '2', // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        isChecked: false,
        validationRules: {
          isRequired: false
        },
        options: [
          {
            name: 'Delete Marker',
            value: '1'
          },
          {
            name: 'Expiry Days:',
            value: '2'
          }
        ],
        validationMessage: '', // Errror message to display to the user
        helperMessage: '',
        columnSize: 2
      },
      expireDays: {
        sectionGroup: 'deleteMarker',
        type: 'integer', // options = 'text ,'textArea'
        label: '',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 4,
        validationRules: {
          isRequired: false,
          onlyCreditNumeric: true,
          checkMinValue: 1,
          checkMaxValue: 2557
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: '',
        columnSize: 6,
        customClass: 'flex-width-20'
      },
      noncurrentExpireDays: {
        sectionGroup: 'noncurrentExpireDays',
        type: 'text', // options = 'text ,'textArea'
        label: 'Non current expiry days:',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 4,
        validationRules: {
          isRequired: false,
          onlyCreditNumeric: true,
          checkMaxValue: 2557
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: '',
        customClass: 'flex-width-20',
        columnSize: 6
      }
    },
    servicePayload: {
      spec: {
        deleteMarker: false,
        expireDays: 0,
        noncurrentExpireDays: 0,
        prefix: ''
      }
    },
    isValidForm: true,
    showReservationModal: false,
    showErrorModal: false,
    errorMessage: '',
    errorTitleMessage: '',
    errorDescription: '',
    navigationBottom: [
      {
        buttonLabel: 'Save',
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
  const showError = useToastStore((state) => state.showError)
  const [bucketRuleToEdit, setBucketRuleToEdit] = useState(null)
  const [isPageReady, setIsPageReady] = useState(false)
  const throwError = useErrorBoundary()

  // Global State
  const objectStorages = useBucketStore((state) => state.objectStorages)
  const setObjectStorages = useBucketStore((state) => state.setObjectStorages)
  const currentSelectedBucket = useBucketStore((state) => state.currentSelectedBucket)
  const setCurrentSelectedBucket = useBucketStore((state) => state.setCurrentSelectedBucket)

  // Navigation
  const navigate = useNavigate()

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      if (objectStorages?.length === 0) await refreshStorages(false)
      setIsPageReady(true)
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    updateDetails()
  }, [objectStorages, isPageReady])

  const updateDetails = () => {
    const storage = objectStorages.find((instance) => instance.name === name)
    if (storage === undefined) {
      if (isPageReady) navigate('/buckets')
      setCurrentSelectedBucket(null)
      return
    }
    const bucketRule = storage.lifecycleRulePolicies.find((x) => x.ruleName === ruleName)
    if (bucketRule === undefined) {
      navigate('/buckets')
      return
    }

    const storageDetail = { ...storage }
    setCurrentSelectedBucket(storageDetail)
    setBucketRuleToEdit({ ...bucketRule })
  }

  useEffect(() => {
    setForm()
  }, [bucketRuleToEdit])

  // functions
  function setForm() {
    const stateUpdated = {
      ...state
    }

    if (bucketRuleToEdit) {
      stateUpdated.form = setFormValue('prefix', bucketRuleToEdit.prefix, stateUpdated.form)
      stateUpdated.form = setFormValue('expireDays', bucketRuleToEdit.expireDays, stateUpdated.form)
      stateUpdated.form = setFormValue('noncurrentExpireDays', bucketRuleToEdit.noncurrentExpireDays, stateUpdated.form)
      stateUpdated.form = setFormValue('deleteMarker', bucketRuleToEdit.deleteMarker ? '1' : '2', stateUpdated.form)
      stateUpdated.form.expireDays.isReadOnly = bucketRuleToEdit.deleteMarker
    }
    setState(stateUpdated)
  }

  function goBack() {
    navigate({
      pathname: `/buckets/d/${name}`,
      search: 'tab=lifecycleRules'
    })
  }

  function onCancel() {
    // Navigates back to the page when this method triggers.
    goBack()
  }

  function onChangeInput(event, formInputName) {
    const value = event.target.value
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)

    if (formInputName === 'deleteMarker') {
      updatedForm.expireDays.value = 0
      updatedForm.expireDays.isReadOnly = value === '1'
      updatedForm.expireDays.isValid = value === '1'
    }

    if (
      formInputName === 'expireDays' &&
      (updatedForm.expireDays.value === '' || updatedForm.expireDays?.value.toString() === '0')
    ) {
      updatedForm.expireDays.isTouched = true
      updatedForm.expireDays.isValid = false
    }

    updatedState.isValidForm = isValidForm(updatedForm)

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

  function validateExpiryDays() {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = stateCopy.form
    if (
      updatedForm.deleteMarker.value === '2' &&
      (updatedForm.expireDays.value === '' || updatedForm.expireDays?.value.toString() === '0')
    ) {
      updatedForm.expireDays.isTouched = true
      updatedForm.expireDays.isValid = false

      setState(stateCopy)
      return false
    }
    return true
  }

  const refreshStorages = async () => {
    try {
      await setObjectStorages(false)
    } catch (error) {
      if (isErrorInAuthorization(error)) {
        const stateUpdated = { ...state }
        stateUpdated.showErrorModal = true
        stateUpdated.errorMessage = error.response.data.message
        setState(stateUpdated)
      } else throwError(error)
    }
  }

  async function onSubmit(e) {
    try {
      const stateCopy = { ...state }
      const isValidForm = stateCopy.isValidForm
      if (!validateExpiryDays() || !isValidForm) {
        showRequiredFields()
        return
      }
      const payloadCopy = { ...stateCopy.servicePayload }
      payloadCopy.spec.prefix = getFormValue('prefix', stateCopy.form)
      payloadCopy.spec.expireDays = formatNumber(getFormValue('expireDays', stateCopy.form), 0)
      payloadCopy.spec.noncurrentExpireDays = formatNumber(getFormValue('noncurrentExpireDays', stateCopy.form), 0)
      payloadCopy.spec.deleteMarker = getFormValue('deleteMarker', stateCopy.form) === '1'
      stateCopy.showReservationModal = true
      setState(stateCopy)
      await editRule(payloadCopy)
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
      stateUpdated.errorTitleMessage = 'Could not update your Lifecycle Rule'
      stateUpdated.errorDescription = 'There was an error while processing your lifecycle rule.'
      setState(stateUpdated)
    }
  }

  function onClickCloseErrorModal() {
    const stateUpdated = { ...state }
    stateUpdated.showErrorModal = false
    setState(stateUpdated)
  }

  async function editRule(servicePayload) {
    await BucketService.updateObjectBucketRule(
      servicePayload,
      bucketRuleToEdit.resourceId,
      currentSelectedBucket.resourceId
    )
    refreshStorages()
    setTimeout(() => {
      const stateUpdated = { ...state }

      stateUpdated.showReservationModal = false

      setState(stateUpdated)

      goBack()
    }, state.timeoutMiliseconds)
  }
  return (
    <ObjectStorageRuleEdit
      loading={!isPageReady || !bucketRuleToEdit}
      state={state}
      onClickCloseErrorModal={onClickCloseErrorModal}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
    />
  )
}

export default ObjectStorageRuleEditContainer
