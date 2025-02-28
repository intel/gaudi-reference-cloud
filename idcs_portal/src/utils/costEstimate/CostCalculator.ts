// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { costTimeFactorEnum } from '../Enums'

export const costPerMinute = (baseCost: number): number => {
  return baseCost * costTimeFactorEnum.minute
}

export const costPerHour = (baseCost: number): number => {
  return baseCost * costTimeFactorEnum.hour
}

export const costPerDay = (baseCost: number): number => {
  return baseCost * costTimeFactorEnum.day
}

export const costPerWeek = (baseCost: number): number => {
  return baseCost * costTimeFactorEnum.week
}
