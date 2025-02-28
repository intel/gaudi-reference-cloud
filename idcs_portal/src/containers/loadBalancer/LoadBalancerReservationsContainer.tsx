// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { BsPencilFill, BsDashCircle, BsCheckCircle, BsTrash3 } from 'react-icons/bs'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useToastStore from '../../store/toastStore/ToastStore'
import useLoadBalancerStore from '../../store/loadBalancerStore/LoadBalancerStore'
import LoadBalancerReservations from '../../components/loadBalancer/LoadBalancerReservations'
import LoadBalancerService from '../../services/LoadBalancerService'
import { useNavigate } from 'react-router'
import StateTooltipCell from '../../utils/gridPagination/cellRender/StateTooltipCell'

interface EmptyGridInterface {
  title: string
  subTitle: string
  action: EmptyGridActionInterface
}

interface EmptyGridActionInterface {
  type: string
  href: string | (() => void)
  label: string
}

const getActionItemLabel = (text: string, statusStep: string = ''): JSX.Element => {
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
    case 'Pending':
      message = <StateTooltipCell statusStep={statusStep} text={text} />
      break
    case 'Deleting':
      message = (
        <>
          {' '}
          <BsDashCircle /> {text}{' '}
        </>
      )
      break
    case 'Active':
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

const LoadBalancerReservationsContainer = (): JSX.Element => {
  // local state
  const columns = [
    {
      columnName: 'Name',
      targetColumn: 'name'
    },
    {
      columnName: 'Virtual IP',
      targetColumn: 'vip'
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

  const emptyGrid = {
    title: 'No load balancers found',
    subTitle: 'Your account currently has no load balancers.',
    action: {
      type: 'redirect',
      href: '/load-balancer/reserve',
      label: 'Launch Load Balancer'
    }
  }

  const emptyGridByFilter = {
    title: 'No load balancer found',
    subTitle: 'The applied filter criteria did not match any items.',
    action: {
      type: 'function',
      href: () => {
        setFilter('', true)
      },
      label: 'Clear filters'
    }
  }

  const throwError = useErrorBoundary()

  // Navigation
  const navigate = useNavigate()

  const [myreservations, setMyreservations] = useState<any[] | null>(null)
  const [showActionModal, setShowActionModal] = useState(false)
  const [actionModalContent, setActionModalContent] = useState(modalContent)
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState<EmptyGridInterface>(emptyGrid)
  const [isPageReady, setIsPageReady] = useState(false)

  // Global State
  const loading = useLoadBalancerStore((state) => state.loading)
  const loadBalancers = useLoadBalancerStore((state) => state.loadBalancers)
  const setLoadBalancers = useLoadBalancerStore((state) => state.setLoadBalancers)
  const setShouldRefreshLoadBalancers = useLoadBalancerStore((state) => state.setShouldRefreshLoadBalancers)

  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  const refreshBalancers = async (background: boolean): Promise<void> => {
    try {
      await setLoadBalancers(background)
      setIsPageReady(true)
    } catch (error) {
      throwError(error)
    }
  }

  // Hooks
  useEffect(() => {
    void refreshBalancers(false)

    setShouldRefreshLoadBalancers(true)
    return () => {
      setShouldRefreshLoadBalancers(false)
    }
  }, [])

  useEffect(() => {
    if (!isPageReady) {
      return
    }
    setGridInfo()
  }, [loadBalancers, isPageReady])

  // functions
  const setGridInfo = (): void => {
    const gridInfo: any[] = []

    for (const index in loadBalancers) {
      const balancer = { ...loadBalancers[index] }

      gridInfo.push({
        name:
          balancer.status !== 'Deleting'
            ? {
                showField: true,
                type: 'hyperlink',
                value: balancer.name,
                function: () => {
                  navigateToDetailsPage(balancer.name)
                }
              }
            : balancer.name,
        vip: balancer.vip,
        status: {
          showField: true,
          type: 'function',
          value: balancer,
          sortValue: balancer.status,
          function: getStatusInfo
        },
        creationTimestamp: {
          showField: true,
          type: 'date',
          toLocalTime: true,
          value: balancer.creationTimestamp,
          format: 'MM/DD/YYYY h:mm a'
        },
        actions: {
          showField: true,
          type: 'Buttons',
          value: balancer,
          selectableValues: getActionsByStatus(balancer.status, actionsOptions),
          function: setAction
        }
      })
    }
    setMyreservations(gridInfo)
  }

  const getStatusInfo = (balancer: any): any => {
    return getActionItemLabel(balancer.status, balancer.message)
  }

  const navigateToDetailsPage = (name: any): void => {
    if (name) navigate({ pathname: `/load-balancer/d/${name}` })
  }

  const setAction = (action: any, item: any): void => {
    switch (action.id) {
      case 'edit':
        navigate({
          pathname: `/load-balancer/d/${item.name}/edit`,
          search: '?backTo=grid'
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
        copyModalContent.actionType = 'terminateBalancer'
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
      showSuccess('Load balancer marked for deletion.', false)
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

  const setFilter = (event: any, clear: boolean): void => {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  return (
    <>
      <LoadBalancerReservations
        myreservations={myreservations ?? []}
        columns={columns}
        showActionModal={showActionModal}
        actionModalContent={actionModalContent}
        emptyGrid={emptyGridObject}
        loading={loading || myreservations === null}
        filterText={filterText}
        setFilter={setFilter}
        setShowActionModal={actionOnModal}
        setAction={setAction}
      />
    </>
  )
}

export default LoadBalancerReservationsContainer
