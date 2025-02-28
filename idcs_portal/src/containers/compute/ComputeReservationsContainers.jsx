// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import ComputeReservations from '../../components/compute/computeReservations/ComputeReservations'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import {
  BsPencilFill,
  BsDashCircle,
  BsCheckCircle,
  BsTrash3,
  BsStopCircle,
  BsPlayCircle,
  BsTerminal
} from 'react-icons/bs'
import { useNavigate } from 'react-router'
import CloudAccountService from '../../services/CloudAccountService'
import useProductStore from '../../store/productStore/ProductStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useToastStore from '../../store/toastStore/ToastStore'
import StateTooltipCell from '../../utils/gridPagination/cellRender/StateTooltipCell'
import InstanceTypeTooltipCell from '../../utils/gridPagination/cellRender/InstanceTypeTooltipCell'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'

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
    case 'Starting':
      message = (
        <>
          {' '}
          <BsPlayCircle /> {text}{' '}
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
    case 'Stopping':
      message = (
        <>
          {' '}
          <BsStopCircle /> {text}{' '}
        </>
      )
      break
    case 'Stopped':
      message = (
        <>
          {' '}
          <BsStopCircle /> {text}{' '}
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
    default:
      message = <> {text} </>
      break
  }

  return message
}

const ComputeReservationsContainers = () => {
  // local state
  const columns = [
    {
      columnName: 'Instance Name',
      targetColumn: 'name'
    },
    {
      columnName: 'Ip',
      targetColumn: 'IpDefault',
      className: 'text-end'
    },
    {
      columnName: 'State',
      targetColumn: 'status'
    },
    {
      columnName: 'Instance Type',
      targetColumn: 'displayName'
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
      id: 'edit',
      name: getActionItemLabel('Edit'),
      status: ['Ready', 'Stopped', 'Provisioning', 'Starting'],
      instanceTypes: ['BareMetalHost', 'VirtualMachine'],
      label: 'Edit instance'
    },

    {
      id: 'terminate',
      name: getActionItemLabel('Delete'),
      status: ['Ready', 'Provisioning', 'Failed', 'Starting', 'Stopped', 'Stopping'],
      instanceTypes: ['BareMetalHost', 'VirtualMachine'],
      label: 'Delete instance',
      buttonLabel: 'Delete'
    },
    {
      id: 'start',
      name: getActionItemLabel('Start'),
      instanceTypes: ['BareMetalHost'],
      status: ['Stopped'],
      label: 'Start instance',
      question: 'Do you want to start instance $<Name> ?',
      buttonVariant: 'primary',
      buttonLabel: 'Start'
    },
    {
      id: 'stop',
      name: getActionItemLabel('Stop'),
      instanceTypes: ['BareMetalHost'],
      status: ['Ready', 'Starting'],
      label: 'Stop instance',
      buttonLabel: 'Stop'
    },
    {
      id: 'connect',
      name: getActionItemLabel('Connect'),
      status: ['Ready'],
      instanceTypes: ['BareMetalHost', 'VirtualMachine'],
      label: 'Connect'
    }
  ]

  const modalContent = {
    action: '',
    label: '',
    buttonLabel: '',
    instanceName: '',
    resourceId: '',
    question: '',
    feedback: '',
    buttonVariant: '',
    name: ''
  }

  const emptyGrid = {
    title: 'No instances found',
    subTitle: 'Your account currently has no instances',
    action: {
      type: 'redirect',
      href: '/compute/reserve',
      label: 'Launch instance'
    }
  }

  const emptyGridByFilter = {
    title: 'No instances found',
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
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [itemSelected, setItemSelected] = useState(null)
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [isPageReady, setIsPageReady] = useState(false)
  // Global State
  const instances = useCloudAccountStore((state) => state.instances)
  const loading = useCloudAccountStore((state) => state.loading)
  const setInstances = useCloudAccountStore((state) => state.setInstances)
  const setShouldRefreshInstances = useCloudAccountStore((state) => state.setShouldRefreshInstances)
  const machineOs = useProductStore((state) => state.machineOs)
  const setMachineOs = useProductStore((state) => state.setMachineOs)
  const showSuccess = useToastStore((state) => state.showSuccess)
  const showError = useToastStore((state) => state.showError)
  const generateCloudConnection = useCloudAccountStore((state) => state.generateCloudConnection)

  const refreshInstances = async (background) => {
    try {
      await setInstances(background)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  // Hooks
  useEffect(() => {
    refreshInstances(false)

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
    setShouldRefreshInstances(true)
    return () => {
      setShouldRefreshInstances(false)
    }
  }, [])

  useEffect(() => {
    if (!isPageReady) {
      return
    }
    setGridInfo()
  }, [instances, isPageReady])

  // Navigation
  const navigate = useNavigate()

  // functions
  function setGridInfo() {
    const gridInfo = []

    for (const index in instances) {
      const instance = { ...instances[index] }
      const interfacesItem = [...instance.interfaces]
      let ipValue = null

      if (interfacesItem.length > 0) {
        ipValue = interfacesItem[0].addresses[0] ? interfacesItem[0].addresses[0] : null
      }
      gridInfo.push({
        name:
          instance.status !== 'Terminating'
            ? {
                showField: true,
                type: 'hyperlink',
                value: instance.name,
                function: () => {
                  navigateToDetailsPage(instance.name)
                }
              }
            : instance.name,
        IpDefault: ipValue,
        status: {
          showField: true,
          type: 'function',
          value: instance,
          sortValue: instance.status,
          function: getStatusInfo
        },
        displayName: {
          showField: true,
          type: 'function',
          value: instance,
          sortValue: instance.instanceTypeDetails.name,
          function: getInstanceTypeInfo
        },
        creationTimestamp: {
          showField: true,
          type: 'date',
          toLocalTime: true,
          value: instance.creationTimestamp,
          format: 'MM/DD/YYYY h:mm a'
        },
        actions: {
          showField: true,
          type: 'Buttons',
          value: instance,
          selectableValues: getActionsByStatus(
            instance.status,
            instance.instanceTypeDetails.instanceCategory,
            instance.quickConnectEnabled
          ),
          function: setAction
        }
      })
    }
    setMyreservations(gridInfo)
  }

  function getStatusInfo(instance) {
    return getActionItemLabel(instance.status, instance.message)
  }

  const getInstanceTypeInfo = (instance) => {
    return (
      <InstanceTypeTooltipCell
        name={instance.instanceTypeDetails.name}
        displayName={instance.instanceTypeDetails.displayName}
      />
    )
  }

  function navigateToDetailsPage(name) {
    if (name) navigate({ pathname: `/compute/d/${name}` })
  }

  function setAction(action, item) {
    switch (action.id) {
      case 'edit':
        navigate({
          pathname: `/compute/d/${item.name}/edit`,
          search: '?backTo=grid'
        })
        break
      case 'connect':
        openCloudConnectlink(item.resourceId)
        break
      default: {
        const copyModalContent = { ...modalContent }
        copyModalContent.action = action.id
        copyModalContent.label = action.label
        copyModalContent.instanceName = item.name
        copyModalContent.feedback = action.feedback
        copyModalContent.buttonLabel = action.buttonLabel
        copyModalContent.buttonVariant = action.buttonVariant
        copyModalContent.resourceId = item.resourceId
        const question = action.question ? action.question.replace('$<Name>', item.name) : ''
        copyModalContent.question = question
        copyModalContent.name = item.name
        setItemSelected(item)
        setActionModalContent(copyModalContent)
        setShowActionModal(true)
        break
      }
    }
  }

  function getActionsByStatus(status, instanceCategory, quickConnectEnabled) {
    const result = []
    const optionsByInstanceType = actionsOptions.filter((item) => item.instanceTypes.includes(instanceCategory))

    for (const index in optionsByInstanceType) {
      const option = { ...optionsByInstanceType[index] }

      if (option.status.find((item) => item === status)) {
        if (option.id === 'connect') {
          if (
            !isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_QUICK_CONNECT) ||
            !quickConnectEnabled ||
            status !== 'Ready'
          ) {
            continue
          }
        }
        result.push(option)
      }
    }

    return result
  }

  const startOrStopInstance = async (runStrategy) => {
    const payloadCopy = { ...serviceEditPayload }
    payloadCopy.spec.runStrategy = runStrategy
    payloadCopy.spec.sshPublicKeyNames = itemSelected.sshPublicKey
    await CloudAccountService.putComputeReservation(itemSelected.resourceId, payloadCopy)
    setShowActionModal(false)
    showSuccess(`${itemSelected.name} ${runStrategy === PowerOnInstance ? 'started' : 'stopped'} successfully`)
    setTimeout(() => {
      refreshInstances(true)
    }, 1000)
    // In case terminating state does dot show inmmediately
    setTimeout(() => {
      refreshInstances(true)
    }, 5000)
  }

  const deleteInstance = async (resourceId) => {
    await CloudAccountService.deleteComputeReservation(resourceId)
    setTimeout(() => {
      refreshInstances(true)
    }, 1000)
    // In case terminating state does dot show inmmediately
    setTimeout(() => {
      refreshInstances(true)
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
            await deleteInstance(actionModalContent.resourceId)
            setShowActionModal(false)
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
      setItemSelected(null)
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

  const openCloudConnectlink = async (resourceId) => {
    try {
      await generateCloudConnection(resourceId)
    } catch (error) {
      showError('Unable to perform action')
    }
  }

  return (
    <ComputeReservations
      myreservations={myreservations ?? []}
      columns={columns}
      showActionModal={showActionModal}
      actionModalContent={actionModalContent}
      emptyGrid={emptyGridObject}
      loading={loading || myreservations === null}
      filterText={filterText}
      setFilter={setFilter}
      setShowActionModal={actionOnModal}
      launchPagePath="/compute/reserve"
    />
  )
}

export default ComputeReservationsContainers
