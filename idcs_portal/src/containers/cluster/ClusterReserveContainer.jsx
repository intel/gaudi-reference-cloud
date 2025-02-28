// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import ClusterReserve from '../../components/cluster/clusterReserve/ClusterReserve'
import {
  UpdateFormHelper,
  getFormValue,
  hideFormElement,
  isValidForm,
  setFormValue,
  setSelectOptions,
  updateDictionary,
  validateDictionaryItems,
  showFormRequiredFields
} from '../../utils/updateFormHelper/UpdateFormHelper'
import { useNavigate } from 'react-router'
import ClusterService from '../../services/ClusterService'
import useClusterStore from '../../store/clusterStore/ClusterStore'
import idcConfig from '../../config/configurator'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import CloudAccountService from '../../services/CloudAccountService'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useToastStore from '../../store/toastStore/ToastStore'
import {
  friendlyErrorMessages,
  isErrorInsufficientCapacity,
  isErrorInsufficientCredits
} from '../../utils/apiError/apiError'
import { toastMessageEnum } from '../../utils/Enums'

const getCustomMessage = (messageType, field = null) => {
  let message = null

  switch (messageType) {
    case 'clusterConfiguration':
      message = (
        <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'}>
          Max length 63 characters. Letters, numbers and ‘- ‘ accepted. Name should start and end with an alphanumeric
          character.
        </div>
      )
      break
    case 'clusterSize':
      message = (
        <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'}>
          Select Cluster size.
          <br />
        </div>
      )
      break
    case 'clusterRegion':
      message = (
        <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'}>
          Select Region.
          <br />
        </div>
      )
      break
    case 'tagMessage':
      message = (
        <div className="valid-feedback" intc-id={'InstanceDetailsValidMessage'}>
          Max length 63 characters. <br />
          Letters, numbers, ‘- ‘, ‘_’ and ‘.’ accepted
        </div>
      )
      break
    case 'tagMessageInvalid':
      message = (
        <div intc-id={'InstanceDetailsValidMessage'}>
          {`${field} cannot be empty`} <br />
        </div>
      )
      break
    default:
      break
  }

  return message
}

