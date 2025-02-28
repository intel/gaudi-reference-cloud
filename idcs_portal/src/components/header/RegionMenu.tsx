// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Nav, Dropdown } from 'react-bootstrap'
import { BsGlobe, BsCheck } from 'react-icons/bs'
import idcConfig from '../../config/configurator'
import useAppStore from '../../store/appStore/AppStore'
import { useSearchParams } from 'react-router-dom'

const RegionMenu: React.FC = (): JSX.Element => {
  const changeRegion = useAppStore((state) => state.changeRegion)
  const [searchParams] = useSearchParams()

  const generateRegionLink = (item: string): string => {
    const newSearchParams = new URLSearchParams(searchParams)
    newSearchParams.set('region', item)
    return '?' + newSearchParams.toString()
  }

  const getRegions = (): JSX.Element[] => {
    const regions = idcConfig.REACT_APP_DEFAULT_REGIONS
    return regions.map(function (item: string) {
      const isSelected = item === idcConfig.REACT_APP_SELECTED_REGION
      return (
        <Dropdown.Item
          key={item}
          intc-id={`region-option-${item}`}
          onClick={(e) => {
            e.preventDefault()
            changeRegion(item)
          }}
          href={generateRegionLink(item)}
        >
          <BsCheck className={isSelected ? 'me-1' : 'invisible'} size="16" intc-id={`selectedIcon-${item}`} />
          <span className={isSelected ? 'active' : ''} aria-label={`Connect to ${item} Region`}>
            {' '}
            {item}{' '}
          </span>
        </Dropdown.Item>
      )
    })
  }

  return (
    <Dropdown intc-id="regionMenu" as={Nav.Item} className="align-self-center">
      <Dropdown.Toggle
        id="dropdown-header-region-toggle"
        role="combobox"
        variant="simple"
        aria-label="Switch region"
        data-bs-toggle="dropdown"
        aria-expanded="false"
        aria-controls="dropdown-header-menu-region"
        className="d-flex align-items-center"
      >
        <BsGlobe intc-id="regionIcon" title="Region Icon" className="d-sm-none" />
        <div className="d-none d-sm-flex">
          <span intc-id="regionLabel" role="note" aria-label={`Selected ${idcConfig.REACT_APP_SELECTED_REGION} region`}>
            {' '}
            {idcConfig.REACT_APP_SELECTED_REGION}{' '}
          </span>
        </div>
      </Dropdown.Toggle>
      <Dropdown.Menu id="dropdown-header-menu-region" renderOnMount aria-labelledby="dropdown-header-region-toggle">
        {getRegions()}
      </Dropdown.Menu>
    </Dropdown>
  )
}

export default RegionMenu
