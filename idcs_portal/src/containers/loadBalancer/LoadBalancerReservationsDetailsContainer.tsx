// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { BsPencilFill, BsTrash3 } from 'react-icons/bs'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useToastStore from '../../store/toastStore/ToastStore'
import useLoadBalancerStore from '../../store/loadBalancerStore/LoadBalancerStore'
import LoadBalancerService from '../../services/LoadBalancerService'
import { useNavigate, useParams } from 'react-router'
import LoadBalancerReservationsDetails from '../../components/loadBalancer/LoadBalancerReservationsDetails'
import useCloudAccountStore from '../../store/cloudAccountStore/CloudAccountStore'

const getActionItemLabel = (text: string): JSX.Element => {
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
    case 'Edit':
      message = (
        <>
          {' '}
          <BsPencilFill /> {text}{' '}
        </>
      )
      break
    default:
      message = <> {text} </>
      break
  }

  return message
}

const LoadBalancerReservationsDetailsContainer = (): JSX.Element => {
  const tabsInitial = [
    {
      label: 'Details',
      id: 'details',
      show: true
    },
    {
      label: 'Source IPs',
      id: 'sourceIPs',
      show: true
    },
    {
      label: 'Listeners',
      id: 'listeners',
      show: true
    }
  ]

  // Column structure for listener tab
  const listenerColumns = {
    externalPort: 'Listener Port',
    internalPort: 'Instance Port',
    monitor: 'Monitor',
    loadBalancingMode: 'Load Balancing Mode',
    instanceSelector: 'Instance Selectors',
    instanceSelectors: 'Instance Selectors',
    instanceLabels: 'Instance Tags',
    instances: 'Instances',
    message: 'Status'
  }

  const tabDetailsInitial = [
    {
      tapTitle: 'Load Balancer information',
      tapConfig: { type: 'columns', columnCount: 4 },
      fields: [
        { label: 'ID:', field: 'resourceId', value: '' },
        { label: 'Virtual IP:', field: 'vip', value: '' },
        { label: 'Firewall Rule Created:', field: 'firewallRuleCreated', value: '' },
        { label: 'Status:', field: 'status', value: '' },
        { label: 'Status Message:', field: 'message', value: '' }
      ]
    },
    {
      tapTitle: 'Source IPs information',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: 'Listeners information',
      tapConfig: { type: 'custom' },
      customContent: null
    }
  ]

  const actionsOptions = [
    {
      id: 'edit',
      name: getActionItemLabel('Edit'),
      status: ['Reconciling', 'Pending', 'Active'],
      label: 'Edit balancer'
    },
    {
      id: 'terminate',
      name: getActionItemLabel('Delete'),
      status: ['Pending', 'Active'],
      label: 'Delete load balancer',
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
    actionType: '',
    name: ''
  }

  const throwError = useErrorBoundary()

  // Navigation
  const navigate = useNavigate()

  const [reserveDetails, setReserveDetails] = useState<any>(null)
  const [reserveListener, setReserveListener] = useState<any[]>([])
  const [actionsReserveDetails, setActionsReserveDetails] = useState<any[]>([])
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [tabDetails, setTabDetails] = useState(tabDetailsInitial)
  const [afterActionShowModal, setAfterActionShowModal] = useState(false)
  const [afterActionModalContent, setAfterActionModalContent] = useState(modalContent)
  const [isPageReady, setIsPageReady] = useState(false)

  // Global State
  const loading = useLoadBalancerStore((state) => state.loading)
  const loadBalancers = useLoadBalancerStore((state) => state.loadBalancers)
  const setLoadBalancers = useLoadBalancerStore((state) => state.setLoadBalancers)
  const setShouldRefreshLoadBalancers = useLoadBalancerStore((state) => state.setShouldRefreshLoadBalancers)
  const loadBalancerActiveTab = useLoadBalancerStore((state) => state.loadBalancerActiveTab)
  const setLoadBalancerActiveTab = useLoadBalancerStore((state) => state.setLoadBalancerActiveTab)
  const instances = useCloudAccountStore((state) => state.instances)
  const setInstances = useCloudAccountStore((state) => state.setInstances)

  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  const { param: name } = useParams()

  const refreshBalancers = async (background: boolean): Promise<void> => {
    try {
      await setLoadBalancers(background)
    } catch (error) {
      throwError(error)
    }
  }

  // Hooks
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        if (loadBalancers?.length === 0) await refreshBalancers(false)
        if (instances?.length === 0) await setInstances(false)
        setIsPageReady(true)
      } catch (error) {
        throwError(error)
      }
    }

    fetch().catch((error) => {
      throwError(error)
    })

    setShouldRefreshLoadBalancers(true)
    return () => {
      setShouldRefreshLoadBalancers(false)
    }
  }, [])

  useEffect(() => {
    updateDetails()
  }, [loadBalancers, instances, isPageReady])

  // functions

  const updateDetails = (): void => {
    const balancer = loadBalancers.find((balancer) => balancer.name === name)
    if (balancer === undefined) {
      if (isPageReady) navigate('/load-balancer')
      setActionsReserveDetails([])
      setReserveDetails(null)
      setReserveListener([])
      return
    }

    const balancerDetail: any = { ...balancer }

    const tabDetailsCopy = []
    for (const tabDetail of tabDetails) {
      const updateFields = []
      for (const index in tabDetail.fields) {
        const field: any = { ...tabDetail.fields[index as any] }
        field.value = String(balancerDetail[field.field])
        updateFields.push(field)
      }
      tabDetail.fields = updateFields
      tabDetailsCopy.push(tabDetail)
    }

    setTabDetails(tabDetailsCopy)
    setActionsReserveDetails(getActionsByStatus(balancerDetail.status, actionsOptions))
    setReserveDetails(balancerDetail)

    const listenerGrid = []
    const listeners = balancerDetail.listeners
    for (const listener of listeners) {
      listenerGrid.push({
        externalPort: listener.externalPort,
        internalPort: listener.internalPort,
        monitor: listener.monitor,
        loadBalancingMode: listener.loadBalancingMode,
        instanceSelector: listener.instanceSelector,
        instanceSelectors: listener.instanceSelectors,
        message: listener.message,
        poolMembers: listener.poolMembers
      })
    }

    setReserveListener(listenerGrid)
  }

  const setAction = (action: any, item: any): void => {
    switch (action.id) {
      case 'edit':
        navigate({
          pathname: `/load-balancer/d/${item.name}/edit`,
          search: '?backTo=detail'
        })
        break
      default: {
        const copyModalContent = { ...modalContent }
        copyModalContent.label = action.label
        copyModalContent.instanceName = item.name
        copyModalContent.feedback = action.feedback
        copyModalContent.buttonLabel = action.buttonLabel
        copyModalContent.resourceId = item.resourceId
        const question = action.question ? action.question.replace('$<Name>', item.name) : ''
        copyModalContent.question = question
        copyModalContent.actionType = copyModalContent.actionType = 'terminateBalancer'
        copyModalContent.name = item.name
        setActionModalContent(copyModalContent)
        setShowActionModal(true)
        break
      }
    }
  }

  const getActionsByStatus = (status: string, options: any): any[] => {
    const result = []

    for (const index in options) {
      const option = { ...options[index] }
      if (option.status.find((item: string) => item === status)) {
        result.push(option)
      }
    }

    return result
  }

  const deleteBalancer = async (resourceId: string): Promise<void> => {
    try {
      await LoadBalancerService.deleteLoadBalancer(resourceId)
      setTimeout(() => {
        void refreshBalancers(true)
      }, 1000)
      showSuccess('Load Balancer deleted successfully.', false)
      const copyModalContent = { ...modalContent }
      copyModalContent.label = 'Deleted load balancer'
      copyModalContent.feedback = 'Your load balancer was deleted.'
      copyModalContent.buttonLabel = 'OK'
      setShowActionModal(false)
      setAfterActionModalContent(copyModalContent)
      setAfterActionShowModal(true)
      // In case terminating state does not show inmmediately
      setTimeout(() => {
        void refreshBalancers(true)
      }, 5000)
    } catch (error) {
      showError('Error while deleting the load balancer. Please try again later.', false)
    }
  }

  const actionOnModal = (result: boolean): void => {
    setShowActionModal(result)
    if (result) {
      if (actionModalContent.actionType === 'terminateBalancer') {
        void deleteBalancer(actionModalContent.resourceId)
      }
      setShowActionModal(false)
    }
  }

  return (
    <LoadBalancerReservationsDetails
      reserveDetails={reserveDetails}
      loadBalancerActiveTab={loadBalancerActiveTab}
      setLoadBalancerActiveTab={setLoadBalancerActiveTab}
      tabDetails={tabDetails}
      actionsReserveDetails={actionsReserveDetails}
      showActionModal={showActionModal}
      actionModalContent={actionModalContent}
      tabs={tabsInitial}
      loading={loading}
      setShowActionModal={actionOnModal}
      setAction={setAction}
      listenerColumns={listenerColumns}
      reserveListener={reserveListener}
      afterActionModalContent={afterActionModalContent}
      afterActionShowModal={afterActionShowModal}
      setAfterActionShowModal={setAfterActionShowModal}
      computeInstances={instances}
    />
  )
}

export default LoadBalancerReservationsDetailsContainer
