// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import { useNavigate, useParams } from 'react-router'
import SuperComputerDetail from '../../components/superComputer/superComputerDetail/SuperComputerDetail'
import { BsFillPlugFill, BsTrash3, BsDownload, BsCopy } from 'react-icons/bs'
import useSuperComputerStore from '../../store/superComputer/SuperComputerStore'
import SuperComputerInfoContainer from './SuperComputerInfoContainer'
import SuperComputerWorkerNodesContainer from './SuperComputerWorkerNodesContainer'
import SuperComputerLoadBalancerContainer from './SuperComputerLoadBalancerContainer'
import SuperComputerStorageContainer from './SuperComputerStorageContainer'
import SuperComputerSecurityRulesContainer from './SuperComputerSecurityRulesContainer'
import idcConfig, { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import SuperComputerService from '../../services/SuperComputerService'
import Spinner from '../../utils/spinner/Spinner'
import { useCopy } from '../../hooks/useCopy'
import { ReactComponent as ExternalLink } from '../../assets/images/ExternalLink.svg'

const getActionItemLabel = (text, statusStep = null, option = null) => {
  let message = null

  switch (text) {
    case 'Delete':
      message = (
        <>
          {' '}
          <BsTrash3 /> {text}{' '}
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

const SuperComputerDetailContainer = () => {
  // *****
  // Params
  // *****
  const { param } = useParams()

  // *****
  // Global state
  // *****
  const { copyToClipboard } = useCopy()
  const setProducts = useSuperComputerStore((state) => state.setProducts)
  const clusterDetail = useSuperComputerStore((state) => state.clusterDetail)
  const setClusterDetail = useSuperComputerStore((state) => state.setClusterDetail)
  const loading = useSuperComputerStore((state) => state.loadingDetail)
  const setShouldRefreshClusters = useSuperComputerStore((state) => state.setShouldRefreshClusters)
  const setShouldRefreshClusterDetail = useSuperComputerStore((state) => state.setShouldRefreshClusterDetail)
  const debounceDetailRefresh = useSuperComputerStore((state) => state.debounceDetailRefresh)
  const setDebounceDetailRefresh = useSuperComputerStore((state) => state.setDebounceDetailRefresh)
  const throwError = useErrorBoundary()
  // *****
  // local state
  // *****
  const navigationTopInitial = [
    {
      type: 'title',
      label: ''
    },
    {
      type: 'dropdown',
      label: 'Actions',
      actions: []
    },
    {
      type: 'documentation',
      label: (
        <>
          <BsFillPlugFill />
          Learn how it works
          <ExternalLink />
        </>
      ),
      redirecTo: idcConfig.REACT_APP_CLUSTER_GUIDE
    }
  ]

  const actionsOptions = [
    {
      id: 'delete',
      status: ['Updating', 'Active', 'Error'],
      name: getActionItemLabel('Delete'),
      label: 'Delete cluster instance',
      question: 'Do you want to delete cluster $<Name> ?',
      feedback: 'If the cluster is running it will be stopped. All your information will be lost.',
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
      id: 'downloadAdmin',
      name: getActionItemLabel('Download'),
      status: ['Active'],
      label: 'downloadAdmin',
      buttonLabel: 'downloadAdmin'
    }
  ]

  const actionsOptionsV2 = [
    {
      id: 'delete',
      status: ['Updating', 'Active', 'Error'],
      name: getActionItemLabel('Delete'),
      label: 'Delete cluster instance',
      question: 'Do you want to delete cluster $<Name> ?',
      feedback: 'If the cluster is running it will be stopped. All your information will be lost.',
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
      id: 'downloadReadOnly',
      name: getActionItemLabel('Download readonly'),
      status: ['Active'],
      label: 'downloadReadOnly',
      buttonLabel: 'downloadReadOnly'
    },
    {
      id: 'downloadAdmin',
      name: getActionItemLabel('Download admin'),
      status: ['Active'],
      label: 'downloadAdmin',
      buttonLabel: 'downloadAdmin'
    }
  ]

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

  const tabsInitial = [
    {
      label: 'Details',
      id: 'details',
      show: true,
      content: (
        <SuperComputerInfoContainer downloadKubeConfig={downloadKubeConfig} getKubeConfigCopy={getKubeConfigCopy} />
      )
    },
    {
      label: 'Worker Node Groups (qty)',
      id: 'workerNodeGroups',
      show: true,
      content: <SuperComputerWorkerNodesContainer />
    },
    {
      label: 'Load Balancers (qty)',
      id: 'loadBalancers',
      show: true,
      content: <SuperComputerLoadBalancerContainer />
    },
    {
      label: 'Storage',
      id: 'storage',
      show: true,
      content: <SuperComputerStorageContainer />
    },
    {
      label: 'Security',
      id: 'security',
      show: false,
      content: <SuperComputerSecurityRulesContainer />
    }
  ]

  const noFoundItem = {
    title: 'No cluster found',
    subTitle: 'The item you are trying to access does not exist. \n You can go to any of the following links:',
    action: {
      type: 'redirect',
      btnType: 'primary',
      href: '/supercomputer',
      label: 'Supercomputing'
    }
  }

  const [navigationTop, setNavigationTop] = useState(navigationTopInitial)
  const [tabs, setTabs] = useState(tabsInitial)
  const [activeTab, setActiveTab] = useState(0)
  const [actionModal, setActionModal] = useState(modalContentInitial)
  const [isPageReady, setIsPageReady] = useState(clusterDetail && clusterDetail.name === name)
  const navigate = useNavigate()
  // *****
  // Hooks
  // *****
  useEffect(() => {
    if (!isPageReady) {
      fetchClusterDetail(false)
    }
    setShouldRefreshClusters(true)
    setShouldRefreshClusterDetail(true)
    return () => {
      setShouldRefreshClusterDetail(false)
      setShouldRefreshClusters(false)
    }
  }, [])

  useEffect(() => {
    if (isPageReady && clusterDetail === null) {
      navigate('/supercomputer')
    }
    if (isPageReady && clusterDetail) {
      getClusterDetailInfo()
    }
  }, [clusterDetail, isPageReady])

  useEffect(() => {
    debounceRefresh()
  }, [debounceDetailRefresh])

  useEffect(() => {
    const fetchProducts = async () => {
      try {
        await setProducts()
      } catch (error) {
        throwError(error)
      }
    }
    fetchProducts()
  }, [])

  // *****
  // functions
  // *****
  const debounceRefresh = () => {
    if (debounceDetailRefresh) {
      setTimeout(() => {
        fetchClusterDetail(true)
      }, 1000)
      setTimeout(() => {
        fetchClusterDetail(true)
      }, 3000)
      setTimeout(() => {
        fetchClusterDetail(true)
      }, 5000)
      setTimeout(() => {
        fetchClusterDetail(true)
        setDebounceDetailRefresh(false)
      }, 10000)
    }
  }

  const fetchClusterDetail = async (isBackground) => {
    try {
      await setClusterDetail(param, isBackground)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  const getClusterDetailInfo = () => {
    if (!clusterDetail) {
      return
    }
    // Update options to display based on status
    const navigation = []
    for (const index in navigationTopInitial) {
      const item = navigationTopInitial[index]

      if (item.type === 'title') {
        item.label = clusterDetail.name
      }

      if (item.type === 'dropdown') {
        const options = isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_IKS_KUBE_CONFIG)
          ? [...actionsOptionsV2]
          : [...actionsOptions]
        const actions = options.filter((item) => item.status.includes(clusterDetail.clusterstate))
        item.actions = actions
      }

      navigation.push(item)
    }
    setNavigationTop(navigation)

    // Update tab info based on service response
    const tabsCopy = [...tabs]
    tabsCopy[1].label = `Worker Node Groups (${clusterDetail.nodegroups.length})`
    tabsCopy[2].label = `Load Balancers (${clusterDetail.vips.length})`
    if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_SC_SECURITY)) {
      tabsCopy[4].show = true
    }
    setTabs(tabsCopy)
  }

  const getKubeConfig = async (clusterUUID, readOnly) => {
    try {
      const kubeconfigResponse = await SuperComputerService.getKubeconfigFile(clusterUUID, readOnly)
      return kubeconfigResponse.data.kubeconfig
    } catch (error) {
      throwError(error)
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

  function setAction(action, item) {
    switch (action.id) {
      case 'downloadReadOnly': {
        downloadKubeConfig(item.name, item.uuid, true)
        break
      }
      case 'downloadAdmin': {
        downloadKubeConfig(item.name, item.uuid, false)
        break
      }
      case 'copy':
        getKubeConfigCopy(item.uuid)
        break
      default:
        {
          const question = action.question.replace('$<Name>', clusterDetail.name)
          setActionModal({
            ...actionModal,
            show: true,
            name: clusterDetail.name,
            uuid: clusterDetail.uuid,
            label: action.label,
            buttonLabel: action.buttonLabel,
            question,
            resourceType: 'cluster'
          })
        }
        break
    }
  }

  function onActionModal(result) {
    if (!result) {
      setActionModal({ ...actionModal, show: false })
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
      navigate({
        pathname: '/supercomputer'
      })
    } catch (error) {}
  }

  if (loading) {
    return <Spinner />
  }

  return (
    <SuperComputerDetail
      loading={loading || !isPageReady}
      setActiveTab={setActiveTab}
      activeTab={activeTab}
      tabs={tabs?.filter((x) => x.show) ?? []}
      navigationTop={navigationTop}
      noFoundItem={noFoundItem}
      clusterDetail={clusterDetail}
      setAction={setAction}
      actionModal={actionModal}
      onActionModal={onActionModal}
    />
  )
}

export default SuperComputerDetailContainer
