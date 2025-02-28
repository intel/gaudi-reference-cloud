// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Button from 'react-bootstrap/Button'
import OverlayTrigger from 'react-bootstrap/OverlayTrigger'
import Popover from 'react-bootstrap/Popover'
import SpinnerIcon from '../../spinner/SpinnerIcon'

interface StateTooltipCellProps {
  statusStep?: string | null
  text?: string
  spinnerAtTheEnd?: boolean
  disableSpinner?: boolean
  customIcon?: React.ReactNode
  disableOverlay?: boolean
}

const popover = (message: string | null): JSX.Element => {
  return (
    <Popover id="popover-basic">
      <Popover.Header as="h3">Provisioning status</Popover.Header>
      <Popover.Body>{message}</Popover.Body>
    </Popover>
  )
}

const StateTooltipCell: React.FC<StateTooltipCellProps> = ({
  statusStep = '',
  text = '',
  spinnerAtTheEnd = false,
  disableSpinner = false,
  customIcon,
  disableOverlay = false
}): JSX.Element => {
  if (disableOverlay) {
    return (
      <span className={`d-flex flex-row gap-s4${spinnerAtTheEnd ? 'flex-row-reverse' : ''}`} intc-id="Tooltip Message">
        {customIcon !== undefined ? <>{customIcon}</> : disableSpinner ? null : <SpinnerIcon />}
        {text}
      </span>
    )
  }

  return (
    <div className="d-flex" intc-id="Tooltip Message">
      <OverlayTrigger trigger="focus" placement="right" overlay={popover(statusStep)}>
        <Button variant="link" className={spinnerAtTheEnd ? 'flex-row-reverse' : ''}>
          {customIcon !== undefined ? <>{customIcon}</> : disableSpinner ? null : <SpinnerIcon />}
          {text}
        </Button>
      </OverlayTrigger>
    </div>
  )
}

export default StateTooltipCell
