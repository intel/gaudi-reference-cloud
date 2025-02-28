// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockBaseEnrollResponse } from '../authentication/authHelper'
import { mockAxios } from '../../../setupTests'

export const mockBaseInvoicesStore = () => {
  return {
    lastUpdated: '2023-06-13T16:56:55.720541785Z',
    invoices: [mockBaseInvoiceDetails(), mockBaseInvoiceDetails(), mockBaseInvoiceDetails()]
  }
}

export const mockBaseInvoiceDetails = () => {
  return {
    cloudAccountId: '583807362652',
    id: '1377837239',
    total: 0,
    paid: 0,
    due: 0,
    start: '2023-05-01T00:00:00Z',
    end: '2023-05-31T00:00:00Z',
    invoiceDate: null,
    dueDate: null,
    notifyDate: null,
    paidDate: null,
    billingPeriod: 'May, 2023',
    status: '',
    statementLink: '',
    downloadLink: ''
  }
}

export const mockInvoices = () => {
  const { cloudAccountId } = mockBaseEnrollResponse()
  when(mockAxios.get)
    .calledWith(expect.stringContaining(`/billing/invoices?cloudAccountId=${cloudAccountId}`), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseInvoicesStore()
      })
    )
}
