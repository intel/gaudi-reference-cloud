// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import SearchBox from '../../utils/searchBox/SearchBox'

interface TopNavBarSearchProps {
  searchValue: string
  setSearchValue: (searchValue: string) => void
  onSearchButtonClicked: () => void
}

const TopNavbarSearch: React.FC<TopNavBarSearchProps> = ({
  searchValue,
  setSearchValue,
  onSearchButtonClicked
}): JSX.Element => {
  return (
    <SearchBox
      value={searchValue}
      placeholder="Search..."
      onClickSearchButton={onSearchButtonClicked}
      onChange={(e) => {
        setSearchValue(e.target.value)
      }}
    />
  )
}

export default TopNavbarSearch
