// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Button, Nav, Navbar } from 'react-bootstrap'
import UserMenu from './UserMenu'
import useLogout from '../../hooks/useLogout'
import SupportMenu from './SupportMenu'
import NotificationButton from './NotificationButton'
import RegionMenu from './RegionMenu'
import { BsSearch } from 'react-icons/bs'
import TopNavbarSearch from './TopNavbarSearch'
import TopNavbarOverlaySearch from './TopNavbarOverlaySearch'
import { appFeatureFlags, isFeatureFlagEnable } from '../../config/configurator'
import { type User } from '../../store/userStore/UserStore'
import TopNavSiteLogo from './TopNavSiteLogo'
import { useCopy } from '../../hooks/useCopy'

interface TopNavBarProps {
  userDetails: User | null
  isOwnCloudAccount: boolean
  searchValue: string
  setSearchValue: (searchValue: string) => void
  onSearchButtonClicked: () => void
  showSearchOverlay: boolean
  setShowSearchOverlay: (showSearchOverlay: boolean) => void
}

const TopNavBar: React.FC<TopNavBarProps> = ({
  userDetails,
  isOwnCloudAccount,
  searchValue,
  setSearchValue,
  onSearchButtonClicked,
  showSearchOverlay,
  setShowSearchOverlay
}): JSX.Element => {
  const searchFeatureEnabled = isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_NAVBAR_SEARCH)
  const { logoutHandler } = useLogout()
  const { copyToClipboard } = useCopy()

  return (
    <Navbar fixed="top" className="w-100 siteNavbar" expand aria-label="Site NavBar">
      {showSearchOverlay && searchFeatureEnabled ? (
        <Nav.Item>
          <TopNavbarOverlaySearch
            showSearchOverlay={showSearchOverlay}
            setShowSearchOverlay={setShowSearchOverlay}
            searchValue={searchValue}
            setSearchValue={setSearchValue}
            onSearchButtonClicked={onSearchButtonClicked}
          />
        </Nav.Item>
      ) : (
        <>
          <TopNavSiteLogo asLink />

          {searchFeatureEnabled ? (
            <Nav.Item className="d-none d-md-flex m-auto header-center-search">
              <TopNavbarSearch
                searchValue={searchValue}
                setSearchValue={setSearchValue}
                onSearchButtonClicked={onSearchButtonClicked}
              />
            </Nav.Item>
          ) : null}

          <Nav>
            {searchFeatureEnabled ? (
              <Button
                className="d-md-none"
                variant="icon-simple"
                onClick={() => {
                  setShowSearchOverlay(true)
                }}
                role="searchbox"
              >
                <BsSearch />
              </Button>
            ) : null}

            <RegionMenu />

            <SupportMenu />

            {isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_NOTIFICATIONS) && <NotificationButton />}

            <UserMenu
              logoutHandler={() => {
                void logoutHandler()
              }}
              userDetails={userDetails}
              isOwnCloudAccount={isOwnCloudAccount}
              copyToClipboard={copyToClipboard}
            />
          </Nav>
        </>
      )}
    </Navbar>
  )
}

export default TopNavBar
