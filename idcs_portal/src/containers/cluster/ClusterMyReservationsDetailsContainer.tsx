// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import {
  BsDashCircle,
  BsCheckCircle,
  BsPlayCircle,
  BsNodePlus,
  BsFillPlugFill,
  BsTrash3,
  BsDatabase,
  BsCopy,
  BsDownload
} from 'react-icons/bs'

import { useNavigate, useParams } from 'react-router'
import ClusterService from '../../services/ClusterService'
import { UpdateFormHelper, getFormValue, isValidForm } from '../../utils/updateFormHelper/UpdateFormHelper'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useClusterStore from '../../store/clusterStore/ClusterStore'
import { getErrorMessageFromCodeAndMessage } from '../../utils/apiError/apiError'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import ClusterMyReservationsDetail from '../../components/cluster/clusterMyReservations/ClusterMyReservationsDetail'
import StateTooltipCell from '../../utils/gridPagination/cellRender/StateTooltipCell'
import useToastStore from '../../store/toastStore/ToastStore'
import { useCopy } from '../../hooks/useCopy'
import ClusterSecurityRulesContainer from './ClusterSecurityRulesContainer'

const getActionItemLabel = (text: string, statusStep: string | null = null, option: any = null): JSX.Element => {
  let message = null

  switch (text) {
    case 'Start':
      message = (
        <>
          {' '}
          <BsPlayCircle /> {text}{' '}
        </>
      )
      break
    case 'Deleting':
    case 'Terminating':
    case 'DeletePending':
      message = (
        <>
          {' '}
          <BsDashCircle /> {text}{' '}
        </>
      )
      break
    case 'Storage':
      message = (
        <>
          <BsDatabase /> Add Storage{' '}
        </>
      )
      break
    case 'Delete':
      message = (
        <>
          {' '}
          <BsTrash3 /> {text}{' '}
        </>
      )
      break
    case 'Pending':
    case 'Provisioning':
    case 'Error':
    case 'Updating':
      message = (
        <div className="d-flex flex-row bd-highlight">
          {option && (
            <>
              <div>{'State: '}</div>
              <StateTooltipCell statusStep={statusStep} text={text} spinnerAtTheEnd />
            </>
          )}
          {!option && <StateTooltipCell statusStep={statusStep} text={text} />}
        </div>
      )
      break
    case 'Active':
    case 'Ready':
      message = (
        <>
          <BsCheckCircle /> {text}{' '}
        </>
      )
      break
    case 'addNodeGroup':
      message = (
        <>
          {' '}
          <BsNodePlus /> Add Node Group{' '}
        </>
      )
      break
    case 'addLoadBalancer':
      message = (
        <>
          {' '}
          <BsNodePlus /> Add Load Balancer{' '}
        </>
      )
      break
    case 'Connect via SSH':
      message = (
        <>
          {' '}
          <BsFillPlugFill /> {text}{' '}
        </>
      )
      break
    case 'Copy admin':
    case 'Copy readonly':
    case 'Copy':
      message = (
        <>
          {' '}
          <BsCopy /> {text}{' '}
        </>
      )
      break
    case 'Download admin':
    case 'Download readonly':
    case 'Download':
      message = (
        <>
          {' '}
          <BsDownload /> {text}{' '}
        </>
      )
      break
    default:
      message = <> {text} </>
      break
  }

  return message
}

