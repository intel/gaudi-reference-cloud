// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import ComputeLaunch from '../../components/compute/computeLauch/ComputeLaunch'
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
import idcConfig from '../../config/configurator'
import { formatNumber } from '../../utils/numberFormatHelper/NumberFormatHelper'
import useSuperComputerStore from '../../store/superComputer/SuperComputerStore'
import SuperComputerService from '../../services/SuperComputerService'
import { friendlyErrorMessages, isErrorInsufficientCapacity } from '../../utils/apiError/apiError'
import {
  superComputerNodeGroupTypes,
  computeCategoriesEnum,
  superComputerProductCatalogTypes,
  toastMessageEnum,
  IDCVendorFamilies
} from '../../utils/Enums'
import useToastStore from '../../store/toastStore/ToastStore'
import { costPerDay, costPerHour, costPerWeek } from '../../utils/costEstimate/CostCalculator'
import PublicService from '../../services/PublicService'

const getCustomMessage = (messageType) => {
  let message = null

  switch (messageType) {
    case 'nodeGroupDetails':
      message = (
        <div className="valid-feedback" intc-id={'nodeGroupDetailsValidMessage'}>
          Name must be 63 characters or less, and can include letters, numbers, and ‘-‘ only.
          <br />
          It should start and end with an alphanumeric character.
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
    default:
      break
  }

  return message
}

const SuperComputerAddWorkerNodeContainer = () => {
  // Navigation
  const navigate = useNavigate()
  const { param: name } = useParams()

  const loadingScStore = useSuperComputerStore((state) => state.loading)
  const loadingDetail = useSuperComputerStore((state) => state.loadingDetail)
  const clusterDetail = useSuperComputerStore((state) => state.clusterDetail)
  const setClusterDetail = useSuperComputerStore((state) => state.setClusterDetail)
  const setNodeTabNumber = useSuperComputerStore((state) => state.setNodeTabNumber)
  const loadingCloudStore = useCloudAccountStore((state) => state.loading)
  const publicKeys = useCloudAccountStore((state) => state.publicKeys)
  const setPublickeys = useCloudAccountStore((state) => state.setPublickeys)
  const products = useSuperComputerStore((state) => state.coreComputeProducts)
  const setProducts = useSuperComputerStore((state) => state.setProducts)
  const vnets = useCloudAccountStore((state) => state.vnets)
  const setVnets = useCloudAccountStore((state) => state.setVnets)
  const setDebounceDetailRefresh = useSuperComputerStore((state) => state.setDebounceDetailRefresh)
  const scControlPlane = useSuperComputerStore((state) => state.scControlPlane)
  const showError = useToastStore((state) => state.showError)
  const isGeneralComputeAvailable = useSuperComputerStore((state) => state.isGeneralComputeAvailable)

  // local state
  const columns = [
    {
      columnName: '',
      targetColumn: 'instanceType',
      hideField: true
    },
    {
      columnName: 'Instance Type',
      targetColumn: 'instanceTypeDetails'
    },
    {
      columnName: 'Price (hour)',
      targetColumn: 'price',
      className: 'text-end',
      width: '8rem'
    },
    {
      columnName: 'Instance Type and Price',
      targetColumn: 'instanceTypePrice',
      showOnBreakpoint: true
    }
  ]

  const selectAllButton = {
    label: 'Select/Deselect All',
    buttonFunction: () => {
      onSelectAllKeys()
    }
  }

  const initialState = {
    mainTitle: `Add node group to Supercomputing cluster ${name}`,
    mainSubtitle: '',
    instanceConfigSectionTitle: 'Node group configuration',
    publicKeysMenuSection: 'Public Keys',
    form: {
      recommendedUseCase: {
        sectionGroup: 'configuration',
        type: 'radio-card', // options = 'text ,'textArea'
        label: '',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: ''
      },
      instanceType: {
        sectionGroup: 'configuration',
        type: 'grid',
        gridBreakpoint: 'sm',
        maxWidth: '100%',
        label: 'Node type: ',
        hiddenLabel: true,
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        columns,
        validationMessage: '',
        singleSelection: true,
        selectedRecords: null,
        setSelectedRecords: null,
        emptyGridMessage: {
          title: 'No available node types',
          subTitle: 'Please choose a different configuration'
        }
      },
      nodeGroupName: {
        sectionGroup: 'configuration',
        type: 'text', // options = 'text ,'textArea'
        label: 'Node group name:',
        placeholder: 'Node group name',
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
        helperMessage: getCustomMessage('nodeGroupDetails')
      },
      nodeQuantity: {
        sectionGroup: 'configuration',
        type: 'integer', // options = 'text ,'textArea'
        label: 'Node quantity:',
        placeholder: 'Nodes',
        value: 1,
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        maxLength: 2,
        validationRules: {
          isRequired: true,
          checkMinValue: 1,
          checkMaxValue: 10
        },
        options: getNodeQuantityOptions(),
        validationMessage: '',
        hidden: false
      },
      cloudConfigInitUrl: {
        sectionGroup: 'configuration',
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
        helperMessage: ''
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
          status: true,
          label: '+ Upload Key',
          buttonFunction: () => onShowHidePublicKeyModal(true)
        },
        emptyOptionsMessage: 'No keys found. Please create a key to continue.'
      }
    },
    costEstimate: {
      title: 'Cost estimate',
      description:
        'This cost estimate is based on the components you have selected. Actual costs may vary depending on usage and additional services.',
      costArray: []
    },
    isValidForm: false,
    showReservationModal: false,
    showErrorModal: false,
    errorMessage: null,
    errorHideRetryMessage: null,
    errorDescription: null,
    navigationBottom: [
      {
        buttonLabel: 'Launch node group',
        buttonVariant: 'primary'
      },
      {
        buttonLabel: 'Cancel',
        buttonVariant: 'link',
        buttonFunction: () => onCancel()
      }
    ],
    servicePayload: {
      count: 0,
      description: null,
      vnets: [],
      instancetypeid: null,
      userdataurl: null,
      instanceType: null,
      nodegrouptype: null,
      name: null,
      tags: [],
      sshkeyname: null
    },
    networkServicePayload: {
      metadata: {
        name: idcConfig.REACT_APP_DEFAULT_REGION_NAME
      },
      spec: {
        availabilityZone: idcConfig.REACT_APP_DEFAULT_REGION_AVAILABILITY_ZONE,
        prefixLength: idcConfig.REACT_APP_DEFAULT_REGION_PREFIX,
        region: idcConfig.REACT_APP_DEFAULT_REGION
      }
    },
    timeoutMiliseconds: 4000
  }

  const emptyCatalogModalInitial = {
    show: false,
    product: 'Supercomputing Node Group',
    goBackPath: '/supercomputer',
    extraExplanation: 'You may still launch AI nodes.',
    launchPath: '/supercomputer/launch',
    launchLabel: 'Launch New Supercomputing'
  }

  const premiumCancelButtonOptions = {
    label: 'Cancel',
    onClick: () => hidePremiumModal()
  }

  const premiumFormActions = {
    afterSuccess: () => afterPremiumSuceess(),
    afterError: togglePremiumError
  }

  const throwError = useErrorBoundary()

  const [state, setState] = useState(initialState)
  const [showPublicKeyModal, setShowPublicKeyModal] = useState(false)
  const [showCostEstimateModal, setShowCostEstimateModal] = useState(false)
  const [showPremiumModal, setShowPremiumModal] = useState(false)
  const [premiumError, setPremiumError] = useState({
    isShow: false,
    errorMessage: ''
  })
  const [showUpgradeNeededModal, setShowUpgradeNeededModal] = useState(false)
  const [productsByCategory, setProductsByCategory] = useState([])
  const [selectedInstanceTypes, setSelectedInstanceTypes] = useState([])
  const [isPageReady, setIsPageReady] = useState(clusterDetail && clusterDetail.name === name)
  const [emptyCatalogModal, setEmptyCatalogModal] = useState(emptyCatalogModalInitial)

  const fetchClusterDetail = async (isBackground) => {
    try {
      await setClusterDetail(name, isBackground)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  // Hooks
  useEffect(() => {
    if (!isPageReady) {
      fetchClusterDetail(false)
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

    if (products.length === 0) {
      const fetchProducts = async () => {
        try {
          await setProducts()
        } catch (error) {
          throwError(error)
        }
      }

      fetchProducts()
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
  }, [])

  useEffect(() => {
    const productsByCategory = products.filter(
      (item) => item.category === computeCategoriesEnum.singleNode || item.category === computeCategoriesEnum.cluster
    )
    setProductsByCategory(productsByCategory)
  }, [products])

  useEffect(() => {
    if (isPageReady && clusterDetail) {
      setForm()
    }
  }, [products, vnets, publicKeys, isPageReady, clusterDetail])

  useEffect(() => {
    if (isPageReady && clusterDetail === null) {
      navigate('/supercomputer')
    }
  }, [clusterDetail, isPageReady])

  useEffect(() => {
    if (!isGeneralComputeAvailable) {
      setEmptyCatalogModal({ ...emptyCatalogModal, show: true })
    }
  }, [isGeneralComputeAvailable])

  useEffect(() => {
    const instanceType = selectedInstanceTypes.length > 0 ? selectedInstanceTypes[0] : ''
    const event = { target: { value: instanceType } }
    onChangeInput(event, 'instanceType')
  }, [selectedInstanceTypes])

  // functions
  const updateInstanceTypeFormState = (form, instanceType) => {
    let formUpdated = {
      ...form
    }

    formUpdated = setFormValue('instanceType', instanceType, formUpdated)
    instanceType !== '' ? setSelectedInstanceTypes([instanceType]) : setSelectedInstanceTypes([])

    return formUpdated
  }

  function setForm() {
    const stateUpdated = {
      ...state
    }
    // Load key pairs
    const keys = publicKeys.map((key) => {
      return { name: key.name + ` (${key.ownerEmail})`, value: key.value }
    })
    stateUpdated.form = setSelectOptions('keyPairList', keys, stateUpdated.form)
    if (keys.length > 0) {
      stateUpdated.form = setFormValue('keyPairList', [keys[0].value], stateUpdated.form)
      stateUpdated.form.keyPairList.selectAllButton = selectAllButton
    }

    // Load Recommended Use Case
    let recommendedUseCaseValue = ''
    const selectableRecommendedUseCase = getRecommendedUseCaseOptions()
    if (selectableRecommendedUseCase.length > 0) {
      recommendedUseCaseValue =
        recommendedUseCaseValue !== '' ? recommendedUseCaseValue : selectableRecommendedUseCase[0].value
      stateUpdated.form = setFormValue('recommendedUseCase', recommendedUseCaseValue, stateUpdated.form)
      stateUpdated.form = setSelectOptions('recommendedUseCase', selectableRecommendedUseCase, stateUpdated.form)
    }

    stateUpdated.form.instanceType.selectedRecords = selectedInstanceTypes
    stateUpdated.form.instanceType.setSelectedRecords = setSelectedInstanceTypes
    const instanceFormValue = getFormValue('instanceType', state.form)
    const selectableInstanceTypes = getInstanceTypeOptions(recommendedUseCaseValue)
    let instanceType = null
    if (selectableInstanceTypes.length > 0) {
      instanceType = instanceFormValue !== '' ? instanceFormValue : selectableInstanceTypes[0].instanceType
      stateUpdated.form = updateInstanceTypeFormState(stateUpdated.form, instanceType)
      stateUpdated.form = setSelectOptions('instanceType', selectableInstanceTypes, stateUpdated.form)
      stateUpdated.costEstimate.costArray = getCostEstimate(instanceType)
    }
    setState(stateUpdated)
  }

  function getRecommendedUseCaseOptions() {
    const uniqueByRecommendedUseCase = productsByCategory.reduce((acc, item) => {
      if (!acc.some((obj) => obj.recommendedUseCase === item.recommendedUseCase)) {
        acc.push(item)
      }
      return acc
    }, [])

    const recommendedUseCaseOptions = []
    uniqueByRecommendedUseCase.forEach((item) => {
      const description = PublicService.getCatalogShortDescription(IDCVendorFamilies.Compute, item.recommendedUseCase)
      recommendedUseCaseOptions.push({
        value: item.recommendedUseCase,
        name: item.recommendedUseCase,
        subTitleHtml: (
          <div className="d-flex flex-column gap-s2 col-lg-10 col-xl-8">
            <span className="fw-semibold text-nowrap">{item.recommendedUseCase}</span>
            {description ? (
              <span>
                <small>{description}</small>
              </span>
            ) : null}
          </div>
        )
      })
    })

    return recommendedUseCaseOptions
  }

  function getInstanceTypeOptions(recommendedUseCase) {
    const instancesTypes = productsByCategory.filter((product) => product.recommendedUseCase === recommendedUseCase)
    const instanceTypeOptions = []

    for (let index = 0; index < instancesTypes.length; index++) {
      instanceTypeOptions.push({
        instanceType: instancesTypes[index].instanceType,
        instanceTypeDetails: {
          showField: true,
          type: 'function',
          canSelectRow: true,
          value: instancesTypes[index],
          function: (value) => {
            return (
              <div className="d-flex flex-column">
                <div>
                  <span className="fw-semibold text-nowrap me-s3">{value.name.toUpperCase()}</span>
                  <span>{value.familyDisplayName}</span>
                </div>
                <div>
                  <span className="fw-semibold">{`${value.service === 'Bare Metal' ? `${value.service}, ` : ''}`}</span>
                  <span>{value.description}</span>
                </div>
              </div>
            )
          }
        },
        price: {
          showField: true,
          type: 'function',
          canSelectRow: true,
          value: getPricing(instancesTypes[index]),
          function: (value) => {
            return (
              <>
                <div className="d-flex justify-content-end w-100">
                  <span className="fw-semibold ">{value}</span>
                </div>
              </>
            )
          }
        },
        instanceTypePrice: {
          showField: true,
          type: 'function',
          canSelectRow: true,
          value: instancesTypes[index],
          function: (value) => {
            return (
              <div className="d-flex flex-column">
                <div>
                  <span className="fw-semibold text-nowrap me-s3">{value.name.toUpperCase()}</span>
                  <span>{value.familyDisplayName}</span>
                </div>
                <div>
                  <span className="fw-semibold">{`${value.service === 'Bare Metal' ? `${value.service}, ` : ''}`}</span>
                  <span>{value.description}</span>
                </div>
                <span className="fw-semibold pt-s4">{getPricing(value)}</span>
              </div>
            )
          }
        }
      })
    }
    return instanceTypeOptions
  }

  function getPricing(product) {
    const rate = product.rate

    let result = null
    const cost = formatNumber(costPerHour(rate), 2)
    result = cost === 0 ? 'Free' : `$${cost}`

    return result
  }

  function getCostEstimate(instanceType) {
    const costArray = []

    const product = products.find((x) => x.name === instanceType)
    if (product) {
      costArray.push(
        {
          label: 'Daily cost:',
          value: `$${formatNumber(costPerDay(product.rate), 2)}`
        },
        {
          label: 'Weekly cost:',
          value: `$${formatNumber(costPerWeek(product.rate), 2)}`
        }
      )
    }

    return costArray
  }

  function getNodeQuantityOptions() {
    const nodeQuantityOptions = []
    for (let index = 1; index <= 10; index++) {
      const dropDownItem = {
        value: index,
        name: index
      }
      nodeQuantityOptions.push(dropDownItem)
    }
    return nodeQuantityOptions
  }

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    if (formInputName === 'recommendedUseCase') {
      // Clear value
      updatedState.form = updateInstanceTypeFormState(updatedState.form, '')
      updatedState.form = setSelectOptions('instanceType', [], updatedState.form)

      const selectableInstanceTypes = getInstanceTypeOptions(event.target.value)
      updatedState.form = setSelectOptions('instanceType', selectableInstanceTypes, updatedState.form)
      if (selectableInstanceTypes.length > 0) {
        const instanceTypeValue = selectableInstanceTypes[0].instanceType
        updatedState.form = updateInstanceTypeFormState(updatedState.form, instanceTypeValue)
      }
    }

    if (formInputName === 'instanceType') {
      updatedState.form.instanceType.selectedRecords = [event.target.value]
      updatedState.costEstimate.costArray = getCostEstimate(event.target.value)
    }

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

  function onShowHidePublicKeyModal(status = false) {
    setShowPublicKeyModal(status)
  }

  function onShowHideCostEstimateModal(status = false) {
    setShowCostEstimateModal(status)
  }

  function afterPubliKeyCreate() {
    setPublickeys()
  }

  function goBack() {
    navigate({
      pathname: `/supercomputer/d/${name}`,
      search: 'tab=workerNodeGroups'
    })
  }

  function onCancel() {
    // Navigates back to the page when this method triggers.
    goBack()
  }

  function onSubmit() {
    submitForm()
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

  async function submitForm() {
    try {
      const isValidForm = state.isValidForm
      if (!isValidForm) {
        showRequiredFields()
        return
      }

      const instanceType = getFormValue('instanceType', state.form)
      const instanceTypeControlPlane = scControlPlane.find(
        (item) => item.instanceType === superComputerProductCatalogTypes.controlPlane
      ).instanceType
      const servicePayload = { ...state.servicePayload }
      servicePayload.name = getFormValue('nodeGroupName', state.form)
      servicePayload.count = state.form.nodeQuantity.hidden ? 1 : Number(getFormValue('nodeQuantity', state.form))
      servicePayload.instancetypeid = instanceType
      servicePayload.instanceType = instanceTypeControlPlane
      servicePayload.nodegrouptype = superComputerNodeGroupTypes.gpCompute
      servicePayload.userdataurl = getFormValue('cloudConfigInitUrl', state.form)
      servicePayload.tags = []
      servicePayload.sshkeyname = getFormValue('keyPairList', state.form).map((sshkeyname) => ({ sshkey: sshkeyname }))
      const stateUpdated = { ...state }
      stateUpdated.showReservationModal = true
      setState(stateUpdated)
      await getVnet(servicePayload)
      await createNodeGroup(servicePayload)
    } catch (error) {
      const stateUpdated = { ...state }

      stateUpdated.showReservationModal = false
      stateUpdated.errorHideRetryMessage = false
      stateUpdated.errorTitleMessage = 'Could not launch your cluster'
      stateUpdated.errorDescription = ''

      if (error.response) {
        const errData = error.response.data
        const errCode = errData.code
        if (errCode === 7) {
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

  function hidePremiumModal() {
    setShowPremiumModal(false)
    togglePremiumError('', false)
  }

  function afterPremiumSuceess() {
    hidePremiumModal()
    submitForm()
  }

  async function getVnet(servicePayload) {
    const networks = []
    const network = {
      vNet: null,
      name: 'eth0'
    }
    if (vnets.length === 0) {
      // Create Vnet
      await CloudAccountService.postVnets(state.networkServicePayload)
      network.vNet = idcConfig.REACT_APP_DEFAULT_REGION_NAME
    } else {
      network.vNet = idcConfig.REACT_APP_DEFAULT_REGION_NAME
    }

    networks.push(network)
    servicePayload.vnets = [
      {
        availabilityzonename: state.networkServicePayload.spec.availabilityZone,
        networkinterfacevnetname: network.vNet
      }
    ]
  }

  async function createNodeGroup(servicePayload) {
    await SuperComputerService.createNodeGroup(servicePayload, clusterDetail.uuid)
    setDebounceDetailRefresh(true)
    setNodeTabNumber(1)
    setTimeout(() => {
      const stateUpdated = { ...state }

      stateUpdated.showReservationModal = false

      setState(stateUpdated)

      goBack()
    }, state.timeoutMiliseconds)
  }

  function onClickCloseErrorModal() {
    const stateUpdated = { ...state }

    stateUpdated.showErrorModal = false

    setState(stateUpdated)
  }

  function togglePremiumError(errorMessage, isShow = true) {
    setPremiumError({
      isShow,
      errorMessage
    })
  }

  if (!clusterDetail || clusterDetail.nodegroups.length === 10) {
    return null
  }

  return (
    <>
      <ComputeLaunch
        state={state}
        loading={loadingDetail || loadingScStore || loadingCloudStore || !isPageReady}
        showPublicKeyModal={showPublicKeyModal}
        showCostEstimateModal={showCostEstimateModal}
        onSubmit={onSubmit}
        onChangeInput={onChangeInput}
        afterPubliKeyCreate={afterPubliKeyCreate}
        onClickCloseErrorModal={onClickCloseErrorModal}
        onChangeDropdownMultiple={onChangeDropdownMultiple}
        onShowHidePublicKeyModal={onShowHidePublicKeyModal}
        onShowHideCostEstimateModal={onShowHideCostEstimateModal}
        showPremiumModal={showPremiumModal}
        premiumError={premiumError}
        premiumCancelButtonOptions={premiumCancelButtonOptions}
        premiumFormActions={premiumFormActions}
        showUpgradeNeededModal={showUpgradeNeededModal}
        setShowUpgradeNeededModal={setShowUpgradeNeededModal}
        computeReservationsPagePath="/cluster"
        emptyCatalogModal={emptyCatalogModal}
      />
    </>
  )
}

export default SuperComputerAddWorkerNodeContainer
