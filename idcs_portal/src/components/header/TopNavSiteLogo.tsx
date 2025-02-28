// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Nav, NavbarBrand } from 'react-bootstrap'
import { NavLink } from 'react-router-dom'
import { ReactComponent as TiberLogo } from '../../assets/images/intelTiberDeveloperCloudLogoMini.svg'

interface TopNavSiteLogoProps {
  asLink?: boolean
}

const TopNavSiteLogo: React.FC<TopNavSiteLogoProps> = ({ asLink }): JSX.Element => {
  return (
    <NavbarBrand intc-id="siteNavBrand">
      {asLink && (
        <Nav.Link
          className={'d-flex text-center align-self-center align-items-center gap-s6'}
          intc-id="siteHeaderButton"
          to="/home"
          aria-label="Go to homepage"
          as={NavLink}
        >
          <TiberLogo />
          <img className="siteLogo d-none d-sm-flex" intc-id="siteLogo" aria-label="Site Logo" />
        </Nav.Link>
      )}
      {!asLink && (
        <div className="d-flex text-center align-self-center align-items-center gap-s6">
          <TiberLogo />
          <img className="siteLogo d-none d-sm-flex" intc-id="siteLogo" aria-label="Site Logo" />
        </div>
      )}
    </NavbarBrand>
  )
}

export default TopNavSiteLogo
