// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

export const formatCurrency = (number, local = 'en-US', currency = 'USD') => {
  return new Intl.NumberFormat(local, { style: 'currency', currency }).format(number)
}

export const formatNumber = (number, decimals) => {
  return Number(Number(number).toFixed(decimals))
}
