// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useSuperComputerStore from '../../store/superComputer/SuperComputerStore'
import { useNavigate, useParams } from 'react-router'
import SuperComputerAddStorage from '../../components/superComputer/superComputerAddStorage/SuperComputerAddStorage'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  showFormRequiredFields
} from '../../utils/updateFormHelper/UpdateFormHelper'
import { formatNumber } from '../../utils/numberFormatHelper/NumberFormatHelper'
import SuperComputerService from '../../services/SuperComputerService'
import useToastStore from '../../store/toastStore/ToastStore'
import { superComputerProductCatalogTypes, toastMessageEnum } from '../../utils/Enums'
import useErrorBoundary from '../../hooks/useErrorBoundary'

const getCustomMessage = (messageType, volume = 'Enter size to calculate.', cost = '0', minSize, maxSize) => {
  let message = null

  switch (messageType) {
    case 'volumeSize':
      message = (
        <div className="valid-feedback" intc-id={'volumeSize'}>
          Enter a value between {minSize} and {maxSize} <br />
          GB Volume cost per hour: <strong>{volume}</strong>
          <br />
          Volume cost per hour: <strong>$ {cost}</strong>
        </div>
      )
      break
    default:
      break
  }

  return message
}

const SuperComputerAddStorageContainer = () => {
  // *****
  // Global state
  // *****
  const loadingDetail = useSuperComputerStore((state) => state.loadingDetail)
  const clusterDetail = useSuperComputerStore((state) => state.clusterDetail)
  const setClusterDetail = useSuperComputerStore((state) => state.setClusterDetail)
  const loadingProducts = useSuperComputerStore((state) => state.loading)
  const fileStorage = useSuperComputerStore((state) => state.fileStorage)
  const setProducts = useSuperComputerStore((state) => state.setProducts)
  const showError = useToastStore((state) => state.showError)
  const throwError = useErrorBoundary()
  // *****
  // local state
  // *****
  const navigate = useNavigate()
  const { param: name } = useParams()

  const initialState = {
    mainTitle: `Add storage to ${name}`,
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
        label: 'Volume size (GB):',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 6,
        validationRules: {
          isRequired: true,
          checkMinValue: 5,
          checkMaxValue: 500000
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
  const [isPageReady, setIsPageReady] = useState(clusterDetail && clusterDetail.name === name)
  const [sizeUnit, setSizeUnit] = useState('')

  const fetchClusterDetail = async (isBackground) => {
    try {
      await setClusterDetail(name, isBackground)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  // *****
  // use effect
  // *****

  useEffect(() => {
    if (!isPageReady) {
      fetchClusterDetail(false)
    }

    if (fileStorage.length === 0) {
      const fetchProducts = async () => {
        try {
          await setProducts()
        } catch (error) {
          throwError(error)
        }
      }

      fetchProducts()
    }
  }, [])

  useEffect(() => {
    if (isPageReady && clusterDetail === null) {
      navigate('/supercomputer')
    }
  }, [clusterDetail, isPageReady])

  // *****
  // functions
  // *****
  function goBack() {
    navigate({
      pathname: `/supercomputer/d/${name}`,
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

    updatedState.form = updatedForm

    const storageGbCount = getFormValue('volumeSize', updatedForm)

    if (fileStorage.length > 0) {
      const storage = fileStorage.find(
        (item) => item.recommendedUseCase === superComputerProductCatalogTypes.fileStorage
      )
      const storageRate = Number(storage.rate)
      const minimumSize = storage.minimumSize
      const maximumSize = storage.maximumSize
      const unitSize = storage.unitSize
      setSizeUnit(unitSize)
      let storageHourlyCost = 0
      if (storageGbCount) {
        storageHourlyCost = storageRate * Number(storageGbCount)
      }
      updatedForm.volumeSize.validationRules = {
        ...updatedForm.volumeSize.validationRules,
        checkMinValue: minimumSize,
        checkMaxValue: maximumSize
      }
      updatedForm.volumeSize.helperMessage = getCustomMessage(
        'volumeSize',
        formatNumber(storageHourlyCost, 2),
        storageRate,
        minimumSize,
        maximumSize
      )
    }
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

  async function onSubmit() {
    try {
      const isValidForm = state.isValidForm
      if (!isValidForm) {
        showRequiredFields()
        return
      }
      const payloadCopy = { ...payload }
      const volumeSizeValue = getFormValue('volumeSize', state.form)
      payloadCopy.storagesize = volumeSizeValue + sizeUnit
      setSubmitModal({ ...submitModal, show: true })
      await SuperComputerService.createStorage(payloadCopy, clusterDetail.uuid)
      setTimeout(() => {
        setSubmitModal({ ...submitModal, show: false })
        setState(initialState)
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
    <SuperComputerAddStorage
      loading={loadingDetail || loadingProducts || !isPageReady}
      navigationBottom={state.navigationBottom}
      form={state.form}
      mainTitle={state.mainTitle}
      isValidForm={state.isValidForm}
      onChangeInput={onChangeInput}
      onSubmit={onSubmit}
      submitModal={submitModal}
      errorModal={errorModal}
    />
  )
}

export default SuperComputerAddStorageContainer
