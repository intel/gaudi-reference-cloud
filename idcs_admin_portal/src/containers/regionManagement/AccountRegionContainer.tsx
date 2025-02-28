// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useNavigate } from 'react-router'
import { BsTrash3 } from 'react-icons/bs'
import React, { useState, useEffect } from 'react'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import useRegionStore from '../../store/regionStore/RegionStore'
import AccountRegion from '../../components/regionManagement/AccountRegion'
import useToastStore from '../../store/toastStore/ToastStore'
import RegionManagementService from '../../services/RegionManagementService'

const AccountRegionContainer = (): JSX.Element => {
  // ****
  // global state
  // *****
  const accountWhitelist = useRegionStore((state) => state.accountWhitelist)
  const setAccountWhitelist = useRegionStore((state) => state.setAccountWhitelist)
  const loading = useRegionStore((state) => state.loading)
  const throwError = useErrorBoundary()
  const showError = useToastStore((state) => state.showError)
  const showSuccess = useToastStore((state) => state.showSuccess)

  // *****
  // local state
  // *****
  const columns = [
    {
      columnName: 'Cloud Account',
      targetColumn: 'cloudaccountId'
    },
    {
      columnName: 'Region',
      targetColumn: 'regionName'
    },
    {
      columnName: 'Assignment Date',
      targetColumn: 'created'
    },
    {
      columnName: 'Assignment Admin',
      targetColumn: 'adminName'
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
      id: 'delete',
      label: 'delete account',
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
    title: 'No data found',
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
    id: '',
    regionName: '',
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
  const [regions, setRegions] = useState<any>([{ name: 'All', value: '' }])
  const [selectedRegion, setSelectedRegion] = useState<any>('')

  const navigate = useNavigate()

  // *****
  // hooks
  // *****
  useEffect(() => {
    setGridRegions()
    setRegionsOptions()
  }, [accountWhitelist])

  useEffect(() => {
    const fetch = async (): Promise<void> => {
      try {
        await setAccountWhitelist()
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
  function setRegionsOptions(): void {
    const regions = [{ name: 'All', value: '' }]
    const uniqueRegions: any = []
    if (accountWhitelist) {
      accountWhitelist.forEach((acl) => {
        if (!uniqueRegions.includes(acl.regionName)) {
          uniqueRegions.push(acl.regionName)
          regions.push({
            name: acl.regionName,
            value: acl.regionName
          })
        }
      })
    }
    setRegions(regions)
  }

  function setRegionFilter(event: any): void {
    setEmptyGridObject(emptyGridByFilter)
    setSelectedRegion(event.target.value)
  }

  function closeConfirmModal(): void {
    setConfirmModalData(initialConfirmData)
  }

  const setGridRegions = (): void => {
    const gridInfo: any[] = []
    for (const index in accountWhitelist) {
      const regionItem = { ...accountWhitelist[Number(index)] }
      gridInfo.push({
        cloudaccountId: regionItem.cloudaccountId,
        regionName: regionItem.regionName,
        created: regionItem.created,
        adminName: regionItem.adminName,
        actions: {
          showField: true,
          type: 'Buttons',
          value: regionItem,
          selectableValues: actionsOptions,
          function: setAction
        }
      })
    }
    setRegionItems(gridInfo)
  }

  const setAction = (item: any, acl: any): void => {
    const type = item.id
    if (type === 'delete') {
      const data: any = [
        {
          col: 'Cloud Account',
          value: acl.cloudaccountId
        },
        {
          col: 'Region',
          value: acl.regionName
        },
        {
          col: 'Assignment Date',
          value: acl.created
        },
        {
          col: 'Assignment Admin',
          value: acl.adminName
        }
      ]
      setConfirmModalData({
        title: 'Remove whitelisted account',
        data,
        isShow: true,
        id: acl.cloudaccountId,
        regionName: acl.regionName,
        onClose: closeConfirmModal
      })
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
      await RegionManagementService.deleteAcl(confirmModalData.id, confirmModalData.regionName)
      showSuccess(
        `Account ${confirmModalData.id} has been removed from the ${confirmModalData.regionName} region whitelist.`,
        false
      )
      await setAccountWhitelist()
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
    <AccountRegion
      onCancel={onCancel}
      regions={regions}
      data={regionItems}
      regionColumns={columns}
      loading={loading}
      emptyGrid={emptyGridObject}
      filterText={filterText}
      confirmModalData={confirmModalData}
      showLoader={showLoader}
      selectedRegion={selectedRegion}
      setFilter={setFilter}
      onSubmit={onSubmit}
      setRegionFilter={setRegionFilter}
    />
  )
}

export default AccountRegionContainer
