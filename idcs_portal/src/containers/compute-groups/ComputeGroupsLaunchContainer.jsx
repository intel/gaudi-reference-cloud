// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import Wrapper from '../../utils/Wrapper'
import ComputeLaunch from '../../components/compute/computeLauch/ComputeLaunch'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  setFormValue,
  setSelectOptions,
  showFormRequiredFields
} from '../../utils/updateFormHelper/UpdateFormHelper'
import { useNavigate } from 'react-router'
import useHardwareStore from '../../store/hardwareStore/HardwareStore'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import CloudAccountService from '../../services/CloudAccountService'
import useImageStore from '../../store/imageStore/ImageStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import idcConfig, { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import { formatNumber } from '../../utils/numberFormatHelper/NumberFormatHelper'
import { friendlyErrorMessages, isErrorInsufficientCapacity } from '../../utils/apiError/apiError'
import { computeCategoriesEnum, IDCVendorFamilies, toastMessageEnum } from '../../utils/Enums'
import useToastStore from '../../store/toastStore/ToastStore'
import { costPerDay, costPerHour, costPerWeek } from '../../utils/costEstimate/CostCalculator'
import PublicService from '../../services/PublicService'
import ImageComponentsWithOverlay from '../../utils/imageComponents/ImageComponentsWithOverlay'

const getCustomMessage = (messageType) => {
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
          A key consists of a public key that the application stores and a private key file that you store, allowing you
          to connect to your instance. The selected key in this step will be added to the set of keys authorized to this
          instance.
        </p>
      )
      break
    case 'operationSystem':
      message = (
        <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'}>
          Choose the desired OS for your reservation
        </div>
      )
      break
    default:
      break
  }

  return message
}

