// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useNavigate } from 'react-router'
import React, { useState, useEffect } from 'react'
import { BsTrash3, BsPencilFill } from 'react-icons/bs'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useRegionStore from '../../store/regionStore/RegionStore'
import RegionManagement from '../../components/regionManagement/RegionManagement'
import useToastStore from '../../store/toastStore/ToastStore'
import RegionManagementService from '../../services/RegionManagementService'

const RegionManagementContainer = (): JSX.Element => {
  // ****
  // global state
  // *****
  const regions = useRegionStore((state) => state.regions)
  const setRegions = useRegionStore((state) => state.setRegions)
  const loading = useRegionStore((state) => state.loading)
  const throwError = useErrorBoundary()
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // *****
  // local state
  // *****
  const columns = [
    {
      columnName: 'Global Name',
      targetColumn: 'name'
    },
    {
      columnName: 'Name',
      targetColumn: 'friendly_name'
    },
    {
      columnName: 'Type',
      targetColumn: 'type'
    },
    {
      columnName: 'Subnet',
      targetColumn: 'subnet'
    },
    {
      columnName: 'Availability Zone',
      targetColumn: 'availability_zone'
    },
    {
      columnName: 'Prefix',
      targetColumn: 'prefix'
    },
    {
      columnName: 'API DNS',
      targetColumn: 'api_dns'
    },
    {
      columnName: 'Updated By',
      targetColumn: 'adminName'
    },
    {
      columnName: 'Created At',
      targetColumn: 'created_at'
    },
    {
      columnName: 'Updated At',
      targetColumn: 'updated_at'
    },
    {
      columnName: 'Actions',
      targetColumn: 'actions',
      columnConfig: {
        behaviorType: 'buttons',
        behaviorFunction: null
      }
    }
  ]
  const actionsOptions = [
    {
      id: 'edit',
      label: 'edit region',
      name: (
        <>
          <BsPencilFill /> Update{' '}
        </>
      )
    },
    {
      id: 'delete',
      label: 'delete region',
      name: (
        <>
          <BsTrash3 /> Delete{' '}
        </>
      )
    }
  ]

  const emptyGrid = {
    title: 'No region found',
    subTitle: 'No region items'
  }

  const emptyGridByFilter = {
    title: 'No region found',
    subTitle: 'The applied filter criteria did not match any items',
    action: {
      type: 'function',
      href: () => {
        setFilter('', true)
      },
      label: 'Clear filters'
    }
  }

  const initialConfirmData = {
    isShow: false,
    title: '',
    data: [],
    id: null,
    onClose: closeConfirmModal
  }

  // Initial state for loader
  const initialLoaderData = {
    isShow: false,
    message: ''
  }

  const [regionItems, setRegionItems] = useState<any[]>([])
  const [filterText, setFilterText] = useState('')
  const [emptyGridObject, setEmptyGridObject] = useState(emptyGrid)
  const [confirmModalData, setConfirmModalData] = useState(initialConfirmData)
  const [showLoader, setShowLoader] = useState<any>(initialLoaderData)
  const [defaultRegion, setDefaultRegion] = useState<any>(null)

  const navigate = useNavigate()

  // *****
  // hooks
  // *****
  useEffect(() => {
    setGridRegions()
  }, [regions])

  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        await setRegions()
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
  function closeConfirmModal(): void {
    setConfirmModalData(initialConfirmData)
  }

  const setGridRegions = (): void => {
    let defaultRegion = null
    const gridInfo: any[] = []
    for (const index in regions) {
      const regionItem = { ...regions[Number(index)] }
      gridInfo.push({
        name: regionItem.name,
        friendly_name: regionItem.friendly_name,
        type: regionItem.type,
        subnet: regionItem.subnet,
        availability_zone: regionItem.availability_zone,
        prefix: regionItem.prefix,
        api_dns: regionItem.api_dns,
        adminName: regionItem.adminName,
        created_at: regionItem.created_at,
        updated_at: regionItem.updated_at,
        actions: {
          showField: true,
          type: 'Buttons',
          value: regionItem,
          selectableValues: actionsOptions,
          function: setAction
        }
      })
      if (regionItem.is_default) defaultRegion = regionItem
    }
    setRegionItems(gridInfo)
    setDefaultRegion(defaultRegion)
  }

  const setDeleteModal = (region: any): void => {
    const data: any = [
      {
        col: 'Global Name',
        value: region.name
      },
      {
        col: 'Name',
        value: region.friendly_name
      },
      {
        col: 'Type',
        value: region.type
      },
      {
        col: 'Subnet',
        value: region.subnet
      },
      {
        col: 'Availability Zone',
        value: region.availability_zone
      },
      {
        col: 'Prefix',
        value: region.prefix
      },
      {
        col: 'API DNS',
        value: region.api_dns
      },
      {
        col: 'Updated By',
        value: region.adminName
      },
      {
        col: 'Created At',
        value: region.created_at
      },
      {
        col: 'Updated At',
        value: region.updated_at
      }
    ]
    setConfirmModalData({
      title: `Delete region: ${region.name}?`,
      data,
      isShow: true,
      id: region.name,
      onClose: closeConfirmModal
    })
  }

  const setAction = (item: any, region: any): void => {
    const type = item.id
    switch (type) {
      case 'edit':
        navigate(`/regionmanagement/regions/d/${region.name}/edit`)
        break
      case 'delete':
        setDeleteModal(region)
        break
      default:
        navigate('/regionmanagement/regions')
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

  const onCancel = (): void => {
    navigate('/')
  }

  async function onSubmit(): Promise<void> {
    closeConfirmModal()
    setShowLoader({ isShow: true, message: 'Working on your request' })
    await submitForm()
  }

  async function submitForm(): Promise<void> {
    try {
      await RegionManagementService.deleteRegion(confirmModalData.id)
      showSuccess(`Region ${confirmModalData.id} has been deleted.`, false)
      await setRegions()
    } catch (error: any) {
      let message = ''
      if (error.response) {
        if (error.response.data.message !== '') {
          message = error.response.data.message
        } else {
          message = error.message
        }
      } else {
        message = error.message
      }
      showError(message, false)
    }
    setShowLoader({ isShow: false })
  }

  return (
    <RegionManagement
      onCancel={onCancel}
      regions={regionItems}
      regionColumns={columns}
      loading={loading}
      emptyGrid={emptyGridObject}
      filterText={filterText}
      confirmModalData={confirmModalData}
      showLoader={showLoader}
      defaultRegion={defaultRegion}
      setFilter={setFilter}
      onSubmit={onSubmit}
    />
  )
}

export default RegionManagementContainer
