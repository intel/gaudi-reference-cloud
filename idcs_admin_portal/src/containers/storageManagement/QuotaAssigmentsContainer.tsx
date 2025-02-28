// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React, { useEffect, useState } from 'react'
import useStorageManagementStore from '../../store/storageManagementStore/StorageManagementStore'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import StorageDetailView from '../../components/storageManagement/storageDetailView/StorageDetailView'
import { useNavigate } from 'react-router'
import {
  BsPencilFill
} from 'react-icons/bs'

const QuotaAssigmentsContainer = (): JSX.Element => {
  // *****
  // Global state
  // *****
  const storageQuotas = useStorageManagementStore((state) => state.storageQuotas)
  const storageDefaultQuotas = useStorageManagementStore((state) => state.storageDefaultQuotas)
  const loading = useStorageManagementStore((state) => state.loading)
  const getQuotaAssigments = useStorageManagementStore((state) => state.getQuotaAssigments)
  const setEditStorageQuota = useStorageManagementStore((state) => state.setEditStorageQuota)

  // *****
  // local state
  // *****
  const columns = [
    {
      columnName: 'Account ID',
      targetColumn: 'cloudAccountId'
    },
    {
      columnName: 'Account Type',
      targetColumn: 'accountType'
    },
    {
      columnName: 'Reason',
      targetColumn: 'reason'
    },
    {
      columnName: 'FileSize Quota',
      targetColumn: 'filesizeQuotaInTB'
    },
    {
      columnName: 'Max Volumes',
      targetColumn: 'filevolumesQuota'
    },
    {
      columnName: 'Bucket Size',
      targetColumn: 'bucketsQuota'
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

  const emptyGrid = {
    title: 'No Custom Quotas found',
    subTitle: 'No Custom Quotas items'
  }

  const emptyGridByFilter = {
    title: 'No Custom Quotas found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => { setFilter('', true) },
      label: 'Clear filters'
    }
  }

  const defaultQuotaColumns = [
    {
      columnName: 'Account Type',
      targetColumn: 'accountType'
    },
    {
      columnName: 'FileSize Quota',
      targetColumn: 'filesizeQuotaInTB'
    },
    {
      columnName: 'Max Volumes',
      targetColumn: 'filevolumesQuota'
    },
    {
      columnName: 'Bucket Size',
      targetColumn: 'bucketsQuota'
    }
  ]

  const defaultQuotaEmptyGrid = {
    title: 'No quota storage found',
    subTitle: 'No quota storage items'
  }

  const defaultQuotaEmptyGridByFilter = {
    title: 'No quota storage found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => { setDefaultQuotaFilter('', true) },
      label: 'Clear filters'
    }
  }

  const moduleName = 'Quotas'

  const throwError = useErrorBoundary()
  const navigate = useNavigate()
  const [bucketItems, setButcketItems] = useState<any[]>([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [defaultQuotaEmptyGridObject, setDefaultQuotaEmptyGridObject] = useState(defaultQuotaEmptyGrid)
  const [defaultQuotaFilterText, setDefaultQuotaFilterText] = useState('')
  const [defaultQuotaItems, setDefaultQuotaItems] = useState<any[]>([])

  const actionsOptions = [
    {
      id: 'edit',
      name: <>
        <BsPencilFill /> Edit{' '}
      </>
    }
  ]

  // *****
  // Hooks
  // *****
  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        await getQuotaAssigments(!!storageQuotas)
      } catch (error) {
        throwError(error)
      }
    }
    fetch().catch((error) => {
      throwError(error)
    })
  }, [])

  useEffect(() => {
    setQuotaAssgiments()
  }, [storageQuotas])

  useEffect(() => {
    setDefaultQuotaAssgiments()
  }, [storageDefaultQuotas])

  // *****
  // functions
  // *****
  const setQuotaAssgiments = (): void => {
    const gridInfo: any[] = []
    for (const index in storageQuotas) {
      const storageQuota = { ...storageQuotas[Number(index)] }
      gridInfo.push({
        cloudAccountId: storageQuota.cloudAccountId,
        accountType: storageQuota.accountType,
        reason: storageQuota.reason,
        filesizeQuotaInTB: storageQuota.filesizeQuotaInTB,
        filevolumesQuota: storageQuota.filevolumesQuota,
        bucketsQuota: storageQuota.bucketsQuota,
        actions: {
          showField: true,
          type: 'Buttons',
          value: storageQuota,
          selectableValues: actionsOptions,
          function: setAction
        }
      })
    }
    setButcketItems(gridInfo)
  }

  const setDefaultQuotaAssgiments = (): void => {
    const gridInfo: any[] = []
    for (const index in storageDefaultQuotas) {
      const storageQuota = { ...storageDefaultQuotas[Number(index)] }
      gridInfo.push({
        accountType: storageQuota.cloudAccountType,
        filesizeQuotaInTB: storageQuota.filesizeQuotaInTB,
        filevolumesQuota: storageQuota.filevolumesQuota,
        bucketsQuota: storageQuota.bucketsQuota
      })
    }
    setDefaultQuotaItems(gridInfo)
  }

  const setAction = (action: any, item: any): void => {
    switch (action.id) {
      default:
        setEditStorageQuota(item)
        navigate('/storagemanagement/managequota')
        break
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

  const setDefaultQuotaFilter = (event: any, clear: boolean): void => {
    if (clear) {
      setDefaultQuotaEmptyGridObject(defaultQuotaEmptyGrid)
      setDefaultQuotaFilterText('')
    } else {
      setDefaultQuotaEmptyGridObject(defaultQuotaEmptyGridByFilter)
      setDefaultQuotaFilterText(event.target.value)
    }
  }

  const onCancel = (): void => {
    navigate('/')
  }

  return (
    <StorageDetailView
      storageUsages={bucketItems}
      columns={columns}
      emptyGrid={emptyGridObject}
      loading={loading}
      filterText={filterText}
      moduleName={moduleName}
      defaultQuotaItems={defaultQuotaItems}
      defaultQuotaColumns={defaultQuotaColumns}
      defaultQuotaEmptyGrid={defaultQuotaEmptyGridObject}
      defaultQuotaFilterText={defaultQuotaFilterText}
      setDefaultQuotaFilter={setDefaultQuotaFilter}
      setFilter={setFilter}
      onCancel={onCancel}
    />
  )
}
export default QuotaAssigmentsContainer
