// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useSuperComputerStore from '../../store/superComputer/SuperComputerStore'
import SuperComputerInfo from '../../components/superComputer/superComputerInfo/SuperComputerInfo'
import { UpdateFormHelper, getFormValue, isValidForm } from '../../utils/updateFormHelper/UpdateFormHelper'
import SuperComputerService from '../../services/SuperComputerService'
import { getErrorMessageFromCodeAndMessage } from '../../utils/apiError/apiError'
import useToastStore from '../../store/toastStore/ToastStore'
import { BsDownload } from 'react-icons/bs'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'

const SuperComputerInfoContainer = (props) => {
  // props
  const downloadKubeConfig = props.downloadKubeConfig
  // *****
  // Global state
  // *****
  const clusterDetail = useSuperComputerStore((state) => state.clusterDetail)
  const setDebounceDetailRefresh = useSuperComputerStore((state) => state.setDebounceDetailRefresh)
  const showError = useToastStore((state) => state.showError)
  // *****
  // local state
  // *****

  const actions = [
    {
      label: 'DownloadAdmin',
      name: (
        <>
          {' '}
          <BsDownload />
          Download{' '}
        </>
      ),
      type: 'button',
      func: (data) => {
        downloadKubeConfig(data.name, data.uuid, false).catch(() => {})
      }
    }
  ]

  const actionsV2 = [
    {
      label: 'DownloadReadOnly',
      name: (
        <>
          {' '}
          <BsDownload /> Download readonly{' '}
        </>
      ),
      type: 'button',
      func: (data) => {
        downloadKubeConfig(data.name, data.uuid, true).catch(() => {})
      }
    },
    {
      label: 'DownloadAdmin',
      name: (
        <>
          {' '}
          <BsDownload /> Download admin{' '}
        </>
      ),
      type: 'button',
      func: (data) => {
        downloadKubeConfig(data.name, data.uuid, false).catch(() => {})
      }
    }
  ]

  const displayInfoInitial = [
    { label: 'Id:', field: 'uuid', value: '' },
    {
      label: 'Kubernetes version:',
      field: 'k8sversion',
      value: '',
      actions: [{ label: 'Upgrade', type: 'link', func: () => onUpgradeModal(true) }]
    },
    {
      label: 'Kubeconfig',
      field: 'kubeconfig',
      value: '',
      actions: isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_IKS_KUBE_CONFIG) ? [...actionsV2] : [...actions]
    },
    { label: 'State:', field: 'clusterstate', value: '', formula: 'status' },
    { label: 'Node group qty:', field: 'nodegroups', value: '', formula: 'length' },
    { label: 'Load balancer qty:', field: 'vips', value: '', formula: 'length' },
    { label: 'Location/Region:', field: 'region', value: '' }
  ]

  const upgradeModalInitital = {
    show: false,
    onHide: () => onUpgradeModal(false),
    centered: true,
    closeButton: true
  }

  const upgradeFormInitial = {
    form: {
      versions: {
        sectionGroup: 'upgrade',
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Select the kubernetes Version:',
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
        helperMessage: ''
      }
    },
    isValidForm: false
  }

  const payloadUpgrade = {
    k8sversionname: ''
  }

  const [displayInfo, setDisplayInfo] = useState(displayInfoInitial)
  const [upgradeModal, setUpgradeModal] = useState(upgradeModalInitital)
  const [upgradeForm, setUpgradeForm] = useState(upgradeFormInitial)

  // *****
  // Hooks
  // *****
  useEffect(() => {
    updateClusterInfo()
  }, [clusterDetail])

  // *****
  // functions
  // *****

  const updateClusterInfo = () => {
    const ClusterInfoUpdated = []
    for (const index in displayInfo) {
      const item = { ...displayInfo[index] }
      const netowrk = { ...clusterDetail.network }
      if (item.formula === 'length') {
        item.value = clusterDetail[item.field].length
      } else if (item.formula === 'status') {
        item.value = getClusterStatus(clusterDetail)
      } else if (item.field === 'region') {
        item.value = netowrk[item.field]
      } else {
        item.value = clusterDetail[item.field]
      }
      ClusterInfoUpdated.push(item)
    }
    setDisplayInfo(ClusterInfoUpdated)

    const upgradeFormUpdated = { ...upgradeForm }
    const form = upgradeFormUpdated.form
    const formElement = form.versions
    const options = []
    for (const item in clusterDetail.upgradek8sversionavailable) {
      options.push({
        name: clusterDetail.upgradek8sversionavailable[item],
        value: clusterDetail.upgradek8sversionavailable[item]
      })
    }
    formElement.options = options
    formElement.value = options.length === 1 ? options[0].value : ''
    formElement.isValid = Boolean(formElement.value)
    form.versions = formElement
    upgradeFormUpdated.form = form
    upgradeFormUpdated.isValidForm = Boolean(formElement.value)
    setUpgradeForm(upgradeFormUpdated)
  }

  function getClusterStatus(cluster) {
    let message = 'No status'
    if (cluster) {
      const { clusterstatus } = cluster
      if (clusterstatus.errorcode) {
        message = clusterstatus.message
          ? getErrorMessageFromCodeAndMessage(clusterstatus.errorcode, clusterstatus.message)
          : 'No Status'
      } else {
        message = clusterstatus.message ? clusterstatus.state : 'No Status'
      }
    }
    return message
  }

  const onUpgradeModal = (show) => {
    setUpgradeModal({ ...upgradeModal, show })
  }

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...upgradeForm
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setUpgradeForm(updatedState)
  }

  async function submitUpgradeK8sVersion() {
    try {
      const payload = { ...payloadUpgrade }
      payload.k8sversionname = getFormValue('versions', upgradeForm.form)
      const response = await SuperComputerService.upgradeCluster(payload, clusterDetail.uuid)
      onUpgradeModal(false)
      setDebounceDetailRefresh(true)
      if (!response.data) {
        displayAlertSection(response.statusText, 'error')
      }
    } catch (error) {
      onUpgradeModal(false)
      if (error.response) {
        displayAlertSection(error.response.data.message, 'error')
      } else {
        displayAlertSection(error.message, 'error')
      }
    }
  }

  function displayAlertSection(message, alertType) {
    switch (alertType) {
      default:
        showError(message, false)
        break
    }
  }

  return (
    <SuperComputerInfo
      displayInfo={displayInfo}
      clusterDetail={clusterDetail}
      upgradeModal={upgradeModal}
      upgradeForm={upgradeForm}
      onChangeInput={onChangeInput}
      submitUpgradeK8sVersion={submitUpgradeK8sVersion}
    />
  )
}

export default SuperComputerInfoContainer
