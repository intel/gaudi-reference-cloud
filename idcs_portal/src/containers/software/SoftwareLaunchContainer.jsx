// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { useNavigate } from 'react-router'
import idcConfig, { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import SoftwareLaunch from '../../components/software/softwareLaunch/SoftwareLaunch'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import {
  UpdateFormHelper,
  getFormValue,
  isValidForm,
  setFormValue,
  setSelectOptions
} from '../../utils/updateFormHelper/UpdateFormHelper'
import useSoftwareStore from '../../store/SoftwareStore/SoftwareStore'
import useHardwareStore from '../../store/hardwareStore/HardwareStore'
import useImageStore from '../../store/imageStore/ImageStore'
import Wrapper from '../../utils/Wrapper'
import { formatCurrency, formatNumber } from '../../utils/numberFormatHelper/NumberFormatHelper'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import {
  friendlyErrorMessages,
  isErrorInsufficientCapacity,
  isErrorInsufficientCredits
} from '../../utils/apiError/apiError'
import CloudAccountService from '../../services/CloudAccountService'
import { computeCategoriesEnum } from '../../utils/Enums'
import useAppStore from '../../store/appStore/AppStore'
import ImageComponentsWithOverlay from '../../utils/imageComponents/ImageComponentsWithOverlay'

const SoftwareLaunchContainer = () => {
  // Params
  const { id } = useParams()

  // Navigation
  const navigate = useNavigate()

  const throwError = useErrorBoundary()

  // Global State
  const setPublickeys = useCloudAccountStore((state) => state.setPublickeys)
  const publicKeys = useCloudAccountStore((state) => state.publicKeys)
  const softwareDetail = useSoftwareStore((state) => state.softwareDetail)
  const getSoftware = useSoftwareStore((state) => state.getSoftware)
  const products = useHardwareStore((state) => state.products)
  const families = useHardwareStore((state) => state.families)
  const setProducts = useHardwareStore((state) => state.setProducts)
  const image = useImageStore((state) => state.image)
  const addBreadcrumCustomTitle = useAppStore((state) => state.addBreadcrumCustomTitle)
  const setImage = useImageStore((state) => state.setImage)
  const setFamilyIdSelected = useHardwareStore((state) => state.setFamilyIdSelected)
  const vnets = useCloudAccountStore((state) => state.vnets)
  const setVnets = useCloudAccountStore((state) => state.setVnets)

  // Local State
  const comingMessage =
    'Get ready for Intel-optimized software stacks hosted on performance optimized Intel compute platforms'
  const isAvailable = isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_SOFTWARE)
  const initialState = {
    instanceConfigSectionTitle: 'Instance configuration',
    publicKeysMenuSection: 'Public Keys',
    form: {
      operationSystem: {
        sectionGroup: 'operationSystem',
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Machine image:',
        placeholder: 'Please select',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: '',
        helperMessage: (
          <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'}>
            Choose the desired OS for your reservation
          </div>
        ),
        columnSize: '12'
      },
      families: {
        sectionGroup: 'configuration',
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Instance family:',
        placeholder: 'Please select',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: '',
        helperMessage: null,
        columnSize: '12'
      },
      instanceType: {
        sectionGroup: 'configuration',
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Instance type:',
        placeholder: 'Instance type',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [],
        validationMessage: '',
        columnSize: '12',
        labelButton: {
          label: 'Compare instance types',
          buttonFunction: () => onShowHideInstanceCompareModal(true)
        }
      },
      instanceName: {
        sectionGroup: 'configuration',
        type: 'text', // options = 'text ,'textArea'
        label: 'Instance name:',
        placeholder: 'Instance name',
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
        helperMessage: (
          <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'}>
            Name must be 63 characters or less, and can include letters, numbers, and ‘-‘ only.
            <br />
            It should start and end with an alphanumeric character.
          </div>
        ),
        columnSize: '12'
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
        refreshButton: {
          label: 'Refresh Keys',
          buttonFunction: () => setPublickeys()
        },
        emptyOptionsMessage: 'No keys found. Please create a key to continue.',
        columnSize: '12'
      }
    },
    isValidForm: false,
    showReservationModal: false,
    showErrorModal: false,
    errorMessage: null,
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
        buttonFunction: () => onCancel()
      }
    ],
    servicePayload: {
      metadata: {
        name: null
      },
      spec: {
        availabilityZone: null,
        instanceType: null,
        machineImage: null,
        runStrategy: 'RerunOnFailure',
        sshPublicKeyNames: [],
        interfaces: [
          {
            name: 'eth0',
            vNet: null
          }
        ]
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

  const dropdownContent = {
    imageSrc: '',
    name: '',
    value: '',
    subTitleHtml: null,
    feedbackHtml: null
  }

  const monthlyEstimate = {
    totalHour: 0,
    totalMonth: 0,
    instanceRate: 0,
    softwareRate: 0
  }

  const category = computeCategoriesEnum.singleNode

  const [showPublicKeyModal, setShowPublicKeyModal] = useState(false)
  const [showInstanceCompareModal, setShowInstanceCompareModal] = useState(false)
  const [compareProducts, setCompareProducts] = useState([])
  const [allowedInstances, setAllowedInstances] = useState([])

  const [state, setState] = useState(initialState)
  const [monthlyEstimateRates, setMonthlyEstimateRates] = useState(monthlyEstimate)
  const [showUpgradeNeededModal, setShowUpgradeNeededModal] = useState(false)

  // Hooks

  useEffect(() => {
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
    if (families.length === 0) {
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
    if (softwareDetail === null) {
      const fetchSoftware = async () => {
        try {
          await getSoftware(id)
        } catch (error) {
          throwError(error)
        }
      }
      fetchSoftware()
    }
  }, [])

  useEffect(() => {
    if (softwareDetail) {
      addBreadcrumCustomTitle(`/software/d/${id}`, softwareDetail.displayName)
      const fetchImage = async () => {
        try {
          await setImage(softwareDetail.launchImage)
        } catch (error) {
          throwError(error)
        }
      }
      fetchImage()
    }
  }, [softwareDetail])

  useEffect(() => {
    if (image) {
      setImageInstances()
    }
  }, [image])

  useEffect(() => {
    if (allowedInstances.length !== 0) {
      setForm()
    }
  }, [allowedInstances, publicKeys])

  // Functions

  function setImageInstances() {
    const imageInstances = image.instanceTypes
    const instancesTypes = products.filter((product) => imageInstances.includes(product.name))
    setAllowedInstances(instancesTypes)
    setCompareProducts(instancesTypes)
  }

  function onCancel() {
    navigate({
      pathname: '/software/d/' + id
    })
  }

  function onShowHidePublicKeyModal(status = false) {
    setShowPublicKeyModal(status)
  }

  function setForm() {
    const stateUpdated = { ...state }
    let selectedInstance = ''

    // Load key pairs
    const keys = publicKeys.map((key) => {
      return { name: key.name + ` (${key.ownerEmail})`, value: key.value }
    })
    stateUpdated.form = setSelectOptions('keyPairList', keys, stateUpdated.form)
    // Load Image
    const selectableImages = getOsOptions()
    if (selectableImages.length > 0) {
      stateUpdated.form = setFormValue('operationSystem', selectableImages[0].value, stateUpdated.form)
      stateUpdated.form = setSelectOptions('operationSystem', selectableImages, stateUpdated.form)
    }

    const selectableFamilies = getFamiliesOptions()
    const familyValue = selectableFamilies[0].value
    if (selectableFamilies.length > 0) {
      stateUpdated.form = setFormValue('families', familyValue, stateUpdated.form)
      stateUpdated.form = setSelectOptions('families', selectableFamilies, stateUpdated.form)
    }

    const instanceTypeOptions = getInstanceTypeOptions(familyValue)
    if (instanceTypeOptions.length > 0) {
      selectedInstance = instanceTypeOptions[0].value
      stateUpdated.form = setFormValue('instanceType', selectedInstance, stateUpdated.form)
      stateUpdated.form = setSelectOptions('instanceType', instanceTypeOptions, stateUpdated.form)
    }
    getMonthlyEstimate(selectedInstance)
    setState(stateUpdated)
  }

  function getFamiliesOptions() {
    const selectableFamilies = []
    const uniqueFamily = []

    for (const index in allowedInstances) {
      const item = { ...allowedInstances[index] }

      if (!uniqueFamily.includes(item.familyDisplayName)) {
        uniqueFamily.push(item.familyDisplayName)

        const copyDropdownItem = { ...dropdownContent }

        copyDropdownItem.value = item.familyDisplayName
        copyDropdownItem.name = item.familyDisplayName

        const feedback = (
          <ReactMarkdown linkTarget="_blank" remarkPlugins={[remarkGfm]} components={{ p: 'span' }}>
            {item.information}
          </ReactMarkdown>
        )
        copyDropdownItem.feedbackHtml = <div className="valid-feedback">{feedback}</div>

        selectableFamilies.push(copyDropdownItem)
      }
    }

    return selectableFamilies
  }

  function getInstanceTypeOptions(familyDisplayName) {
    const instancesTypes = allowedInstances.filter((product) => product.familyDisplayName === familyDisplayName)
    const instanceTypeOptions = []

    for (const index in instancesTypes) {
      const item = instancesTypes[index]
      instanceTypeOptions.push({
        name: item.displayName,
        dropSelect: getPricing(item),
        value: item.instanceType,
        subTitleHtml: (
          <>
            <span>{item.description} </span>
          </>
        )
      })
    }

    return instanceTypeOptions
  }

  function getMonthlyEstimate(instanceType) {
    const instance = allowedInstances.find((x) => x.instanceType === instanceType)
    const copyMonthlyEstimate = { ...monthlyEstimate }
    const hourInstanceRate = instance.rate * 60
    const hourSoftwareRate = softwareDetail.rate * 60

    const monthInstanceRate = hourInstanceRate * 30
    const monthSoftwareRate = hourSoftwareRate * 30

    const totalHourly = formatNumber(hourInstanceRate + hourSoftwareRate, 2)
    const totalMonthly = formatNumber(monthInstanceRate + monthSoftwareRate, 2)

    copyMonthlyEstimate.totalHour = totalHourly === 0 ? 'Free' : formatCurrency(totalHourly)
    copyMonthlyEstimate.totalMonth = totalMonthly === 0 ? 'Free' : formatCurrency(totalMonthly)
    copyMonthlyEstimate.instanceRate =
      monthInstanceRate === 0 ? 'Free' : formatCurrency(formatNumber(monthInstanceRate))
    copyMonthlyEstimate.softwareRate =
      monthSoftwareRate === 0 ? 'Free' : formatCurrency(formatNumber(monthSoftwareRate))

    setMonthlyEstimateRates(copyMonthlyEstimate)
  }

  function getPricing(instanceType) {
    const rate = instanceType.rate

    let result = null
    const cost = formatNumber(rate * 60, 2)
    result = cost === 0 ? 'Free' : `$${cost} / hour`

    return result
  }

  function getOsOptions() {
    const selectableImages = []

    const imageItem = { ...image }

    const copyDropdownItem = { ...dropdownContent }

    copyDropdownItem.imageSrc = softwareDetail.imageSource
    copyDropdownItem.value = imageItem.name
    copyDropdownItem.name = softwareDetail.displayName
    copyDropdownItem.dropSelect = getPricing(softwareDetail)

    copyDropdownItem.subTitleHtml = (
      <Wrapper>
        <span className="small"> Architecture: {imageItem.architecture}</span>
      </Wrapper>
    )

    copyDropdownItem.feedbackHtml = <ImageComponentsWithOverlay components={imageItem.components} />

    selectableImages.push(copyDropdownItem)

    return selectableImages
  }

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    if (formInputName === 'families') {
      // Clear value
      updatedState.form = setFormValue('instanceType', '', updatedState.form)
      updatedState.form = setSelectOptions('instanceType', [], updatedState.form)

      const selectableInstanceTypes = getInstanceTypeOptions(event.target.value)
      updatedState.form = setSelectOptions('instanceType', selectableInstanceTypes, updatedState.form)
      if (selectableInstanceTypes.length > 0) {
        const instanceTypeValue = selectableInstanceTypes[0].value
        updatedState.form = setFormValue('instanceType', instanceTypeValue, updatedState.form)
        getMonthlyEstimate(instanceTypeValue)
      }
    }

    if (formInputName === 'instanceType') {
      // Clear value
      getMonthlyEstimate(event.target.value)
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

  function afterPubliKeyCreate() {
    setPublickeys()
  }

  function onSubmit() {
    submitForm()
  }

  async function submitForm() {
    try {
      const instanceType = getFormValue('instanceType', state.form)

      const servicePayload = { ...state.servicePayload }
      servicePayload.metadata.name = getFormValue('instanceName', state.form)
      servicePayload.spec.instanceType = instanceType
      servicePayload.spec.machineImage = getFormValue('operationSystem', state.form)
      servicePayload.spec.sshPublicKeyNames = getFormValue('keyPairList', state.form)

      const stateUpdated = { ...state }
      stateUpdated.showReservationModal = true
      setState(stateUpdated)
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
        if (isErrorInsufficientCredits(error)) {
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
    servicePayload.spec.interfaces = networks
    servicePayload.spec.availabilityZone = state.networkServicePayload.spec.availabilityZone
  }

  async function createReservation(servicePayload) {
    await CloudAccountService.postComputeReservation(servicePayload)

    setTimeout(() => {
      const stateUpdated = { ...state }

      stateUpdated.showReservationModal = false

      setState(stateUpdated)

      setFamilyIdSelected('')

      navigate({
        pathname: '/compute'
      })
    }, state.timeoutMiliseconds)
  }

  function onClickCloseErrorModal() {
    const stateUpdated = { ...state }

    stateUpdated.showErrorModal = false

    setState(stateUpdated)
  }

  function onShowHideInstanceCompareModal(status = false) {
    setShowInstanceCompareModal(status)
  }

  function afterInstanceSelected(instance) {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(instance.familyDisplayName, 'families', updatedState.form)
    updatedState.isValidForm = isValidForm(updatedForm)
    updatedState.form = updatedForm

    const selectableInstanceTypes = getInstanceTypeOptions(instance.familyDisplayName)
    updatedState.form = setSelectOptions('instanceType', selectableInstanceTypes, updatedState.form)

    if (selectableInstanceTypes.length > 0) {
      updatedState.form = setFormValue('instanceType', instance.name, updatedState.form)
      const selectableImages = getOsOptions(instance.name, instance.familyDisplayName)
      if (selectableImages.length > 0) {
        const imageValue = selectableImages[0].value
        updatedState.form = setSelectOptions('operationSystem', selectableImages, updatedState.form)
        updatedState.form = setFormValue('operationSystem', imageValue, updatedState.form)
      }
    }

    setState(updatedState)
  }

  return (
    <SoftwareLaunch
      isAvailable={isAvailable}
      mainTitle={`Launch software ${softwareDetail?.displayName ?? ''}`}
      softwareDetail={softwareDetail}
      comingMessage={comingMessage}
      monthlyEstimateRates={monthlyEstimateRates}
      state={state}
      showPublicKeyModal={showPublicKeyModal}
      onShowHidePublicKeyModal={onShowHidePublicKeyModal}
      afterPubliKeyCreate={afterPubliKeyCreate}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      onClickCloseErrorModal={onClickCloseErrorModal}
      onChangeDropdownMultiple={onChangeDropdownMultiple}
      showUpgradeNeededModal={showUpgradeNeededModal}
      setShowUpgradeNeededModal={setShowUpgradeNeededModal}
      category={category}
      showInstanceCompareModal={showInstanceCompareModal}
      onShowHideInstanceCompareModal={onShowHideInstanceCompareModal}
      afterInstanceSelected={afterInstanceSelected}
      products={compareProducts}
    />
  )
}

export default SoftwareLaunchContainer
