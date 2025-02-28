// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { BsBoxArrowRight } from 'react-icons/bs'
import { Nav, Navbar } from 'react-bootstrap'
import useLogout from '../../hooks/useLogout'
import TopNavSiteLogo from './TopNavSiteLogo'

const SingleTopNavBar: React.FC = (): JSX.Element => {
  const { logoutHandler } = useLogout()

  return (
    <Navbar fixed="top" className="w-100 siteNavbar" expand aria-label="Site NavBar">
      <TopNavSiteLogo />
      <Nav.Item
        onClick={() => {
          void logoutHandler()
        }}
        intc-id="signOutHeaderButton"
        aria-label="Sign out user"
        as={Nav.Link}
        className="btn btn-link"
      >
        <BsBoxArrowRight intc-id="signoutIcon" title="sign out Icon" />
        Sign-out
      </Nav.Item>
    </Navbar>
  )
}

export default SingleTopNavBar
