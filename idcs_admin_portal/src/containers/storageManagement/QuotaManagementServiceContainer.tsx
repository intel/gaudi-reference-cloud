// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React, { useState, useEffect } from 'react'
import QuotaManagementService from '../../components/storageManagement/quotaManagementService/QuotaManagementService'
import { useNavigate } from 'react-router'
import useStorageManagementStore from '../../store/storageManagementStore/StorageManagementStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { BsPencilFill, BsTrash3, BsCardList } from 'react-icons/bs'
import useToastStore from '../../store/toastStore/ToastStore'
import StorageManagementService from '../../services/StorageManagementService'

const QuotaManagementServiceContainer = (): JSX.Element => {
  // *****
  // Global state
  // *****
  const services = useStorageManagementStore((state) => state.services)
  const getServices = useStorageManagementStore((state) => state.getServices)
  const loading = useStorageManagementStore((state) => state.loading)
  const throwError = useErrorBoundary()
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // *****
  // local state
  // *****
  const seviceColumns = [
    {
      columnName: 'Service Name',
      targetColumn: 'serviceName'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'Buttons',
        behaviorFunction: null
      }
    }
  ]

  const deleteModalInitial = {
    show: false,
    title: 'Delete',
    itemToDelete: '',
    question: 'Are you sure you want to delete the service name {serviceName}?',
    feedback: 'Delete the service item cannot be undone',
    buttonLabel: 'Delete'
  }

  const actionsOptions = [
    {
      id: 'edit',
      name: <>
        <BsPencilFill /> Edit limits{' '}
      </>
    },
    {
      id: 'quota',
      name: <>
        <BsCardList /> View Quotas{' '}
      </>
    },
    {
      id: 'delete',
      name: <>
        <BsTrash3 /> Delete service{' '}
      </>
    }
  ]

  const emptyGrid = {
    title: 'No service found',
    subTitle: 'No service items'
  }

  const emptyGridByFilter = {
    title: 'No Service found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => { setFilter('', true) },
      label: 'Clear filters'
    }
  }

  const navigate = useNavigate()
  const [serviceItems, setServiceItems] = useState<any[]>([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [deleteModal, setDeleteModal] = useState(deleteModalInitial)

  // *****
  // Hooks
  // *****
  useEffect(() => {
    setGridServices()
  }, [services])

  useEffect(() => {
    const isBackgroundUptade = services === null
    const fetch = async (): Promise<void> => {
      try {
        await getServices(isBackgroundUptade)
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])
  // *****
  // functions
  // *****
  const setGridServices = (): void => {
    const gridInfo: any[] = []
    for (const index in services) {
      const serviceItem = { ...services[Number(index)] }
      gridInfo.push({
        serviceName: serviceItem.serviceName,
        actions: {
          showField: true,
          type: 'Buttons',
          value: serviceItem,
          selectableValues: actionsOptions,
          function: setAction
        }
      })
    }
    setServiceItems(gridInfo)
  }

  const setAction = (action: any, item: any): void => {
    switch (action.id) {
      case 'quota': {
        navigate({
          pathname: `/quotamanagement/services/d/${item.serviceId}/quotas`
        })
        break
      }
      case 'delete': {
        setDeleteModal({ ...deleteModalInitial, show: true, question: deleteModalInitial.question.replace('{serviceName}', item.serviceName), itemToDelete: item })
        break
      }
      default:
        navigate({
          pathname: `/quotamanagement/services/d/${item.serviceId}/edit`
        })
        break
    }
  }

  const onCancel = (): void => {
    navigate('/')
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

  const onCloseDeleteModal = async (action: boolean): Promise<void> => {
    if (action) {
      await deleteService(deleteModal.itemToDelete)
    } else {
      setDeleteModal(deleteModalInitial)
    }
  }

  const deleteService = async (item: any): Promise<void> => {
    try {
      await StorageManagementService.deleteService(item.serviceId)
      await getServices(false)
      showSuccess('Service deleted successfully', false)
      setDeleteModal(deleteModalInitial)
    } catch (error: any) {
      const message = String(error.message)
      if (error.response) {
        const errData = error.response.data
        const errMessage = errData.message
        showError(errMessage, false)
      } else {
        showError(message, false)
      }
      setDeleteModal(deleteModalInitial)
    }
  }

  return <QuotaManagementService onCancel={onCancel} seviceColumns={seviceColumns} serviceItems={serviceItems} emptyGrid={emptyGridObject} loading={loading} filterText={filterText} setFilter={setFilter} deleteModal={deleteModal} onCloseDeleteModal={onCloseDeleteModal}/>
}

export default QuotaManagementServiceContainer
