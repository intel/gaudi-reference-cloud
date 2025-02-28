// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { AxiosInstance } from '../utils/AxiosInstance'
import useUserStore from '../store/userStore/UserStore'
import idcConfig from '../config/configurator'

class InvoicesService {
  // Method to retrieve all allocations by cloud account.

  getInvoices() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/billing/invoices?cloudAccountId=${cloudAccountNumber}`
    return AxiosInstance.get(route)
  }

  async getInvoicePDF(invoiceId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/billing/invoices/statement?cloudAccountId=${cloudAccountNumber}&invoiceId=${invoiceId}`
    return AxiosInstance.get(route)
  }
}

export default new InvoicesService()
