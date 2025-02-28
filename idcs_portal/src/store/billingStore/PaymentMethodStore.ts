// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import PaymentMethodService from '../../services/PaymentMethodService'

export interface Countries {
  name: string
  value: string
}

export interface States {
  name: string
  value: string
}

export interface CreditCard {
  id: number
  created: string | null
  name: string
  cloudAccountId: number
  firstName: string
  middleInitial: string
  lastName: string
  address1: string
  address2: string
  address3: string
  city: string
  locality: string
  postCode: string
  country: string
  company: string
  email: string
  cellPhone: string
  birthDate: string
  paymentType: string
  creditCard: {
    suffix: number
    expiration: string
  }
}

export interface CreditCardResponse {
  success: boolean
  message: string
}

interface PaymentMethodStore {
  countries: Countries[] | null
  states: States[] | null
  creditCard: CreditCard | null
  creditCardResponse: CreditCardResponse | null
  loading: boolean
  setCountries: () => Promise<void>
  setStates: (country: any) => Promise<void>
  setCreditCard: () => Promise<void>
  setCreditCardResponse: (creditCardResponse: CreditCardResponse | null) => void
  reset: () => void
}

const initialState = {
  countries: null,
  states: null,
  creditCard: null,
  creditCardResponse: null,
  loading: false
}

const usePaymentMethodStore = create<PaymentMethodStore>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  setCountries: async () => {
    set({ loading: true })
    const response = await PaymentMethodService.getCountries()

    set({ countries: response })
    set({ loading: false })
  },
  setStates: async (country) => {
    set({ loading: true })
    const response = await PaymentMethodService.getStates(country)
    set({ states: response })
    set({ loading: false })
  },
  setCreditCard: async () => {
    set({ loading: true })
    const { data } = await PaymentMethodService.getCreditCardInfo()

    set({ creditCard: buildCreditCardResponse(data) })

    set({ loading: false })
  },
  setCreditCardResponse: (newCreditCardResponse: CreditCardResponse | null) => {
    if (newCreditCardResponse === null) {
      set({ creditCardResponse: null })
    } else {
      set({ creditCardResponse: { success: newCreditCardResponse.success, message: newCreditCardResponse.message } })
    }
  }
}))

const buildCreditCardResponse = (data: any): any => {
  if (data === '') {
    return null
  }

  return data[0].result
}

export default usePaymentMethodStore
