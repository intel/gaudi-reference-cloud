// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect } from 'react'
import Nav from 'react-bootstrap/Nav'
import { NavLink, useSearchParams } from 'react-router-dom'

interface TabNavigation {
  id: string
  label: string
  to?: string
}

interface TabsNavigationProps {
  tabs: TabNavigation[]
  activeTab: string | number
  setTabActive: (value: string | number) => void
}

const TabsNavigation: React.FC<TabsNavigationProps> = ({ tabs, activeTab, setTabActive }): JSX.Element => {
  const [searchParams, setSearchParams] = useSearchParams()

  useEffect(() => {
    const tab: any = searchParams.get('tab') ? searchParams.get('tab') : tabs[0].id
    setTab(tab)
  }, [])

  function setTab(tabId: string): void {
    let tabIndex = tabs.findIndex((tab) => tab.id.toLowerCase() === tabId.toLowerCase())
    if (tabIndex < 0) tabIndex = 0
    setTabActive(tabIndex)
    setSearchParams(
      (params) => {
        params.set('tab', tabs[tabIndex].id)
        return params
      },
      { replace: true }
    )
  }

  return (
    <Nav variant="tabs" className="tabs-secondary" activeKey={activeTab}>
      {tabs.map((tab, index) => (
        <Nav.Item key={index}>
          {tab.to === undefined ? (
            <Nav.Link
              className={activeTab === index ? 'tap-active' : 'tap-inactive'}
              onClick={() => {
                setTab(tab.id)
              }}
              aria-current="page"
              intc-id={`${tab.label}Tab`}
            >
              {tab.label}
            </Nav.Link>
          ) : (
            <Nav.Link
              className={activeTab === index ? 'tap-active' : 'tap-inactive'}
              onClick={() => {
                setTab(tab.id)
              }}
              aria-current="page"
              intc-id={`${tab.label}Tab`}
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
