// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  setSelectOptions,
  setFormValue,
  showFormRequiredFields,
  validateInput
} from '../../utils/updateFormHelper/UpdateFormHelper'

import useErrorBoundary from '../../hooks/useErrorBoundary'

import { isErrorInsufficientCapacity, isErrorInsufficientCredits } from '../../utils/apiError/apiError'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import LoadBalancerLaunch from '../../components/loadBalancer/LoadBalancerLaunch'
import LoadBalancerService from '../../services/LoadBalancerService'
import useLoadBalancerStore from '../../store/loadBalancerStore/LoadBalancerStore'
import useToastStore from '../../store/toastStore/ToastStore'
import { toastMessageEnum } from '../../utils/Enums'

interface ServicePayloadInterface {
  metadata: PayloadMetadataInterface
  spec: PayloadSpecInterface
}

interface PayloadMetadataInterface {
  name: string
}

interface PayloadSpecInterface {
  listeners: PayloadListenersInterface[]
  security: PayloadSecurityInterface
}

interface PayloadSecurityInterface {
  sourceips: string[]
}

interface PayloadListenersInterface {
  port: string
  pool: PayloadPoolInstanceSelectorsInterface | PayloadPoolInstanceResourceIdsInterface
}

interface PayloadPoolInstanceSelectorsInterface {
  port: string
  monitor: string
  loadBalancingMode: string
  instanceSelectors: PayloadInstanceSelectorsInterface
}

interface PayloadPoolInstanceResourceIdsInterface {
  port: string
  monitor: string
  loadBalancingMode: string
  instanceResourceIds: string[]
}

type PayloadInstanceSelectorsInterface = Record<string, string>

const getCustomMessage = (messageType: string): JSX.Element | null => {
  let message = null

  switch (messageType) {
    case 'name':
      message = (
        <div className="valid-feedback" intc-id="LoadBalancerNameValidMessage">
          Max length 50 characters. Letters, numbers and ‘- ‘ accepted.
          <br />
          Name should start and end with an alphanumeric character.
        </div>
      )
      break
    case 'ips':
      message = (
        <div className="valid-feedback" intc-id="LoadBalanceripsValidMessage">
          Use “any” to allow access from anywhere. Specify a single IP (ex.: 10.0.0.1) or CIDR-format (ex.: 10.0.0.1/24)
        </div>
      )

      break
    default:
      break
  }

  return message
}

