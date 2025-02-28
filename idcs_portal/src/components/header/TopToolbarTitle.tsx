// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Button, NavbarBrand } from 'react-bootstrap'
import { BsList } from 'react-icons/bs'
import { type IdcNavigation } from '../../containers/navigation/Navigation.types'
import idcConfig from '../../config/configurator'
import './TopToolbar.scss'

interface TopToolbarTitleProps {
  idcNavigation?: IdcNavigation
  showExpandMenu?: boolean
  showCollapseMenu?: boolean
  isXsScreen: boolean
  setShowSideMenu?: (show: boolean, savePreference: boolean) => void
}

const TopToolbarTitle: React.FC<TopToolbarTitleProps> = ({
  idcNavigation,
  showExpandMenu,
  showCollapseMenu,
  isXsScreen,
  setShowSideMenu
}): JSX.Element => {
  const showMenu = (value: boolean): void => {
    if (setShowSideMenu !== undefined) {
      setShowSideMenu(value, true)
    }
  }

  const Toolbartitle = idcNavigation?.toolbarTitle ?? idcNavigation?.name ?? idcConfig.REACT_APP_CONSOLE_SHORT_NAME

  return (
    <NavbarBrand intc-id="toolbarNavBrand" className={`align-self-center ${!isXsScreen ? 'w-220px' : ''}`}>
      {showExpandMenu && (
        <Button
          intc-id="btn-site-nav-expand"
          variant="icon-simple"
          aria-label="Expand side menu"
          onClick={() => {
            showMenu(true)
          }}
        >
          <BsList />
        </Button>
      )}
      {showCollapseMenu && (
        <Button
          intc-id="btn-site-nav-expand"
          variant="icon-simple"
          aria-label="Collapse side menu"
          className="show"
          onClick={() => {
            showMenu(false)
          }}
        >
          <BsList />
        </Button>
      )}
      {!isXsScreen && <h1 className="h4">{Toolbartitle}</h1>}
    </NavbarBrand>
  )
}

export default TopToolbarTitle
