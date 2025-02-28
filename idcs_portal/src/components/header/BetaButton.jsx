// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Link } from 'react-router-dom'
import { BsStars } from 'react-icons/bs'
import idcConfig from '../../config/configurator'
import Badge from 'react-bootstrap/Badge'

const BetaButton = () => {
  return (
    <Link
      intc-id="topNavBarTryBeta"
      className="betaButton"
      to={idcConfig.REACT_APP_GUI_BETA_DOMAIN}
      role="button"
      aria-label="Go to IDC Console Beta"
    >
      <BsStars className="img-fluid" color="black" />
      <span className="mx-2">Try our new look</span>
      <Badge>Beta</Badge>
    </Link>
  )
}

export default BetaButton
