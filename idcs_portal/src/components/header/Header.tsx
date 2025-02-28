// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import TopNavBar from './TopNavBar'
import React, { useState } from 'react'
import { type User } from '../../store/userStore/UserStore'
import SingleTopNavBar from './SingleTopNavBar'
import TopToolbarContainer from '../../containers/navigation/TopToolbarContainer'
import BannerContainer from '../../containers/banner/BannerContainer'

interface HeaderProps {
  userDetails: User | null
  isOwnCloudAccount: boolean
  hideMenu: boolean
  pathname: string
}

const Header: React.FC<HeaderProps> = ({ userDetails, isOwnCloudAccount, pathname, hideMenu }): JSX.Element => {
  const [searchValue, setSearchValue] = useState('')
  const [showSearchOverlay, setShowSearchOverlay] = useState(false)

  const onSearchButtonClicked = (): void => {
    setShowSearchOverlay(false)
    // TO DO: implement global search
  }

  return (
    <>
      {!hideMenu ? (
        <>
          <TopNavBar
            userDetails={userDetails}
            isOwnCloudAccount={isOwnCloudAccount}
            searchValue={searchValue}
            setSearchValue={setSearchValue}
            onSearchButtonClicked={onSearchButtonClicked}
            showSearchOverlay={showSearchOverlay}
            setShowSearchOverlay={setShowSearchOverlay}
          />
          <TopToolbarContainer />
          <BannerContainer />
        </>
      ) : (
        <>
          <SingleTopNavBar />
        </>
      )}
    </>
  )
}

export default Header
