// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import {
  BsDashCircle,
  BsCheckCircle,
  BsTrash3,
  BsStopCircle,
  BsPlayCircle,
  BsPencilFill,
  BsTerminal
} from 'react-icons/bs'
import CloudAccountService from '../../services/CloudAccountService'
import useProductStore from '../../store/productStore/ProductStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import ComputeGroupsReservations from '../../components/compute-groups/computeGroupsReservations/ComputeGroupsReservations'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import { useNavigate } from 'react-router'
import useToastStore from '../../store/toastStore/ToastStore'
import StateTooltipCell from '../../utils/gridPagination/cellRender/StateTooltipCell'
import InstanceTypeTooltipCell from '../../utils/gridPagination/cellRender/InstanceTypeTooltipCell'

const getActionItemLabel = (text, statusStep = null) => {
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
    case 'Stop':
      message = (
        <>
          {' '}
          <BsStopCircle /> {text}{' '}
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
    case 'Edit':
      message = (
        <>
          {' '}
          <BsPencilFill /> {text}{' '}
        </>
      )
      break
    case 'Connect':
      message = (
        <>
          {' '}
          <BsTerminal />
          {text}{' '}
        </>
      )
      break
    case 'Provisioning':
      message = <StateTooltipCell statusStep={statusStep} text={text} />
      break
    case 'Terminating':
      message = (
        <>
          {' '}
          <BsDashCircle /> {text}{' '}
        </>
      )
      break
    case 'Ready':
      message = (
        <>
          {' '}
          <BsCheckCircle /> {text}{' '}
        </>
      )
      break
    case 'Connect via SSH':
      message = <> {text} </>
      break
    default:
      message = <> {text} </>
      break
  }

  return message
}

