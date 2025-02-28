// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import { BsPencilFill, BsTrash3, BsStopCircle, BsPlayCircle, BsTerminal } from 'react-icons/bs'
import ComputeReservationsDetails from '../../components/compute/computeDetails/ComputeReservationsDetails'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useProductStore from '../../store/productStore/ProductStore'
import useToastStore from '../../store/toastStore/ToastStore'
import CloudAccountService from '../../services/CloudAccountService'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'

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
    case 'Connect':
      message = (
        <>
          {' '}
          <BsTerminal />
          {text}{' '}
        </>
      )
      break

    default:
      message = <> {text} </>
      break
  }

  return message
}

const ComputeReservationsDetailsContainer = (): JSX.Element => {
  const initialTaps = [
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

  const tapDetailsInitial: any = [
    {
      tapTitle: 'Instance type information',
      tapConfig: { type: 'custom' },
      fields: [
        { label: 'Instance Type:', field: 'displayName', value: '' },
        { label: 'Memory:', field: 'memory', value: '' },
        { label: 'Status:', field: 'status', value: '' },
        { label: 'Instance Description:', field: 'typeDescription', value: '' },
        { label: 'Memory Speed', field: 'speedMemory', value: '' },
        { label: 'Storage:', field: 'disk', value: '' },
        { label: 'Status details:', field: 'message', value: '' },
        { label: 'CPU Cores:', field: 'cpu', value: '' },
        { label: 'DIMMs:', field: 'dimCount', value: '' },
        { label: 'Run strategy:', field: 'runStrategy', value: '' },
        { label: 'Resource ID', field: 'resourceId', value: '' },
        { label: 'Instance Category:', field: 'instanceCategory', value: '' },
        { label: 'Dimm Size:', field: 'dimmSize', value: '' },
        { label: 'Machine Image:', field: 'machineImage', value: '' }
      ],
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
      tapTitle: 'Instance Public Keys',
      tapConfig: { type: 'custom' },
      fields: [{ label: 'Name', field: 'sshPublicKey', value: '' }]
    },
    {
      tapTitle: '',
      tapConfig: { type: 'custom' },
      customContent: null
    }
  ]

  if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_COMPUTE_SHOW_LABELS)) {
    initialTaps.push({
      label: 'Tags',
      id: 'tags'
    })
    tapDetailsInitial.splice(3, 0, {
      tapTitle: 'Instance Tags',
      tapConfig: { type: 'custom' },
      fields: [{ field: 'labels', value: '' }]
    })
  }

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

  const actionsOptions = [
    {
      id: 'edit',
      name: getActionItemLabel('Edit'),
      status: ['Ready', 'Stopped', 'Provisioning', 'Starting'],
      instanceTypes: ['BareMetalHost', 'VirtualMachine'],
      label: 'Edit instance'
    },
    {
      id: 'connect',
      name: getActionItemLabel('Connect'),
      status: ['Ready'],
      instanceTypes: ['BareMetalHost', 'VirtualMachine'],
      label: 'Connect'
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
    }
  ]

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

  const allowedInstanceStatusMetrics = ['Ready']
  const allowedInstanceCategoriesMetrics = ['VirtualMachine']

  if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_METRICS_BARE_METAL)) {
    allowedInstanceCategoriesMetrics.push('BareMetalHost')
  }

  const throwError = useErrorBoundary()

  const [taps, setTaps] = useState<any>(initialTaps)
  const [reserveDetails, setReserveDetails] = useState<any>(null)
  const [actionsReserveDetails, setActionsReserveDetails] = useState<any[]>([])
  const [activeTap, setActiveTap] = useState<string | number>(0)
  const [tapDetails, setTapDetails] = useState(tapDetailsInitial)
  const [showHowToConnectModal, setShowHowToConnectModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [showActionModal, setShowActionModal] = useState(false)
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

  // Navigation
  const navigate = useNavigate()
  const { param: name } = useParams()
  const refreshInstances = async (background: boolean): Promise<void> => {
    try {
      await setInstances(background)
    } catch (error) {
      throwError(error)
    }
  }

  // Hooks
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        if (instances?.length === 0) await refreshInstances(false)
        if (machineOs?.length === 0) await setMachineOs()
        setIsPageReady(true)
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })

    setShouldRefreshInstances(true)
    return () => {
      setShouldRefreshInstances(false)
    }
  }, [])

  useEffect(() => {
    updateDetails()
  }, [instances, isPageReady, machineOs])

  const updateDetails = (): void => {
    const instance = instances.find((instance) => instance.name === name)
    if (instance === undefined) {
      if (isPageReady) navigate('/compute')
      setActionsReserveDetails([])
      setReserveDetails(null)
      return
    }

    const interfaces = instance.interfaces

    const instanceTypeDetails = instance.instanceTypeDetails

    const instanceDetail: any = {
      ...instance,
      typeDescription: instanceTypeDetails ? instanceTypeDetails.description : null,
      instanceCategory: instanceTypeDetails ? instanceTypeDetails.instanceCategory : null,
      memory: instanceTypeDetails ? instanceTypeDetails.memory : null,
      displayName: instanceTypeDetails ? instanceTypeDetails.displayName : null,
      dimCount: instanceTypeDetails ? instanceTypeDetails.dimCount : null,
      disk: instanceTypeDetails ? instanceTypeDetails.disk : null,
      cpu: instanceTypeDetails ? instanceTypeDetails.cpu : null,
      speedMemory: instanceTypeDetails ? instanceTypeDetails.speedMemory : null,
      dimmSize: instanceTypeDetails ? instanceTypeDetails.dimmSize : null,
      vNetName: interfaces.length > 0 ? instance.interfaces[0].vNet : null,
      vNet: interfaces.length > 0 ? instance.interfaces[0].name : null,
      ipNbr: interfaces.length > 0 ? instance.interfaces[0].addresses[0] : null
    }

    const tapDetailsCopy = []
    for (const tapIndex in tapDetails) {
      const tapDetail = { ...tapDetails[tapIndex] }
      const updateFields = []
      for (const index in tapDetail.fields) {
        const field = { ...tapDetail.fields[index] }
        field.value = instanceDetail[field.field]
        updateFields.push(field)
      }
      tapDetail.fields = updateFields
      tapDetail.machineImageInfo = machineOs.find((item) => item.name === instanceDetail.machineImage)
      tapDetailsCopy.push(tapDetail)
    }

    const isClusterBM =
      instanceDetail.nodegroupType === 'worker' && instanceDetail?.instanceCategory === 'BareMetalHost'

    if (
      allowedInstanceCategoriesMetrics.includes(instanceDetail?.instanceCategory as string) &&
      allowedInstanceStatusMetrics.includes(instanceDetail.status) &&
      !isClusterBM &&
      taps.filter((x: any) => x.id === 'metrics').length === 0
    ) {
      const newTaps = [...taps]
      newTaps.push({
        label: 'Cloud Monitor',
        id: 'metrics'
      })
      setTaps(newTaps)
    }

    setTapDetails(tapDetailsCopy)
    setActionsReserveDetails(getActionsByStatus(instanceDetail.status, instanceDetail.instanceCategory))
    setReserveDetails(instanceDetail)
  }

  function setAction(action: any, item: any): void {
    switch (action.id) {
      case 'edit':
        navigate({
          pathname: `/compute/d/${item.name}/edit`,
          search: '?backTo=detail'
        })
        break
      case 'connect':
        openCloudConnectlink()
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
        setActionModalContent(copyModalContent)
        setShowActionModal(true)
        break
      }
    }
  }

  const startOrStopInstance = async (runStrategy: string): Promise<void> => {
    const payloadCopy: any = { ...serviceEditPayload }
    payloadCopy.spec.runStrategy = runStrategy
    payloadCopy.spec.sshPublicKeyNames = reserveDetails?.sshPublicKey
    await CloudAccountService.putComputeReservation(reserveDetails?.resourceId, payloadCopy)
    setShowActionModal(false)
    showSuccess(
      `${reserveDetails?.name} ${runStrategy === PowerOnInstance ? 'started' : 'stopped'} successfully`,
      false
    )
    setTimeout((): void => {
      refreshInstances(true).catch((error) => {
        throwError(error)
      })
    }, 1000)
    // In case terminating state does dot show immediately
    setTimeout((): void => {
      refreshInstances(true).catch((error) => {
        throwError(error)
      })
    }, 5000)
  }

  const deleteInstance = async (resourceId: string): Promise<void> => {
    await CloudAccountService.deleteComputeReservation(resourceId)
    setTimeout((): void => {
      refreshInstances(true).catch((error) => {
        throwError(error)
      })
    }, 1000)

    setTimeout((): void => {
      navigate({ pathname: '/compute' })
    }, 2000)
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
            await deleteInstance(actionModalContent.resourceId)
            setShowActionModal(false)
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

  function getActionsByStatus(status: string, instanceCategory: string | null): any[] {
    const result = []
    if (instanceCategory) {
      const optionsByInstanceType = actionsOptions.filter((item) => item.instanceTypes.includes(instanceCategory))
      for (const index of optionsByInstanceType) {
        const option = { ...index }
        if (option.status.find((item) => item === status)) {
          if (option.id === 'connect') {
            if (
              !isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_QUICK_CONNECT) ||
              !reserveDetails?.quickConnectEnabled ||
              reserveDetails?.status !== 'Ready'
            ) {
              continue
            }
          }
          result.push(option)
        }
      }
    }
    return result
  }

  const openCloudConnectlink = (): void => {
    void createUrlQuickConnect()
  }

  async function createUrlQuickConnect(): Promise<void> {
    try {
      await generateCloudConnection(reserveDetails.resourceId)
    } catch (error) {
      showError('Unable to perform action', false)
    }
  }

  return (
    <ComputeReservationsDetails
      taps={taps}
      tapDetails={tapDetails}
      loading={loading}
      activeTap={activeTap}
      reserveDetails={reserveDetails}
      actionsReserveDetails={actionsReserveDetails}
      showHowToConnectModal={showHowToConnectModal}
      showActionModal={showActionModal}
      actionModalContent={actionModalContent}
      setShowHowToConnectModal={setShowHowToConnectModal}
      setActiveTap={setActiveTap}
      setAction={setAction}
      setShowActionModal={actionOnModal}
      openCloudConnectlink={openCloudConnectlink}
    />
  )
}

export default ComputeReservationsDetailsContainer
