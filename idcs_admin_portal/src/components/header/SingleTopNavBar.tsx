// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { BsBoxArrowRight } from 'react-icons/bs'
import { Nav, Navbar } from 'react-bootstrap'
import useLogout from '../../hooks/useLogout'

const SingleTopNavBar: React.FC = (): JSX.Element => {
  const { logoutHandler } = useLogout()

  return (
    <Navbar fixed="top" className="w-100 siteNavbar" expand aria-label="Site NavBar">
      <div className="h4" style={{ cursor: 'pointer' }}>
        Intel Tiber AI Cloud Admin Console
      </div>
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