const LoadBalancerLaunchContainer = (): JSX.Element => {
  // local state

  const instancesAllowedStatus = ['Ready']

  const monitorTypes = [
    {
      name: 'TCP',
      value: 'tcp'
    },
    {
      name: 'HTTP',
      value: 'http'
    },
    {
      name: 'HTTPS',
      value: 'https'
    }
  ]

  const loadBalancerModes = [
    {
      name: 'Round Robin',
      value: 'roundRobin'
    }
  ]

  const instanceLabelsOption = {
    key: {
      label: 'Key:',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      maxLength: 50,
      validationRules: {
        isRequired: true
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: getCustomMessage('externalPort')
    },
    value: {
      label: 'Value:',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      maxLength: 50,
      validationRules: {
        isRequired: true
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: getCustomMessage('externalPort')
    }
  }

  const listenersIntials = {
    externalPort: {
      sectionGroup: 'listeners1',
      type: 'text', // options = 'text ,'textArea'
      maxWidth: '10rem',
      label: 'Listener Port:',
      placeholder: 'e.g. 80',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 5,
      validationRules: {
        isRequired: true,
        onlyCreditNumeric: true,
        checkMaxLength: true,
        checkMinValue: 1,
        checkMaxValue: 65535
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: getCustomMessage('externalPort')
    },
    internalPort: {
      sectionGroup: 'listeners1',
      type: 'text', // options = 'text ,'textArea'
      maxWidth: '10rem',
      label: 'Instance Port:',
      placeholder: 'e.g. 80',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 5,
      validationRules: {
        isRequired: true,
        onlyCreditNumeric: true,
        checkMaxLength: true,
        checkMinValue: 1,
        checkMaxValue: 65535
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: getCustomMessage('internalPort')
    },
    monitorType: {
      sectionGroup: 'listeners1',
      type: 'dropdown', // options = 'text ,'textArea'
      label: 'Monitor type:',
      placeholder: 'Select type',
      value: '',
      isValid: false,
      isTouched: false,
      isReadOnly: false,
      validationRules: {
        isRequired: true
      },
      maxWidth: '20rem',
      options: monitorTypes,
      validationMessage: '',
      helperMessage: getCustomMessage('monitorType')
    },
    loadBalancingMode: {
      sectionGroup: 'listeners2',
      type: 'dropdown', // options = 'text ,'textArea'
      label: 'Mode:',
      placeholder: 'Select mode',
      value: '',
      isValid: false,
      isTouched: false,
      isReadOnly: false,
      validationRules: {
        isRequired: true
      },
      options: [],
      validationMessage: '',
      helperMessage: getCustomMessage('loadBalancingMode')
    },
    instanceSelectors: {
      sectionGroup: 'pool',
      type: 'radio', // options = 'text ,'textArea'
      label: 'Selector type:',
      placeholder: '',
      value: 'instances', // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isChecked: false,
      radioGroupHorizontal: true,
      validationRules: {
        isRequired: false
      },
      options: [
        {
          name: 'Instance Labels',
          value: 'labels',
          default: false,
          disabled: true,
          hidden: true
        },
        {
          name: 'Instances',
          value: 'instances',
          default: true
        }
      ],
      validationMessage: '', // Errror message to display to the user
      helperMessage: ''
    },
    instancesLabels: {
      sectionGroup: 'pool',
      type: 'dictionary', // options = 'text ,'textArea'
      label: 'Instances tags:',
      placeholder: 'Select Instances',
      value: '',
      isValid: false,
      isTouched: false,
      isReadOnly: false,
      validationRules: {
        isRequired: false
      },
      dictionaryOptions: [instanceLabelsOption],
      validationMessage: '',
      maxLength: 20,
      helperMessage: getCustomMessage('instances'),
      hidden: true
    },
    instances: {
      sectionGroup: 'pool',
      type: 'multi-select', // options = 'text ,'textArea'
      label: 'Instances:',
      placeholder: 'Select Instances',
      value: [],
      isValid: false,
      isTouched: false,
      isReadOnly: false,
      borderlessDropdownMultiple: true,
      validationRules: {
        isRequired: true
      },
      options: [],
      validationMessage: '',
      helperMessage: getCustomMessage('instances'),
      hidden: false,
      emptyOptionsMessage: 'No instances found. Please create an instance to continue.'
    }
  }

  const initialState = {
    showReservationModal: false,
    showErrorModal: false,
    errorMessage: '',
    errorHideRetryMessage: null,
    errorDescription: null,
    navigationBottom: [
      {
        buttonLabel: 'Launch',
        buttonVariant: 'primary'
      },
      {
        buttonLabel: 'Cancel',
        buttonVariant: 'link',
        buttonFunction: () => {
          onCancel()
        }
      }
    ],
    timeoutMiliseconds: 4000
  }

  const ipInput = {
    ip: {
      sectionGroup: 'configuration',
      type: 'text', // options = 'text ,'textArea'
      label: 'Source IP:',
      placeholder: 'e.g. 10.0.0.1 or 10.0.0.1/24 or any',
      value: '', // Value enter by the user
      isValid: false, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 50,
      validationRules: {
        isRequired: true,
        checkMaxLength: true,
        isLoadBalancerSourceIP: true
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: getCustomMessage('ips')
    }
  }

  const initialForm = {
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
      helperMessage: getCustomMessage('name')
    },

    ips: {
      items: [{ ...ipInput }],
      isValid: false
    },
    listeners: { items: [{ ...listenersIntials }], isValid: false }
  }

  const [state, setState] = useState(initialState)
  const [form, setForm] = useState(initialForm)
  const [isFormValid, setIsFormValid] = useState(false)

  const [emptyCatalogModal, setEmptyCatalogModal] = useState(false)
  const [showUpgradeNeededModal, setShowUpgradeNeededModal] = useState(false)
  const [maxListeners, setMaxListeners] = useState('')
  const [maxSourceIps, setMaxSourceIps] = useState('')

  // Store
  const instances = useCloudAccountStore((state) => state.instances)
  const setInstances = useCloudAccountStore((state) => state.setInstances)
  const networkProducts = useLoadBalancerStore((state) => state.networkProducts)
  const setNetworkProducts = useLoadBalancerStore((state) => state.setNetworkProducts)
  const productLoading = useLoadBalancerStore((state) => state.productLoading)
  const showError = useToastStore((state) => state.showError)

  const throwError = useErrorBoundary()

  // Hooks
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        await setNetworkProducts()
        if (instances?.length === 0) await setInstances(false)
      } catch (error) {
        throwError(error)
      }
    }

    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    setFormValues()
  }, [instances])

  useEffect(() => {
    if (networkProducts.length > 0) {
      setLoadBalancerOptions()
    }

    if (!productLoading && networkProducts.length < 1) {
      setEmptyCatalogModal(true)
    }
  }, [networkProducts])

  // Navigation
  const navigate = useNavigate()

  // Functions

  const onCancel = (): void => {
    // Navigates back to the page when this method triggers.
    navigate('/load-balancer')
  }

  const setLoadBalancerOptions = (): void => {
    setMaxListeners(networkProducts[0].maxListeners)
    setMaxSourceIps(networkProducts[0].maxSourceIps)
    setEmptyCatalogModal(false)
  }

  const setFormValues = (): void => {
    const updatedFormState = { ...form }
    let listenersIntialsUpdate = { ...listenersIntials }
    const instancesValues = setInstanceOption()

    if (instancesValues.length > 0) {
      listenersIntialsUpdate = setSelectOptions('instances', instancesValues, listenersIntials)
    }
    listenersIntialsUpdate = setSelectOptions('loadBalancingMode', loadBalancerModes, listenersIntials)
    listenersIntialsUpdate = setFormValue('loadBalancingMode', loadBalancerModes[0].value, listenersIntials)

    updatedFormState.listeners.items[0] = listenersIntialsUpdate

    setForm(updatedFormState)
  }

  const setInstanceOption = (): any[] => {
    let instancesValues: any[] = []

    if (instances.length > 0) {
      instancesValues = instances
        .filter((x) => x.nodegroupType !== 'worker' && instancesAllowedStatus.includes(x.status))
        .map((instance) => {
          const instanceTypeDetails = instance.instanceTypeDetails
          const displayName = instanceTypeDetails ? instanceTypeDetails.displayName : ''
          return { name: instance.name + ` (${displayName})`, value: instance.resourceId }
        })
    }
    return instancesValues
  }

  const onChangeInput = (event: any, formInputName: string, key: number, listenersItemName: string): void => {
    const value = event.target.value
    const updatedFormState = { ...form }

    let updatedForm = updatedFormState

    if (listenersItemName === 'listeners') {
      const listeners = updatedForm.listeners
      const listenersItems = [...listeners.items]
      const listenersItem = listenersItems[key]
      const listenerItemForm = UpdateFormHelper(value, formInputName, listenersItem)
      listenersItems[key] = listenerItemForm

      if (formInputName === 'instanceSelectors') {
        if (value === 'labels') {
          listenersItem.instances.hidden = true
          listenersItem.instances.isValid = true
          listenersItem.instances.validationRules.isRequired = false
          listenersItem.instancesLabels.hidden = false
          listenersItem.instancesLabels.validationRules.isRequired = true
          listenersItem.instancesLabels.isValid = checkIsFormValid(listenersItem.instancesLabels.dictionaryOptions)
        } else {
          if (listenersItem.instances.value.length > 0) {
            listenersItem.instances.isTouched = true
            listenersItem.instances.isValid = true
          } else {
            listenersItem.instances.isTouched = false
            listenersItem.instances.isValid = false
          }

          listenersItem.instances.hidden = false
          listenersItem.instances.validationRules.isRequired = true

          listenersItem.instancesLabels.hidden = true
          listenersItem.instancesLabels.validationRules.isRequired = false
          listenersItem.instancesLabels.isValid = true
        }
      }

      if (formInputName === 'externalPort' && listenerItemForm[formInputName].isValid && listenersItems.length > 1) {
        const allExternalPorts = listenersItems.map((x) => x.externalPort.value)

        const duplicateValue = allExternalPorts.filter((x) => x === value)

        if (duplicateValue.length > 1) {
          listenersItems[key][formInputName].isValid = false
          listenersItems[key][formInputName].validationMessage = 'Duplicate port'
        }
      }

      updatedForm.listeners.items = listenersItems
      updatedForm.listeners.isValid = checkIsFormValid(listenersItems)
    } else {
      if (formInputName === 'ip') {
        const ipsItems = [...updatedForm.ips.items]
        const ipItemForm = UpdateFormHelper(value, formInputName, ipsItems[key])
        ipsItems[key] = ipItemForm

        if (ipItemForm[formInputName].isValid && updatedForm.ips.items.length > 1) {
          const allValues = updatedForm.ips.items.map((x) => x.ip.value)

          const duplicateValue = allValues.filter((x) => x === value)

          if (duplicateValue.length > 1) {
            ipItemForm[formInputName].isValid = false
            ipItemForm[formInputName].validationMessage = 'Duplicate IP'
          }
        }

        updatedForm.ips.items = ipsItems
        updatedForm.ips.isValid = checkIsFormValid(ipsItems)
      } else {
        updatedForm = UpdateFormHelper(value, formInputName, updatedFormState)
      }
    }

    setIsFormValid(isValidForm(updatedForm))

    setForm(updatedForm)
  }

  const onChangeTagValue = (event: any, formInputName: string, tagIndex: number, itemKey: number): void => {
    const value = event.target.value
    const updatedFormState = { ...form }

    const updatedForm = updatedFormState

    const listeners = updatedForm.listeners
    const listenersItems = [...listeners.items]
    const listenersItem = listenersItems[itemKey]
    const tagOptions = [...listenersItem.instancesLabels.dictionaryOptions]
    const tagOption = tagOptions[tagIndex]
    const tagOptionForm = UpdateFormHelper(value, formInputName, tagOption)

    tagOptions[tagIndex] = tagOptionForm
    listenersItem.instancesLabels.dictionaryOptions = tagOptions
    listenersItem.instancesLabels.isValid = checkIsFormValid(tagOptions)

    listenersItems[itemKey] = listenersItem

    updatedForm.listeners.items = listenersItems
    updatedForm.listeners.isValid = checkIsFormValid(listenersItems)

    setIsFormValid(isValidForm(updatedForm))

    setForm(updatedForm)
  }

  const onChangeDropdownMultiple = (values: [], key: number, listenersItemName: string): void => {
    const updatedFormState = { ...form }
    const updatedForm = updatedFormState

    const listeners = updatedForm.listeners
    const listenersItems = [...listeners.items]
    const listenersItem = listenersItems[key]

    const updatedListnerItem = UpdateFormHelper(values, 'instances', listenersItem)

    listenersItems[key] = updatedListnerItem
    updatedForm.listeners.items = listenersItems

    updatedForm.listeners.isValid = checkIsFormValid(listenersItems)
    setIsFormValid(isValidForm(updatedForm))

    setForm(updatedForm)
  }

  const selectAllInstances = (key: number): void => {
    const updatedFormState = { ...form }
    const updatedForm = updatedFormState

    const listeners = updatedForm.listeners
    const listenersItems = [...listeners.items]
    const listenersItem = listenersItems[key]
    const listenerItemInstance = { ...listenersItem.instances }

    const values: [] = instances.map((x) => x.resourceId) as []
    const shouldDeselect = values.every((x) => listenerItemInstance.value.includes(x))

    if (!shouldDeselect) {
      listenerItemInstance.value = values
      listenerItemInstance.isTouched = true
      listenerItemInstance.isValid = true
    } else {
      listenerItemInstance.value = []
      listenerItemInstance.isTouched = true
      listenerItemInstance.isValid = false
      listenerItemInstance.validationMessage = listenerItemInstance.label + ' is required'
    }

    listenersItem.instances = listenerItemInstance
    listenersItems[key] = listenersItem
    updatedForm.listeners.items = listenersItems

    updatedForm.listeners.isValid = checkIsFormValid(listenersItems)
    setIsFormValid(isValidForm(updatedForm))

    setForm(updatedForm)
  }

  const checkIsFormValid = (formsItems: any): boolean => {
    let isValid = true

    if (formsItems.length === 0) {
      isValid = false
    }

    for (const formItem of formsItems) {
      const isFormItemValid = isValidForm(formItem)
      if (!isFormItemValid) {
        isValid = false
        break
      }
    }
    return isValid
  }

  const showRequiredError = (formsItems: any): boolean => {
    for (const key in formsItems) {
      const formItem = formsItems[key]
      const { validationRules } = formItem
      if (validationRules && !formItem.isValid) {
        formItem.isTouched = true
        validateInput(formItem)
      }
    }

    return formsItems
  }

  const showRequiredFields = (): void => {
    let updatedFormState = { ...form }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(updatedFormState)

    const listeners = updatedForm.listeners
    const listenersItems = [...listeners.items]
    const allExternalPorts = listenersItems.map((x) => x.externalPort.value)

    for (let i = 0; i < listenersItems.length; i++) {
      const updatedListnersItem = showRequiredError(listenersItems[i])
      listenersItems[i] = updatedListnersItem

      if (listenersItems[i].externalPort.value !== '') {
        const duplicateValue = allExternalPorts.filter((x) => x === listenersItems[i].externalPort.value)
        if (duplicateValue.length > 1) {
          listenersItems[i].externalPort.isValid = false
          listenersItems[i].externalPort.validationMessage = 'Duplicate port'
        }
      }
    }
    updatedForm.listeners.items = listenersItems

    const ips = updatedForm.ips
    const ipsItems = [...ips.items]
    const allSourceIps = ipsItems.map((x) => x.ip.value)
    for (let i = 0; i < ipsItems.length; i++) {
      const updatedIpsItem = showRequiredError(ipsItems[i])
      ipsItems[i] = updatedIpsItem

      if (ipsItems[i].ip.value !== '') {
        const duplicateValue = allSourceIps.filter((x) => x === ipsItems[i].ip.value)
        if (duplicateValue.length > 1) {
          ipsItems[i].ip.isValid = false
          ipsItems[i].ip.validationMessage = 'Duplicate IP'
        }
      }
    }
    updatedForm.ips.items = ipsItems

    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    updatedFormState = updatedForm
    setForm(updatedFormState)
  }

  const onSubmit = async (e: any): Promise<void> => {
    try {
      const stateCopy = { ...state }

      if (form.listeners.items.length === 0) {
        showError('At-least one listener is required.', false)
        return
      }

      if (maxListeners !== '' && form.listeners.items.length > Number(maxListeners)) {
        showError('The number of listeners cannot exceed the limit.', false)
        return
      }

      if (form.ips.items.length === 0) {
        showError('At-least one source IP is required.', false)
        return
      }

      if (maxSourceIps !== '' && form.ips.items.length > Number(maxSourceIps)) {
        showError('The number of source ips cannot exceed the limit.', false)
        return
      }

      if (!isFormValid) {
        showRequiredFields()
        return
      }

      const getInstanceSelectors = (instancesLabels: any): PayloadInstanceSelectorsInterface => {
        const payloadInstanceSelectors: any = {}
        for (const label of instancesLabels.dictionaryOptions) {
          const key = getFormValue('key', label)
          const value = getFormValue('value', label)
          payloadInstanceSelectors[key] = value
        }

        return payloadInstanceSelectors
      }

      const listenersPayload = []

      for (const listenerItem of form.listeners.items) {
        const pool: any = {
          port: getFormValue('internalPort', listenerItem),
          monitor: getFormValue('monitorType', listenerItem),
          loadBalancingMode: getFormValue('loadBalancingMode', listenerItem)
        }

        const instanceSelectors = getFormValue('instanceSelectors', listenerItem)

        if (instanceSelectors === 'labels') {
          pool.instanceSelectors = getInstanceSelectors(listenerItem.instancesLabels)
        } else {
          pool.instanceResourceIds = getFormValue('instances', listenerItem)
        }

        const listenerPayload: PayloadListenersInterface = {
          port: getFormValue('externalPort', listenerItem),
          pool
        }
        listenersPayload.push(listenerPayload)
      }

      const sourceIpsPayload: any[] = []

      for (const ipsItem of form.ips.items) {
        const ip = getFormValue('ip', ipsItem)
        sourceIpsPayload.push(ip)
      }

      const metadata: PayloadMetadataInterface = {
        name: getFormValue('name', form)
      }

      const sourceips: PayloadSecurityInterface = {
        sourceips: sourceIpsPayload
      }

      const specPayload: PayloadSpecInterface = {
        listeners: listenersPayload,
        security: sourceips
      }

      const payload: ServicePayloadInterface = {
        metadata,
        spec: specPayload
      }

      stateCopy.showReservationModal = true
      setState(stateCopy)
      await createReservation(payload)
    } catch (error: any) {
      const stateUpdated = { ...state }

      stateUpdated.showReservationModal = false
      // stateUpdated.errorHideRetryMessage = false
      stateUpdated.errorDescription = null

      if (error.response) {
        const errData = error.response.data
        const errCode = errData.code
        if (isErrorInsufficientCredits(error)) {
          // No Credits
          setShowUpgradeNeededModal(true)
        } else if (errCode === 11) {
          // No Quota
          stateUpdated.showErrorModal = true
          stateUpdated.errorMessage = error.response.data.message
          // stateUpdated.errorHideRetryMessage = true
        } else if (isErrorInsufficientCapacity(error.response.data.message)) {
          stateUpdated.showErrorModal = true
          // stateUpdated.errorDescription = friendlyErrorMessages.insufficientCapacity
          stateUpdated.errorMessage = ''
          // stateUpdated.errorHideRetryMessage = true
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

  const onClickCloseErrorModal = (): void => {
    const stateUpdated = { ...state }
    stateUpdated.showErrorModal = false
    setState(stateUpdated)
  }

  const onClickFormAction = (actionType: string, index: number): void => {
    const updateForm = { ...form }
    const formListeners = updateForm.listeners

    if (actionType === 'Add') {
      let listeners = { ...listenersIntials }
      listeners = setSelectOptions('instances', setInstanceOption(), listeners)
      listeners = setSelectOptions('loadBalancingMode', loadBalancerModes, listeners)
      listeners = setFormValue('loadBalancingMode', loadBalancerModes[0].value, listeners)
      formListeners.items.push(listeners)
      formListeners.isValid = false
    }

    if (actionType === 'Delete') {
      formListeners.items.splice(index, 1)
      formListeners.isValid = checkIsFormValid(formListeners.items)
    }

    if (formListeners.items.length === 0) {
      formListeners.isValid = false
      setIsFormValid(false)
    }

    updateForm.listeners = formListeners

    setIsFormValid(isValidForm(updateForm))
    setForm(updateForm)
  }

  const onClickActionTag = (actionType: string, listenerItemKey: number, tagIndex: number): void => {
    const updateForm = { ...form }
    const formListeners = updateForm.listeners
    const listenersItems = [...formListeners.items]
    const listenersItem = listenersItems[listenerItemKey]

    if (actionType === 'Add') {
      listenersItem.instancesLabels.dictionaryOptions.push(instanceLabelsOption)
      listenersItem.instancesLabels.isValid = false
    }

    if (actionType === 'Delete') {
      const options = [...listenersItem.instancesLabels.dictionaryOptions]
      options.splice(tagIndex, 1)
      listenersItem.instancesLabels.dictionaryOptions = options
      listenersItem.instancesLabels.isValid = checkIsFormValid(options)
      listenersItem.instancesLabels.isTouched = true
    }

    updateForm.listeners.items = listenersItems

    updateForm.listeners.isValid = checkIsFormValid(listenersItems)
    setIsFormValid(isValidForm(updateForm))

    setForm(updateForm)
  }

  const onClickSourceIpsAction = (actionType: string, index: number): void => {
    const updateForm = { ...form }
    const formIps = updateForm.ips

    if (actionType === 'Add') {
      formIps.items.push({ ...ipInput })
      formIps.isValid = false
    }

    if (actionType === 'Delete') {
      formIps.items.splice(index, 1)
      formIps.isValid = checkIsFormValid(formIps.items)
    }

    if (formIps.items.length === 0) {
      formIps.isValid = false
      setIsFormValid(false)
    }

    updateForm.ips = formIps

    setIsFormValid(isValidForm(updateForm))
    setForm(updateForm)
  }

  const createReservation = async (servicePayload: any): Promise<void> => {
    await LoadBalancerService.createLoadBalancer(servicePayload)
    setTimeout(() => {
      const stateUpdated = { ...state }

      stateUpdated.showReservationModal = false

      setState(stateUpdated)

      navigate({
        pathname: '/load-balancer'
      })
    }, state.timeoutMiliseconds)
  }
  return (
    <LoadBalancerLaunch
      state={state}
      form={form}
      onClickCloseErrorModal={onClickCloseErrorModal}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      showUpgradeNeededModal={showUpgradeNeededModal}
      setShowUpgradeNeededModal={setShowUpgradeNeededModal}
      onClickFormAction={onClickFormAction}
      onChangeTagValue={onChangeTagValue}
      onClickActionTag={onClickActionTag}
      onChangeDropdownMultiple={onChangeDropdownMultiple}
      selectAllInstances={selectAllInstances}
      emptyCatalogModal={emptyCatalogModal}
      onClickSourceIpsAction={onClickSourceIpsAction}
      maxListeners={maxListeners}
      maxSourceIps={maxSourceIps}
    />
  )
}

export default LoadBalancerLaunchContainer
