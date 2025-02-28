// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useMemo } from 'react'
import { Navbar, Nav, Button, Dropdown } from 'react-bootstrap'
import { NavLink } from 'react-router-dom'
import { type IdcNavigation } from '../../containers/navigation/Navigation.types'
import { BsBook } from 'react-icons/bs'
import useAppStore from '../../store/appStore/AppStore'
import { useMediaQuery } from 'react-responsive'
import TopToolbarTitle from './TopToolbarTitle'
import './TopToolbar.scss'

interface ToolBarProps {
  idcNavigation?: IdcNavigation
  currentPath: string
}

interface TopToolbarTabsProps {
  idcNavigation: IdcNavigation
  currentPath: string
  isMdScreen: boolean
}

const TopToolbarTabs: React.FC<TopToolbarTabsProps> = ({ idcNavigation, currentPath, isMdScreen }) => {
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

  if (!idcNavigation.children || idcNavigation.children.length === 0 || !idcNavigation.children.some((x) => x.name)) {
    return null
  }

  const toolbarNavigations = idcNavigation.children.filter((x) => x.showInToolbar === undefined || x.showInToolbar)

  if (toolbarNavigations.length === 0) {
    return null
  }

  if (isMdScreen) {
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
  const showLearningBar = useAppStore((state) => state.showLearningBar)
  const setShowLearningBar = useAppStore((state) => state.setShowLearningBar)
  const showSideNavBar = useAppStore((state) => state.showSideNavBar)
  const setShowSideNavBar = useAppStore((state) => state.setShowSideNavBar)
  const learningArticlesAvailable = useAppStore((state) => state.learningArticlesAvailable)

  const isMdScreen = useMediaQuery({
    query: '(max-width: 991px)'
  })

  const isXsScreen = useMediaQuery({
    query: '(max-width: 575px)'
  })

  return (
    <Navbar fixed="top" className="w-100 siteToolbar pb-0" expand>
      <>
        <TopToolbarTitle
          idcNavigation={idcNavigation}
          setShowSideMenu={setShowSideNavBar}
          showExpandMenu={!showSideNavBar}
          showCollapseMenu={showSideNavBar}
          isXsScreen={isXsScreen}
        />
        {idcNavigation !== undefined && (
          <TopToolbarTabs idcNavigation={idcNavigation} currentPath={currentPath} isMdScreen={isMdScreen} />
        )}
        <Nav.Item className={`align-self-center ${!isXsScreen ? 'd-flex justify-content-end w-220px' : ''}`}>
          <Button
            intc-id="btn-toolbar-learning"
            variant="simple"
            className={`d-none d-sm-flex ${showLearningBar ? 'show' : ''}`}
            onClick={() => {
              setShowLearningBar(!showLearningBar, true)
            }}
            aria-label="Open learning bar"
            disabled={!learningArticlesAvailable}
          >
            <BsBook />
            Documentation
          </Button>
          <Button
            intc-id="btn-ico-toolbar-learning"
            variant="icon-simple"
            className={`d-xs-flex d-sm-none mb-1 ${showLearningBar ? 'show' : ''}`}
            aria-label="Open learning bar"
            onClick={() => {
              setShowLearningBar(!showLearningBar, true)
            }}
            disabled={!learningArticlesAvailable}
          >
            <BsBook />
          </Button>
        </Nav.Item>
      </>
    </Navbar>
  )
}

export default TopToolbar
