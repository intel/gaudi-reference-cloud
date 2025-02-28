// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Nav from 'react-bootstrap/Nav'
import { NavLink } from 'react-router-dom'

interface TabNavigation {
  label: string
  to?: string
}

interface TabsNavigationProps {
  tabs: TabNavigation[]
  activeTab: string | number
  setTabActive: (value: string | number) => void
}

const TabsNavigation: React.FC<TabsNavigationProps> = ({ tabs, activeTab, setTabActive }): JSX.Element => {
  return (
    <Nav variant="tabs" className="tabs-secondary" activeKey={activeTab}>
      {tabs.map((tab, index) => (
        <Nav.Item key={index}>
          {tab.to === undefined
            ? (
            <Nav.Link
              className={activeTab === index ? 'tap-active' : 'tap-inactive'}
              onClick={() => {
                setTabActive(index)
              }}
              aria-current="page"
            >
              {tab.label}
            </Nav.Link>
              )
            : (
            <Nav.Link
              className={activeTab === index ? 'tap-active' : 'tap-inactive'}
              onClick={() => {
                setTabActive(index)
              }}
              aria-current="page"
              to={tab.to}
              as={NavLink}
            >
              {tab.label}
            </Nav.Link>
              )}
        </Nav.Item>
      ))}
    </Nav>
  )
}

export default TabsNavigation
