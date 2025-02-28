// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import SearchBox from '../../utility/searchBox/SearchBox'
import { Button } from 'react-bootstrap'
import TabsNavigation from '../../utility/tabsNavigation/TabsNavagation'
import TapContent from '../../utility/TapContent/TapContent'
import AccountSettingsView from './AccountSettingsView'
import SkuQuotaDetailsContainer from '../../containers/skuManagement/SkuQuotaDetailsContainer'
import CreditUsageContainer from '../../containers/userSummary/CreditUsageContainer'
import DeployedServicesContainer from '../../containers/userSummary/DeployedServicesContainer'
import CloudCreditsContainers from '../../containers/userSummary/CloudCreditsContainers'

interface UserSummaryProps {
  loadingUser: boolean
  cloudAccount: string
  cloudAccountError: string
  selectedCloudAccount: any
  tabs: any[]
  activeTab: string | number
  tabDetails: any[]
  setActiveTab: (tab: string | number) => void
  handleSearchInputChange: (e: any) => void
  onClickSearchButton: (e: any) => Promise<void>
  onCancel: () => void
}

const UserSummary: React.FC<UserSummaryProps> = (props: any): JSX.Element => {
  // props
  const loadingUser = props.loadingUser
  const cloudAccount = props.cloudAccount
  const cloudAccountError = props.cloudAccountError
  const selectedCloudAccount = props.selectedCloudAccount
  const tabs = props.tabs
  const activeTab = props.activeTab
  const tabDetails = props.tabDetails
  const setActiveTab = props.setActiveTab
  const handleSearchInputChange = props.handleSearchInputChange
  const onClickSearchButton = props.onClickSearchButton
  const onCancel = props.onCancel

  const emptyUserDisplay = (
    <div className="section align-items-center">
      <span className="h4">User not found/selected</span>
      <p className="add-break-line lead">Please search with Cloud Accound Email/ID</p>
    </div>
  )

  // functions
  const getTabInfo = (tab: number): any => {
    const tabLabel = tabs[Number(tab)]?.id
    let content = null
    switch (tabLabel) {
      case 'information':
        tabDetails[tab].customContent =
          selectedCloudAccount || loadingUser ? (
            <AccountSettingsView user={selectedCloudAccount} loading={loadingUser} />
          ) : (
            emptyUserDisplay
          )
        content = tabDetails[tab]
        break
      case 'usage':
        tabDetails[tab].customContent = selectedCloudAccount ? (
          <CreditUsageContainer userId={selectedCloudAccount?.id} />
        ) : (
          emptyUserDisplay
        )
        content = tabDetails[tab]
        break
      case 'credit':
        tabDetails[tab].customContent = selectedCloudAccount ? (
          <CloudCreditsContainers userId={selectedCloudAccount?.id} />
        ) : (
          emptyUserDisplay
        )
        content = tabDetails[tab]
        break
      case 'sku':
        tabDetails[tab].customContent = selectedCloudAccount ? (
          <SkuQuotaDetailsContainer viewMode={true} userId={selectedCloudAccount?.id} />
        ) : (
          emptyUserDisplay
        )
        content = tabDetails[tab]
        break
      case 'services':
        tabDetails[tab].customContent = selectedCloudAccount ? (
          <DeployedServicesContainer userId={selectedCloudAccount?.id} userEmail={selectedCloudAccount?.name} />
        ) : (
          emptyUserDisplay
        )
        content = tabDetails[tab]
        break
      default:
        content = tabDetails[tab]
        break
    }

    return content
  }

  return (
    <>
      <div className="section">
        <Button variant="link" className="p-s0" onClick={onCancel}>
          ⟵ Back to Home
        </Button>
        <div className="filter p-0">
          <h2 className="h4">User: {selectedCloudAccount?.name ?? 'N/A'}</h2>
          <div>
            <SearchBox
              intc-id="searchAccounts"
              value={cloudAccount || ''}
              onChange={handleSearchInputChange}
              onClickSearchButton={onClickSearchButton}
              placeholder="Enter cloud account email or ID"
              aria-label="Enter cloud account email or ID"
            />
            {cloudAccountError && (
              <div className="text-danger text-center w-100" intc-id="cloudAccountSearchError">
                {cloudAccountError}
              </div>
            )}
          </div>
        </div>
      </div>
      <div className="section">
        <TabsNavigation tabs={tabs} activeTab={activeTab} setTabActive={setActiveTab} />
        <TapContent infoToDisplay={getTabInfo(activeTab)} />
      </div>
    </>
  )
}

export default UserSummary