const ClusterReserveContainer = (props) => {
  // local state
  const mainTitle = 'Launch Kubernetes cluster'
  const mainSubtitle = ''
  const clusterMenuSection = 'Cluster details and configuration'
  const providerMenuSection = 'Provider'
  const configurationMenuSection = ''
  const tagsMenuSection = 'Tags'
  const tagsSubtitle =
    'A tag is a label that you assign to an instance. Each tag consists of a key and an optional value, both of which you define.'
  const timeoutMiliseconds = 4000
  const vnets = useCloudAccountStore((state) => state.vnets)
  const setVnets = useCloudAccountStore((state) => state.setVnets)

  const errorModalInitialState = {
    message: '',
    errorMessage: '',
    showErrorModal: false,
    titleMessage: 'Could not launch your cluster',
    errorDescription: 'There was an error while processing your cluster',
    hideRetryMessage: false
  }
  const servicePayload = {
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
  const clusterCreatePayload = {
    description: '',
    k8sversionname: '',
    name: '',
    runtimename: '',
    tags: []
  }
  const navigationBottom = [
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

  const tagInput = {
    isValidRow: false,
    key: {
      label: 'Key:',
      value: '',
      isValid: false,
      isTouched: false,
      maxLength: 63,
      validationRules: {
        isRequired: true,
        checkMaxLength: true,
        onlyAlphaNumLower: true
      },
      helperMessage: getCustomMessage('tagMessage'),
      validationMessage: ''
    },
    value: {
      label: 'Value:',
      value: '',
      isValid: false,
      isTouched: false,
      maxLength: 63,
      validationRules: {
        isRequired: true,
        checkMaxLength: true,
        onlyAlphaNumLower: true
      },
      helperMessage: getCustomMessage('tagMessage'),
      validationMessage: ''
    }
  }

  const initialForm = {
    clusterName: {
      sectionGroup: 'clusterConfiguration',
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
        checkMaxLength: true,
        onlyAlphaNumLower: true
      },
      validationMessage: '', // Error message to display to the user
      helperMessage: getCustomMessage('clusterConfiguration')
    },
    clusterRuntime: {
      sectionGroup: 'clusterConfiguration',
      type: 'dropdown', // options = 'text ,'textArea'
      label: 'Container runtime:',
      placeholder: 'Please select runtime',
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
      hidden: true
    },
    clusterK8sVersion: {
      sectionGroup: 'clusterConfiguration',
      type: 'dropdown', // options = 'text ,'textArea'
      label: 'Select cluster kubernetes version:',
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
      helperMessage: ''
    }
  }

  const [isValid, setIsValid] = useState(false)

  const [form, setForm] = useState(initialForm)
  const [showReservationModal, setShowReservationModal] = useState(false)
  const setClustersRuntimes = useClusterStore((state) => state.setClustersRuntimes)
  const clustersRuntimes = useClusterStore((state) => state.clustersRuntimes)
  const clusterProducts = useClusterStore((state) => state.clusterProducts)
  const setClusterProducts = useClusterStore((state) => state.setClusterProducts)
  const setCurrentSelectedCluster = useClusterStore((state) => state.setCurrentSelectedCluster)
  const [showUpgradeNeededModal, setShowUpgradeNeededModal] = useState(false)
  const [errorModalState, setErrorModalState] = useState(errorModalInitialState)
  const [emptyCatalogModal, setEmptyCatalogModal] = useState(false)
  const [isPageReady, SetIsPageReady] = useState(false)
  const showError = useToastStore((state) => state.showError)
  const throwError = useErrorBoundary()

  // Navigation
  const navigate = useNavigate()

  // functions

  function onCancel() {
    navigate({
      pathname: '/cluster'
    })
  }

  function handleClickActionTag(index, actionType) {
    const formUpdated = { ...form }
    const clusterTagsCopy = { ...formUpdated.clusterTags }

    if (actionType.toUpperCase() === 'DELETE') {
      clusterTagsCopy.options.splice(index, 1)
      const isValid = clusterTagsCopy.options.filter((option) => !option.isValidRow)
      clusterTagsCopy.isValid = isValid.length === 0
    } else {
      // Add
      const maxLength = clusterTagsCopy.maxLength
      if (clusterTagsCopy.options.length >= maxLength) {
        return
      }
      // Validate previuos rows
      const optionsUpdated = []
      for (const index in clusterTagsCopy.options) {
        const option = clusterTagsCopy.options[index]
        option.key.isTouched = !option.key.isValid
        option.key.validationMessage = getCustomMessage('tagMessageInvalid', 'key')
        option.value.isTouched = !option.value.isValid
        option.value.validationMessage = getCustomMessage('tagMessageInvalid', 'key')
        optionsUpdated.push(option)
      }
      clusterTagsCopy.options = optionsUpdated
      clusterTagsCopy.options.push(tagInput)
    }

    formUpdated.clusterTags = clusterTagsCopy
    setIsValid(isValidForm(formUpdated))
    setForm(formUpdated)
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
    servicePayload.spec.interfaces = networks
    servicePayload.spec.availabilityZone = networkServicePayload.spec.availabilityZone
  }

  function handleTagChange(event, field, index) {
    const value = event.target.value
    const formUpdated = { ...form }
    const clusterTagsCopy = formUpdated.clusterTags
    const clusterOptionsCopy = clusterTagsCopy.options
    const tagUpdated = updateDictionary(value, field, clusterOptionsCopy[index])
    clusterOptionsCopy[index] = tagUpdated
    clusterTagsCopy.options = clusterOptionsCopy
    clusterTagsCopy.isValid = validateDictionaryItems(clusterOptionsCopy)
    formUpdated.clusterTags = clusterTagsCopy
    setIsValid(isValidForm(formUpdated))
    setForm(formUpdated)
  }

  function onChangeInput(event, formInputName) {
    let value = null
    value = event.target.value
    const formCopy = {
      ...form
    }

    const updatedForm = UpdateFormHelper(value, formInputName, formCopy)

    const isValid = isValidForm(updatedForm)

    setForm(updatedForm)

    setIsValid(isValid)
  }

  function showRequiredFields() {
    let formCopy = { ...form }
    // Mark regular Inputs
    const updatedForm = showFormRequiredFields(formCopy)
    // Create toast
    showError(toastMessageEnum.formValidationError, false)
    formCopy = updatedForm
    setForm(formCopy)
  }

  async function onSubmit() {
    try {
      if (!isValid) {
        showRequiredFields()
        return
      }
      // Making sure the process Modal will appear until cluster creation success/failure.
      setShowReservationModal(true)
      await createCluster()
    } catch (error) {
      setShowReservationModal(false)
      if (error.response) {
        setErrorModalState({ ...errorModalInitialState, showErrorModal: true, message: error.response.data.message })
      } else {
        setErrorModalState({ ...errorModalInitialState, showErrorModal: true, message: error.message })
      }
    }
  }
  async function createCluster() {
    try {
      const ok = [200, 201]
      await getVnet(servicePayload)
      const clusterCreatePayloadCopy = { ...clusterCreatePayload }
      clusterCreatePayloadCopy.name = getFormValue('clusterName', form)
      clusterCreatePayloadCopy.description = ''
      clusterCreatePayloadCopy.k8sversionname = getFormValue('clusterK8sVersion', form)
      clusterCreatePayloadCopy.runtimename = getFormValue('clusterRuntime', form)
      clusterCreatePayloadCopy.instanceType = clusterProducts[0].instanceType
      const response = await ClusterService.createClusterFunc(clusterCreatePayloadCopy)
      setTimeout(() => {
        navigate({
          pathname: '/cluster'
        })
      }, timeoutMiliseconds)
      setCurrentSelectedCluster(clusterCreatePayloadCopy.name)
      if (ok.includes(response.status)) {
        setErrorModalState({ ...errorModalInitialState, showErrorModal: false })
        setShowReservationModal(true)
        return response
      }
      if (response.code) {
        const message = response.message ? response.message : ''
        setErrorModalState({ ...errorModalInitialState, showErrorModal: true, message })
        setShowReservationModal(false)
      }
      return response
    } catch (error) {
      setShowReservationModal(false)
      const errorModalStateUpdated = { ...errorModalInitialState }
      errorModalStateUpdated.errorHideRetryMessage = false
      errorModalStateUpdated.errorTitleMessage = 'Could not launch your cluster'
      errorModalStateUpdated.errorDescription = ''

      if (error.response) {
        const errData = error.response.data
        const errCode = errData.code
        if (isErrorInsufficientCredits(error)) {
          // No Credits
          setShowUpgradeNeededModal(true)
        } else if (errCode === 11) {
          // No Quota
          errorModalStateUpdated.showErrorModal = true
          errorModalStateUpdated.errorMessage = error.response.data.message
          errorModalStateUpdated.errorHideRetryMessage = true
        } else if (isErrorInsufficientCapacity(error.response.data.message)) {
          errorModalStateUpdated.showErrorModal = true
          errorModalStateUpdated.errorDescription = friendlyErrorMessages.insufficientCapacity
          errorModalStateUpdated.errorMessage = ''
          errorModalStateUpdated.errorHideRetryMessage = true
        } else {
          errorModalStateUpdated.showErrorModal = true
          errorModalStateUpdated.errorMessage = error.response.data.message
        }
      } else {
        errorModalStateUpdated.showErrorModal = true
        errorModalStateUpdated.errorMessage = error.message
      }

      setErrorModalState(errorModalStateUpdated)
    }
  }

  function onClickCloseErrorModal() {
    setErrorModalState({ ...errorModalState, showErrorModal: false })
  }

  function getK8sVersionsByRuntime(data, runtimeToMatch) {
    for (const item of data) {
      if (item.runtimename === runtimeToMatch) {
        return item.k8sversionname
      }
    }
    return []
  }

  function sortObjectsByKey(arr, key, direction) {
    if (direction === 'asc') {
      return arr.sort((a, b) => parseFloat(a[key]) - parseFloat(b[key]))
    } else if (direction === 'desc') {
      return arr.sort((a, b) => parseFloat(b[key]) - parseFloat(a[key]))
    } else {
      throw new Error('Invalid direction. Use "asc" or "desc".')
    }
  }

  const setK8sVersionOptions = (updatedForm) => {
    let k8sversionlist = []
    const k8sversionOptions = []
    k8sversionlist = getK8sVersionsByRuntime(clustersRuntimes, updatedForm.clusterRuntime.value)
    for (let j = 0; j < k8sversionlist.length; j++) {
      k8sversionOptions.push({
        name: k8sversionlist[j],
        value: k8sversionlist[j],
        displayName: k8sversionlist[j]
      })
    }
    let newFormUpgraded = setSelectOptions(
      'clusterK8sVersion',
      sortObjectsByKey(k8sversionOptions, 'value', 'desc'),
      updatedForm
    )
    if (k8sversionOptions.length === 1) {
      newFormUpgraded = setFormValue('clusterK8sVersion', k8sversionOptions[0].value, newFormUpgraded)
    }
    return newFormUpgraded
  }

  useEffect(() => {
    const fetch = async () => {
      try {
        const promises = [setClustersRuntimes()]
        if (clusterProducts.length === 0) promises.push(setClusterProducts())
        if (vnets.length === 0) promises.push(setVnets())
        await Promise.all(promises)
        SetIsPageReady(true)
      } catch (error) {
        throwError(error)
      }
    }
    fetch()
  }, [])

  useEffect(() => {
    if (isPageReady && clusterProducts.length === 0) {
      setEmptyCatalogModal(true)
    }
  }, [isPageReady])

  useEffect(() => {
    const runtimeOptions = []
    for (let i = 0; i < clustersRuntimes.length; i++) {
      runtimeOptions.push({
        name: clustersRuntimes[i].runtimename,
        value: clustersRuntimes[i].runtimename,
        displayName: clustersRuntimes[i].runtimename
      })
    }
    let updatedForm = setSelectOptions('clusterRuntime', runtimeOptions, form)
    if (runtimeOptions.length === 1) {
      updatedForm = setFormValue('clusterRuntime', runtimeOptions[0].value, updatedForm)
      updatedForm = hideFormElement('clusterRuntime', true, updatedForm)
    } else if (runtimeOptions.length > 1) {
      updatedForm = hideFormElement('clusterRuntime', false, updatedForm)
    }
    updatedForm = setK8sVersionOptions(updatedForm)
    setForm(updatedForm)
  }, [clustersRuntimes])

  return (
    <ClusterReserve
      mainSubtitle={mainSubtitle}
      mainTitle={mainTitle}
      navigationBottom={navigationBottom}
      clusterMenuSection={clusterMenuSection}
      form={form}
      errorDescription={errorModalState.errorDescription}
      providerMenuSection={providerMenuSection}
      configurationMenuSection={configurationMenuSection}
      tagsMenuSection={tagsMenuSection}
      tagsSubtitle={tagsSubtitle}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput}
      onClickActionTag={handleClickActionTag}
      onChangeTagValue={handleTagChange}
      titleMessage={errorModalState.titleMessage}
      errorMessage={errorModalState.errorMessage}
      showUpgradeNeededModal={showUpgradeNeededModal}
      setShowUpgradeNeededModal={setShowUpgradeNeededModal}
      showReservationModal={showReservationModal}
      showErrorModal={errorModalState.showErrorModal}
      errorHideRetryMessage={errorModalState.errorHideRetryMessage}
      emptyCatalogModal={emptyCatalogModal}
      onClickCloseErrorModal={onClickCloseErrorModal}
      clusterProduct={clusterProducts[0] ?? null}
      isPageReady={isPageReady}
    />
  )
}

export default ClusterReserveContainer
