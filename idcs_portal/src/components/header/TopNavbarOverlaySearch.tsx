// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Offcanvas from 'react-bootstrap/Offcanvas'
import TopNavbarSearch from './TopNavbarSearch'

interface TopNavBarOverlaySearchProps {
  showSearchOverlay: boolean
  setShowSearchOverlay: (showSearchOverlay: boolean) => void
  searchValue: string
  setSearchValue: (searchValue: string) => void
  onSearchButtonClicked: () => void
}

const TopNavbarOverlaySearch: React.FC<TopNavBarOverlaySearchProps> = ({
  showSearchOverlay,
  setShowSearchOverlay,
  searchValue,
  setSearchValue,
  onSearchButtonClicked
}): JSX.Element => {
  return (
    <Offcanvas
      style={{ height: '64px' }}
      show={showSearchOverlay}
      placement="top"
      onHide={() => {
        setShowSearchOverlay(false)
      }}
    >
      <Offcanvas.Header>
        <TopNavbarSearch
          searchValue={searchValue}
          setSearchValue={setSearchValue}
          onSearchButtonClicked={onSearchButtonClicked}
        />
      </Offcanvas.Header>
    </Offcanvas>
  )
}

export default TopNavbarOverlaySearch
