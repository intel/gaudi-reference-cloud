// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Card } from 'react-bootstrap'
import DashboardCardGenAi from '../../../assets/images/DashboardCardGenAi.png'

const ImageWidget = (): JSX.Element => {
  return (
    <Card className="h-100">
      <Card.Img className="dashboard-image" src={DashboardCardGenAi} />
    </Card>
  )
}

export default ImageWidget
