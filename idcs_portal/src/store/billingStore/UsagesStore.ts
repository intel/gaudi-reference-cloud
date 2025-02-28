// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import UsagesService from '../../services/UsagesService'
import moment from 'moment'
import { formatNumber } from '../../utils/numberFormatHelper/NumberFormatHelper'

const dateFormat = 'MM/DD/YYYY h:mm a'

export interface UsagesReport {
  totalAmount: number
  lastUpdated: string | Date
  downloadUrl: string
  period: string
  usages: UsageDetails[] | []
  loading: boolean
  setUsage: () => Promise<void>
  reset: () => void
}

export interface UsageDetails {
  serviceName: string
  productType: string
  startDate: string
  endDate: string
  amount: number
  rate: string | number
  regionName: string
  usageQuantity: number
  usageUnitName: string
  usageQuantityUnitName: string
  productFamily: string
}

const initialState = {
  totalAmount: 0.0,
  lastUpdated: new Date(),
  downloadUrl: '',
  period: '',
  usages: [],
  loading: false
}

const useUsagesReport = create<UsagesReport>()((set) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setUsage: async () => {
    try {
      set({ loading: true })

      const { data } = await UsagesService.getUsages()

      set({ totalAmount: formatNumber(data.totalAmount, 2) })
      set({ lastUpdated: moment(data.lastUpdated).format(dateFormat) })
      set({ downloadUrl: data.downloadUrl })
      set({ period: data.period })

      const usages = buildUsagesDetails(data.usages)
      usages.sort((a, b) => a.productFamily.localeCompare(b.productFamily))

      set({ usages })

      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  }
}))

// TODO use interface instead of any
const buildUsagesDetails = (usages: any): UsageDetails[] => {
  return usages.map((usage: any) => {
    const billingUsageMetrics = usage.billingUsageMetrics

    return {
      serviceName: usage.serviceName,
      productType: usage.productType,
      startDate: usage.start,
      endDate: usage.end,
      amount: formatNumber(usage.amount, 2),
      rate: formatNumber(usage.rate, 6),
      regionName: usage.regionName,
      usageQuantity: formatNumber(billingUsageMetrics.usageQuantity, 2),
      usageUnitName: billingUsageMetrics.usageUnitName,
      usageQuantityUnitName: billingUsageMetrics.usageQuantityUnitName,
      productFamily: usage.productFamily
    }
  })
}
export default useUsagesReport
