// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import './Spinner.scss'

export interface SpinnerProps {
  showSpinner?: boolean
  className?: string
}

const Spinner: React.FC<SpinnerProps> = ({ showSpinner = true, className = '' }): JSX.Element => {
  if (!showSpinner) {
    return <></>
  }

  return (
    <div
      className={`section align-self-center justify-content-center align-items-center ${className}`}
      role="progressbar"
      aria-label="loading..."
    >
      <div className="align-self-center">
        <div className="spinner-border text-primary center">
          <div className="invisible visually-hidden">Loading...</div>
        </div>
      </div>
    </div>
  )
}

export default Spinner
