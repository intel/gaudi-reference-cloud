// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  showFormRequiredFields,
  setFormValue
} from '../../utils/updateFormHelper/UpdateFormHelper'
import useClusterStore from '../../store/clusterStore/ClusterStore'
import ClusterAddStorage from '../../components/cluster/clusterAddStorage/ClusterAddStorage'
import ClusterService from '../../services/ClusterService'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useStorageStore from '../../store/storageStore/StorageStore'
import { formatCurrency, formatNumber } from '../../utils/numberFormatHelper/NumberFormatHelper'

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
    case 'volumeSize':
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

const ClusterEditStorageContainer = () => {
  // *****
  // Global state
  // *****
  const clusters = useClusterStore((state) => state.clustersData)
  const setClusters = useClusterStore((state) => state.setClustersData)
  const showError = useToastStore((state) => state.showError)
  const storages = useStorageStore((state) => state.storages)
  const setStorages = useStorageStore((state) => state.setStorages)

  // *****
  // local state
  // *****
  const navigate = useNavigate()
  const { param: name } = useParams()

  const initialState = {
    mainTitle: `Edit storage for cluster ${name}`,
    mainSubtitle: '',
    form: {
      fileVolumeFlag: {
        type: 'checkbox', // options = 'text ,'textArea'
        label: '',
        placeholder: '',
        value: '0', // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: true, // Input create as read only
        options: [
          {
            name: 'Enable file volume',
            value: '0'
          }
        ],
        validationRules: {
          isRequired: true
        },
        sectionGroup: 'storage',
        subGroup: 'storage',
        columnSize: '12',
        validationMessage: '',
        helperMessage: '',
        hidden: false
      },
      volumeSize: {
        type: 'integer', // options = 'text ,'textArea'
        label: 'Storage size (GB):',
        placeholder: '',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        maxLength: 6,
        validationRules: {
          isRequired: true,
          checkMaxValue: 500000,
          checkMinValue: 5
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: getCustomMessage('volumeSize'),
        sectionGroup: 'storage',
        subGroup: 'storage',
        columnSize: '8',
        hidden: false
      }
    },
    isValidForm: false,
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
    ]
  }

  const payload = {
    enablestorage: true,
    storagesize: ''
  }

  const submitModalInitial = {
    show: false
  }

  const errorModalInitial = {
    show: false,
    title: 'Could not create your storage',
    message: null,
    hideRetryMessage: false,
    description: null,
    onClose: () => onCloseErrorModal()
  }

  const [state, setState] = useState(initialState)
  const [submitModal, setSubmitModal] = useState(submitModalInitial)
  const [errorModal, setErrorModal] = useState(errorModalInitial)
  const [isPageReady, setIsPageReady] = useState(false)
  const [clusterToEdit, setClusterToEdit] = useState(null)
  const [storageUnitSize, setStorageUnitSize] = useState('')
  const [storageUsageUnit, setStorageUsageUnit] = useState('')
  const [costPerMonth, setCostPerMonth] = useState('')
  const [emptyCatalogModal, setEmptyCatalogModal] = useState(false)
  const throwError = useErrorBoundary()

  const fetchClusters = async (isBackground) => {
    try {
      await setClusters(isBackground)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  const fetchStorageProducts = async () => {
    try {
      await setStorages()
    } catch (error) {
      throwError(error)
    }
  }

  const fetchItems = async () => {
    if (clusters.length === 0) {
      await fetchClusters()
    }
    if (!storages) {
      await fetchStorageProducts()
    }
    setIsPageReady(true)
  }

  useEffect(() => {
    fetchItems()
    return () => {
      setClusterToEdit(null)
    }
  }, [])

  useEffect(() => {
    let shouldExit = false
    if (isPageReady) {
      if (clusters.length > 0) {
        const cluster = clusters.find((item) => item.name === name)
        if (cluster !== undefined) {
          setClusterToEdit(cluster)
        } else {
          shouldExit = true
        }
      } else {
        shouldExit = true
      }
    }
    if (shouldExit) {
      navigate({
        pathname: '/cluster'
      })
    }
  }, [isPageReady])

  useEffect(() => {
    if (clusterToEdit) {
      const storageItem = { ...clusterToEdit.storages[0] }
      const sizeValue = storageItem.size
      const size = sizeValue.replace('GB', '').replace('TB', '')
      const stateCopy = { ...state }
      const formUpdated = setFormValue('volumeSize', size, state.form)
      stateCopy.form = formUpdated
      stateCopy.isValidForm = true
      setState(stateCopy)
    }
  }, [clusterToEdit])

  useEffect(() => {
    if (storages === null) {
      return
    }
    if (storages.length > 0) {
      setStorageOptions()
    } else {
      setEmptyCatalogModal(true)
    }
  }, [storages])

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
    const element = formCopy.volumeSize

    element.maxLength = maximumSize.toString().length
    element.validationRules = {
      ...element.validationRules,
      checkMinValue: minimumSize,
      checkMaxValue: maximumSize
    }
    element.helperMessage = getCustomMessage('volumeSize', minimumSize, maximumSize, unitSize)
    element.label = `Storage Size (${unitSize}):`

    stateCopy.form = formCopy
    setState(stateCopy)
  }

  function goBack() {
    navigate({
      pathname: `/cluster/d/${name}`,
      search: 'tab=storage'
    })
  }

  function onCancel() {
    // Navigates back to the page when this method triggers.
    goBack()
  }

  function onChangeInput(event, formInputName) {
    let value = ''
    if (formInputName === 'fileVolumeFlag') {
      value = event.target.checked | ''
    } else {
      value = event.target.value
    }

    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    const storageSizeInput = updatedForm.volumeSize

    if (storageSizeInput.isValid) {
      const costPerHour = getPricing(storageSizeInput.value)
      setCostPerMonth(costPerHour)
    } else {
      setCostPerMonth('error')
    }

    updatedState.form = updatedForm

    setState(updatedState)
  }

  function getPricing(spaceInGb) {
    const storage = { ...storages[0] }
    const costPerMonth = storage.rate * Number(spaceInGb)
    return formatCurrency(formatNumber(costPerMonth, 2))
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

  async function onSubmit() {
    try {
      const isValidForm = state.isValidForm
      if (!isValidForm) {
        showRequiredFields()
        return
      }
      const payloadCopy = { ...payload }
      const volumeSizeValue = getFormValue('volumeSize', state.form)
      payloadCopy.storagesize = volumeSizeValue + storageUnitSize
      setSubmitModal({ ...submitModal, show: true })
      await ClusterService.updateStorage(payloadCopy, clusterToEdit.uuid)
      setTimeout(() => {
        setSubmitModal({ ...submitModal, show: false })
        setState(initialState)
        fetchClusters(false)
        goBack()
      }, 4000)
    } catch (error) {
      setSubmitModal({ ...submitModal, show: false })
      let errorMessage = ''
      if (error.response) {
        const errData = error.response.data
        const errCode = errData.code
        if (errCode) {
          errorMessage = error.response.data.message
        } else {
          errorMessage = error.message
        }
      } else {
        errorMessage = error.message
      }
      setErrorModal({ ...errorModal, show: true, message: errorMessage })
    }
  }

  function onCloseErrorModal() {
    setErrorModal(errorModalInitial)
  }

  return (
    <ClusterAddStorage
      navigationBottom={state.navigationBottom}
      form={state.form}
      mainTitle={state.mainTitle}
      loading={!isPageReady}
      onChangeInput={onChangeInput}
      onSubmit={onSubmit}
      submitModal={submitModal}
      errorModal={errorModal}
      storageUsageUnit={storageUsageUnit}
      costPerMonth={costPerMonth}
      emptyCatalogModal={emptyCatalogModal}
    />
  )
}

export default ClusterEditStorageContainer
