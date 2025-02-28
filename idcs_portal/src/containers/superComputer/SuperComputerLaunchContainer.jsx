// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import SuperComputerLaunch from '../../components/superComputer/superComputerLaunch/SuperComputerLaunch'
import { useNavigate } from 'react-router'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import {
  UpdateFormHelper,
  getFormValue,
  isValidForm,
  setSelectOptions,
  showFormRequiredFields,
  setFormValue
} from '../../utils/updateFormHelper/UpdateFormHelper'
import useSuperComputerStore from '../../store/superComputer/SuperComputerStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { formatNumber } from '../../utils/numberFormatHelper/NumberFormatHelper'
import SuperComputerService from '../../services/SuperComputerService'
import idcConfig from '../../config/configurator'
import { superComputerNodeGroupTypes, superComputerProductCatalogTypes, toastMessageEnum } from '../../utils/Enums'
import CloudAccountService from '../../services/CloudAccountService'
import useToastStore from '../../store/toastStore/ToastStore'
import { isErrorInsufficientCredits } from '../../utils/apiError/apiError'

const getCustomMessage = (
  messageType,
  volume = 'Enter size to calculate.',
  cost = '0',
  minSize,
  maxSize,
  unit,
  usageUnit
) => {
  let message = null

  switch (messageType) {
    case 'instanceDetails':
      message = (
        <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'}>
          Name must be 63 characters or less, and can include letters, numbers, and ‘-‘ only.
          <br />
          It should start and end with an alphanumeric character.
        </div>
      )
      break
    case 'publicKey':
      message = (
        <p className="lead">
          A key consists of a public key that Intel stores and a private key file that you store, allowing you to
          connect to your instance. The selected key in this step will be added to the set of keys authorized to this
          instance.
        </p>
      )
      break
    case 'volumeSize':
      message = (
        <div className="valid-feedback" intc-id={'volumeSize'}>
          Enter a value between {minSize} and {maxSize} <br />
          {unit} Volume cost {usageUnit}: <strong>{volume}</strong>
          <br />
          Volume cost {usageUnit}: <strong>$ {cost}</strong>
        </div>
      )
      break
    default:
      break
  }

  return message
}

