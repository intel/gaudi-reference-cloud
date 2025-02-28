// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import './LineDivider.scss'

interface LineDividerProps {
  horizontal?: boolean
  vertical?: boolean
  className?: string
}

const LineDivider: React.FC<LineDividerProps> = ({
  horizontal = false,
  vertical = false,
  className = ''
}): JSX.Element => {
  if (horizontal) {
    return (
      <div className={`section px-0 ${className}`}>
        <div className="horizontalLine"></div>
      </div>
    )
  }

  if (vertical) {
    return (
      <div className={`section py-0 ${className}`}>
        <div className="verticalLine"></div>
      </div>
    )
  }

  return <></>
}

export default LineDivider
