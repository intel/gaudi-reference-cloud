import { create } from 'zustand'
import PublicService from '../../services/PublicService'
import type Product from '../models/Product/Product'

export interface ProductStore {
  products: Product[] | []
  controlledProducts: Product[] | []
  families: Product[] | []
  controlledFamilies: Product[] | []
  loading: boolean
  setProducts: () => Promise<void>
  getProductByName: (name: string) => Product | null
  getProductById: (id: string) => Product | null
}

const getFamiliesByProducts = (products: Product[]): Product[] | [] => {
  const families = products.reduce((acc: Product[], current: Product) => {
    if (
      !acc.find(
        (item: Product) =>
          item.familyDisplayName === current.familyDisplayName && item.familyId === current.familyId
      )
    ) {
      acc.push(current)
    }
    return acc
  }, [])
  return families
}

const useProductStore = create<ProductStore>()((set, get) => ({
  products: [],
  controlledProducts: [],
  families: [],
  controlledFamilies: [],
  loading: false,
  setProducts: async () => {
    set({ loading: true })

    try {
      const response = await PublicService.getProductCatalog()
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
            accountType = rate.accountType
            unit = rate.unit
            rateValue = rate.rate
            usageExpr = rate.usageExpr
          }
        }

        const product: Product = {
          name: productDetail.name,
          id: productDetail.id,
          vendorId: productDetail.vendorId,
          familyId: productDetail.familyId,
          description: productDetail.description,
          category: metadata.category,
          recommendedUseCase: metadata.recommendedUseCase,
          cpuCores: metadata['cpu.cores'],
          cpuSockets: metadata['cpu.sockets'],
          diskSize: metadata['disks.size'],
          displayName: metadata.displayName,
          familyDisplayDescription: metadata['family.displayDescription'],
          familyDisplayName: metadata['family.displayName'],
          information: metadata.information,
          instanceType: metadata.instanceType,
          instanceCategories: metadata.instanceCategories,
          memorySize: metadata['memory.size'],
          processor: metadata.processor,
          region: metadata.region,
          service: metadata.service,
          eccn: productDetail.eccn,
          pcq: productDetail.pcq,
          accountType,
          unit,
          rate: rateValue,
          usageExpr,
          access: metadata.access,
          nodesCount: metadata.nodesCount
        }

        products.push(product)
      }

      const families = getFamiliesByProducts(products)

      const controlledProducts = products.filter(x => x.access === 'controlled')

      const controlledFamilies = getFamiliesByProducts(controlledProducts)

      set({ families, products, loading: false, controlledFamilies, controlledProducts })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  getProductByName: (name: string) => {
    const product = get().products?.find((item) => item.name === name)

    return product !== undefined ? product : null
  },
  getProductById: (id: string) => {
    const product = get().products?.find((item) => item.id === id)

    return product !== undefined ? product : null
  }
}))

export default useProductStore
