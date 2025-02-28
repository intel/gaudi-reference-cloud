// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { iso31661, iso31662 } from 'iso-3166'
import useUserStore from '../store/userStore/UserStore'
import { AxiosInstance } from '../utils/AxiosInstance'
import idcConfig from '../config/configurator'
import { countryCodesForAcceptedCountries } from '../utils/Enums'

class PaymentMethodService {
  // Method to retrieve all allocations by cloud account.

  async getCountries() {
    const countriesSorted = iso31661.sort((c1, c2) => (c1.name > c2.name ? 1 : c1.name < c2.name ? -1 : 0))

    return countriesSorted
      .filter((x) => countryCodesForAcceptedCountries.includes(x.alpha2))
      .map((x) => {
        return {
          name: x.name,
          value: x.alpha2
        }
      })
  }

  async getStates(country) {
    if (!country) {
      return []
    }
    const states = iso31662.filter((x) => x.parent === country)
    if (states.length === 0) {
      return [{ name: 'N/A', value: 'N/A' }]
    }
    const statesSorted = states.sort((c1, c2) => (c1.name > c2.name ? 1 : c1.name < c2.name ? -1 : 0))
    return statesSorted.map((x) => {
      return {
        name: x.name,
        value: x.code.split('-')[1]
      }
    })
  }

  getPrePayment() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/billing/payments/prepayment?cloudAccountId=${cloudAccountNumber}`
    return AxiosInstance.get(route)
  }

  getPostPayment(paymentMethod) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/billing/payments/postpayment?cloudAccountId=${cloudAccountNumber}&primaryPaymentMethodNo=${paymentMethod}`
    return AxiosInstance.get(route)
  }

  getCreditCardInfo() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/billing/options?cloudAccountId=${cloudAccountNumber}`
    return AxiosInstance.get(route)
  }

  creditMigrate() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_GLOBAL_SERVICE}/cloudcredits/credit/creditmigrate`
    const payload = {
      cloudAccountId: cloudAccountNumber
    }
    return AxiosInstance.post(route, payload)
  }
}

export default new PaymentMethodService()
