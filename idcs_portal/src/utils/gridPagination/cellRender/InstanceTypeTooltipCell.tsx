// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Tooltip, OverlayTrigger } from 'react-bootstrap'
import { BsInfoCircle } from 'react-icons/bs'

interface InstanceTypeTooltipCellProps {
  name: string
  displayName: string
}

const InstanceTypeTooltipCell: React.FC<InstanceTypeTooltipCellProps> = ({ name, displayName }) => {
  const renderTooltip = (props: any): JSX.Element => (
    <Tooltip {...props} id={`tooltip-instance-type-displayname-${name}`}>
      {displayName}
    </Tooltip>
  )
  return (
    <>
      {name.toUpperCase()}
      <OverlayTrigger placement="bottom" flip overlay={renderTooltip}>
        <div tabIndex={0} className="tooltipinfo-icon" aria-label={displayName} role="note">
          <BsInfoCircle />
        </div>
      </OverlayTrigger>
    </>
  )
}

export default InstanceTypeTooltipCell
