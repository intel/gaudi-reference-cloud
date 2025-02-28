// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React, { useState, useEffect } from 'react'
import StorageAddQuota from '../../components/storageManagement/storageAddQuota/StorageAddQuota'
import { useNavigate } from 'react-router-dom'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  setFormValue,
  markErrorOnElement,
  showFormRequiredFields
} from '../../utility/updateFormHelper/UpdateFormHelper'
import StorageManagementService from '../../services/StorageManagementService'
import CloudAccountService from '../../services/CloudAccountService'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utility/Enums'
import useStorageManagementStore from '../../store/storageManagementStore/StorageManagementStore'

const AddQuotaContainer = (): JSX.Element => {
  // *****
  // Global state
  // *****
  const showError = useToastStore((state) => state.showError)
  const editStorageQuota = useStorageManagementStore((state) => state.editStorageQuota)
  const setEditStorageQuota = useStorageManagementStore((state) => state.setEditStorageQuota)

  // *****
  // local state
  // *****
  const initialState = {
    title: 'Manage Storage Quota',
    form: {
      cloudAccount: {
        sectionGroup: 'cloudAccount',
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Cloud Account:',
        placeholder: 'Enter Cloud Account Id',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: <div className="valid-feedback" intc-id={'cloudAccount'}>
          Enter a the user cloudaccount
        </div>
      },
      reason: {
        sectionGroup: 'configuration',
        type: 'textArea', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Reason:',
        placeholder: 'Enter justification',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        validationMessage: '' // Errror message to display to the user
      },
      filesizeQuota: {
        sectionGroup: 'configuration',
        type: 'integer', // options = 'text ,'textArea'
        label: 'File System quota size (TB):',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 6,
        validationRules: {
          isRequired: true,
          checkMinValue: 1,
          checkMaxValue: 1000
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: '',
        subGroup: 'storage',
        hidden: false
      },
      volumeQuota: {
        sectionGroup: 'configuration',
        type: 'integer', // options = 'text ,'textArea'
        label: 'Volumes quota:',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 6,
        validationRules: {
          isRequired: true,
          checkMinValue: 1,
          checkMaxValue: 1000
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: '',
        subGroup: 'storage',
        hidden: false
      },
      bucketQuota: {
        sectionGroup: 'configuration',
        type: 'integer', // options = 'text ,'textArea'
        label: 'Buckets quota:',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 6,
        validationRules: {
          isRequired: true,
          checkMinValue: 1,
          checkMaxValue: 1000
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: '',
        subGroup: 'storage',
        hidden: false
      }
    },
    isValidForm: false,
    servicePayload: {
      cloudAccountId: '',
      reason: '',
      filesizeQuotaInTB: '',
      filevolumesQuota: '',
      bucketsQuota: ''
    },
    navigationBottom: [
      {
        buttonLabel: 'Request',
        buttonVariant: 'primary'
      },
      {
        buttonLabel: 'Cancel',
        buttonVariant: 'link',
        buttonFunction: () => {
          onCancel()
        }
      }
    ]
  }

  const cloudAccountSelectedInitial = {
    show: false,
    name: '',
    id: '',
    type: '',
    defaultQuota: {},
    updatedQuota: {}
  }

  const searchModalInitial = {
    show: false,
    message: ''
  }

  const navigate = useNavigate()
  const [state, setState] = useState(initialState)
  const [cloudAccountSelected, setCloudAccountSelected] = useState(cloudAccountSelectedInitial)
  const [searchModal, setSearchModal] = useState(searchModalInitial)
  const [editMode, setEditMode] = useState(false)
  // *****
  // Hooks
  // *****
  useEffect(() => {
    if (editStorageQuota) {
      const fetch = async (): Promise<void> => {
        await loadEditForm()
      }
      fetch().catch((error) => {
        showError(error, false)
      })
    }
    return () => {
      setEditStorageQuota(null)
    }
  }, [editStorageQuota])

  // *****
  // Functions
  // *****
  const loadEditForm = async (): Promise<void> => {
    if (editStorageQuota) {
      const updateState = { ...state }
      await getCloudAccountInfo(editStorageQuota?.cloudAccountId)

      updateState.form = setFormValue('cloudAccount', editStorageQuota?.cloudAccountId, updateState.form)
      updateState.form = setFormValue('reason', editStorageQuota?.reason, updateState.form)
      updateState.form = setFormValue('bucketQuota', Number(editStorageQuota?.bucketsQuota), updateState.form)
      updateState.form = setFormValue('volumeQuota', Number(editStorageQuota?.filevolumesQuota), updateState.form)
      updateState.form = setFormValue('filesizeQuota', Number(editStorageQuota?.filesizeQuotaInTB), updateState.form)
      setState(updateState)
      setEditMode(true)
    }
  }

  const onChangeInput = (event: any, formInputName: string): void => {
    const stateUpdated = {
      ...state
    }
    const value = event.target.value
    const updatedForm = UpdateFormHelper(value, formInputName, stateUpdated.form)
    stateUpdated.isValidForm = isValidForm(updatedForm)
    stateUpdated.form = updatedForm

    if (formInputName === 'cloudAccount') {
      setCloudAccountSelected(cloudAccountSelectedInitial)
    }

    setState(stateUpdated)
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
      const quotaResponse = await StorageManagementService.getQuotaById(id)
      const dataQuota = quotaResponse.data
      const defaultQuota = dataQuota.DefaultQuota
      const updatedQuota = dataQuota.UpdatedQuota
      setCloudAccountSelected({
        ...cloudAccountSelected,
        show: true,
        id,
        name,
        type,
        defaultQuota,
        updatedQuota
      })
      if (dataQuota.UpdatedQuota) {
        setEditMode(true)
      }
      setSearchModal(searchModalInitial)
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

  const showRequiredFields = async (): Promise<void> => {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(stateCopy.form)
    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    stateCopy.form = updatedForm
    setState(stateCopy)
  }

  const onSubmit = async (): Promise<void> => {
    const stateCopy = { ...state }
    const isValidForm = stateCopy.isValidForm
    if (!isValidForm) {
      void showRequiredFields()
      return
    }

    if (!cloudAccountSelected.id) {
      const stateCopy = { ...state }
      const updatedForm = markErrorOnElement(state.form, 'cloudAccount', 'Cloud Account: is not found')
      stateCopy.form = updatedForm
      setState(stateCopy)
      return
    }
    const paylodCopy = { ...state.servicePayload }
    paylodCopy.cloudAccountId = cloudAccountSelected.id
    paylodCopy.reason = getFormValue('reason', state.form)
    paylodCopy.filesizeQuotaInTB = getFormValue('filesizeQuota', state.form)
    paylodCopy.filevolumesQuota = getFormValue('volumeQuota', state.form)
    paylodCopy.bucketsQuota = getFormValue('bucketQuota', state.form)
    await createUpdateStorage(paylodCopy)
  }

  const createUpdateStorage = async (paylodCopy: any): Promise<void> => {
    try {
      setSearchModal({ ...searchModal, show: true, message: 'Working on your request' })
      if (editMode) {
        await StorageManagementService.putUsages(paylodCopy)
      } else {
        await StorageManagementService.postUsages(paylodCopy)
      }
      setSearchModal(searchModalInitial)
      navigate(-1)
    } catch (error: any) {
      let message = ''
      if (error?.response?.data?.message) {
        if (error?.response?.data?.code === 13) {
          if (cloudAccountSelected.updatedQuota) {
            message = 'Quota cannot be decreased'
          } else {
            message = 'The new quota cannot be smaller than the default quota.'
          }
        } else {
          message = error.response.data.message
        }
      } else {
        message = error.message
      }
      setSearchModal(searchModalInitial)
      showError(message, false)
    }
  }

  const onCancel = (): void => {
    navigate(-1)
  }

  return (
    <StorageAddQuota
      state={state}
      searchModal={searchModal}
      onChangeInput={onChangeInput}
      onSearchCloudAccount={onSearchCloudAccount}
      cloudAccountSelected={cloudAccountSelected}
      editMode={editMode}
      onSubmit={onSubmit}
      onCancel={onCancel}
    />
  )
}

export default AddQuotaContainer
