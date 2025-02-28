// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import PublicService from '../../services/PublicService'
import type Product from '../models/Product/Product'
import IntelXenon from '../../assets/images/IntelXenon.png'
import IntelGPUFlexPVC from '../../assets/images/IntelGPUFlexPVC.png'
import IntelGPUFlex from '../../assets/images/IntelGPUFlex.png'
import Gaudi from '../../assets/images/Gaudi.png'
import moment from 'moment'

export interface HardwareStore {
  products: Product[] | []
  families: Product[] | []
  familySelected: string
  loading: boolean
  getPricingFamily: (familyDisplayName: string) => boolean
  setFamilyIdSelected: (familyDisplayName: string) => void
  setProducts: (background?: boolean) => Promise<void>
  reset: () => void
}

const getImage = (familyDisplayName: string): any => {
  let imageSrc = null

  if (familyDisplayName.toLowerCase().includes('gaudi')) {
    imageSrc = Gaudi
  } else if (['max', 'gpu'].every((x) => familyDisplayName.toLowerCase().includes(x))) {
    imageSrc = IntelGPUFlexPVC
  } else if (['flex', 'gpu'].every((x) => familyDisplayName.toLowerCase().includes(x))) {
    imageSrc = IntelGPUFlex
  } else {
    imageSrc = IntelXenon
  }

  return imageSrc
}

const initialState = {
  products: [],
  families: [],
  familySelected: '',
  loading: false
}

const useHardwareStore = create<HardwareStore>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  getPricingFamily: (familyDisplayName: string) => {
    const products = get().products.filter(
      (item) => item.familyDisplayName === familyDisplayName && item.rate === '.00'
    )

    let isFreeProduct = false

    if (products.length > 0) {
      isFreeProduct = true
    }

    return isFreeProduct
  },
  setFamilyIdSelected: (familyDisplayName: string) => {
    set({ familySelected: familyDisplayName })
  },
  setProducts: async (background) => {
    if (!background) {
      set({ loading: true })
    }

    try {
      const response = await PublicService.getHardwareCatalog()
      const productDetails = { ...response.data.products }

      const products: Product[] = []

      for (const index in productDetails) {
        const productDetail = {
          ...productDetails[index]
        }

        const metadata = {
          ...productDetail.metadata
        }

        const rates = productDetail.rates
        let accountType = null
        let unit = null
        let rateValue = null
        let usageExpr = null

        if (rates.length > 0) {
          const rate = rates[0]

          if (rate) {
            accountType = filterValue(rate, 'accountType')
            unit = filterValue(rate, 'unit')
            rateValue = filterValue(rate, 'rate')
            usageExpr = filterValue(rate, 'usageExpr')
          }
        }

        const product: Product = {
          name: filterValue(productDetail, 'name'),
          id: filterValue(productDetail, 'id'),
          created: moment(filterValue(productDetail, 'created')).format('MM/DD/YYYY h:mm a'),
          vendorId: filterValue(productDetail, 'vendorId'),
          familyId: filterValue(productDetail, 'familyId'),
          description: filterValue(productDetail, 'description'),
          category: filterValue(metadata, 'category'),
          recommendedUseCase: filterValue(metadata, 'recommendedUseCase'),
          cpuSockets: filterValue(metadata, 'cpu.sockets'),
          cpuCores: filterValue(metadata, 'cpu.cores'),
          diskSize: filterValue(metadata, 'disks.size'),
          displayName: filterValue(metadata, 'displayName'),
          familyDisplayDescription: filterValue(metadata, 'family.displayDescription'),
          releaseStatus: filterValue(metadata, 'releaseStatus'),
          familyDisplayName: filterValue(metadata, 'family.displayName'),
          information: filterValue(metadata, 'information'),
          instanceType: filterValue(metadata, 'instanceType'),
          instanceCategories: filterValue(metadata, 'instanceCategories'),
          memorySize: filterValue(metadata, 'memory.size'),
          nodesCount: filterValue(metadata, 'nodesCount'),
          processor: filterValue(metadata, 'processor'),
          region: filterValue(metadata, 'region'),
          service: filterValue(metadata, 'service'),
          eccn: filterValue(productDetail, 'eccn'),
          pcq: filterValue(productDetail, 'pcq'),
          accountType,
          unit,
          rate: rateValue,
          usageExpr,
          imageSource: getImage(filterValue(metadata, 'family.displayName'))
        }
        products.push(product)
      }

      const families = products.reduce((acc: Product[], current: Product) => {
        if (!acc.find((item: Product) => item.familyDisplayName === current.familyDisplayName)) {
          acc.push(current)
        }
        return acc
      }, [])

      set({ families })
      set({ loading: false })
      set({ products })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  }
}))

const filterValue = (object: any, property: any): any => {
  return object[property] ? object[property] : ''
}

export default useHardwareStore
