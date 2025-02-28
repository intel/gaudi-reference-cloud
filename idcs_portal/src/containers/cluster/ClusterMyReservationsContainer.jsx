// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import {
  BsDashCircle,
  BsCheckCircle,
  BsPlayCircle,
  BsNodePlus,
  BsFillPlugFill,
  BsDownload,
  BsTrash3,
  BsDatabase,
  BsCopy
} from 'react-icons/bs'
import { useNavigate } from 'react-router'
import ClusterMyReservations from '../../components/cluster/clusterMyReservations/ClusterMyReservations'
import ClusterService from '../../services/ClusterService'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useClusterStore from '../../store/clusterStore/ClusterStore'
import { getErrorMessageFromCodeAndMessage } from '../../utils/apiError/apiError'
import { useCopy } from '../../hooks/useCopy'
import StateTooltipCell from '../../utils/gridPagination/cellRender/StateTooltipCell'
import Spinner from '../../utils/spinner/Spinner'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'

const getActionItemLabel = (text, statusStep = null, option = null) => {
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
          {!option && (
            <>
              <StateTooltipCell statusStep={statusStep} text={text} />
            </>
          )}
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
    case 'Copy read only':
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

const ClusterMyReservationsContainer = () => {
  const emptyGrid = {
    title: 'No clusters found',
    subTitle: 'Your account currently has no clusters',
    action: {
      type: 'redirect',
      href: '/cluster/reserve',
      label: 'Launch cluster'
    }
  }

  const emptyGridByFilter = {
    title: 'No clusters found',
    subTitle: 'The applied filter criteria did not match any items.',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  // *****
  // cluster table structure
  // *****

  const columns = [
    {
      columnName: 'Cluster Name',
      targetColumn: 'cluster-name'
    },
    {
      columnName: 'State',
      targetColumn: 'status'
    },
    {
      columnName: 'Worker Node Groups',
      targetColumn: 'nodeGroupsIndicator'
    },
    {
      columnName: 'Kubernetes Version',
      targetColumn: 'k8sversion'
    },
    {
      columnName: 'Kube Config',
      targetColumn: 'kubeconfig'
    },
    {
      columnName: 'Created at',
      targetColumn: 'creationTimestamp'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions'
    }
  ]

  const actionsOptions = [
    {
      id: 'deletecluster',
      name: getActionItemLabel('Delete'),
      status: ['Updating', 'Active', 'Error'],
      label: 'Delete cluster instance',
      buttonLabel: 'Delete'
    }
  ]

  const kubeConfigOptions = [
    { id: 'DownloadReadWrite', name: getActionItemLabel('Download'), status: ['Active'], label: 'Download' }
  ]

  const kubeConfigOptionsV2 = [
    {
      id: 'DownloadReadOnly',
      name: getActionItemLabel('Download readonly'),
      status: ['Active'],
      label: 'Download read only'
    },
    {
      id: 'DownloadReadWrite',
      name: getActionItemLabel('Download admin'),
      status: ['Active'],
      label: 'Download read write'
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

  const { copyToClipboard } = useCopy()
  // store information
  const clusters = useClusterStore((state) => state.clustersData)
  const loading = useClusterStore((state) => state.loading)
  const setShouldRefreshClusters = useClusterStore((state) => state.setShouldRefreshClusters)
  const setClusters = useClusterStore((state) => state.setClustersData)
  const currentSelectedCluster = useClusterStore((state) => state.currentSelectedCluster)
  const setCurrentSelectedCluster = useClusterStore((state) => state.setCurrentSelectedCluster)
  // cluster information
  const [myreservations, setMyreservations] = useState([])
  const [filterText, setFilterText] = useState('')
  // modals
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)

  const throwError = useErrorBoundary()
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [isPageReady, setIsPageReady] = useState(false)
  // Navigation
  const navigate = useNavigate()

  const [upgradeModal, setUpgradeModal] = useState({
    show: false,
    onHide: () => onUpgradeModal(false),
    centered: true,
    closeButton: true
  })
  const clusterLimit = 3

  // *****
  // use effect
  // *****

  useEffect(() => {
    fetchClusters(false)
    setShouldRefreshClusters(true)
    return () => {
      setShouldRefreshClusters(false)
    }
  }, [])

  useEffect(() => {
    setGridInfo()
  }, [clusters])

  const fetchClusters = async (isBackground) => {
    try {
      await setClusters(isBackground)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  const debounceClusterRefresh = () => {
    setTimeout(() => {
      fetchClusters(true)
    }, 1000)
    setTimeout(() => {
      fetchClusters(true)
    }, 3000)
    setTimeout(() => {
      fetchClusters(true)
    }, 5000)
    setTimeout(() => {
      fetchClusters(true)
    }, 10000)
  }

  // *****
  // functions
  // *****

  function getStatusInfo(cluster) {
    return getActionItemLabel(cluster.clusterstate, getClusterStatus(cluster))
  }

  function setFilter(event, clear) {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  const closeDeleteModal = () => {
    setShowActionModal(false)
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
        message = clusterstatus.message ? clusterstatus.message : 'No Status'
      }
    }
    return message
  }

  function setGridInfo() {
    const gridInfo = []
    for (const item in clusters) {
      const clusterItem = { ...clusters[item] }
      clusterItem.value = clusterItem.clusterstate
      gridInfo.push({
        'cluster-name':
          clusterItem.clusterstate !== 'DeletePending'
            ? {
                showField: true,
                type: 'hyperlink',
                value: clusterItem.name,
                function: () => {
                  setDetails(clusterItem.name)
                }
              }
            : clusterItem.name,
        status: {
          showField: true,
          type: 'function',
          value: clusterItem,
          sortValue: clusterItem.clusterstate,
          function: getStatusInfo
        },
        nodeGroupsIndicator:
          clusterItem?.nodegroups.length > 0 || clusterItem.clusterstate !== 'Active'
            ? clusterItem.nodegroups.length
            : {
                showField: true,
                type: 'hyperlink',
                value: 'Add group',
                noHyperLinkValue: '0 ',
                function: () => {
                  setDetails(clusterItem.name)
                  goToAddNodeGroup(clusterItem.name)
                }
              },
        k8sversion: clusterItem.k8sversion,
        kubeconfig: {
          showField: true,
          type: 'Buttons',
          value: clusterItem,
          selectableValues: getActionsByStatus(
            clusterItem.clusterstate,
            isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_IKS_KUBE_CONFIG)
              ? [...kubeConfigOptionsV2]
              : [...kubeConfigOptions]
          ),
          function: setAction
        },
        creationTimestamp: {
          showField: true,
          type: 'date',
          toLocalTime: true,
          value: clusterItem.createddate,
          format: 'MM/DD/YYYY h:mm a'
        },
        actions: {
          showField: true,
          type: 'Buttons',
          value: clusterItem,
          selectableValues: getActionsByStatus(clusterItem.clusterstate, actionsOptions),
          function: setAction
        }
      })
    }
    setMyreservations(gridInfo)
  }

  function getActionsByStatus(status, options) {
    const result = []

    for (const index in options) {
      const option = { ...options[index] }
      if (option.status.find((item) => item === status)) {
        result.push(option)
      }
    }

    return result
  }

  function actionOnModal(result) {
    if (!result) {
      closeDeleteModal()
      return
    }
    switch (actionModalContent.resourceType) {
      case 'cluster':
        deleteCluster(actionModalContent.uuid)
        break
      default:
        break
    }
  }

  async function setDetails(name = null) {
    setCurrentSelectedCluster(name)
    if (name) navigate(`/cluster/d/${name}`)
  }

  const getKubeConfig = async (clusterUUID, readOnly) => {
    try {
      const kubeconfigResponse = await ClusterService.getKubeconfigFile(clusterUUID, readOnly)
      return kubeconfigResponse.data.kubeconfig
    } catch (error) {
      throwError(error)
    }
  }

  async function getKubeConfigCopy(clusterUUID, readOnly) {
    const kubeconfig = await getKubeConfig(clusterUUID, readOnly)
    copyToClipboard(kubeconfig)
  }

  async function downloadKubeConfig(name, clusterUUID, readOnly) {
    const kubeconfig = await getKubeConfig(clusterUUID, readOnly)
    const clusterName = name
    const element = document.createElement('a')
    const file = new Blob([kubeconfig], { type: 'text/plain' })
    element.href = URL.createObjectURL(file)
    const fileName = `kubeconfig-${clusterName}-${readOnly ? 'readonly' : 'admin'}.yaml`
    element.download = fileName
    document.body.appendChild(element)
    element.click()
  }

  function updateValueByKey(keyToSearch, newValue) {
    for (const obj of myreservations) {
      if (obj['cluster-id'] === keyToSearch) {
        obj[status].value.clusterstate = newValue
        setMyreservations(obj) // or handle this case as needed
      }
    }
  }

  function onUpgradeModal(show) {
    const upgradeModalCopy = { ...upgradeModal }
    upgradeModalCopy.show = show
    setUpgradeModal(upgradeModalCopy)
  }

  // *****
  // API functions
  //

  async function deleteCluster(clusterUUID) {
    try {
      updateValueByKey(clusterUUID, 'DeletePending')
      await ClusterService.deleteCluster(clusterUUID)
      debounceClusterRefresh()
      closeDeleteModal()
    } catch (error) {
      closeDeleteModal()
      throwError(error)
    }
  }

  function setAction(action, item) {
    switch (action.id) {
      case 'DownloadReadOnly': {
        downloadKubeConfig(item.name, item.uuid, true)
        break
      }
      case 'CopyReadOnly':
        getKubeConfigCopy(item.uuid, true)
        break
      case 'DownloadReadWrite': {
        downloadKubeConfig(item.name, item.uuid, false)
        break
      }
      case 'CopyReadWrite':
        getKubeConfigCopy(item.uuid, false)
        break
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
      default:
        break
    }
  }

  const goToAddNodeGroup = (name) => {
    navigate({
      pathname: `/cluster/d/${name}/addnodegroup`,
      search: '?backTo=grid'
    })
  }

  const goToAddLoadBalancer = () => {
    navigate({
      pathname: `/cluster/d/${currentSelectedCluster}/reserveLoadbalancer`
    })
  }
  const goToAddStorage = () => {
    navigate({
      pathname: `/cluster/d/${currentSelectedCluster}/addstorage`
    })
  }

  if (!isPageReady) {
    return <Spinner />
  }

  return (
    <ClusterMyReservations
      columns={columns}
      myreservations={myreservations}
      showActionModal={showActionModal}
      actionOnModal={actionOnModal}
      actionModalContent={actionModalContent}
      emptyGrid={emptyGridObject}
      filterText={filterText}
      loading={loading}
      clusterLimit={clusterLimit}
      setFilter={setFilter}
    />
  )
}

export default ClusterMyReservationsContainer