const SuperComputerLaunchContainer = () => {
  // *****
  // Global state
  // *****
  const loading = useSuperComputerStore((state) => state.loading)
  const aiFamilies = useSuperComputerStore((state) => state.aiFamilies)
  const isWhitelisted = useSuperComputerStore((state) => state.isWhitelisted)
  const isGeneralComputeAvailable = useSuperComputerStore((state) => state.isGeneralComputeAvailable)
  const aiProducts = useSuperComputerStore((state) => state.aiProducts)
  const coreComputeProducts = useSuperComputerStore((state) => state.coreComputeProducts)
  const fileStorage = useSuperComputerStore((state) => state.fileStorage)
  const scControlPlane = useSuperComputerStore((state) => state.scControlPlane)
  const setProducts = useSuperComputerStore((state) => state.setProducts)
  const publicKeys = useCloudAccountStore((state) => state.publicKeys)
  const setPublickeys = useCloudAccountStore((state) => state.setPublickeys)
  const setClustersRuntimes = useSuperComputerStore((state) => state.setClustersRuntimes)
  const clustersRuntimes = useSuperComputerStore((state) => state.clustersRuntimes)
  const setClusters = useSuperComputerStore((state) => state.setClusters)
  const clusterResourceLimit = useSuperComputerStore((state) => state.clusterResourceLimit)
  const vnets = useCloudAccountStore((state) => state.vnets)
  const setVnets = useCloudAccountStore((state) => state.setVnets)
  const showError = useToastStore((state) => state.showError)
  const throwError = useErrorBoundary()
  // *****
  // local state
  // *****

  const computeNodeInitial = {
    computeInstanceType: {
      type: 'dropdown', // options = 'text ,'textArea'
      label: 'Compute Instance type:',
      placeholder: 'Please select instance type',
      value: '',
      isValid: true,
      isTouched: false,
      isReadOnly: false,
      validationRules: {
        isRequired: false
      },
      options: [],
      validationMessage: '',
      helperMessage: '',
      sectionGroup: 'configuration',
      subGroup: 'computeNodesType',
      hidden: false,
      labelButton: {
        label: 'Delete',
        buttonFunction: () => onClickComputeAction() // To be overwrite in component
      }
    },
    nodeCount: {
      type: 'integer', // options = 'text ,'textArea'
      label: 'Nodes:',
      placeholder: '0 to 10',
      value: '',
      isValid: true,
      isTouched: false,
      isReadOnly: false,
      maxLength: 2,
      validationRules: {
        isRequired: false,
        checkMinValue: 0,
        checkMaxValue: 10
      },
      validationMessage: '',
      helperMessage: '',
      sectionGroup: 'configuration',
      subGroup: 'computeNodesConfig',
      hidden: false
    },
    cloudConfigInitUrl: {
      type: 'text', // options = 'text ,'textArea'
      label: 'User data URL:',
      placeholder: 'https://',
      value: '', // Value enter by the user
      isValid: true, // Flag to validate if the input is ready
      isTouched: false, // Flag to validate if the user has modified the input
      isReadOnly: false, // Input create as read only
      maxLength: 150,
      validationRules: {
        isRequired: false,
        isValidURL: true
      },
      validationMessage: '', // Errror message to display to the user
      helperMessage: '',
      sectionGroup: 'configuration',
      subGroup: 'computeNodesConfig',
      hidden: false
    }
  }

  const initialState = {
    mainTitle: 'Launch a supercomputing cluster',
    form: {
      aiInstanceType: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'AI Instance type:',
        placeholder: 'Please select instance type',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: '',
        helperMessage: '',
        sectionGroup: 'configuration',
        subGroup: 'instanceType',
        hidden: false
      },
      aiNodes: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Nodes:',
        placeholder: 'Please select a quantity',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: '',
        helperMessage: '',
        sectionGroup: 'configuration',
        subGroup: 'config',
        hidden: false
      },
      aiCloudConfigInitUrl: {
        type: 'text', // options = 'text ,'textArea'
        label: 'User data URL:',
        placeholder: 'https://',
        value: '', // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 150,
        validationRules: {
          isRequired: false,
          isValidURL: true
        },
        validationMessage: '', // Errror message to display to the user
        helperMessage: '',
        sectionGroup: 'configuration',
        subGroup: 'config',
        hidden: false
      },
      computeNodes: {
        items: [],
        label: 'Compute nodes',
        isValid: true,
        validationRules: {
          isRequired: false
        }
      },
      instanceName: {
        sectionGroup: 'clusterProperties',
        subGroup: 'clusterProperties',
        type: 'text', // options = 'text ,'textArea'
        label: 'Cluster name:',
        placeholder: 'Cluster name',
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
        helperMessage: getCustomMessage('instanceDetails'),
        hidden: false
      },
      clusterK8sVersion: {
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Cluster kubernetes version:',
        placeholder: 'Please select kubernetes version',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: '',
        helperMessage: '',
        sectionGroup: 'clusterProperties',
        subGroup: 'clusterProperties',
        hidden: false
      },
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
          isRequired: false
        },
        sectionGroup: 'storage',
        subGroup: 'storage',
        validationMessage: '',
        helperMessage: '',
        hidden: false
      },
      volumeSize: {
        type: 'integer', // options = 'text ,'textArea'
        label: 'Volume size ():',
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
        hidden: false
      },
      keyPairList: {
        type: 'multi-select', // options = 'text ,'textArea'
        label: 'keys ',
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
          status: true,
          label: '+ Upload Key',
          buttonFunction: () => onShowHidePublicKeyModal(true)
        },
        selectAllButton: {
          label: 'Select/Deselect All',
          buttonFunction: () => onSelectAllKeys()
        },
        emptyOptionsMessage: 'No keys found. Please create a key to continue.',
        sectionGroup: 'keys',
        subGroup: 'keys',
        hidden: false
      }
    },
    isValidForm: false,
    navigationBottom: [
      {
        buttonLabel: 'Launch',
        buttonVariant: 'primary'
      },
      {
        buttonLabel: 'Cancel',
        buttonVariant: 'link',
        buttonFunction: () => onCancel()
      }
    ]
  }

  const emptyCatalogModalInitial = {
    show: false,
    product: 'Supercomputing',
    goBackPath: '/supercomputer',
    extraExplanation: 'You can still launch individual compute instances'
  }

  const defaultClusterRuntime = 'Containerd'

  const nodeGroupKey = {
    sshkey: ''
  }

  const nodegroupItemInitial = {
    count: 1,
    vnets: [],
    instancetypeid: '',
    instanceType: '',
    name: '',
    description: '',
    tags: [],
    sshkeyname: [
      {
        sshkey: ''
      }
    ],
    nodegrouptype: ''
  }

  const servicePayload = {
    clusterspec: {
      description: '',
      k8sversionname: '',
      name: '',
      runtimename: 'Containerd',
      tags: []
    },
    nodegroupspec: [],
    storagespec: {},
    instanceType: '',
    clustertype: 'supercompute'
  }

  const networkServicePayload = {
    metadata: {
      name: idcConfig.REACT_APP_DEFAULT_REGION_NAME
    },
    spec: {
      availabilityZone: idcConfig.REACT_APP_DEFAULT_REGION_AVAILABILITY_ZONE,
      prefixLength: idcConfig.REACT_APP_DEFAULT_REGION_PREFIX,
      region: idcConfig.REACT_APP_DEFAULT_REGION
    }
  }

  const submitModalInit = {
    show: false
  }

  const errorModalInitial = {
    show: false,
    title: 'Could not create your Supercomputing cluster',
    message: null,
    hideRetryMessage: false,
    description: null
  }

  // Navigation
  const navigate = useNavigate()
  const [state, setState] = useState(initialState)
  const [showInstanceCompareModal, setShowInstanceCompareModal] = useState({ type: 'computeInstance', show: false })
  const [showPublicKeyModal, setShowPublicKeyModal] = useState(false)
  const [costEstimate, setCostEstimate] = useState({
    controlPlaneRate: '',
    controlPlaneHourlyCost: '',
    aiInstanceTypeRate: '',
    aiNodeCount: '',
    aiHourlyCost: '',
    computeInstanceTypeRate: '',
    computeNodeCount: '',
    computeHourlyCost: '',
    storageRate: '',
    storageGbCount: '',
    storageHourlyCost: '',
    costTotal: ''
  })
  const [submitModal, setSubmitModal] = useState(submitModalInit)
  const [errorModal, setErrorModal] = useState(errorModalInitial)
  const [emptyCatalogModal, setEmptyCatalogModal] = useState(emptyCatalogModalInitial)
  const [showUpgradeNeededModal, setShowUpgradeNeededModal] = useState(false)
  const [sizeUnit, setSizeUnit] = useState('')

  // *****
  // Hooks
  // *****
  useEffect(() => {
    const fetchProducts = async () => {
      try {
        await setProducts()
      } catch (error) {
        throwError(error)
      }
    }
    fetchProducts()

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

    const fetchRuntimes = async () => {
      try {
        await setClustersRuntimes()
      } catch (error) {
        throwError(error)
      }
    }
    fetchRuntimes()
    if (vnets.length === 0) {
      const fetchVnets = async () => {
        try {
          await setVnets()
        } catch (error) {
          throwError(error)
        }
      }
      fetchVnets()
    }
    if (vnets.length === 0) {
      const fetchVnets = async () => {
        try {
          await setVnets()
        } catch (error) {
          throwError(error)
        }
      }
      fetchVnets()
    }

    if (!clusterResourceLimit) {
      const fetchClusters = async () => {
        try {
          await setClusters(false)
        } catch (error) {
          throwError(error)
        }
      }
      fetchClusters()
    }
  }, [])

  useEffect(() => {
    if (!isWhitelisted && !loading) {
      setEmptyCatalogModal({ ...emptyCatalogModal, show: true })
    }
  }, [isWhitelisted, loading])

  useEffect(() => {
    if (!loading && isWhitelisted) {
      setForm()
    }
  }, [aiFamilies, coreComputeProducts, fileStorage, isGeneralComputeAvailable, loading])

  useEffect(() => {
    updatePublicKeysForm()
  }, [publicKeys])

  // *****
  // functions
  // *****
  function onCancel() {
    // Navigates back to the page when this method triggers.
    navigate('/supercomputer')
  }

  function onShowHidePublicKeyModal(status = false) {
    setShowPublicKeyModal(status)
  }

  function onShowHideInstanceCompareModal(status = false, type) {
    setShowInstanceCompareModal({ ...showInstanceCompareModal, show: status, type })
  }

  const updatePublicKeysForm = () => {
    const stateUpdated = {
      ...state
    }
    const keys = publicKeys.map((key) => {
      return { name: key.name + ` (${key.ownerEmail})`, value: key.value }
    })
    // Load key pairs
    stateUpdated.form = setSelectOptions('keyPairList', keys, stateUpdated.form)
    stateUpdated.form = setFormValue('keyPairList', keys.length > 0 ? [keys[0].value] : [], stateUpdated.form)
    setState(stateUpdated)
  }

  const setForm = () => {
    const stateUpdated = {
      ...state
    }
    stateUpdated.form.keyPairList.selectAllButton = {
      label: 'Select/Deselect All',
      buttonFunction: () => onSelectAllKeys()
    }

    // Load AI Nodes
    const selectableAiTypes = getDropdownOptions(aiFamilies)
    if (selectableAiTypes.length > 0) {
      stateUpdated.form = setSelectOptions('aiInstanceType', selectableAiTypes, stateUpdated.form)
      stateUpdated.form = setFormValue('aiInstanceType', selectableAiTypes[0].value, stateUpdated.form)
      const aiItems = aiProducts.filter((item) => item.name === selectableAiTypes[0].value)
      const selectableAiNodes = getSelectOptions(aiItems)
      if (selectableAiNodes.length > 0) {
        stateUpdated.form = setSelectOptions('aiNodes', selectableAiNodes, stateUpdated.form)
        stateUpdated.form = setFormValue('aiNodes', selectableAiNodes[0].value, stateUpdated.form)
      }
    }

    const selectableClusterVersion = getClusterVersionOptions()
    if (selectableClusterVersion.length > 0) {
      stateUpdated.form = setSelectOptions('clusterK8sVersion', selectableClusterVersion, stateUpdated.form)
    }

    getCostEstimate(stateUpdated.form)
    setState(stateUpdated)
  }

  function onChangeInput(event, formInputName, idParent = '', index) {
    let value = ''
    if (formInputName === 'fileVolumeFlag') {
      value = event.target.checked | ''
    } else {
      value = event.target.value
    }
    const updatedState = {
      ...state
    }
    let updatedForm = updatedState.form

    if (idParent === 'computeNodes') {
      const computeNodes = updatedForm.computeNodes
      const computeNodeItems = [...computeNodes.items]
      const computeNodeItem = computeNodeItems[index]
      const updatedNode = UpdateFormHelper(value, formInputName, computeNodeItem)
      computeNodeItems[index] = updatedNode

      updatedForm.computeNodes.items = computeNodeItems
      // Validate rows
      updatedForm.computeNodes.isValid = validateComputeRows(computeNodes)
    } else {
      updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)
    }

    if (formInputName === 'fileVolumeFlag') {
      if (event.target.checked) {
        updatedForm.volumeSize.hidden = false
      } else {
        updatedForm.volumeSize.hidden = true
        updatedForm.volumeSize.isValid = true
      }
    }

    if (formInputName === 'aiInstanceType') {
      const aiItems = aiProducts.filter((item) => item.name === value)
      const selectableAiNodes = getSelectOptions(aiItems)
      if (selectableAiNodes.length > 0) {
        updatedForm = setSelectOptions('aiNodes', selectableAiNodes, updatedForm)
        updatedForm = setFormValue('aiNodes', selectableAiNodes[0].value, updatedForm)
      }
    }

    updatedState.form = updatedForm

    updatedState.isValidForm = isValidForm(updatedForm)

    getCostEstimate(updatedForm)

    setState(updatedState)
  }

  function getCostEstimate(updatedForm) {
    // const monthyRate = 24 * 30

    // AI Nodes
    const aiInstanceTypeSelected = getFormValue('aiInstanceType', updatedForm)
    const aiNodeCountSelected = getFormValue('aiNodes', updatedForm)

    let aiHourlyCost = 0
    let aiInstanceTypeRate = 0

    if (aiProducts.length > 0) {
      if (aiInstanceTypeSelected) {
        const allAiNodes = aiProducts.filter((item) => item.name === aiInstanceTypeSelected)
        aiInstanceTypeRate = Number(allAiNodes.find((item) => item.nodesCount === aiNodeCountSelected).rate)
      }

      if (aiNodeCountSelected) {
        aiHourlyCost = aiInstanceTypeRate * 60
      }
    }

    // Storage
    let storageHourlyCost = 0
    let storageRate = 0
    const storageGbCount = getFormValue('volumeSize', updatedForm)
    if (fileStorage.length > 0) {
      const storage = fileStorage.find(
        (item) => item.recommendedUseCase === superComputerProductCatalogTypes.fileStorage
      )

      storageRate = Number(storage.rate)
      const minimumSize = storage.minimumSize
      const maximumSize = storage.maximumSize
      const unitSize = storage.unitSize
      const usageUnit = storage.usageUnit
      setSizeUnit(unitSize)
      if (storageGbCount) {
        storageHourlyCost = storageRate * Number(storageGbCount)
      }
      updatedForm.volumeSize.validationRules = {
        ...updatedForm.volumeSize.validationRules,
        checkMinValue: minimumSize,
        checkMaxValue: maximumSize
      }
      updatedForm.volumeSize.maxLength = maximumSize.toString().length
      updatedForm.volumeSize.label = `Volume size (${unitSize}):`
      updatedForm.volumeSize.helperMessage = getCustomMessage(
        'volumeSize',
        formatNumber(storageHourlyCost, 2),
        storageRate,
        minimumSize,
        maximumSize,
        unitSize,
        usageUnit
      )
    }

    // Control Plane
    let controlPlaneHourlyCost = 0
    let controlPlaneRate = 0
    if (fileStorage.length > 0) {
      controlPlaneRate = Number(
        scControlPlane.find((item) => item.instanceType === superComputerProductCatalogTypes.controlPlane).rate
      )
      controlPlaneHourlyCost = controlPlaneRate * 60
    }
    // Compute Nodes
    let computeHourlyCost = 0
    let computeNodeCount = 0
    const computeNodes = updatedForm.computeNodes
    for (const index in computeNodes.items) {
      const item = { ...computeNodes.items[index] }
      const computeInstanceSelected = getFormValue('computeInstanceType', item)
      const computeNodesCount = getFormValue('nodeCount', item)
      if (computeNodesCount && computeInstanceSelected) {
        const instanceRate = Number(
          coreComputeProducts.find((item) => item.instanceType === computeInstanceSelected).rate
        )
        const itemCostHourly = instanceRate * Number(computeNodesCount) * 60
        computeHourlyCost = computeHourlyCost + itemCostHourly
        computeNodeCount = computeNodeCount + Number(computeNodesCount)
      }
    }

    const costTotal = computeHourlyCost + storageHourlyCost + aiHourlyCost + controlPlaneHourlyCost

    setCostEstimate({
      ...costEstimate,
      aiNodeCount: aiNodeCountSelected,
      aiHourlyCost,
      computeNodeCount,
      computeHourlyCost,
      controlPlaneHourlyCost,
      controlPlaneRate,
      storageGbCount,
      storageHourlyCost,
      storageRate,
      costTotal
    })
  }

  function getDropdownOptions(items) {
    const options = []
    for (const index in items) {
      options.push({
        name: items[index].displayName,
        dropSelect: getPricing(items[index]),
        value: items[index].name
      })
    }
    return options
  }

  function getSelectOptions(items) {
    const options = []
    for (const index in items) {
      options.push({
        name: items[index].nodesCount + ' Nodes',
        value: items[index].nodesCount
      })
    }

    return options.sort((a, b) => a.value - b.value)
  }

  function getPricing(instanceType) {
    const rate = instanceType.rate

    let result = null
    const cost = formatNumber(rate * 60, 2)
    result = cost === 0 ? 'Free' : `$${cost} / hour`

    return result
  }

  function getClusterVersionOptions() {
    const runTimes = []
    const items = clustersRuntimes.filter((item) => item.runtimename === defaultClusterRuntime)
    for (let index = 0; index < items.length; index++) {
      runTimes.push({
        name: items[index].k8sversionname,
        value: items[index].k8sversionname
      })
    }
    return runTimes
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
    setPublickeys()
  }

  function onClickComputeAction(index, action) {
    const updatedState = {
      ...state
    }
    const form = state.form
    const computeNodes = form.computeNodes
    if (action === 'Add') {
      const selectableComputeTypes = getDropdownOptions(coreComputeProducts)
      if (selectableComputeTypes.length > 0) {
        let computeNodeUpdated = { ...computeNodeInitial }
        if (selectableComputeTypes.length > 0) {
          computeNodeUpdated = setSelectOptions('computeInstanceType', selectableComputeTypes, computeNodeInitial)
          computeNodeUpdated = setFormValue('computeInstanceType', selectableComputeTypes[0].value, computeNodeInitial)
          computeNodes.items.push(computeNodeUpdated)
        }
      }
    } else {
      // Delete
      computeNodes.items.splice(index, 1)
      computeNodes.isValid = validateComputeRows(computeNodes)
    }

    form.computeNodes = computeNodes
    getCostEstimate(form)
    updatedState.isValidForm = isValidForm(form)
    updatedState.form = form
    setState(updatedState)
  }

  function validateComputeRows(computeNodes) {
    let isValidArray = true
    for (const index in computeNodes.items) {
      const computeItem = { ...computeNodes.items[index] }
      const isValidRow = isValidForm(computeItem)
      if (!isValidRow) {
        isValidArray = false
        break
      }
    }
    return isValidArray
  }

  async function getVnet(servicePayload) {
    const networks = []
    const network = {
      vNet: null,
      name: 'eth0'
    }
    if (vnets.length === 0) {
      // Create Vnet
      await CloudAccountService.postVnets(networkServicePayload)
      network.vNet = idcConfig.REACT_APP_DEFAULT_REGION_NAME
    } else {
      network.vNet = idcConfig.REACT_APP_DEFAULT_REGION_NAME
    }

    networks.push(network)
    servicePayload.vnets = [
      {
        availabilityzonename: networkServicePayload.spec.availabilityZone,
        networkinterfacevnetname: network.vNet
      }
    ]
  }

  function showRequiredFields() {
    const stateCopy = { ...state }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(stateCopy.form)
    // Mark Ai Inputs
    const computeNodes = updatedForm.computeNodes
    const items = computeNodes.items
    const itemsUpdated = []
    for (const index in items) {
      const item = { ...items[index] }
      const updatedItem = showFormRequiredFields(item)
      itemsUpdated.push(updatedItem)
    }
    computeNodes.items = itemsUpdated
    updatedForm.computeNodes = computeNodes

    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    stateCopy.form = updatedForm
    setState(stateCopy)
  }

  async function onSubmit() {
    const isValidForm = state.isValidForm
    if (!isValidForm) {
      showRequiredFields()
      return
    }
    const payload = { ...servicePayload }
    const clusterSpec = { ...payload.clusterspec }
    const keys = []
    const publicKeys = getFormValue('keyPairList', state.form)
    const instanceType = scControlPlane.find(
      (item) => item.instanceType === superComputerProductCatalogTypes.controlPlane
    ).instanceType
    for (const index in publicKeys) {
      const keyCopy = { ...nodeGroupKey }
      keyCopy.sshkey = publicKeys[index]
      keys.push(keyCopy)
    }
    const superclusterName = getFormValue('instanceName', state.form)
    clusterSpec.name = superclusterName
    clusterSpec.k8sversionname = getFormValue('clusterK8sVersion', state.form)
    clusterSpec.runtimename = defaultClusterRuntime
    payload.instanceType = instanceType
    payload.clusterspec = clusterSpec

    const nodegroupItems = []
    const nodegroupItemCopy = { ...nodegroupItemInitial }
    await getVnet(nodegroupItemCopy)
    nodegroupItemCopy.instancetypeid = getFormValue('aiInstanceType', state.form)
    nodegroupItemCopy.count = '1' // Nodes From Product Catalog.
    nodegroupItemCopy.nodegrouptype = superComputerNodeGroupTypes.aiCompute
    nodegroupItemCopy.instanceType = instanceType
    nodegroupItemCopy.name = superclusterName + '-group-ai'
    nodegroupItemCopy.userdataurl = getFormValue('aiCloudConfigInitUrl', state.form)
    nodegroupItemCopy.sshkeyname = keys
    nodegroupItems.push(nodegroupItemCopy)
    payload.nodegroupspec = nodegroupItems

    const computeNodes = [...state.form.computeNodes.items]
    if (isGeneralComputeAvailable && computeNodes?.length > 0) {
      const aiInstances = []
      for (const index in computeNodes) {
        const computeNode = { ...computeNodes[index] }
        const instanceTypeValue = getFormValue('computeInstanceType', computeNode)
        const instanceCountValue = Number(getFormValue('nodeCount', computeNode))
        const userdataurlValue = getFormValue('cloudConfigInitUrl', computeNode)
        if (instanceCountValue > 0) {
          aiInstances.push({
            instancetypeid: instanceTypeValue,
            nodeCount: instanceCountValue,
            userdataurl: userdataurlValue
          })
        }
      }
      const computeNodeItems = Object.values(
        aiInstances.reduce((value, object) => {
          if (value[object.instancetypeid]) {
            value[object.instancetypeid].nodeCount += object.nodeCount
            value[object.instancetypeid].count++
          } else {
            value[object.instancetypeid] = { ...object, count: 1 }
          }
          return value
        }, {})
      )

      for (const index in computeNodeItems) {
        const item = { ...computeNodeItems[index] }
        const nodegroupItemCopy = { ...nodegroupItemInitial }
        nodegroupItemCopy.count = String(item.nodeCount)
        nodegroupItemCopy.instancetypeid = item.instancetypeid
        nodegroupItemCopy.nodegrouptype = superComputerNodeGroupTypes.gpCompute
        nodegroupItemCopy.nodegrouptype = superComputerNodeGroupTypes.gpCompute
        nodegroupItemCopy.userdataurl = item.userdataurl
        nodegroupItemCopy.instanceType = instanceType
        nodegroupItemCopy.name = superclusterName + '-gp-' + index
        nodegroupItemCopy.sshkeyname = keys
        await getVnet(nodegroupItemCopy)
        nodegroupItemCopy.instanceType = instanceType
        nodegroupItemCopy.name = superclusterName + '-gp-' + index
        nodegroupItemCopy.sshkeyname = keys
        await getVnet(nodegroupItemCopy)
        nodegroupItems.push(nodegroupItemCopy)
      }
    }
    payload.storagespec.storagesize = getFormValue('volumeSize', state.form) + sizeUnit
    payload.storagespec.enablestorage = true

    createSuperCluster(payload)
  }

  const createSuperCluster = async (payload) => {
    try {
      setSubmitModal({ ...submitModal, show: true })
      await SuperComputerService.createSuperCluster(payload)
      setTimeout(() => {
        setSubmitModal(submitModalInit)
        navigate({
          pathname: '/supercomputer'
        })
      }, 4000)
    } catch (error) {
      setSubmitModal(submitModalInit)
      let errorMessage = ''
      if (error.response) {
        const errorMessage = error.response.data.message
        if (isErrorInsufficientCredits(error)) {
          // No Credits
          setShowUpgradeNeededModal(true)
        } else {
          setErrorModal({ ...errorModal, show: true, message: errorMessage })
        }
      } else {
        errorMessage = error.message
        setErrorModal({ ...errorModal, show: true, message: errorMessage })
      }
      setErrorModal({ ...errorModal, show: true, message: errorMessage })
    }
  }

  const onCloseErrorModal = () => {
    setErrorModal(errorModalInitial)
  }
  return (
    <SuperComputerLaunch
      loading={loading}
      setState={setState}
      mainTitle={state.mainTitle}
      form={state.form}
      onChangeInput={onChangeInput}
      showInstanceCompareModal={showInstanceCompareModal}
      onShowHideInstanceCompareModal={onShowHideInstanceCompareModal}
      aiProducts={aiProducts}
      coreComputeProducts={coreComputeProducts}
      navigationBottom={state.navigationBottom}
      isValidForm={state.isValidForm}
      onChangeDropdownMultiple={onChangeDropdownMultiple}
      onSubmit={onSubmit}
      costEstimate={costEstimate}
      showPublicKeyModal={showPublicKeyModal}
      onShowHidePublicKeyModal={onShowHidePublicKeyModal}
      afterPubliKeyCreate={afterPubliKeyCreate}
      onClickComputeAction={onClickComputeAction}
      submitModal={submitModal}
      errorModal={errorModal}
      onCloseErrorModal={onCloseErrorModal}
      emptyCatalogModal={emptyCatalogModal}
      setEmptyCatalogModal={setEmptyCatalogModal}
      setShowUpgradeNeededModal={setShowUpgradeNeededModal}
      showUpgradeNeededModal={showUpgradeNeededModal}
      clusterResourceLimit={clusterResourceLimit}
      isGeneralComputeAvailable={isGeneralComputeAvailable}
      sizeUnit={sizeUnit}
    />
  )
}

export default SuperComputerLaunchContainer
