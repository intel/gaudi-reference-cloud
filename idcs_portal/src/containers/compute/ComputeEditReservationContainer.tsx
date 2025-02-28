// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import ComputeEditReservation from '../../components/compute/computeEdit/ComputeEditReservation'
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
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'

type InstanceLabelsInterface = Record<string, string>

const getCustomMessage = (messageType: string, defaultLabels: any[] = []): JSX.Element | null => {
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
    case 'instancesLabels':
      message = (
        <p>
          Reserved words: <strong>{defaultLabels.join(', ')}</strong>
        </p>
      )
      break
  }

  return message
}

const ComputeEditReservationContainer = (): JSX.Element => {
  const { param: name } = useParams()
  // local state
  const instanceLabelsOption = {
    key: {
      label: 'Key:',
      value: '', // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false,
      maxLength: 50,
      validationRules: {
        isRequired: false
      },
      validationMessage: 'Key is required'
    },
    value: {
      label: 'Value:',
      value: '', // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false,
      maxLength: 50,
      validationRules: {
        isRequired: false
      },
      validationMessage: 'Value is required'
    }
  }

  const defaultLabels = ['clusterName', 'nodegroupName', 'nodegroupType', 'ipAddress']

  const initialState = {
    mainTitle: `Edit instance ${name}`,
    instanceDetailsMenuSection: 'Instance details',
    publicKeysMenuSection: 'Public Keys',
    instancLabelsMenuSection: isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_COMPUTE_EDIT_LABELS)
      ? 'Instance Tags'
      : '',
    form: {
      instanceName: {
        sectionGroup: 'instanceDetails',
        type: 'text', // options = 'text ,'textArea'
        label: 'Instance name:',
        placeholder: 'Instance name',
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
          buttonFunction: () => {
            onShowHidePublicKeyModal(true)
          }
        },
        selectAllButton: {
          label: 'Select/Deselect All',
          buttonFunction: () => {
            onSelectAllKeys()
          }
        },
        emptyOptionsMessage: 'No keys found. Please upload a key to continue.'
      },
      instancesLabels: {
        sectionGroup: 'instancesLabels',
        type: 'dictionary', // options = 'text ,'textArea'
        label: '',
        placeholder: '',
        value: '',
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        hidden: !isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_COMPUTE_EDIT_LABELS),
        validationRules: {
          isRequired: false
        },
        dictionaryOptions: [instanceLabelsOption],
        validationMessage: '',
        maxLength: 20,
        helperMessage: getCustomMessage('instancesLabels', defaultLabels)
      }
    },
    isValidForm: true,
    servicePayload: {
      metadata: {
        labels: {}
      },
      spec: {
        sshPublicKeyNames: []
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
        buttonFunction: () => {
          onCancel()
        }
      }
    ]
  }

  const throwError = useErrorBoundary()
  const [searchParams] = useSearchParams()

  const [state, setState] = useState(initialState)
  const [showPublicKeyModal, setShowPublicKeyModal] = useState(false)
  const [showUpdateInstanceSshModal, setShowUpdateInstanceSshModal] = useState(false)
  const [instanceToEdit, setInstanceToEdit] = useState<any>(null)
  const [isPageReady, setIsPageReady] = useState(false)

  // Global State
  const publicKeys = useCloudAccountStore((state) => state.publicKeys)
  const setPublickeys = useCloudAccountStore((state) => state.setPublickeys)
  const instances = useCloudAccountStore((state) => state.instances)
  const setInstances = useCloudAccountStore((state) => state.setInstances)
  const loading = useCloudAccountStore((state) => state.loading)
  const showSuccess = useToastStore((state) => state.showSuccess)
  const showError = useToastStore((state) => state.showError)

  // Navigation
  const navigate = useNavigate()

  const refreshInstances = async (background: boolean): Promise<void> => {
    try {
      await setInstances(background)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  // Hooks
  useEffect(() => {
    if (instances.length === 0) {
      void refreshInstances(false)
    } else {
      setIsPageReady(true)
    }

    if (publicKeys.length === 0) {
      const fetchKeys = async (): Promise<void> => {
        try {
          await setPublickeys()
        } catch (error) {
          throwError(error)
        }
      }
      void fetchKeys()
    }
  }, [])

  useEffect(() => {
    let shouldExit = false
    if (isPageReady) {
      if (instances.length > 0) {
        const instance = instances?.find((item) => item.name === name)
        if (instance !== undefined) {
          setInstanceToEdit(instance)
        } else {
          shouldExit = true
        }
      } else {
        shouldExit = true
      }
    }

    if (shouldExit) {
      navigate({
        pathname: '/compute'
      })
    }
  }, [isPageReady])

  useEffect(() => {
    if (instanceToEdit !== null) {
      setForm()
    }
  }, [publicKeys, instanceToEdit])

  // functions
  const setForm = (): void => {
    const stateUpdated = {
      ...state
    }

    if (instanceToEdit && publicKeys) {
      const validPublicKeys = instanceToEdit.sshPublicKey.filter((x: any) => publicKeys.map((x) => x.name).includes(x))
      stateUpdated.form = setFormValue('instanceName', instanceToEdit.name, stateUpdated.form)
      stateUpdated.form = setFormValue('keyPairList', validPublicKeys, stateUpdated.form)

      const keys = publicKeys.map((key) => {
        return { name: key.name + ` (${key.ownerEmail})`, value: key.value }
      })
      stateUpdated.form.keyPairList.selectAllButton.buttonFunction = () => {
        onSelectAllKeys()
      }
      stateUpdated.form = setSelectOptions('keyPairList', keys, stateUpdated.form)
      stateUpdated.form.keyPairList.isValid = keys.length !== 0

      // Labels
      const instanceLabels = instanceToEdit.labels
      const tagKeys = Object.keys(instanceLabels)
      const tagValues = Object.values(instanceLabels)

      for (let i = 0; i < tagValues.length; i++) {
        if (!stateUpdated.form.instancesLabels.dictionaryOptions[i]) {
          stateUpdated.form.instancesLabels.dictionaryOptions[i] = structuredClone(instanceLabelsOption)
        }

        stateUpdated.form.instancesLabels.dictionaryOptions[i].key.value = tagKeys[i]
        stateUpdated.form.instancesLabels.dictionaryOptions[i].key.isValid = true
        stateUpdated.form.instancesLabels.dictionaryOptions[i].value.value = String(tagValues[i])
        stateUpdated.form.instancesLabels.dictionaryOptions[i].value.isValid = true

        if (defaultLabels.includes(tagKeys[i])) {
          stateUpdated.form.instancesLabels.dictionaryOptions[i].key.isReadOnly = true
          stateUpdated.form.instancesLabels.dictionaryOptions[i].value.isReadOnly = true
        }
      }

      stateUpdated.form.instancesLabels.isValid = tagKeys.length === tagValues.length
    }

    stateUpdated.isValidForm = isValidForm(stateUpdated.form)

    setState(stateUpdated)
  }

  const onShowHidePublicKeyModal = (status = false): void => {
    setShowPublicKeyModal(status)
  }

  const onShowHideUpdateInstanceSshModal = (status = false): void => {
    setShowUpdateInstanceSshModal(status)

    if (!status) {
      void refreshInstances(false)
      goBack()

      showSuccess('Instance updated successfully', false)
    }
  }

  const onChangeInput = (event: any, formInputName: string): void => {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  const onChangeDropdownMultiple = (values: any): void => {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(values, 'keyPairList', updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  const onSelectAllKeys = (): void => {
    const stateUpdated = {
      ...state
    }

    const allKeys = publicKeys.map((key) => key.value)
    const selectedValues: string[] = stateUpdated.form.keyPairList.value
    const shouldDeselect = allKeys.every((x: string) => selectedValues.includes(x))

    onChangeDropdownMultiple(shouldDeselect ? [] : allKeys)
  }

  const afterPubliKeyCreate = (): void => {
    try {
      void setPublickeys()
    } catch (error) {
      throwError(error)
    }
  }

  const goBack = (): void => {
    const backTo = searchParams.get('backTo')
    switch (backTo) {
      case 'detail':
        navigate({
          pathname: `/compute/d/${name}`
        })
        break
      default:
        navigate({
          pathname: '/compute'
        })
        break
    }
  }

  const onCancel = (): void => {
    goBack()
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
      if (!state.isValidForm) {
        showRequiredFields()
        return
      }

      const getInstanceLabels = (instancesLabels: any): InstanceLabelsInterface => {
        const payloadInstanceLabels: any = {}
        for (const label of instancesLabels.dictionaryOptions) {
          const key = getFormValue('key', label)
          const value = getFormValue('value', label)
          if (key !== '' && value !== '') payloadInstanceLabels[key] = value
        }

        return payloadInstanceLabels
      }

      const resourceId = instanceToEdit?.resourceId
      const servicePayload = { ...state.servicePayload }

      if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_COMPUTE_EDIT_LABELS)) {
        servicePayload.metadata.labels = getInstanceLabels(state.form.instancesLabels)
      } else {
        delete (servicePayload as any).metadata
      }

      const keys = getFormValue('keyPairList', state.form)
      servicePayload.spec.sshPublicKeyNames = keys

      await CloudAccountService.putComputeReservation(resourceId, servicePayload)

      if (keys.every((x: any) => instanceToEdit?.sshPublicKey.includes(x))) {
        onShowHideUpdateInstanceSshModal(false)
      } else {
        onShowHideUpdateInstanceSshModal(true)
      }
    } catch (error: any) {
      let message = ''

      if (error.response) {
        message = error.response.data.message
      } else {
        message = error.message
      }

      showError(message, false)
    }
  }

  const isTagValidValue = (value: string): boolean => {
    return value !== '' && !defaultLabels.includes(value)
  }

  const onChangeTagValue = (event: any, formInputName: string, tagIndex: number): void => {
    const value = event.target.value
    const updatedState = { ...state }

    const tagOptions = [...updatedState.form.instancesLabels.dictionaryOptions]
    const tagOption = tagOptions[tagIndex]
    const tagOptionForm = UpdateFormHelper(value, formInputName, tagOption)

    tagOptionForm.key.isTouched = true
    tagOptionForm.value.isTouched = true
    tagOptionForm.key.isValid = true
    tagOptionForm.value.isValid = true

    let inputKey = 'key'
    if (formInputName === 'key') {
      inputKey = 'value'
    }

    tagOptionForm[formInputName].isValid = isTagValidValue(tagOptionForm[formInputName].value)
    tagOptionForm[inputKey].isValid = isTagValidValue(tagOptionForm[inputKey].value)

    if (tagOptionForm[formInputName].value === '' && tagOptionForm[inputKey].value === '') {
      tagOptionForm[inputKey].isValid = false
      tagOptionForm[formInputName].isValid = false
      tagOptionForm[inputKey].isTouched = false
      tagOptionForm[formInputName].isTouched = false
    }

    if (formInputName === 'key' && tagOptionForm[formInputName].value !== '' && !tagOptionForm[formInputName].isValid) {
      tagOptionForm[formInputName].isValid = false
      tagOptionForm[formInputName].validationMessage = 'Reserved words not allowed.'
    }

    if (formInputName === 'value' && tagOptionForm[inputKey].value !== '' && !tagOptionForm[inputKey].isValid) {
      tagOptionForm[inputKey].isValid = false
      tagOptionForm[inputKey].validationMessage = 'Reserved words not allowed.'
    }

    tagOptions[tagIndex] = tagOptionForm
    updatedState.form.instancesLabels.dictionaryOptions = tagOptions
    updatedState.form.instancesLabels.isValid = checkIsFormValid(tagOptions)

    updatedState.isValidForm = isValidForm(updatedState.form)

    setState(updatedState)
  }

  const checkIsFormValid = (formsItems: any): boolean => {
    let isValid = true

    for (const formItem of formsItems) {
      const isFormItemValid = isValidForm(formItem)
      if (!isFormItemValid) {
        isValid = false
        break
      }
    }
    return isValid
  }

  const onClickActionTag = (tagIndex: number, actionType: string): void => {
    const updatedState = { ...state }

    if (actionType === 'Add') {
      updatedState.form.instancesLabels.dictionaryOptions.push(instanceLabelsOption)
    }

    if (actionType === 'Delete') {
      const options = [...updatedState.form.instancesLabels.dictionaryOptions]
      options.splice(tagIndex, 1)
      updatedState.form.instancesLabels.dictionaryOptions = options
    }

    updatedState.form.instancesLabels.isValid = checkIsFormValid(updatedState.form.instancesLabels.dictionaryOptions)

    updatedState.isValidForm = isValidForm(updatedState.form)

    setState(updatedState)
  }

  return (
    <ComputeEditReservation
      state={state}
      showPublicKeyModal={showPublicKeyModal}
      showUpdateInstanceSshModal={showUpdateInstanceSshModal}
      loading={loading || !isPageReady || !instanceToEdit}
      updateKeysData={{
        instanceData: instanceToEdit,
        keysData: publicKeys.filter((x) => getFormValue('keyPairList', state.form).includes(x.name))
      }}
      afterPubliKeyCreate={afterPubliKeyCreate}
      onChangeDropdownMultiple={onChangeDropdownMultiple}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      onShowHidePublicKeyModal={onShowHidePublicKeyModal}
      onShowHideUpdateInstanceSshModal={onShowHideUpdateInstanceSshModal}
      onChangeTagValue={onChangeTagValue}
      onClickActionTag={onClickActionTag}
    />
  )
}

export default ComputeEditReservationContainer