const ComputeGroupsReservationsContainers = () => {
  // local state
  const columns = [
    {
      columnName: 'Instance Group Name',
      targetColumn: 'name'
    },
    {
      columnName: 'Nodes ready',
      targetColumn: 'instanceCount'
    },
    {
      columnName: 'Type',
      targetColumn: 'displayName'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions'
    }
  ]

  const actionsOptions = [
    {
      id: 'edit',
      name: getActionItemLabel('Edit'),
      status: ['Ready', 'Stopped', 'Provisioning'],
      label: 'Edit instance group'
    },
    {
      id: 'terminate',
      name: getActionItemLabel('Delete'),
      status: ['Ready', 'Provisioning', 'Failed'],
      label: 'Delete instance group',
      buttonLabel: 'Delete'
    }
  ]

  const modalContent = {
    label: '',
    buttonLabel: '',
    instanceName: '',
    resourceId: '',
    question: '',
    feedback: '',
    action: '',
    name: ''
  }

  const emptyGrid = {
    title: 'No instance groups found',
    subTitle: 'Your account currently has no instance groups',
    action: {
      type: 'redirect',
      href: '/compute-groups/reserve',
      label: 'Launch instance group'
    }
  }

  const emptyGridByFilter = {
    title: 'No instance groups found',
    subTitle: 'The applied filter criteria did not match any items.',
    action: {
      type: 'function',
      href: () => setFilter('', true),
      label: 'Clear filters'
    }
  }

  const serviceEditPayload = {
    metadata: {
      name: null
    },
    spec: {
      runStrategy: null,
      sshPublicKeyNames: []
    }
  }

  const PowerOffInstance = 'Halted'
  const PowerOnInstance = 'Always'

  const throwError = useErrorBoundary()

  const [myreservations, setMyreservations] = useState(null)
  const [selectedNodeDetails, setSelectedNodeDetails] = useState(null)
  const [showHowToConnectModal, setShowHowToConnectModal] = useState(false)
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [selectedNode, setSelectedNode] = useState(null)
  const [isPageReady, setIsPageReady] = useState(false)

  // Global State
  const instanceGroups = useCloudAccountStore((state) => state.instanceGroups)
  const loading = useCloudAccountStore((state) => state.loading)
  const setInstanceGroups = useCloudAccountStore((state) => state.setInstanceGroups)
  const setShouldRefreshInstanceGroups = useCloudAccountStore((state) => state.setShouldRefreshInstanceGroups)
  const machineOs = useProductStore((state) => state.machineOs)
  const setMachineOs = useProductStore((state) => state.setMachineOs)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)
  const generateCloudConnection = useCloudAccountStore((state) => state.generateCloudConnection)

  const refreshInstanceGroups = async (background) => {
    try {
      await setInstanceGroups(background)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  // Hooks
  useEffect(() => {
    refreshInstanceGroups(false)

    if (machineOs?.length === 0) {
      const fetchOs = async () => {
        try {
          await setMachineOs()
        } catch (error) {
          throwError(error)
        }
      }
      fetchOs()
    }
    setShouldRefreshInstanceGroups(true)
    return () => {
      setShouldRefreshInstanceGroups(false)
    }
  }, [])

  useEffect(() => {
    if (!isPageReady) {
      return
    }
    setGridInfo()
  }, [instanceGroups, isPageReady])

  // Navigation
  const navigate = useNavigate()

  // functions
  const getInstanceTypeInfo = (instance) => {
    return (
      <InstanceTypeTooltipCell
        name={instance.instanceTypeDetails.name}
        displayName={instance.instanceTypeDetails.displayName}
      />
    )
  }

  function setGridInfo() {
    const gridInfo = []
    for (const index in instanceGroups) {
      const instanceGroup = { ...instanceGroups[index] }

      gridInfo.push({
        name: {
          showField: true,
          type: 'hyperlink',
          value: instanceGroup.name,
          function: () => {
            navigateToDetailsPage(instanceGroup.name)
          }
        },
        instanceCount: `${instanceGroup.readyCount}/${instanceGroup.instanceCount}`,
        displayName: {
          showField: true,
          type: 'function',
          value: instanceGroup,
          sortValue: instanceGroup.instanceTypeDetails.name,
          function: getInstanceTypeInfo
        },
        actions: {
          showField: true,
          type: 'Buttons',
          value: instanceGroup,
          selectableValues: getActionsInGrid(instanceGroup),
          function: setAction
        }
      })
    }
    setMyreservations(gridInfo)
  }

  function getActionsInGrid(instanceGroup) {
    const instanceGroupStatus = instanceGroup.instanceCount === instanceGroup.readyCount ? 'Ready' : 'Not Ready'
    const options = []
    for (const index in actionsOptions) {
      const option = { ...actionsOptions[index] }
      if (option.id === 'connect') {
        if (
          !isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_QUICK_CONNECT) ||
          !instanceGroup.quickConnectEnabled ||
          instanceGroupStatus !== 'Ready'
        ) {
          continue
        }
      }
      options.push(option)
    }
    return options
  }

  function getModalContent(action, item) {
    const copyModalContent = { ...modalContent }
    copyModalContent.label = action.label
    copyModalContent.instanceName = item.name
    copyModalContent.feedback = action.feedback
    copyModalContent.buttonLabel = action.buttonLabel
    copyModalContent.resourceId = item.resourceId
    copyModalContent.action = action.id
    const question = action.question ? action.question.replace('$<Name>', item.name) : ''
    copyModalContent.question = question
    copyModalContent.name = item.name
    return copyModalContent
  }

  function navigateToDetailsPage(name) {
    if (name) navigate({ pathname: `/compute-groups/d/${name}` })
  }

  async function setAction(action, item) {
    switch (action.id) {
      case 'edit':
        navigate({
          pathname: `/compute-groups/d/${item.name}/edit`,
          search: '?backTo=grid'
        })
        break
      case 'terminate': {
        const modalContent = getModalContent(action, item)
        setActionModalContent(modalContent)
        setShowActionModal(true)
        break
      }
      case 'sshconnect': {
        const selectedNode = { ...item }
        selectedNode.ipNbr = item.interfaces.length > 0 ? item.interfaces[0].addresses[0] : null
        setSelectedNodeDetails(selectedNode)
        setShowHowToConnectModal(true)
        break
      }
      case 'start': {
        const modalContent = getModalContent(action, item)
        setSelectedNode(item)
        setActionModalContent(modalContent)
        setShowActionModal(true)
        break
      }
      case 'connect':
        CloudConnect(item.name)
        break
      case 'stop': {
        const modalContent = getModalContent(action, item)
        setSelectedNode(item)
        setActionModalContent(modalContent)
        setShowActionModal(true)
        break
      }
      default:
        break
    }
  }

  const startOrStopInstance = async (runStrategy) => {
    const payloadCopy = { ...serviceEditPayload }
    payloadCopy.spec.runStrategy = runStrategy
    payloadCopy.spec.sshPublicKeyNames = selectedNode.sshPublicKey
    await CloudAccountService.putComputeReservation(selectedNode.resourceId, payloadCopy)
    setShowActionModal(false)
    showSuccess(`${selectedNode.name} ${runStrategy === PowerOnInstance ? 'started' : 'stopped'} successfully`)
    setTimeout(() => {
      refreshInstanceGroups(true)
    }, 1000)
    setTimeout(() => {
      refreshInstanceGroups(true)
    }, 5000)
  }

  const deleteInstanceGroup = async (instanceGroupName) => {
    await CloudAccountService.deleteInstanceGroupByName(instanceGroupName)
    setTimeout(() => {
      refreshInstanceGroups(true)
    }, 1000)
    // In case terminating state does dot show inmmediately
    setTimeout(() => {
      refreshInstanceGroups(true)
    }, 5000)
  }

  async function actionOnModal(result) {
    if (result) {
      try {
        switch (actionModalContent.action) {
          case 'start':
            await startOrStopInstance(PowerOnInstance)
            break
          case 'stop':
            await startOrStopInstance(PowerOffInstance)
            break
          case 'terminate':
            deleteInstanceGroup(actionModalContent.instanceName)
              .then(() => {
                setShowActionModal(false)
              })
              .catch((error) => {
                setShowActionModal(false)
                throwError(error)
              })
            break
          default:
            setShowActionModal(false)
            break
        }
      } catch (error) {
        setShowActionModal(false)
        showError('Unable to perform action')
      }
    } else {
      setShowActionModal(result)
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

  async function CloudConnect(resourceId) {
    try {
      await generateCloudConnection(resourceId)
    } catch (error) {
      showError('Unable to perform action')
    }
  }

  return (
    <ComputeGroupsReservations
      columns={columns}
      myreservations={myreservations ?? []}
      selectedNodeDetails={selectedNodeDetails}
      showHowToConnectModal={showHowToConnectModal}
      setShowHowToConnectModal={setShowHowToConnectModal}
      showActionModal={showActionModal}
      setShowActionModal={actionOnModal}
      actionModalContent={actionModalContent}
      emptyGrid={emptyGridObject}
      loading={loading || myreservations === null}
      filterText={filterText}
      setFilter={setFilter}
    />
  )
}

export default ComputeGroupsReservationsContainers
