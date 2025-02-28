// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import Card from 'react-bootstrap/Card'
import { type CostEstimateProps } from './CostEstimate.types'

const CostEstimateCard: React.FC<CostEstimateProps> = (props): JSX.Element => {
  const { title, description, costArray } = props

  return (
    <Card className="cost-estimate-card">
      <Card.Body>
        <h5>{title}</h5>
        <small>{description}</small>
        {costArray.map((item, index) => (
          <div key={index} className="d-flex justify-content-between w-100">
            <span className="fw-semibold">{item.label}</span>
            <span>{item.value}</span>
          </div>
        ))}
      </Card.Body>
    </Card>
  )
}

export default CostEstimateCard
