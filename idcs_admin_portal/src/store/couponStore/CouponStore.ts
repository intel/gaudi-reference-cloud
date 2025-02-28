import { create } from 'zustand'
import CouponService from '../../services/CouponService'

export interface Coupons {
  code: string
  creator: string
  created: string
  start: string
  expires: string
  disabled: boolean
  amount: string
  numUses: number
  numRedeemed: number
  isStandard: boolean
  redemptions: string[]
}

interface CouponsStore {
  coupons: Coupons[] | []
  loading: boolean
  setCoupons: () => Promise<void>
  setLoading: (isLoad: boolean) => void
}

const useCouponsStore = create<CouponsStore>()((set) => ({
  coupons: [],
  loading: false,
  setCoupons: async () => {
    set({ loading: true })
    const { data } = await CouponService.getCoupons()
    set({ coupons: buildCouponsResponse(data.coupons) })
    set({ loading: false })
  },
  setLoading: (isLoad: boolean) => {
    set({ loading: isLoad })
  }
}))

const buildCouponsResponse = (data: any): Coupons[] => {
  // Sorting coupons based on created date.
  return data?.sort(function (a: any, b: any) {
    return new Date(b.created).getTime() - new Date(a.created).getTime()
  })
}

export default useCouponsStore
