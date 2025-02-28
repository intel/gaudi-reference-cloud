import { create } from 'zustand'
import moment from 'moment'

import type SkuQuota from '../models/skuQuota/SkuQuota'
import SkuQuotaService from '../../services/SkuQuotaService'
import useProductStore from '../productStore/ProductStore'

const dateFormat = 'MM/DD/YYYY hh:mm a'

interface SkuQuotaStore {
  skuQuotas: SkuQuota[] | null
  loading: boolean
  setSkuQuotas: (userId: string) => Promise<void>
}

const useSkuQuotaStore = create<SkuQuotaStore>()((set) => ({
  skuQuotas: [],
  loading: false,
  setSkuQuotas: async (userId: string = '') => {
    set({ loading: true })
    // eslint-disable-next-line @typescript-eslint/await-thenable
    const { data } = await SkuQuotaService.getSkuQuotas()

    const products = useProductStore.getState().products

    if (products.length === 0) {
      await useProductStore.getState().setProducts()
    }

    let formattedSkuQuotas = buildSkuQuotaResponse(data.acl)
    if (userId) formattedSkuQuotas = formattedSkuQuotas.filter(quota => quota.cloudAccountId === userId)
    set({ skuQuotas: formattedSkuQuotas })
    set({ loading: false })
  }
}))

const buildSkuQuotaResponse = (data: any): SkuQuota[] => {
  return data?.map((item: any) => {
    const instance = useProductStore.getState().getProductById(item.productId)
    return {
      resourceId: item?.productId ? item.productId : '',
      family: instance?.familyDisplayName ? instance.familyDisplayName : '',
      instanceName: instance?.displayName ? instance.displayName : '',
      instanceType: instance?.instanceType ? instance.instanceType : '',
      service: instance?.service ? instance.service : '',
      cloudAccountId: item?.cloudaccountId ? item.cloudaccountId : '',
      vendorId: item?.vendorId ? item.vendorId : '',
      creationTimestamp: item?.created ? moment(item.created).format(dateFormat) : '',
      creator: item?.adminName ? item.adminName : '',
      name: instance?.name ? instance.name : '',
      familyId: instance?.familyId ? instance.familyId : ''
    }
  })
}

export default useSkuQuotaStore
