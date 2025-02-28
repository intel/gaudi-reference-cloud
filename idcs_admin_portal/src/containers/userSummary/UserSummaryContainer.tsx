// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useState, useEffect } from 'react'
import UserSummary from '../../components/userSummary/UserSummary'
import CloudAccountService from '../../services/CloudAccountService'
import { useNavigate, useSearchParams } from 'react-router-dom'

const UserSummaryContainer = (): JSX.Element => {
  // local variables
  const initialTabs = [
    {
      label: 'Information',
      id: 'information'
    },
    {
      label: 'Credit Usage',
      id: 'usage'
    },
    {
      label: 'Cloud Credits',
      id: 'credit'
    },
    {
      label: 'Whitelisted SKUs',
      id: 'sku'
    },
    {
      label: 'Deployed Services',
      id: 'services'
    }
  ]

  const tabDetailsInitial: any = [
    {
      tapTitle: 'User Information',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: 'User Credit Usage',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: 'User Cloud Credits',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: 'Cloud Account Instance Whitelist',
      tapConfig: { type: 'custom' },
      customContent: null
    },
    {
      tapTitle: 'User Deployed Services',
      tapConfig: { type: 'custom' },
      customContent: null
    }
  ]

  // navigation
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()

  // States
  const [cloudAccount, setCloudAccount] = useState('')
  const [cloudAccountError, setCloudAccountError] = useState('')
  const [selectedCloudAccount, setSelectedCloudAccount] = useState(null)
  const [activeTab, setActiveTab] = useState<string | number>(0)
  const [loadingUser, setLoadingUser] = useState(false)

  // hooks
  useEffect(() => {
    const fetchCloudAccount = async (): Promise<void> => {
      const cloudAccount = searchParams.get('cloudAccount')
      if (cloudAccount) {
        setLoadingUser(true)
        try {
          const data = await CloudAccountService.getCloudAccountDetailsByName(cloudAccount)
          setCloudAccount(cloudAccount)
          setSelectedCloudAccount(data?.data)
        } catch (error) {
          setSearchParams(
            (params) => {
              params.set('cloudAccount', 'undefined')
              return params
            },
            { replace: true }
          )
        }
        setLoadingUser(false)
      }
    }
    fetchCloudAccount().catch(() => {})
  }, [])

  // Functions
  // Function to handle form submission
  const onCancel = (): void => {
    navigate('/')
  }

  const handleSubmit = async (e: any): Promise<void> => {
    setCloudAccountError('')
    setSelectedCloudAccount(null)
    if (cloudAccount !== '') {
      setLoadingUser(true)
      try {
        let data: any
        if (cloudAccount.includes('@') || /[a-zA-Z]/.test(cloudAccount)) {
          data = await CloudAccountService.getCloudAccountDetailsByName(cloudAccount)
        } else {
          data = await CloudAccountService.getCloudAccountDetailsById(cloudAccount)
        }

        setCloudAccount(cloudAccount)
        setSelectedCloudAccount(data?.data)
        setCloudAccountError('')
        setSearchParams(
          (params) => {
            params.set('cloudAccount', data?.data?.name)
            return params
          },
          { replace: true }
        )
      } catch (e: any) {
        const code = e.response.data?.code
        const errorMsg = e.response.data?.message
        const message =
          code && [3, 5].includes(code)
            ? String(errorMsg.charAt(0).toUpperCase()) + String(errorMsg.slice(1))
            : 'Cloud Account ID is not found'
        setCloudAccountError(message)
        setSelectedCloudAccount(null)
        setSearchParams(
          (params) => {
            params.set('cloudAccount', 'undefined')
            return params
          },
          { replace: true }
        )
      }
      setLoadingUser(false)
    } else {
      setCloudAccountError('Cloud Account Number is required')
    }
  }

  const handleSearchInputChange = (e: any): void => {
    setCloudAccount(e.target.value)
  }

  return (
    <UserSummary
      loadingUser={loadingUser}
      cloudAccount={cloudAccount}
      cloudAccountError={cloudAccountError}
      selectedCloudAccount={selectedCloudAccount}
      tabs={initialTabs}
      activeTab={activeTab}
      tabDetails={tabDetailsInitial}
      setActiveTab={setActiveTab}
      handleSearchInputChange={handleSearchInputChange}
      onClickSearchButton={handleSubmit}
      onCancel={onCancel}
    />
  )
}

export default UserSummaryContainer
