// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { create } from 'zustand'
import PublicService from '../../services/PublicService'
import IntelXenonBronze from '../../assets/images/IntelXenonBronze.png'
import IntelXenonGold from '../../assets/images/IntelXenonGold.png'
import IntelXenonMaxSeries from '../../assets/images/IntelXenonMaxSeries.png'

export interface Product {
  name: string
  description: string
  displayName: string
  threads: string
  instanceCategory: string
  cpu: string
  memory: string
  dimCount: string
  disk: string
  rates: string
  familyId: string
  vendorId: string
  speedMemory: string
  dimmSize: string
  imageSource: string | null
  objectStorage?: boolean
}

export interface MachineOs {
  name: string
  value: string
  displayName: string
  description: string
  labels: string
  components: string
}

interface ProductCatalog {
  products: Product[] | []
  machineOs: MachineOs[] | []
  loading: boolean
  getProductByName: (name: string) => Product | null
  setProducts: () => Promise<void>
  setMachineOs: () => Promise<void>
  reset: () => void
}

export const getImage = (instanceType: string): string | null => {
  let imageSrc = null

  switch (instanceType) {
    case 'm3i.metal':
      imageSrc = IntelXenonGold
      break
    case 'm4i.metal':
      imageSrc = IntelXenonGold
      break
    case 'medium':
      imageSrc = IntelXenonMaxSeries
      break
    case 'large':
      imageSrc = IntelXenonMaxSeries
      break
    default:
      imageSrc = IntelXenonBronze
      break
  }

  return imageSrc
}

const initialState = {
  products: [],
  machineOs: [],
  loading: false
}

const useProductStore = create<ProductCatalog>()((set, get) => ({
  ...initialState,
  reset: () => {
    set(initialState)
  },
  getProductByName: (name: string) => {
    const product = get().products?.find((item) => item.name === name)

    return product !== undefined ? product : null
  },
  setProducts: async () => {
    try {
      set({ loading: true })

      const instanceTypes = await PublicService.getInstanceTypes()

      const productDetails = await PublicService.getHardwareCatalog()

      const products: Product[] = []

      for (const index in instanceTypes.data.items) {
        const instanceType = { ...instanceTypes.data.items[index] }

        const productDetail = productDetails.data.products.find((item: any) => item.metadata.instanceType)

        const product: Product = {
          name: instanceType.metadata.name,
          description: instanceType.spec.description,
          displayName: instanceType.spec.displayName,
          threads: instanceType.spec.cpu.threads,
          instanceCategory: instanceType.spec.instanceCategory,
          cpu: instanceType.spec.cpu.cores,
          memory: instanceType.spec.memory.size,
          speedMemory: instanceType.spec.memory.speed,
          dimmSize: instanceType.spec.memory.dimmSize,
          disk: instanceType.spec.disks[0].size,
          rates: productDetail.rates,
          familyId: productDetail.familyId,
          vendorId: productDetail.vendorId,
          dimCount: instanceType.spec.memory.dimmCount,
          imageSource: getImage(instanceType.metadata.name)
        }

        products.push(product)
      }

      products.sort((p1, p2) => (p1.name < p2.name ? 1 : p1.name > p2.name ? -1 : 0))

      set({ products })

      set({ loading: false })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },
  setMachineOs: async () => {
    const response = await PublicService.getMachineImages()

    const machineOss: MachineOs[] = []
    for (const index in response.data.items) {
      const machineItem = { ...response.data.items[index] }
      const machineOsIm: MachineOs = {
        name: machineItem.metadata.name,
        value: machineItem.metadata.name,
        displayName: machineItem.spec.displayName,
        description: machineItem.spec.description,
        components: machineItem.spec.components,
        labels: machineItem.spec.labels
      }
      machineOss.push(machineOsIm)
    }

    set({ machineOs: machineOss })
  }
}))

export default useProductStore