const ClusterMyReservationsDetailsContainer = (): JSX.Element => {
  const emptyStorageInitial = {
    title: 'Storage not found',
    subTitle: 'Your cluster has no volume storage',
    action: {
      type: 'redirect',
      disabled: true,
      href: '/cluster/addstorage',
      label: 'Add Storage',
      btnClass: ''
    }
  }

  // *****
  // cluster table structure
  // *****

  const actionsOptionsV2 = [
    {
      id: 'deletecluster',
      name: getActionItemLabel('Delete'),
      status: ['Updating', 'Active', 'Error'],
      label: 'Delete cluster instance',
      buttonLabel: 'Delete'
    },
    {
      id: 'kubeconfigItemText',
      name: getActionItemLabel('Kubeconfig'),
      status: ['Active'],
      label: 'Kubeconfig',
      buttonLabel: 'Kubeconfig',
      isTextItem: true
    },
    {
      id: 'DownloadReadOnly',
      name: getActionItemLabel('Download readonly'),
      status: ['Active'],
      label: 'DownloadReadOnly',
      buttonLabel: 'Download'
    },
    {
      id: 'DownloadReadWrite',
      name: getActionItemLabel('Download admin'),
      status: ['Active'],
      label: 'DownloadReadWrite',
      buttonLabel: 'Download'
    }
  ]

  const actionsOptions = [
    {
      id: 'deletecluster',
      name: getActionItemLabel('Delete'),
      status: ['Updating', 'Active', 'Error'],
      label: 'Delete cluster instance',
      buttonLabel: 'Delete'
    },
    {
      id: 'kubeconfigItemText',
      name: getActionItemLabel('Kubeconfig'),
      status: ['Active'],
      label: 'Kubeconfig',
      buttonLabel: 'Kubeconfig',
      isTextItem: true
    },
    {
      id: 'DownloadReadWrite',
      name: getActionItemLabel('Download'),
      status: ['Active'],
      label: 'DownloadReadOnly',
      buttonLabel: 'Download'
    }
  ]

  const ilbactionsOptions = [
    {
      id: 'deleteilb',
      name: getActionItemLabel('Delete'),
      status: ['Active', 'Error'],
      label: 'Delete load balancer',
      buttonLabel: 'Delete'
    }
  ]

  const detailPanelOptions = [
    {
      id: 'addNodeGroup',
      name: getActionItemLabel('addNodeGroup'),
      status: ['Active'],
      label: 'Add Node Group',
      buttonLabel: 'Add Node Group'
    },
    {
      id: 'addLoadBalancer',
      name: getActionItemLabel('addLoadBalancer'),
      status: ['Active'],
      label: 'Add Load Balancer',
      buttonLabel: 'Add Load Balancer'
    },
    {
      id: 'addStorage',
      name: getActionItemLabel('Storage'),
      status: ['Active'],
      label: 'Add Storage',
      buttonLabel: 'Add Storage'
    }
  ]

  // *****
  // load balancers table structure
  // *****

  const lbcolumns = [
    {
      columnName: 'Load Balancer Name',
      targetColumn: 'lb-name'
    },
    {
      columnName: 'State',
      targetColumn: 'state'
    },
    {
      columnName: 'IP',
      targetColumn: 'ip',
      className: 'text-end'
    },
    {
      columnName: 'Type',
      targetColumn: 'type'
    },
    {
      columnName: 'Port',
      targetColumn: 'port'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions'
    }
  ]

  // *****
  // tabs structure
  // *****

  const tabsTitleInitial = [
    {
      label: 'Details',
      visible: true,
      id: 'details'
    },
    {
      label: 'Compute',
      visible: true,
      id: 'workerNodeGroups'
    },
    {
      label: 'Load Balancers (qty)',
      visible: true,
      id: 'loadBalancers'
    },
    {
      label: 'Storage',
      visible: false,
      id: 'storage'
    },
    {
      label: 'Security',
      visible: false,
      id: 'security'
    },
    {
      label: 'Cloud Monitor',
      visible: false,
      id: 'metrics'
    }
  ]

  const optionsKubeConfigV2 = [
    {
      label: 'DownloadReadOnly',
      name: getActionItemLabel('Download readonly'),
      type: 'link',
      func: (data: any) => {
        downloadKubeConfig(data.name, data.uuid, true).catch(() => {})
      }
    },
    {
      label: 'DownloadReadWrite',
      name: getActionItemLabel('Download admin'),
      type: 'link',
      func: (data: any) => {
        downloadKubeConfig(data.name, data.uuid, false).catch(() => {})
      }
    }
  ]

  const optionsKubeConfig = [
    {
      label: 'DownloadReadWrite',
      name: getActionItemLabel('Download'),
      type: 'link',
      func: (data: any) => {
        downloadKubeConfig(data.name, data.uuid, false).catch(() => {})
      }
    }
  ]

  const tabDetailsInitial: any = [
    {
      // Details Tab
      tapTitle: null,
      tapConfig: { type: 'custom' },
      fields: [
        { label: 'Id:', field: 'uuid', value: '' },
        {
          label: 'K8s version:',
          field: 'k8sversion',
          value: '',
          actions: [
            {
              label: 'Upgrade',
              type: 'link',
              func: () => {
                onUpgradeModal(true)
              }
            }
          ]
        },
        {
          label: 'Kubeconfig',
          field: 'kubeconfig',
          value: '',
          actions: isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_IKS_KUBE_CONFIG)
            ? [...optionsKubeConfigV2]
            : [...optionsKubeConfig]
        },
        { label: 'State:', field: 'clusterstatus', value: '', formula: 'status' },
        { label: 'Node group qty:', field: 'nodegroups', value: '', formula: 'length' },
        { label: 'Load balancer qty:', field: 'vips', value: '', formula: 'length' }
      ],
      customContent: null
    },
    {
      // Worker Node Groups Tab
      tapTitle: null,
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      // Load Balancers Tab
      tapTitle: null,
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: '',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: '',
      tapConfig: { type: 'custom' },
      customContent: <ClusterSecurityRulesContainer />
    },
    {
      tapTitle: '',
      tapConfig: { type: 'custom' },
      customContent: null
    }
  ]

  // *****
  // Upgrade Cluster tab structure
  // *****

  const modalContent = {
    label: '',
    buttonLabel: '',
    uuid: '',
    name: '',
    resourceId: '',
    question: '',
    feedback: '',
    resourceType: ''
  }

  const upgradeFormInitial = {
    form: {
      versions: {
        sectionGroup: 'upgrade',
        type: 'dropdown', // options = 'text ,'textArea'
        label: 'Select the k8s Version:',
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

  const displayInfoInitial: any = [
    { label: 'Size:', field: 'size', value: '' },
    { label: 'State:', field: 'state', value: '' },
    { label: 'Provider:', field: 'storageprovider', value: '' },
    { label: 'Message:', field: 'message', value: '' }
  ]

  const allowedStatusForMetrics = ['Active']

  // alerts
  const showError = useToastStore((state) => state.showError)

  const { copyToClipboard } = useCopy()
  // store information
  const clusters: any = useClusterStore((state) => state.clustersData)
  const loading = useClusterStore((state) => state.loading)
  const clusterNodegroups = useClusterStore((state) => state.clusterNodegroups)
  const clusterResourceLimit = useClusterStore((state) => state.clusterResourceLimit)
  const setShouldRefreshClusters = useClusterStore((state) => state.setShouldRefreshClusters)
  const setClusters = useClusterStore((state) => state.setClustersData)
  const setClusterNodegroups = useClusterStore((state) => state.setClusterNodegroups)
  const setClusterSecurityRules = useClusterStore((state) => state.setClusterSecurityRules)

  // cluster information
  const [myloadbalancers, setMyloadbalancers] = useState<any[]>([])
  const [myStorages, setMyStorages] = useState<any[]>([])
  const [reserveDetails, setReserveDetails] = useState<any>(null)
  // modals
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)

  const throwError = useErrorBoundary()
  const [actionsReserveDetails, setActionsReserveDetails] = useState<any[]>([])
  const [isPageReady, setIsPageReady] = useState(false)
  const [emptyStorage, setEmptyStorage] = useState(emptyStorageInitial)
  // Navigation
  const navigate = useNavigate()
  const { param: name } = useParams()

  const [activeTab, setActiveTab] = useState(0)
  const [tabDetails, setTabDetails] = useState(tabDetailsInitial)
  const [tabTitles, setTabTitles] = useState([...tabsTitleInitial])
  const [upgradeModal, setUpgradeModal] = useState({
    show: false,
    onHide: () => {
      onUpgradeModal(false)
    },
    centered: true,
    closeButton: true
  })
  const [upgradeForm, setUpgradeForm] = useState<any>(upgradeFormInitial)

  // *****
  // use effect
  // *****
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      if (clusters.length === 0) await fetchClusters(false)
    }
    fetch().catch((error) => {
      throwError(error)
    })
    setShouldRefreshClusters(true)
    return () => {
      setShouldRefreshClusters(false)
    }
  }, [])

  useEffect(() => {
    updateDetails()
  }, [myloadbalancers.length, clusters, isPageReady])

  useEffect(() => {
    if (!reserveDetails) {
      return
    }
    const { clusterstate } = reserveDetails
    if (clusterstate === 'Active') {
      const action = { ...emptyStorage.action }
      action.btnClass = 'btn btn-secondary'
      action.disabled = false
      setEmptyStorage({ ...emptyStorage, action })
    } else {
      setEmptyStorage(emptyStorageInitial)
    }
    const fetchNodeGroups = async (uuid: string): Promise<void> => {
      try {
        await setClusterNodegroups(uuid)
      } catch (error) {
        throwError(error)
      }
    }
    fetchNodeGroups(reserveDetails.uuid).catch((error) => {
      throwError(error)
    })

    const fetchClusterSecurity = async (uuid: string): Promise<void> => {
      try {
        if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_IKS_SECURITY)) {
          await setClusterSecurityRules(uuid)
        }
      } catch (error) {
        throwError(error)
      }
    }
    fetchClusterSecurity(reserveDetails.uuid).catch((error) => {
      throwError(error)
    })

    setGridLBInfo()
    setMyStorageInfo(reserveDetails)
  }, [reserveDetails])

  const fetchClusters = async (isBackground: boolean): Promise<void> => {
    try {
      await setClusters(isBackground)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  const debounceClusterRefresh = (): void => {
    setTimeout(() => {
      fetchClusters(true).catch((error) => {
        throwError(error)
      })
    }, 1000)
    setTimeout(() => {
      fetchClusters(true).catch((error) => {
        throwError(error)
      })
    }, 3000)
    setTimeout(() => {
      fetchClusters(true).catch((error) => {
        throwError(error)
      })
    }, 5000)
    setTimeout(() => {
      fetchClusters(true).catch((error) => {
        throwError(error)
      })
    }, 10000)
  }

  // *****
  // functions
  // *****
  function getStatusInfoLB(loadbalancer: any): JSX.Element {
    return getActionItemLabel(loadbalancer.vipstate, getILBStatus(loadbalancer))
  }

  const closeDeleteModal = (): void => {
    setShowActionModal(false)
  }

  function getClusterStatus(cluster: any): string {
    let message = 'No status'
    if (cluster) {
      const { clusterstatus } = cluster
      if (clusterstatus.errorcode) {
        message = clusterstatus.message
          ? getErrorMessageFromCodeAndMessage(clusterstatus.errorcode, clusterstatus.message)
          : 'No Status'
      } else {
        message = clusterstatus.message ? clusterstatus.message : 'No Status'
      }
    }
    return message
  }

  function getILBStatus(loadbalancer: any): string {
    let message = 'No status'
    if (loadbalancer) {
      const { vipstatus } = loadbalancer
      if (vipstatus.errorcode) {
        message = vipstatus.message
          ? getErrorMessageFromCodeAndMessage(vipstatus.errorcode, vipstatus.message)
          : 'No Status'
      } else {
        message = vipstatus.message ? vipstatus.message : 'No Status'
      }
    }
    return message
  }
  function setMyStorageInfo(reserveDetails: any): void {
    const ClusterInfoUpdated = []
    const storagesData = reserveDetails.storages || []
    for (const storageIndex in storagesData) {
      const storage = { ...storagesData[storageIndex] }
      for (const index in displayInfoInitial) {
        const item = { ...displayInfoInitial[index] }
        if (item.formula === 'length') {
          item.value = storage[item.field].length
        } else {
          item.value = storage[item.field]
        }
        ClusterInfoUpdated.push(item)
      }
    }

    setMyStorages(ClusterInfoUpdated)
  }

  function setGridLBInfo(): void {
    const gridInfo: any[] = []
    if (!reserveDetails.vips || reserveDetails.vips.length === 0) {
      setMyloadbalancers(gridInfo)
      return
    }
    for (const item in reserveDetails.vips) {
      const clusterItem = { ...reserveDetails.vips[item] }
      clusterItem.value = clusterItem.clusterstate
      gridInfo.push({
        'lb-name': clusterItem.name,
        state: {
          showField: true,
          type: 'function',
          value: clusterItem,
          sortValue: clusterItem.vipstate,
          function: getStatusInfoLB
        },
        ip: clusterItem.vipIp,
        type: clusterItem.viptype,
        port: clusterItem.port,
        actions: {
          showField: true,
          type: 'Buttons',
          value: clusterItem,
          selectableValues: getActionsByStatus(clusterItem.vipstate, ilbactionsOptions),
          function: setAction
        }
      })
    }

    setMyloadbalancers(gridInfo)
  }

  function getActionsByStatus(status: string, options: any): any[] {
    const result = []

    for (const index in options) {
      const option = { ...options[index] }
      if (option.status.find((item: string) => item === status)) {
        result.push(option)
      }
    }

    return result
  }

  async function actionOnModal(result: boolean): Promise<void> {
    if (!result) {
      closeDeleteModal()
      return
    }
    switch (actionModalContent.resourceType) {
      case 'cluster':
        await deleteCluster(actionModalContent.uuid)
        break
      case 'loadbalancer':
        await deleteLoadBalancer(actionModalContent.uuid)
        break
      default:
        break
    }
  }

  const updateDetails = (): void => {
    const cluster = clusters.find((cluster: any) => cluster.name === name)
    if (cluster === undefined) {
      if (isPageReady) navigate('/cluster')
      setActionsReserveDetails([])
      setReserveDetails(null)
      return
    }

    let newDetailsPanelOptions = [...detailPanelOptions]

    if (cluster.nodegroups.length === 0 || myloadbalancers.length >= (clusterResourceLimit?.maxvipspercluster ?? 2)) {
      newDetailsPanelOptions = detailPanelOptions.filter((x) => x.id !== 'addLoadBalancer')
    }

    if (
      !isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_IKS_STORAGE) ||
      !cluster.storages ||
      cluster.storages.length > 0
    ) {
      newDetailsPanelOptions = detailPanelOptions.filter((x) => x.id !== 'addStorage')
    }

    const actionOptionsUpdated = isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_IKS_KUBE_CONFIG)
      ? [...actionsOptionsV2]
      : [...actionsOptions]

    setActionsReserveDetails(
      getActionsByStatus(cluster.clusterstate, [...newDetailsPanelOptions, ...actionOptionsUpdated])
    )
    setReserveDetails(cluster)

    const tabDetailsCopy = []
    for (const tabIndex in tabDetails) {
      const tabDetail = { ...tabDetails[tabIndex] }
      const updateFields = []
      for (const index in tabDetail.fields) {
        const field = { ...tabDetail.fields[index] }
        if (field.formula === 'length') {
          field.value = cluster[field.field].length
        } else if (field.formula === 'status') {
          field.value = getClusterStatus(cluster)
        } else {
          field.value = cluster[field.field]
        }
        updateFields.push(field)
      }
      tabDetail.fields = updateFields
      tabDetailsCopy.push(tabDetail)
    }
    setTabDetails(tabDetailsCopy)

    const tabTitlesCopy = [...tabsTitleInitial]
    tabTitlesCopy[1].label = `Worker Node Groups (${cluster.nodegroups.length})`
    tabTitlesCopy[2].label = `Load Balancers (${cluster.vips.length})`

    if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_IKS_STORAGE)) {
      tabTitlesCopy[3].label = `Storage (${cluster.storages.length})`
      tabTitlesCopy[3].visible = true
    }

    if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_IKS_SECURITY)) {
      tabTitlesCopy[4].visible = true
    }

    const clusterAnnotations = cluster.annotations
    let isCloudMonitorEnable = false
    if (clusterAnnotations.length > 0) {
      const metricsEnable = clusterAnnotations.filter(
        (x: any) => x.key === 'cloudmonitorEnable' && String(true).toLowerCase() === x.value.toLowerCase()
      )
      isCloudMonitorEnable = metricsEnable.length > 0
    }

    if (
      isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_METRICS_CLUSTER) &&
      allowedStatusForMetrics.includes(cluster.clusterstate) &&
      isCloudMonitorEnable
    ) {
      tabTitlesCopy[5].visible = true
    }

    setTabTitles(tabTitlesCopy)

    const upgradeFormUpdated = { ...upgradeForm }
    const form = upgradeFormUpdated.form
    const formElement = form.versions
    const options = []
    for (const item in cluster.upgradek8sversionavailable) {
      options.push({
        name: cluster.upgradek8sversionavailable[item],
        value: cluster.upgradek8sversionavailable[item]
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

  function displayAlertSection(message: string, alertType: string): void {
    switch (alertType) {
      default:
        showError(message, false)
        break
    }
  }

  function onUpgradeModal(show: boolean): void {
    const upgradeModalCopy = { ...upgradeModal }
    upgradeModalCopy.show = show
    setUpgradeModal(upgradeModalCopy)
  }

  function onChangeUpgradeForm(event: any, formInputName: string): void {
    const updatedState = {
      ...upgradeForm
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setUpgradeForm(updatedState)
  }

  // *****
  // API functions
  //

  async function deleteCluster(clusterUUID: string): Promise<void> {
    try {
      await ClusterService.deleteCluster(clusterUUID)
      debounceClusterRefresh()
      closeDeleteModal()
      setTimeout((): void => {
        navigate('/cluster')
      }, 1000)
    } catch (error) {
      closeDeleteModal()
      throwError(error)
    }
  }

  const getKubeConfig = async (clusterUUID: string, readOnly: boolean): Promise<string> => {
    let kubeconfig = ''
    try {
      const kubeconfigResponse = await ClusterService.getKubeconfigFile(clusterUUID, readOnly)
      kubeconfig = kubeconfigResponse.data.kubeconfig
    } catch (error) {
      throwError(error)
    }
    return kubeconfig
  }

  async function downloadKubeConfig(name: string, clusterUUID: string, readOnly: boolean): Promise<void> {
    const kubeconfig = await getKubeConfig(clusterUUID, readOnly)
    const element = document.createElement('a')
    const file = new Blob([kubeconfig], { type: 'text/plain' })
    element.href = URL.createObjectURL(file)
    const fileName = `kubeconfig-${name}-${readOnly ? 'readonly' : 'admin'}.yaml`
    element.download = fileName
    document.body.appendChild(element)
    element.click()
  }

  async function getKubeConfigCopy(clusterUUID: string, readOnly: boolean): Promise<void> {
    const kubeconfig = await getKubeConfig(clusterUUID, readOnly)
    copyToClipboard(kubeconfig)
  }

  function setAction(action: any, item: any): void {
    switch (action.id) {
      case 'deletecluster': {
        const copyModalContent = { ...modalContent }
        copyModalContent.label = action.label
        copyModalContent.name = item.name
        copyModalContent.uuid = item.uuid
        copyModalContent.buttonLabel = action.buttonLabel
        copyModalContent.resourceType = 'cluster'
        setActionModalContent(copyModalContent)
        setShowActionModal(true)
        break
      }
      case 'deleteilb': {
        const copyModalContent = { ...modalContent }
        copyModalContent.label = action.label
        copyModalContent.name = item.name
        copyModalContent.uuid = item.vipid
        copyModalContent.buttonLabel = action.buttonLabel
        copyModalContent.resourceType = 'loadbalancer'
        setActionModalContent(copyModalContent)
        setShowActionModal(true)
        break
      }
      case 'addNodeGroup':
        goToAddNodeGroup()
        break
      case 'addLoadBalancer':
        goToAddLoadBalancer()
        break
      case 'addStorage':
        goToAddStorage()
        break
      case 'editStorage':
        goToEditStorage()
        break
      case 'DownloadReadOnly': {
        downloadKubeConfig(item.name, item.uuid, true).catch(() => {})
        break
      }
      case 'CopyReadOnly':
        getKubeConfigCopy(item.uuid, true).catch(() => {})
        break
      case 'DownloadReadWrite': {
        downloadKubeConfig(item.name, item.uuid, false).catch(() => {})
        break
      }
      case 'CopyReadWrite':
        getKubeConfigCopy(item.uuid, false).catch(() => {})
        break
      default:
        break
    }
  }

  async function deleteLoadBalancer(vipid: string): Promise<void> {
    try {
      await ClusterService.deleteLoadBalancer(reserveDetails?.uuid, vipid)
      debounceClusterRefresh()
      closeDeleteModal()
    } catch (error: any) {
      closeDeleteModal()
      if (error?.response?.data?.code) {
        displayAlertSection(error.response.data.message, 'error')
      } else {
        displayAlertSection(error.message, 'error')
      }
    }
  }

  const goToAddNodeGroup = (): void => {
    navigate({
      pathname: `/cluster/d/${name}/addnodegroup`,
      search: 'backTo=detail'
    })
  }

  async function submitUpgradeK8sVersion(): Promise<void> {
    try {
      const payload = { ...payloadUpgrade }
      payload.k8sversionname = getFormValue('versions', upgradeForm.form)
      const response = await ClusterService.upgradeCluster(payload, reserveDetails?.uuid)
      onUpgradeModal(false)
      if (!response.data) {
        displayAlertSection(response.statusText, 'error')
      } else {
        debounceClusterRefresh()
      }
    } catch (error: any) {
      onUpgradeModal(false)
      if (error.response) {
        displayAlertSection(error.response.data.message, 'error')
      } else {
        displayAlertSection(error.message, 'error')
      }
    }
  }
  const goToAddLoadBalancer = (): void => {
    navigate({
      pathname: `/cluster/d/${name}/reserveLoadbalancer`
    })
  }
  const goToAddStorage = (): void => {
    navigate({
      pathname: `/cluster/d/${name}/addstorage`
    })
  }

  const goToEditStorage = (): void => {
    navigate({
      pathname: `/cluster/d/${name}/editstorage`
    })
  }

  return (
    <ClusterMyReservationsDetail
      lbcolumns={lbcolumns}
      myloadbalancers={myloadbalancers}
      reserveDetails={reserveDetails}
      actionsReserveDetails={actionsReserveDetails}
      tabs={tabTitles}
      tabDetails={tabDetails}
      activeTab={activeTab}
      nodegroupsInfo={clusterNodegroups}
      showActionModal={showActionModal}
      actionModalContent={actionModalContent}
      loading={loading}
      upgradeModal={upgradeModal}
      upgradeForm={upgradeForm}
      myStorages={myStorages}
      emptyStorage={emptyStorage}
      clusterResourceLimit={clusterResourceLimit}
      setAction={setAction}
      setActiveTab={setActiveTab}
      actionOnModal={actionOnModal}
      getActionItemLabel={getActionItemLabel}
      goToAddNodeGroup={goToAddNodeGroup}
      goToAddLoadBalancer={goToAddLoadBalancer}
      onChangeUpgradeForm={onChangeUpgradeForm}
      submitUpgradeK8sVersion={submitUpgradeK8sVersion}
      debounceClusterRefresh={debounceClusterRefresh}
    />
  )
}

export default ClusterMyReservationsDetailsContainer
