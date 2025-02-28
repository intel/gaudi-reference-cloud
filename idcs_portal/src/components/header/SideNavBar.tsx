// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { NavLink } from 'react-router-dom'

import { type IdcNavigation } from '../../containers/navigation/Navigation.types'
import { BsChevronDown, BsChevronUp } from 'react-icons/bs'
import Nav from 'react-bootstrap/Nav'
import { Badge } from 'react-bootstrap'

interface IdcNavItemLinkProps {
  idcNavigation: IdcNavigation
  dataTarget: string
  isActive: boolean
  isSecondLevel?: boolean
}

const IdcNavItemLink: React.FC<IdcNavItemLinkProps> = ({
  idcNavigation,
  dataTarget,
  isActive,
  isSecondLevel
}): JSX.Element => {
  const newBadge = (
    <Badge bg="primary" className="mb-0">
      New
    </Badge>
  )

  const button = (
    <Nav.Item
      intc-id={`sidebarnavLink${idcNavigation.path?.replaceAll(' ', '')}`}
      active={isActive}
      aria-label={`Go to ${idcNavigation.name} Page`}
      className={'d-flex'}
      href={dataTarget}
      as={Nav.Link}
      data-bs-toggle="collapse"
      aria-expanded={isActive}
    >
      <div className="d-flex wrapper">
        {idcNavigation.icon ? <idcNavigation.icon className="icon-wrapper" /> : null}
        <span className={'w-100 d-show-animate'}>{idcNavigation.name}</span>
        {idcNavigation.showBadge ? newBadge : ''}
      </div>
      <BsChevronDown className="ms-s2 caret-down" size={14} />
      <BsChevronUp className="ms-s2 caret-up" size={14} />
    </Nav.Item>
  )

  const isExternalPath = Boolean(idcNavigation.externalPath)

  const extraProps = isExternalPath
    ? {
        href: idcNavigation.externalPath
      }
    : {
        to: idcNavigation.path ?? '/',
        end: true
      }

  const link = (
    <Nav.Item
      intc-id={`sidebarnavLink${idcNavigation.path?.replaceAll(' ', '')}`}
      aria-label={`Go to ${idcNavigation.name} Page`}
      as={isExternalPath ? 'a' : NavLink}
      className={`d-flex nav-link  ${isSecondLevel ? 'second-level' : ''} ${isActive ? 'active' : ''}`}
      {...extraProps}
    >
      <div className="d-flex wrapper">
        {idcNavigation.icon ? <idcNavigation.icon className="icon-wrapper" /> : null}
        <span className={'w-100 d-show-animate'}>{idcNavigation.name}</span>
        {idcNavigation.showBadge ? newBadge : ''}
      </div>
    </Nav.Item>
  )

  return dataTarget ? button : link
}

interface SideNavBarProps {
  currentPath: string
  showSideNavBar: boolean
  navigations: IdcNavigation[]
}
const SideNavBar: React.FC<SideNavBarProps> = ({ navigations, showSideNavBar, currentPath }) => {
  const isRouteActive = (pathname: string, children: IdcNavigation[] | undefined): boolean => {
    return (
      currentPath.toLowerCase() === pathname.toLowerCase() ||
      (!children?.some((x) => currentPath.toLowerCase() === x.path) &&
        currentPath.toLowerCase().startsWith(`${pathname.toLowerCase()}/`))
    )
  }

  const getItemChildren = (firstLevel: IdcNavigation): any => {
    const items: any[] = []
    if (firstLevel.children) {
      firstLevel.children.forEach((secondLevel, index) => {
        if (secondLevel.showInMenu !== false) {
          items.push(
            <IdcNavItemLink
              idcNavigation={secondLevel}
              dataTarget=""
              isActive={isRouteActive(secondLevel.path, firstLevel.children)}
              key={index}
              isSecondLevel
            />
          )
        }
      })
    }
    return items
  }

  const listItems: any = []
  navigations.forEach((firstLevel) => {
    if (firstLevel.showInMenu !== false) {
      if ((firstLevel.children === undefined || !firstLevel.children.some((x) => x.showInMenu)) && firstLevel.path) {
        listItems.push(
          <IdcNavItemLink
            idcNavigation={firstLevel}
            isActive={isRouteActive(firstLevel.path, undefined)}
            dataTarget=""
          />
        )
      }
      if (firstLevel.children && !firstLevel.path) {
        const menuId = firstLevel.name?.replaceAll(' ', '')
        const isActive = firstLevel.children.some((x) => isRouteActive(x.path, firstLevel.children))
        const menu = (
          <>
            <IdcNavItemLink idcNavigation={firstLevel} isActive={isActive} dataTarget={`#subMenu${menuId}`} />
            <ul className={`${isActive ? 'show' : ''} collapse submenu`} id={`subMenu${menuId}`}>
              {getItemChildren(firstLevel)}
            </ul>
          </>
        )
        listItems.push(menu)
      }
    }
  })

  return (
    <div
      intc-id="SideBarNavigationMain"
      role="navigation"
      className={`sideNavBar ${showSideNavBar ? 'showSideNavBar' : ''} offcanvas-sm-width header-scroll-menu`}
    >
      <div intc-id="SideBarNavigationBody" className="offcanvas-body py-s6">
        <Nav className="flex-column">
          {listItems.map((item: JSX.Element, index: number) => (
            <div key={index}>{item}</div>
          ))}
        </Nav>
      </div>
    </div>
  )
}

export default SideNavBar
