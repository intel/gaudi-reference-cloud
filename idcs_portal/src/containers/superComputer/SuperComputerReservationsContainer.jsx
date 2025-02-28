// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useSuperComputerStore from '../../store/superComputer/SuperComputerStore'
import { useNavigate } from 'react-router'
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
import { useCopy } from '../../hooks/useCopy'
import SuperComputerReservations from '../../components/superComputer/superComputerReservations/SuperComputerReservations'
import SuperComputerService from '../../services/SuperComputerService'
import StateTooltipCell from '../../utils/gridPagination/cellRender/StateTooltipCell'
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
        <>
          {option && (
            <>
              <div>{'State: '}</div>
              <StateTooltipCell statusStep={statusStep} text={text} spinnerAtTheEnd />
            </>
          )}
          {!option && <StateTooltipCell statusStep={statusStep} text={text} />}
        </>
      )
      break
    case 'Active':
    case 'Ready':
      message = (
        <>
          {' '}
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
    case 'Copy':
      message = (
        <>
          {' '}
          <BsCopy /> {text}{' '}
        </>
      )
      break
    case 'Download readonly':
    case 'Download admin':
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
      message = <> {text} </>
      break
  }

  return message
}

const SuperComputerReservationsContainer = () => {
  // *****
  // Global state
  // *****
  const clusters = useSuperComputerStore((state) => state.clusters)
  const setClusters = useSuperComputerStore((state) => state.setClusters)
  const loading = useSuperComputerStore((state) => state.loading)
  const setShouldRefreshClusters = useSuperComputerStore((state) => state.setShouldRefreshClusters)
  // *****
  // local state
  // *****
  const columns = [
    {
      columnName: 'Cluster Name',
      targetColumn: 'cluster-name'
    },
    {
      columnName: 'State',
      targetColumn: 'clusterstate'
    },
    {
      columnName: 'Storage',
      targetColumn: 'storage'
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
      targetColumn: 'createddate'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions'
    }
  ]

  const kubeConfigOptions = [
    {
      id: 'DownloadAdmin',
      label: 'DownloadAdmin',
      name: getActionItemLabel('Download'),
      status: ['Ready', 'Active']
    }
  ]

  const kubeConfigOptionsV2 = [
    {
      id: 'DownloadReadOnly',
      label: 'DownloadReadOnly',
      name: getActionItemLabel('Download readonly'),
      status: ['Ready', 'Active']
    },
    {
      id: 'DownloadAdmin',
      label: 'DownloadAdmin',
      name: getActionItemLabel('Download admin'),
      status: ['Ready', 'Active']
    }
  ]

  const actionsOptions = [
    {
      id: 'deletecluster',
      name: getActionItemLabel('Delete'),
      status: ['Updating', 'Active', 'Error'],
      label: 'Delete cluster instance',
      question: 'Do you want to delete cluster $<Name> ?',
      feedback: 'If the cluster is running it will be stopped. All your information will be lost.',
      buttonLabel: 'Delete'
    }
  ]

  const emptyGrid = {
    title: 'No clusters found',
    subTitle: 'Your account currently has no clusters',
    action: {
      type: 'redirect',
      href: '/supercomputer/launch',
      label: 'Launch Supercomputing Cluster'
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

  const modalContentInitial = {
    show: false,
    label: '',
    buttonLabel: '',
    uuid: '',
    name: '',
    resourceId: '',
    question: '',
    feedback: '',
    resourceType: ''
  }

  const throwError = useErrorBoundary()
  const navigate = useNavigate()
  const { copyToClipboard } = useCopy()
  const [myreservations, setMyreservations] = useState(null)
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [filterText, setFilterText] = useState('')
  const [actionModal, setActionModal] = useState(modalContentInitial)
  const [isPageReady, setIsPageReady] = useState(false)
  // *****
  // Hooks
  // *****
  useEffect(() => {
    fetchClusters(clusters.length > 0)
    setShouldRefreshClusters(true)
    return () => {
      setShouldRefreshClusters(false)
    }
  }, [])

  useEffect(() => {
    if (!isPageReady) {
      return
    }
    setGridInfo()
  }, [clusters, isPageReady])

  // *****
  // functions
  // *****
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
  }

  const fetchClusters = async (isBackground) => {
    try {
      await setClusters(isBackground)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  function setGridInfo() {
    const gridInfo = []
    for (const item in clusters) {
      const clusterItem = { ...clusters[item] }
      const storages = clusterItem.storages ? [...clusterItem.storages] : []
      const storage = storages.length > 0 ? storages[0].size : ''
      gridInfo.push({
        'cluster-name': {
          showField: true,
          type: 'hyperlink',
          value: clusterItem.name,
          function: () => {
            setDetails(clusterItem.name)
          }
        },
        status: {
          showField: true,
          type: 'function',
          value: clusterItem,
          sortValue: clusterItem.clusterstate,
          function: getStatusInfo
        },
        storage,
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
        createddate: {
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

  const setDetails = (route = null) => {
    navigate(`/supercomputer/d/${route}`)
  }

  function getStatusInfo(cluster) {
    return getActionItemLabel(cluster.clusterstate, cluster.clusterstatus.message)
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

  function setAction(action, item) {
    switch (action.id) {
      case 'DownloadReadOnly': {
        downloadKubeConfig(item.name, item.uuid, true)
        break
      }
      case 'DownloadAdmin': {
        downloadKubeConfig(item.name, item.uuid, false)
        break
      }
      case 'copy':
        getKubeConfigCopy(item.uuid)
        break
      case 'deletecluster': {
        const question = action.question.replace('$<Name>', item.name)
        setActionModal({
          ...actionModal,
          show: true,
          name: item.name,
          uuid: item.uuid,
          label: action.label,
          buttonLabel: action.buttonLabel,
          question,
          resourceType: 'cluster'
        })
        break
      }
      default:
        break
    }
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

  async function getKubeConfigCopy(clusterUUID) {
    const kubeconfig = await getKubeConfig(clusterUUID)
    copyToClipboard(kubeconfig)
  }

  const getKubeConfig = async (clusterUUID, readOnly) => {
    try {
      const kubeconfigResponse = await SuperComputerService.getKubeconfigFile(clusterUUID, readOnly)
      return kubeconfigResponse.data.kubeconfig
    } catch (error) {
      throwError(error)
    }
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

  function actionOnModal(result) {
    if (!result) {
      setActionModal(modalContentInitial)
    } else {
      switch (actionModal.resourceType) {
        case 'cluster':
          deleteCluster(actionModal.uuid)
          break
        default:
          break
      }
    }
  }

  async function deleteCluster(clusterUUID) {
    try {
      await SuperComputerService.deleteCluster(clusterUUID)
      debounceClusterRefresh()
      setActionModal(modalContentInitial)
    } catch (error) {
      setActionModal(modalContentInitial)
      throwError(error)
    }
  }

  return (
    <SuperComputerReservations
      myreservations={myreservations ?? []}
      columns={columns}
      loading={loading || myreservations === null}
      setFilter={setFilter}
      emptyGrid={emptyGridObject}
      filterText={filterText}
      actionModal={actionModal}
      actionOnModal={actionOnModal}
    />
  )
}

export default SuperComputerReservationsContainer
