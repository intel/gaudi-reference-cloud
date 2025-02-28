// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router'
import useStorageManagementStore from '../../store/storageManagementStore/StorageManagementStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useToastStore from '../../store/toastStore/ToastStore'
import QuotaManagementServiceQuotas from '../../components/storageManagement/quotaManagementService/QuotaManagementServiceQuotas'
import { BsPencilFill, BsTrash3 } from 'react-icons/bs'
import StorageManagementService from '../../services/StorageManagementService'

const QuotaManagementServiceQuotasContainer = (): JSX.Element => {
  // *****
  // Params
  // *****
  const { param } = useParams()

  // *****
  // local state
  // *****
  const serviceQuotaColumns = [
    {
      columnName: 'Name',
      targetColumn: 'resourceType'
    },
    {
      columnName: 'Reason',
      targetColumn: 'reason'
    },
    {
      columnName: 'Scope type',
      targetColumn: 'scopeType'
    },
    {
      columnName: 'User type/Id',
      targetColumn: 'scopeValue'
    },
    {
      columnName: 'Limit',
      targetColumn: 'maxLimit'
    },
    {
      columnName: 'Unit',
      targetColumn: 'quotaUnit'
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

  const actionsOptions = [
    {
      id: 'edit',
      name: <>
        <BsPencilFill /> Edit quota{' '}
      </>
    },
    {
      id: 'delete',
      name: <>
        <BsTrash3 /> Delete quota{' '}
      </>
    }
  ]

  const emptyGrid = {
    title: 'No service quota found',
    subTitle: 'No service quotas items'
  }

  const emptyGridByFilter = {
    title: 'No Quota found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => { setFilter('', true) },
      label: 'Clear filters'
    }
  }

  const deleteModalInitial = {
    show: false,
    title: 'Delete',
    itemToDelete: '',
    question: 'Are you sure you want to delete the resource name {resource}?',
    feedback: 'Delete the resource item cannot be undone',
    buttonLabel: 'Delete'
  }

  const navigate = useNavigate()
  const [isPageReady, setIsPageReady] = useState(false)
  const [serviceQuotaItems, setServiceQuotaItems] = useState<any[]>([])
  const [deleteModal, setDeleteModal] = useState(deleteModalInitial)
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)

  // *****
  // Global state
  // *****
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)
  const serviceQuota = useStorageManagementStore((state) => state.serviceQuota)
  const getServiceQuotas = useStorageManagementStore((state) => state.getServiceQuotas)
  const loading = useStorageManagementStore((state) => state.loading)
  const throwError = useErrorBoundary()

  // *****
  // Hooks
  // *****
  useEffect(() => {
    getQuotas()
  }, [])

  useEffect(() => {
    if (serviceQuota) {
      setGrid()
    }
  }, [serviceQuota])

  // *****
  // functions
  // *****

  const getQuotas = (): void => {
    const fetch = async (): Promise<void> => {
      try {
        if (param) {
          await getServiceQuotas(param)
        }
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }

  const setGrid = (): void => {
    const gridInfo: any[] = []
    for (const index in serviceQuota?.serviceQuotaResources) {
      const serviceQuotaItem = { ...serviceQuota?.serviceQuotaResources[Number(index)] }
      const item = {
        serviceId: serviceQuota?.serviceId,
        ...serviceQuotaItem
      }
      gridInfo.push({
        resourceType: serviceQuotaItem.resourceType,
        reason: serviceQuotaItem.reason,
        scopeType: serviceQuotaItem.scopeType,
        scopeValue: serviceQuotaItem.scopeValue,
        maxLimit: serviceQuotaItem.maxLimit,
        quotaUnit: serviceQuotaItem.quotaUnit,
        actions: {
          showField: true,
          type: 'Buttons',
          value: item,
          selectableValues: actionsOptions,
          function: setAction
        }
      })
    }
    setIsPageReady(true)
    setServiceQuotaItems(gridInfo)
  }

  const setAction = (action: any, item: any): void => {
    switch (action.id) {
      case 'delete': {
        setDeleteModal({ ...deleteModalInitial, show: true, question: deleteModalInitial.question.replace('{resource}', item.resourceType), itemToDelete: item })
        break
      }
      default:
        navigate({
          pathname: `/quotamanagement/services/d/${item.serviceId}/quotas/edit/${item.resourceType}/${item.ruleId}`
        })
        break
    }
  }

  const onCancel = (): void => {
    navigate('/quotamanagement/services')
  }

  const onCloseDeleteModal = async (action: boolean): Promise<void> => {
    if (action) {
      await deleteResourceType(deleteModal.itemToDelete)
    } else {
      setDeleteModal(deleteModalInitial)
    }
  }

  const deleteResourceType = async (item: any): Promise<void> => {
    try {
      await StorageManagementService.deleteServiceQuota(param, item.resourceType, item.ruleId)
      getQuotas()
      showSuccess('Resource type deleted successfully', false)
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

  const setFilter = (event: any, clear: boolean): void => {
    if (clear) {
      setEmptyGridObject(emptyGrid)
      setFilterText('')
    } else {
      setEmptyGridObject(emptyGridByFilter)
      setFilterText(event.target.value)
    }
  }

  return <>
    {
      isPageReady ? <QuotaManagementServiceQuotas serviceDetail={serviceQuota}
        serviceQuotaItems={serviceQuotaItems}
        serviceQuotaColumns={serviceQuotaColumns}
        emptyGrid={emptyGridObject} loading={loading}
        onCancel={onCancel}
        deleteModal={deleteModal}
        onCloseDeleteModal={onCloseDeleteModal}
        filterText={filterText} setFilter={setFilter} />
        : <div className="col-12 row mt-s2">
          <div className="spinner-border text-primary center"></div>
        </div>
    }

  </>
}

export default QuotaManagementServiceQuotasContainer