const ComputeLaunchContainer = () => {
  // local state
  const columns = [
    {
      columnName: '',
      targetColumn: 'productName',
      hideField: true
    },
    {
      columnName: 'Instance Type',
      targetColumn: 'instanceType'
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
    mainTitle: 'Launch instance group',
    mainSubtitle: '',
    instanceConfigSectionTitle: 'Group configuration',
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
      productName: {
        sectionGroup: 'configuration',
        type: 'grid',
        gridBreakpoint: 'sm',
        maxWidth: '100%',
        label: 'Instance type: ',
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
          title: 'No available instance types',
          subTitle: 'Please choose a different configuration'
        }
      },
      operationSystem: {
        sectionGroup: 'configuration',
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Machine image:',
        placeholder: 'Please select',
        maxWidth: '100%',
        maxInputWidth: '48rem',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: '',
        helperMessage: getCustomMessage('operationSystem')
      },
      instanceName: {
        sectionGroup: 'configuration',
        type: 'text', // options = 'text ,'textArea'
        label: 'Group name:',
        placeholder: 'Group name',
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
        selectAllButton: undefined,
        emptyOptionsMessage: 'No keys found. Please create a key to continue.'
      },
      quickConnectFlag: {
        sectionGroup: 'quickConnect',
        type: 'switch', // options = 'text ,'textArea'
        label: '',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        options: [
          {
            name: 'Enable JupyterLab in my instance. Enabling this allows you to connect to your instance in a single click via JupyterLab, no further installations required.',
            value: '0'
          }
        ],
        validationRules: {
          isRequired: false
        },
        validationMessage: '',
        helperMessage: '',
        hidden: true
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
        buttonLabel: 'Launch instance group',
        buttonVariant: 'primary'
      },
      {
        buttonLabel: 'Cancel',
        buttonVariant: 'link',
        buttonFunction: () => onCancel()
      }
    ],
    servicePayload: {
      metadata: {
        name: null
      },
      spec: {
        instanceCount: null,
        instanceSpec: {
          availabilityZone: null,
          instanceType: null,
          machineImage: null,
          quickConnectEnabled: false,
          runStrategy: 'RerunOnFailure',
          sshPublicKeyNames: [],
          interfaces: [
            {
              name: 'eth0',
              vNet: null
            }
          ]
        }
      }
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
    product: 'instance groups',
    goBackPath: '/compute-groups',
    extraExplanation: 'You can still launch individual compute instances',
    launchPath: '/compute',
    launchLabel: 'Launch instance'
  }

  const category = computeCategoriesEnum.cluster

  const premiumCancelButtonOptions = {
    label: 'Cancel',
    onClick: () => hidePremiumModal()
  }

  const premiumFormActions = {
    afterSuccess: () => afterPremiumSuceess(),
    afterError: togglePremiumError
  }

  const throwError = useErrorBoundary()
  const [searchParams, setSearchParams] = useSearchParams()

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
  const [selectedProductNames, setSelectedProductNames] = useState([])
  const [emptyCatalogModal, setEmptyCatalogModal] = useState(emptyCatalogModalInitial)
  const [isPageReady, setIsPageReady] = useState(false)
  const initialInstanceType = searchParams.get('instance-type') ?? ''

  // Global State
  const ImagesOs = useImageStore((state) => state.ImagesOs)
  const setImagesOs = useImageStore((state) => state.setImagesOs)
  const publicKeys = useCloudAccountStore((state) => state.publicKeys)
  const setPublickeys = useCloudAccountStore((state) => state.setPublickeys)
  const products = useHardwareStore((state) => state.products)
  const setProducts = useHardwareStore((state) => state.setProducts)
  const familySelected = useHardwareStore((state) => state.familySelected)
  const setFamilyIdSelected = useHardwareStore((state) => state.setFamilyIdSelected)
  const vnets = useCloudAccountStore((state) => state.vnets)
  const setVnets = useCloudAccountStore((state) => state.setVnets)
  const showError = useToastStore((state) => state.showError)

  // Hooks
  useEffect(() => {
    const fetch = async () => {
      try {
        const promiseArray = []
        if (ImagesOs.length === 0) {
          promiseArray.push(setImagesOs())
        }
        if (publicKeys.length === 0) {
          promiseArray.push(setPublickeys())
        }
        if (products.length === 0) {
          promiseArray.push(setProducts())
        }
        if (vnets.length === 0) {
          promiseArray.push(setVnets())
        }
        await Promise.all(promiseArray)
        setIsPageReady(true)
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    if (products.length > 0) {
      const productFilter = products.filter((item) => item.category === category)
      setProductsByCategory(productFilter)
      if (productFilter.length === 0) {
        setEmptyCatalogModal({ ...emptyCatalogModal, show: true })
      }
    }
  }, [products])

  useEffect(() => {
    setInstanceSearchParams()
  }, [state.form.productName.value])

  useEffect(() => {
    setForm()
  }, [productsByCategory, vnets, ImagesOs, publicKeys])

  useEffect(() => {
    const productName = selectedProductNames.length > 0 ? selectedProductNames[0] : ''
    const event = { target: { value: productName } }
    onChangeInput(event, 'productName')
  }, [selectedProductNames])

  // Navigation
  const navigate = useNavigate()

  // functions
  const updateProductNameFormState = (form, productName) => {
    let formUpdated = {
      ...form
    }

    formUpdated = setFormValue('productName', productName, formUpdated)
    productName !== '' ? setSelectedProductNames([productName]) : setSelectedProductNames([])

    return formUpdated
  }

  function setInstanceSearchParams() {
    const instanceType = getFormValue('productName', state.form)
    const searchParamInstanceType = searchParams.get('instance-type')
    if (instanceType && searchParamInstanceType !== instanceType) {
      setSearchParams(
        (params) => {
          params.set('instance-type', instanceType)
          return params
        },
        { replace: true }
      )
    }
  }

  function setForm() {
    const stateUpdated = {
      ...state
    }
    const keys = publicKeys.map((key) => {
      return { name: key.name + ` (${key.ownerEmail})`, value: key.value }
    })
    // Load key pairs
    stateUpdated.form = setSelectOptions('keyPairList', keys, stateUpdated.form)
    if (keys.length > 0) {
      stateUpdated.form = setFormValue('keyPairList', [keys[0].value], stateUpdated.form)
      stateUpdated.form.keyPairList.selectAllButton = selectAllButton
    }

    let preselectedProduct
    if (initialInstanceType) {
      preselectedProduct = productsByCategory.find((product) => product.name === initialInstanceType)
    } else {
      if (familySelected) {
        preselectedProduct = productsByCategory.find((x) => x.familyDisplayName === familySelected)
      }
    }
    // Load Recommended Use Case
    let recommendedUseCaseValue = preselectedProduct
      ? preselectedProduct.recommendedUseCase
      : getFormValue('recommendedUseCase', state.form)
    const selectableRecommendedUseCase = getRecommendedUseCaseOptions()
    if (selectableRecommendedUseCase.length > 0) {
      recommendedUseCaseValue =
        recommendedUseCaseValue !== '' ? recommendedUseCaseValue : selectableRecommendedUseCase[0].value
      stateUpdated.form = setFormValue('recommendedUseCase', recommendedUseCaseValue, stateUpdated.form)
      stateUpdated.form = setSelectOptions('recommendedUseCase', selectableRecommendedUseCase, stateUpdated.form)
    }

    stateUpdated.form.productName.selectedRecords = selectedProductNames
    stateUpdated.form.productName.setSelectedRecords = setSelectedProductNames
    const instanceFormValue = preselectedProduct ? preselectedProduct.name : getFormValue('productName', state.form)
    const selectableInstanceTypes = getInstanceTypeOptions(recommendedUseCaseValue)
    let productName = null
    if (selectableInstanceTypes.length > 0) {
      stateUpdated.costEstimate.costArray = getCostEstimate(instanceFormValue)
      productName = instanceFormValue !== '' ? instanceFormValue : selectableInstanceTypes[0].productName
      stateUpdated.form = updateProductNameFormState(stateUpdated.form, productName)
      stateUpdated.form = setSelectOptions('productName', selectableInstanceTypes, stateUpdated.form)
      stateUpdated.costEstimate.costArray = getCostEstimate(productName)
    }

    const iosFormValue = getFormValue('operationSystem', state.form)
    const selectableImages = getOsOptions(productName, recommendedUseCaseValue)
    if (selectableImages.length > 0) {
      const ioValue = iosFormValue !== '' ? iosFormValue : selectableImages[0].value
      stateUpdated.form = setFormValue('operationSystem', ioValue, stateUpdated.form)
      stateUpdated.form = setSelectOptions('operationSystem', selectableImages, stateUpdated.form)
    }

    if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_QUICK_CONNECT)) {
      stateUpdated.form.quickConnectFlag.hidden = false
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
        productName: instancesTypes[index].name,
        instanceType: {
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

  function getCostEstimate(productName) {
    const costArray = []

    const product = products.find((x) => x.name === productName)
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

  function getOsOptions(productName, recommendedUseCase = null) {
    const selectableImages = []

    const selectedRecommendedUseCase =
      recommendedUseCase !== null ? recommendedUseCase : getFormValue('recommendedUseCase', state.form)

    const selectedInstance = products.find(
      (x) => x.recommendedUseCase === selectedRecommendedUseCase && x.name === productName
    )

    let filteredImageOS = []

    const filteredImageOSByInstanceCategory = ImagesOs.filter(
      (image) =>
        selectedInstance !== undefined && image.instanceCategories.includes(selectedInstance.instanceCategories)
    )
    const filteredImageOSByInstanceType = filteredImageOSByInstanceCategory.filter((item) =>
      item.instanceTypes.includes(selectedInstance.instanceType)
    )

    if (filteredImageOSByInstanceType.length > 0) {
      filteredImageOS = filteredImageOSByInstanceType
    } else {
      filteredImageOS = filteredImageOSByInstanceCategory.filter((item) => item.instanceTypes.length === 0)
    }

    for (let index = 0; index < filteredImageOS.length; index++) {
      const imageItem = { ...filteredImageOS[index] }

      const dropdownItem = {
        value: imageItem.name,
        name: imageItem.name,
        subTitleHtml: <></>,
        feedbackHtml: <></>
      }

      dropdownItem.subTitleHtml = (
        <Wrapper>
          <span className="small"> Architecture: {imageItem.architecture}</span>
        </Wrapper>
      )

      dropdownItem.feedbackHtml = <ImageComponentsWithOverlay components={imageItem.components} />

      selectableImages.push(dropdownItem)
    }

    return selectableImages
  }

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...state
    }

    let value = event.target.value

    if (formInputName === 'quickConnectFlag') {
      value = event.target.checked
    }

    const updatedForm = UpdateFormHelper(value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    if (formInputName === 'recommendedUseCase') {
      // Clear value
      updatedState.form = updateProductNameFormState(updatedState.form, '')
      updatedState.form = setFormValue('operationSystem', '', updatedState.form)
      updatedState.form = setSelectOptions('productName', [], updatedState.form)
      updatedState.form = setSelectOptions('operationSystem', [], updatedState.form)

      const selectableInstanceTypes = getInstanceTypeOptions(event.target.value)
      updatedState.form = setSelectOptions('productName', selectableInstanceTypes, updatedState.form)
      if (selectableInstanceTypes.length > 0) {
        const instanceTypeValue = selectableInstanceTypes[0].productName
        updatedState.form = updateProductNameFormState(updatedState.form, instanceTypeValue)
        const selectableImages = getOsOptions(instanceTypeValue, event.target.value)
        if (selectableImages.length > 0) {
          const imageValue = selectableImages[0].value
          updatedState.form = setSelectOptions('operationSystem', selectableImages, updatedState.form)
          updatedState.form = setFormValue('operationSystem', imageValue, updatedState.form)
        }
      }
    }
    if (formInputName === 'productName') {
      // Clear value
      updatedState.form.productName.selectedRecords = [value]
      updatedState.form = setFormValue('operationSystem', '', updatedState.form)
      updatedState.form = setSelectOptions('operationSystem', [], updatedState.form)
      const selectableImageOS = getOsOptions(event.target.value)

      if (selectableImageOS.length > 0) {
        const imageValue = selectableImageOS[0].value

        updatedState.form = setSelectOptions('operationSystem', selectableImageOS, updatedState.form)
        updatedState.form = setFormValue('operationSystem', imageValue, updatedState.form)
      }

      updatedState.costEstimate.costArray = getCostEstimate(value)
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
    const backTo = searchParams.get('backTo')
    switch (backTo) {
      case 'catalog':
        navigate({
          pathname: '/hardware'
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
      const productName = getFormValue('productName', state.form)
      const product = products.find((product) => product.name === productName)
      const servicePayload = { ...state.servicePayload }
      servicePayload.metadata.name = getFormValue('instanceName', state.form)
      servicePayload.metadata.productId = product.id
      servicePayload.spec.instanceCount = product.nodesCount
      servicePayload.spec.instanceSpec.instanceType = product.instanceType
      servicePayload.spec.instanceSpec.machineImage = getFormValue('operationSystem', state.form)
      servicePayload.spec.instanceSpec.sshPublicKeyNames = getFormValue('keyPairList', state.form)
      const quickConnectValue = getFormValue('quickConnectFlag', state.form) ? 1 : 2
      servicePayload.spec.instanceSpec.quickConnectEnabled = quickConnectValue
      const stateUpdated = { ...state }
      stateUpdated.showReservationModal = true
      setState(stateUpdated)

      if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_METRICS_BARE_METAL)) {
        // Call Cloud monitor api to enable the metrics for BM
        await getUsrDataForBM(product, servicePayload)
      }

      await getVnet(servicePayload)
      await createReservation(servicePayload)
    } catch (error) {
      const stateUpdated = { ...state }

      stateUpdated.showReservationModal = false
      stateUpdated.errorHideRetryMessage = false
      stateUpdated.errorDescription = null

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

  async function getUsrDataForBM(product, servicePayload) {
    if (product.instanceCategories === 'BareMetalHost') {
      try {
        const { data } = await CloudAccountService.enableCloudMonitorForBM()
        if (data) {
          servicePayload.spec.userData = data.config
        }
      } catch {
        return false
      }
    }
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
    servicePayload.spec.instanceSpec.interfaces = networks
    servicePayload.spec.instanceSpec.availabilityZone = state.networkServicePayload.spec.availabilityZone
  }

  async function createReservation(servicePayload) {
    await CloudAccountService.postComputeGroupReservation(servicePayload)

    setTimeout(() => {
      const stateUpdated = { ...state }

      stateUpdated.showReservationModal = false

      setState(stateUpdated)

      setFamilyIdSelected('')

      navigate({
        pathname: '/compute-groups'
      })
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

  return (
    <Wrapper>
      <ComputeLaunch
        state={state}
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
        emptyCatalogModal={emptyCatalogModal}
        setShowUpgradeNeededModal={setShowUpgradeNeededModal}
        category={category}
        computeReservationsPagePath="/compute"
        loading={!isPageReady}
      />
    </Wrapper>
  )
}

export default ComputeLaunchContainer
