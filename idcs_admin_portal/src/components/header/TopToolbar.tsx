// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useMemo } from 'react'
import { Navbar, NavbarBrand, Nav, Dropdown } from 'react-bootstrap'
import { NavLink } from 'react-router-dom'
import { type IdcNavigation } from '../../containers/navigation/Navigation.types'
import { useMediaQuery } from 'react-responsive'

interface ToolBarProps {
  idcNavigation?: IdcNavigation
  currentPath: string
}

interface TopToolbarTabsProps {
  idcNavigation: IdcNavigation
  currentPath: string
  isSmScreen: boolean
}

const TopToolbarTabs: React.FC<TopToolbarTabsProps> = ({ idcNavigation, currentPath, isSmScreen }) => {
  if (!idcNavigation.children || idcNavigation.children.length === 0) {
    return null
  }

  const toolbarNavigations = idcNavigation.children.filter((x) => x.showInToolbar === undefined || x.showInToolbar)

  if (toolbarNavigations.length === 0) {
    return null
  }

  const isRouteActive = (pathname: string): boolean => {
    return (
      currentPath.toLowerCase() === pathname.toLowerCase() ||
      (!idcNavigation.children?.some((x) => currentPath.toLowerCase() === x.path) &&
        currentPath.toLowerCase().startsWith(`${pathname.toLowerCase()}/`))
    )
  }

  const activePage = useMemo(() => {
    return idcNavigation.children?.find((x) => isRouteActive(x.path))
  }, [idcNavigation, currentPath])

  if (isSmScreen) {
    return (
      <Nav>
        <Dropdown intc-id="sitetoolbar-navigation-dropdown" as={Nav.Item}>
          <Dropdown.Toggle
            id="dropdown-toolbar-navigation-toggle"
            role="combobox"
            variant="simple"
            aria-label="switch page"
            data-bs-toggle="dropdown"
            aria-expanded="false"
            aria-controls="dropdown-toolbar-menu-navigation"
            className="d-flex align-items-center"
          >
            <div className="d-sm-flex">{activePage?.name}</div>
          </Dropdown.Toggle>
          <Dropdown.Menu
            id="dropdown-toolbar-menu-navigation"
            renderOnMount
            aria-labelledby="dropdown-header-region-toggle"
          >
            {toolbarNavigations.map((idcNavigation, index) => (
              <Dropdown.Item
                key={index}
                className={`${isRouteActive(idcNavigation.path) ? 'active' : ''}`}
                to={idcNavigation.path}
                as={NavLink}
              >
                {idcNavigation.name}
              </Dropdown.Item>
            ))}
          </Dropdown.Menu>
        </Dropdown>
      </Nav>
    )
  }

  return (
    <Nav variant="tabs" activeKey={currentPath} className="align-self-end">
      {toolbarNavigations.map((idcNavigation, index) => (
        <Nav.Item key={index}>
          <Nav.Link
            className={`${isRouteActive(idcNavigation.path) ? 'tap-active' : 'tap-inactive'}`}
            to={idcNavigation.path}
            as={NavLink}
          >
            {idcNavigation.name}
          </Nav.Link>
        </Nav.Item>
      ))}
    </Nav>
  )
}

const TopToolbar: React.FC<ToolBarProps> = ({ idcNavigation, currentPath }) => {
  const Toolbartitle = idcNavigation?.name ?? 'IDC Admin Console'
  const isSmScreen = useMediaQuery({
    query: '(max-width: 768px)'
  })
  return (
    <Navbar fixed="top" className="w-100 siteToolbar pb-0" expand>
      <>
        <NavbarBrand intc-id="toolbarNavBrand" className="align-self-center">
          {idcNavigation?.icon && <idcNavigation.icon className="d-xs-none d-sm-flex" />}
          <h1 className="h4">{Toolbartitle}</h1>
        </NavbarBrand>
        {idcNavigation !== undefined && (
          <TopToolbarTabs idcNavigation={idcNavigation} currentPath={currentPath} isSmScreen={isSmScreen} />
        )}
        <div className='pe-s10'></div>
      </>
    </Navbar>
  )
}

export default TopToolbar
