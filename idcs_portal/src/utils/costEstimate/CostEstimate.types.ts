// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

interface CostItem {
  label: string
  value: string
}

export interface CostEstimateProps {
  title: string
  description: string
  costArray: CostItem[]
}
