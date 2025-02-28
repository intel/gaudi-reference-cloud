// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'

interface LabelValuePairProps {
  label: string
  children: React.ReactElement
  className?: string
  labelClassName?: string
  small?: boolean
}

const LabelValuePair: React.FC<LabelValuePairProps> = ({ label, children, className, labelClassName, small }) => {
  return (
    <div className={`d-flex flex-column ${className ?? ''}`}>
      <span className={labelClassName ?? `fw-semibold ${small ? 'small' : ''}`}>{label}</span>
      <span className={`d-flex align-items-center text-wrap gap-s4 ${small ? 'small' : ''}`}>{children}</span>
    </div>
  )
}

export default LabelValuePair
