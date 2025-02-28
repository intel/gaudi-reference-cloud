// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import InvoicesService from '../../services/InvoicesService'
import moment from 'moment'
import { formatNumber } from '../../utils/numberFormatHelper/NumberFormatHelper'

export interface Invoices {
  id: number
  period: string
  startDate: string
  endDate: string
  dueDate: string
  status: string
  total: number
  totalPaid: number
  totalDue: number
}

export interface InvoiceUsage {
  amount: number
  end: string
  minsUsed: number
  productType: string
  serviceName: string
  start: string
}

interface InvoicesStore {
  lastUpdated: string | Date
  invoices: Invoices[] | null
  invoiceTotal: number
  loading: boolean
  setInvoices: () => Promise<void>
  reset: () => void
}

const initialState = {
  lastUpdated: new Date(),
  invoices: [],
  invoiceTotal: 0.0,
  loading: false
}

const useInvoicesStore = create<InvoicesStore>()((set) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setInvoices: async () => {
    set({ loading: true })
    const { data } = await InvoicesService.getInvoices()

    set({ lastUpdated: data.lastUpdated })
    set({ invoices: buildInvoiceResponse(data.invoices) })
    set({ loading: false })
  }
}))

// TODO use interface instead of any
const buildInvoiceResponse = (invoices: any): any => {
  return invoices?.map((invoice: any) => {
    return {
      id: invoice.id,
      period: moment(invoice.billingPeriod, 'MMMM ,yyyy').utc().toISOString(),
      startDate: invoice.start,
      endDate: invoice.end,
      dueDate: invoice.dueDate,
      status: invoice.status,
      total: formatNumber(invoice.total, 2),
      totalPaid: formatNumber(invoice.paid, 2),
      totalDue: formatNumber(invoice.due, 2)
    }
  })
}

export default useInvoicesStore
