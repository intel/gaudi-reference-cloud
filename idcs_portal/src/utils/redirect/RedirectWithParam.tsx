// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Navigate, useParams } from 'react-router-dom'

interface RedirectProps {
  path: string
}

const RedirectWithParam: React.FC<RedirectProps> = ({ path }): JSX.Element => {
  const { param } = useParams()
  return <Navigate to={`${path}/${param}`} replace />
}

export default RedirectWithParam
