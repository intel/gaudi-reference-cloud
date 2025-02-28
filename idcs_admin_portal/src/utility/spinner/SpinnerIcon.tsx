// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'

export interface SpinnerIconProps {
  showSpinner?: boolean
  className?: string
}

const SpinnerIcon: React.FC<SpinnerIconProps> = ({ showSpinner = true, className = '' }): JSX.Element => {
  if (!showSpinner) {
    return <></>
  }

  return (
    <div className={`spinner-border spinner-border-sm ${className}`} role="progressbar" aria-label="loading...">
      <div className="invisible visually-hidden">Loading...</div>
    </div>
  )
}

export default SpinnerIcon
