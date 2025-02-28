// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  setFormValue,
  setSelectOptions,
  showFormRequiredFields
} from '../../utils/updateFormHelper/UpdateFormHelper'
import { useNavigate, useParams } from 'react-router'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import CloudAccountService from '../../services/CloudAccountService'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import ComputeGroupsEditReservation from '../../components/compute-groups/computeGroupsEdit/ComputeGroupsEditReservation'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'

const getCustomMessage = (messageType) => {
  let message = null

  switch (messageType) {
    case 'runStrategy':
      message = (
        <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'}>
          No values available
        </div>
      )
      break
    case 'publicKey':
      message = (
        <p className="lead">
          A key consists of a public key that the application stores and a private key file that you store, allowing you
          to connect to your instance. The selected key in this step will be added to the set of keys authorized to this
          instance.
        </p>
      )
      break
  }

  return message
}

const ComputeGroupsEditReservationContainer = () => {
  const { param: name } = useParams()
  // local state
  const initialState = {
    mainTitle: `Edit instance group ${name}`,
    instanceDetailsMenuSection: 'Instance group details',
    publicKeysMenuSection: 'Public Keys',
    form: {
      instanceName: {
        sectionGroup: 'instanceDetails',
        type: 'text', // options = 'text ,'textArea'
        label: 'Group name:',
        placeholder: 'Group name',
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
        helperMessage: getCustomMessage('instanceDetails')
      },
      keyPairList: {
        sectionGroup: 'keys',
        type: 'multi-select', // options = 'text ,'textArea'
        label: 'Select keys:',
        placeholder: 'Please select',
        value: [], // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        validationRules: {
          isRequired: true
        },
        options: [], // Only required for select inputs, contains the seletable options
        validationMessage: '', // Errror message to display to the user
        extraButton: {
          label: '+ Upload Key',
          buttonFunction: () => onShowHidePublicKeyModal(true)
        },
        selectAllButton: {
          label: 'Select/Deselect All',
          buttonFunction: () => onSelectAllKeys()
        },
        emptyOptionsMessage: 'No keys found. Please upload a key to continue.'
      }
    },
    isValidForm: true,
    servicePayload: {
      spec: {
        instanceSpec: {
          sshPublicKeyNames: []
        }
      }
    },
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

  const throwError = useErrorBoundary()
  const [searchParams] = useSearchParams()

  const [state, setState] = useState(initialState)
  const [showPublicKeyModal, setShowPublicKeyModal] = useState(false)
  const [instanceGroupToEdit, setInstanceGroupToEdit] = useState(null)
  const [isPageReady, setIsPageReady] = useState(false)

  // Global State
  const publicKeys = useCloudAccountStore((state) => state.publicKeys)
  const setPublickeys = useCloudAccountStore((state) => state.setPublickeys)
  const instanceGroups = useCloudAccountStore((state) => state.instanceGroups)
  const setInstanceGroups = useCloudAccountStore((state) => state.setInstanceGroups)
  const loading = useCloudAccountStore((state) => state.loading)
  const showError = useToastStore((state) => state.showError)

  // Navigation
  const navigate = useNavigate()

  const refreshInstanceGroups = async (background) => {
    try {
      await setInstanceGroups(background)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  // Hooks
  useEffect(() => {
    if (instanceGroups.length === 0) {
      refreshInstanceGroups(false)
    } else {
      setIsPageReady(true)
    }

    if (publicKeys.length === 0) {
      const fetchKeys = async () => {
        try {
          await setPublickeys()
        } catch (error) {
          throwError(error)
        }
      }
      fetchKeys()
    }
  }, [])

  useEffect(() => {
    let shouldExit = false
    if (isPageReady) {
      if (instanceGroups.length > 0) {
        const instanceGroup = instanceGroups?.find((item) => item.name === name)
        if (instanceGroup !== undefined) {
          setInstanceGroupToEdit(instanceGroup)
        } else {
          shouldExit = true
        }
      } else {
        shouldExit = true
      }
    }

    if (shouldExit) {
      navigate({
        pathname: '/compute-groups'
      })
    }
  }, [isPageReady])

  useEffect(() => {
    if (instanceGroupToEdit !== null) {
      setForm()
    }
  }, [publicKeys, instanceGroupToEdit])

  // functions
  function setForm() {
    const stateUpdated = {
      ...state
    }

    if (instanceGroupToEdit) {
      stateUpdated.form = setFormValue('instanceName', instanceGroupToEdit.name, stateUpdated.form)
      stateUpdated.form = setFormValue('keyPairList', instanceGroupToEdit.sshPublicKey, stateUpdated.form)
    }
    const keys = publicKeys.map((key) => {
      return { name: key.name + ` (${key.ownerEmail})`, value: key.value }
    })
    stateUpdated.form.keyPairList.selectAllButton.buttonFunction = () => onSelectAllKeys()
    stateUpdated.form = setSelectOptions('keyPairList', keys, stateUpdated.form)

    setState(stateUpdated)
  }

  function onShowHidePublicKeyModal(status = false) {
    setShowPublicKeyModal(status)
  }

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  function onChangeDropdownMultiple(values) {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(values, 'keyPairList', updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  function onSelectAllKeys() {
    const stateUpdated = {
      ...state
    }

    const allKeys = publicKeys.map((key) => key.value)
    const selectedValues = stateUpdated.form.keyPairList.value
    const shouldDeselect = allKeys.every((x) => selectedValues.includes(x))

    onChangeDropdownMultiple(shouldDeselect ? [] : allKeys)
  }

  function afterPubliKeyCreate() {
    try {
      setPublickeys()
    } catch (error) {
      throwError(error)
    }
  }

  function goBack() {
    const backTo = searchParams.get('backTo')
    switch (backTo) {
      case 'detail':
        navigate({
          pathname: `/compute-groups/d/${name}`
        })
        break
      default:
        navigate({
          pathname: '/compute-groups'
        })
        break
    }
  }

  function onCancel() {
    goBack()
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
      if (!state.isValidForm) {
        showRequiredFields()
        return
      }
      const resourceName = instanceGroupToEdit.name
      const servicePayload = { ...state.servicePayload }
      const keys = getFormValue('keyPairList', state.form)
      servicePayload.spec.instanceSpec.sshPublicKeyNames = keys

      await CloudAccountService.putComputeGroupReservation(resourceName, servicePayload)
      refreshInstanceGroups(false)
      goBack()
    } catch (error) {
      let message = ''

      if (error.response) {
        message = error.response.data.message
      } else {
        message = error.message
      }

      showError(message)
    }
  }

  return (
    <ComputeGroupsEditReservation
      state={state}
      loading={loading || !isPageReady || !instanceGroupToEdit}
      showPublicKeyModal={showPublicKeyModal}
      afterPubliKeyCreate={afterPubliKeyCreate}
      onChangeDropdownMultiple={onChangeDropdownMultiple}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      onShowHidePublicKeyModal={onShowHidePublicKeyModal}
    />
  )
}

export default ComputeGroupsEditReservationContainer
