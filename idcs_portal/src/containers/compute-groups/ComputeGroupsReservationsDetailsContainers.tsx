// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
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
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import { useNavigate, useParams } from 'react-router'
import useToastStore from '../../store/toastStore/ToastStore'
import ComputeGroupsReservationsDetails from '../../components/compute-groups/computeGroupsDetails/ComputeGroupsReservationsDetails'
import StateTooltipCell from '../../utils/gridPagination/cellRender/StateTooltipCell'

interface Field {
  label?: string
  field: string
  value: any
}

interface TapConfig {
  type: string
}

export interface TapDetail {
  tapTitle?: string
  tapConfig: TapConfig
  fields: Field[]
  nodesInformation?: any[] | null
  nodeIpsInformation?: any[] | null
  machineImageInfo?: any | null
  customContent?: unknown
}

interface Metadata {
  name: string | null
}

interface Spec {
  runStrategy: string | null
  sshPublicKeyNames: string[]
}

interface ServiceEditPayload {
  metadata: Metadata
  spec: Spec
}

const getActionItemLabel = (text: string, statusStep: string | null = null): JSX.Element => {
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
    case 'Connect':
      message = (
        <>
          {' '}
          <BsTerminal />
          {text}{' '}
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

const ComputeGroupsReservationsDetailsContainers = (): JSX.Element => {
  // local state
  const detailsColumns: any[] = [
    {
      columnName: 'Instance',
      targetColumn: 'name'
    },
    {
      columnName: 'IP',
      targetColumn: 'IpDefault',
      className: 'text-end'
    },
    {
      columnName: 'State',
      targetColumn: 'status'
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

  const initialTaps: any[] = [
    {
      label: 'Details',
      id: 'details'
    },
    {
      label: 'Networking',
      id: 'networking'
    },
    {
      label: 'Security',
      id: 'security'
    }
  ]

  const tapDetailsInitial: TapDetail[] = [
    {
      tapConfig: { type: 'custom' },
      fields: [
        { label: 'Node Instance Type:', field: 'displayName', value: '' },
        { label: 'Node Memory:', field: 'memory', value: '' },
        { label: 'Node Description:', field: 'typeDescription', value: '' },
        { label: 'Memory Speed', field: 'speedMemory', value: '' },
        { label: 'Node Storage:', field: 'disk', value: '' },
        { label: 'Node CPU Cores:', field: 'cpu', value: '' },
        { label: 'DIMMs:', field: 'dimCount', value: '' },
        { label: 'Run strategy:', field: 'runStrategy', value: '' },
        { label: 'Instance Category:', field: 'instanceCategory', value: '' },
        { label: 'Dimm Size:', field: 'dimmSize', value: '' },
        { label: 'Machine Image:', field: 'machineImage', value: '' }
      ],
      nodesInformation: null,
      nodeIpsInformation: null,
      machineImageInfo: null,
      customContent: null
    },
    {
      tapTitle: 'Networking interfaces',
      tapConfig: { type: 'custom' },
      fields: [
        { field: 'vNet', value: '' },
        { field: 'vNetName', value: '' }
      ]
    },
    {
      tapTitle: 'Group Public Keys',
      tapConfig: { type: 'custom' },
      fields: [{ label: 'Name', field: 'sshPublicKey', value: '' }]
    },
    {
      tapTitle: '',
      tapConfig: { type: 'custom' },
      customContent: null,
      fields: []
    }
  ]

  const actionsOptions: any[] = [
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

  const detailsActionsOptions = [
    {
      id: 'connect',
      name: getActionItemLabel('Connect'),
      status: ['Ready'],
      label: 'Connect'
    },
    {
      id: 'sshconnect',
      name: getActionItemLabel('Connect via SSH'),
      status: ['Ready', 'Provisioning', 'Failed'],
      label: 'Connect via SSH',
      buttonLabel: 'Connect via SSH'
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

  const serviceEditPayload: ServiceEditPayload = {
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

  const allowedInstanceCategoriesMetrics = ['VirtualMachine']
  const allowedInstanceStatus = ['Ready']

  if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_METRICS_BARE_METAL)) {
    allowedInstanceCategoriesMetrics.push('BareMetalHost')
  }

  const throwError = useErrorBoundary()

  // Navigation
  const navigate = useNavigate()
  const { param: instanceGroupName } = useParams()

  const [taps, setTaps] = useState<any>(initialTaps)
  const [reserveDetails, setReserveDetails] = useState<any>(null)
  const [selectedNodeDetails, setSelectedNodeDetails] = useState(null)
  const [activeTap, setActiveTap] = useState<any>(0)
  const [showHowToConnectModal, setShowHowToConnectModal] = useState(false)
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [tapDetails, setTapDetails] = useState(tapDetailsInitial)
  const [actionsReserveDetails, setActionsReserveDetails] = useState<any[]>([])
  const [selectedNode, setSelectedNode] = useState<any>(null)
  const [isPageReady, setIsPageReady] = useState(false)
  const [areNodesReady, setAreNodesReady] = useState(false)

  // Global State
  const instanceGroups = useCloudAccountStore((state) => state.instanceGroups)
  const instanceGroupInstances = useCloudAccountStore((state) => state.instanceGroupInstances)
  const loading = useCloudAccountStore((state) => state.loading)
  const setInstanceGroups = useCloudAccountStore((state) => state.setInstanceGroups)
  const setInstanceGroupInstances = useCloudAccountStore((state) => state.setInstanceGroupInstances)
  const setShouldRefreshInstanceGroups = useCloudAccountStore((state) => state.setShouldRefreshInstanceGroups)
  const machineOs = useProductStore((state) => state.machineOs)
  const setMachineOs = useProductStore((state) => state.setMachineOs)
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)
  const generateCloudConnection = useCloudAccountStore((state) => state.generateCloudConnection)

  const refreshInstanceGroups = async (background: boolean): Promise<void> => {
    try {
      await setInstanceGroups(background)
    } catch (error) {
      throwError(error)
    }
  }

  const fetchOs = async (): Promise<void> => {
    try {
      await setMachineOs()
    } catch (error) {
      throwError(error)
    }
  }

  const validateInstanceGroup = (): void => {
    const instanceGroup = instanceGroups.find((instanceGroup) => instanceGroup.name === instanceGroupName)
    if (instanceGroup === undefined) {
      if (isPageReady) navigate('/compute-groups')
    }
  }

  // Hooks
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        if (instanceGroups?.length === 0) await refreshInstanceGroups(false)
        if (machineOs?.length === 0) await fetchOs()
        setIsPageReady(true)
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })

    setShouldRefreshInstanceGroups(true)
    return () => {
      setShouldRefreshInstanceGroups(false)
      void setInstanceGroupInstances('')
    }
  }, [])

  const fetchInstanceGroupInstances = async (): Promise<void> => {
    try {
      await setInstanceGroupInstances(instanceGroupName)
    } catch (error) {
      throwError(error)
    }
  }

  useEffect(() => {
    validateInstanceGroup()

    const fetch = async (): Promise<void> => {
      await fetchInstanceGroupInstances()
      setAreNodesReady(true)
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [instanceGroups, isPageReady])

  useEffect(() => {
    if (areNodesReady) {
      updateDetails()
    }
  }, [instanceGroupInstances, areNodesReady])

  // functions
  function getStatusInfo(instance: any): JSX.Element {
    return getActionItemLabel(instance.status, instance.message)
  }

  const getNodesIpsInformation = (): any[] => {
    return instanceGroupInstances.map((x) => ({
      instance: x.name,
      ip: x.interfaces.length > 0 ? x.interfaces[0].addresses[0] : null
    }))
  }

  const getNodesGridInfo = (): any[] => {
    const gridInfo = []

    for (const index in instanceGroupInstances) {
      const instance = { ...instanceGroupInstances[index] }
      const interfacesItem = [...instance.interfaces]

      let ipValue = null

      if (interfacesItem.length > 0) {
        ipValue = interfacesItem[0].addresses[0] ? interfacesItem[0].addresses[0] : null
      }

      gridInfo.push({
        name: instance.name,
        IpDefault: ipValue,
        status: {
          showField: true,
          type: 'function',
          value: instance,
          sortValue: instance.status,
          function: getStatusInfo
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
          selectableValues: getActionsByStatus(instance.status, detailsActionsOptions, instance.quickConnectEnabled),
          function: setAction
        }
      })
    }
    return gridInfo
  }

  const updateDetails = (): void => {
    if (instanceGroupName === null || instanceGroupInstances.length === 0) {
      setActionsReserveDetails([])
      setSelectedNodeDetails(null)
      setReserveDetails(null)
      return
    }

    const instanceGroup = instanceGroups.find((instanceGroup) => instanceGroup.name === instanceGroupName)
    if (instanceGroup === undefined) {
      setActionsReserveDetails([])
      setSelectedNodeDetails(null)
      setReserveDetails(null)
      return
    }

    const interfaces = instanceGroup.interfaces

    const instanceTypeDetails = instanceGroup.instanceTypeDetails

    const instanceDetail: any = {
      ...instanceGroup,
      typeDescription: instanceTypeDetails ? instanceTypeDetails.description : null,
      instanceCategory: instanceTypeDetails ? instanceTypeDetails.instanceCategory : null,
      memory: instanceTypeDetails ? instanceTypeDetails.memory : null,
      displayName: instanceTypeDetails ? instanceTypeDetails.displayName : null,
      dimCount: instanceTypeDetails ? instanceTypeDetails.dimCount : null,
      disk: instanceTypeDetails ? instanceTypeDetails.disk : null,
      cpu: instanceTypeDetails ? instanceTypeDetails.cpu : null,
      speedMemory: instanceTypeDetails ? instanceTypeDetails.speedMemory : null,
      dimmSize: instanceTypeDetails ? instanceTypeDetails.dimmSize : null,
      vNetName: interfaces.length > 0 ? instanceGroup.interfaces[0].vNet : null,
      vNet: interfaces.length > 0 ? instanceGroup.interfaces[0].name : null
    }

    const tapDetailsCopy: any[] = []
    tapDetails.forEach((originalTapDetail) => {
      const tapDetail = { ...originalTapDetail }
      const updateFields: any[] = []
      tapDetail.fields.forEach((originalField) => {
        const field = { ...originalField }
        field.value = instanceDetail[field.field]
        updateFields.push(field)
      })
      tapDetail.fields = updateFields
      tapDetail.nodesInformation = getNodesGridInfo()
      tapDetail.nodeIpsInformation = getNodesIpsInformation()
      tapDetail.machineImageInfo = machineOs.find((item) => item.name === instanceDetail.machineImage)
      tapDetailsCopy.push(tapDetail)
    })

    if (
      isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_METRICS_GROUPS) &&
      allowedInstanceCategoriesMetrics.includes(instanceDetail?.instanceCategory as string) &&
      instanceDetail.readyCount > 0 &&
      taps.filter((x: any) => x.label === 'Cloud Monitor').length === 0
    ) {
      const newTaps = [...taps]
      newTaps.push({
        label: 'Cloud Monitor',
        id: 'metrics'
      })
      setTaps(newTaps)
    }

    setTapDetails(tapDetailsCopy)
    setActionsReserveDetails(actionsOptions)
    setReserveDetails(instanceDetail)
  }

  function getModalContent(action: any, item: any): any {
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

  function setAction(action: any, item: any): void {
    switch (action.id) {
      case 'edit':
        navigate({
          pathname: `/compute-groups/d/${item.name}/edit`,
          search: '?backTo=detail'
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
      case 'connect':
        openCloudConnectlink(item.resourceId)
        break
      case 'start': {
        const modalContent = getModalContent(action, item)
        setSelectedNode(item)
        setActionModalContent(modalContent)
        setShowActionModal(true)
        break
      }
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

  function getActionsByStatus(status: string, gridActionOptions: any[], quickConnectEnabled?: boolean): any[] {
    const result: any[] = []

    gridActionOptions.forEach((originalOption) => {
      const option = { ...originalOption }
      if (option.status.find((item: string) => item === status)) {
        if (option.id === 'connect') {
          if (
            !isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_QUICK_CONNECT) ||
            !quickConnectEnabled ||
            status !== 'Ready'
          ) {
            return
          }
        }
        result.push(option)
      }
    })

    return result
  }

  const startOrStopInstance = async (runStrategy: string): Promise<void> => {
    const payloadCopy = { ...serviceEditPayload }
    payloadCopy.spec.runStrategy = runStrategy
    payloadCopy.spec.sshPublicKeyNames = selectedNode.sshPublicKey
    await CloudAccountService.putComputeReservation(selectedNode.resourceId, payloadCopy)
    setShowActionModal(false)
    showSuccess(`${selectedNode.name} ${runStrategy === PowerOnInstance ? 'started' : 'stopped'} successfully`, false)
    setTimeout(() => {
      void refreshInstanceGroups(true)
    }, 1000)
    setTimeout(() => {
      void refreshInstanceGroups(true)
    }, 5000)
  }

  const deleteInstanceGroup = async (instanceGroupName: string): Promise<void> => {
    await CloudAccountService.deleteInstanceGroupByName(instanceGroupName)
    setTimeout(() => {
      void refreshInstanceGroups(true)
    }, 1000)
    setTimeout(() => {
      void refreshInstanceGroups(true)
    }, 5000)
  }

  async function actionOnModal(result: boolean): Promise<void> {
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
        showError('Unable to perform action', false)
      }
    } else {
      setShowActionModal(result)
    }
  }

  const openCloudConnectlink = (resourceId: string): void => {
    void createUrlQuickConnect(resourceId)
  }

  async function createUrlQuickConnect(resourceId: string): Promise<void> {
    try {
      await generateCloudConnection(resourceId)
    } catch (error) {
      showError('Unable to perform action', false)
    }
  }

  return (
    <ComputeGroupsReservationsDetails
      detailsColumns={detailsColumns}
      reserveDetails={reserveDetails}
      selectedNodeDetails={selectedNodeDetails}
      setActiveTap={setActiveTap}
      activeTap={activeTap}
      tapDetails={tapDetails}
      actionsReserveDetails={actionsReserveDetails}
      setAction={setAction}
      showHowToConnectModal={showHowToConnectModal}
      setShowHowToConnectModal={setShowHowToConnectModal}
      showActionModal={showActionModal}
      setShowActionModal={actionOnModal}
      actionModalContent={actionModalContent}
      taps={taps}
      loading={loading || !isPageReady || !areNodesReady}
      instanceGroupInstances={instanceGroupInstances}
      allowedInstanceCategoriesMetrics={allowedInstanceCategoriesMetrics}
      allowedInstanceStatus={allowedInstanceStatus}
    />
  )
}

export default ComputeGroupsReservationsDetailsContainers
