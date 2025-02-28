// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Navigate } from 'react-router-dom'

interface RedirectProps {
  path: string
}

const Redirect: React.FC<RedirectProps> = ({ path }): JSX.Element => {
  return <Navigate to={path} replace />
}

export default Redirect
