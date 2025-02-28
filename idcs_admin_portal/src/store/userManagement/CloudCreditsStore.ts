// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import UserManagementService from '../../services/UserManagementService'
import { formatNumber } from '../../utility/numberFormatHelper/NumberFormatHelper'

export interface CloudCredits {
  creditType: string
  createdAt: string
  ExpiryDate: string
  total: number
  totalUsed: number
  totalRemaining: number
}

interface CloudCreditsStore {
  lastUpdated: string | Date
  remainingCredits: number
  usedCredits: number
  cloudCredits: CloudCredits[] | null
  loading: boolean
  setCloudCredits: (cloudAccountId: string) => Promise<void>
  reset: () => void
}

const initialState = {
  lastUpdated: new Date(),
  remainingCredits: 0.0,
  usedCredits: 0.0,
  cloudCredits: [],
  loading: false
}

const useCloudCreditsStore = create<CloudCreditsStore>()((set) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setCloudCredits: async (cloudAccountId = '') => {
    try {
      set({ loading: true })
      const { data } = await UserManagementService.getCloudCredits(cloudAccountId)

      set({ lastUpdated: data.lastUpdated })
      set({ remainingCredits: formatNumber(data.totalRemainingAmount, 2) })
      set({ usedCredits: formatNumber(data.totalUsedAmount, 2) })
      set({ cloudCredits: buildCreditResponse(data.credits) })

      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  }
}))

// TODO use interface instead of any
const buildCreditResponse = (credits: any): any => {
  return credits.map((credit: any) => {
    return {
      creditType: credit.reason,
      createdAt: credit.created,
      ExpiryDate: credit.expiration,
      total: formatNumber(credit.originalAmount, 2),
      totalUsed: formatNumber(credit.amountUsed, 2),
      totalRemaining: formatNumber(credit.remainingAmount, 2)
    }
  })
}

export default useCloudCreditsStore
