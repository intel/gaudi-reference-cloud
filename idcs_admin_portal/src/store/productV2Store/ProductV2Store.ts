import { create } from 'zustand'
import PublicService from '../../services/PublicService'
import moment from 'moment'

export interface Product {
  id?: string
  name: string
  serviceName: string
  regionName: string
  familyName: string
  rateId?: string
  usage: string
  status?: string
  createdAt?: string
  updatedAt?: string
  metaDataSets: productMetaDataSet[]
}

export interface productMetaDataSet {
  id: string
  name: string
  context: string
  metaData?: productMetaData[]
  createdAt?: string
  updatedAt?: string
}

export interface productMetaData {
  id: string
  metadataSetId: string
  key: string
  value: string
  type: string
  createdAt?: string
  updatedAt?: string
}

export interface serviceRegistration {
  name: string
  description: string
  adminName: string
  createdAt?: string
  updatedAt?: string
}

export interface ProductStore {
  products: Product[] | null
  productsByFamily: Product[] | null
  product: Product | null
  newProduct: Product | null
  editMode: boolean
  editProduct: Product | null
  productServices: serviceRegistration[] | null
  getProducts: () => Promise<void>
  getProductByFamily: (familyName: string) => void
  getProductByName: (productName: string) => Promise<void>
  setproductServices: () => Promise<void>
  setNewProduct: (newProduct: any) => void
  setEditProduct: (editProduct: any) => void
  setEditMode: (editMode: boolean) => void
  loading: boolean
}

const buildProductArray = (data: any): Product[] => {
  const products: Product[] = []
  data.forEach((item: any) => {
    const product = buildProductItem(item)
    products.push(product)
  })

  return products
}

const buildProductItem = (data: any): Product => {
  const dataProductResponse = { ...data.product }

  const metadataSetsResponse = [...dataProductResponse.metadatasets]
  const metaDataSets: productMetaDataSet[] = []
  metadataSetsResponse.forEach((metaDataSetItem, index) => {
    const metaDataResponse = [...metaDataSetItem.metadata]
    const metaData: productMetaData[] = []
    metaDataResponse.forEach((metaDataItem) => {
      metaData.push({
        id: metaDataItem.id,
        key: metaDataItem.key,
        value: metaDataItem.value,
        type: metaDataItem.type,
        metadataSetId: String(index),
        createdAt: moment(metaDataItem.created_at).format('MM/DD/YYYY hh:mm:ss'),
        updatedAt: moment(metaDataItem.updated_at).format('MM/DD/YYYY hh:mm:ss')
      })
    })
    metaDataSets.push({
      id: String(index),
      name: metaDataSetItem.name,
      context: metaDataSetItem.context,
      createdAt: moment(metaDataSetItem.created_at).format('MM/DD/YYYY hh:mm:ss'),
      updatedAt: moment(metaDataSetItem.updated_at).format('MM/DD/YYYY hh:mm:ss'),
      metaData
    })
  })

  const product: Product = {
    id: dataProductResponse.id,
    name: dataProductResponse.name,
    serviceName: dataProductResponse.service_name,
    regionName: dataProductResponse.region_name,
    familyName: dataProductResponse.product_family_name,
    metaDataSets,
    rateId: dataProductResponse.rate_set_id,
    usage: dataProductResponse.usage,
    status: dataProductResponse.status,
    createdAt: moment(dataProductResponse.created_at).format('MM/DD/YYYY hh:mm:ss'),
    updatedAt: moment(dataProductResponse.updated_at).format('MM/DD/YYYY hh:mm:ss')
  }

  return product
}

const buildServiceRegistrationArray = (data: any): serviceRegistration[] => {
  const response: serviceRegistration[] = []

  data.forEach((item: any) => {
    response.push({
      name: item.name,
      description: item.description,
      adminName: item.admin_name,
      createdAt: moment(item.created_at).format('MM/DD/YYYY hh:mm:ss'),
      updatedAt: moment(item.updated_at).format('MM/DD/YYYY hh:mm:ss')
    })
  })

  return response
}

const useProductV2Store = create<ProductStore>()((set, get) => ({
  products: null,
  productsByFamily: null,
  product: null,
  newProduct: null,
  editMode: false,
  editProduct: null,
  productServices: null,
  getProducts: async () => {
    set({ loading: true })
    const response = await PublicService.getProducts()
    const productDetails = [...response.data.product_details]
    const products = buildProductArray(productDetails)
    set({ products })
    set({ loading: false })
  },
  getProductByFamily: (familyName: string) => {
    set({ loading: true })
    const products = get().products?.filter((item) => item.familyName === familyName)
    set({ productsByFamily: products })
    set({ loading: false })
  },
  getProductByName: async (productName: string) => {
    set({ loading: true })
    const response = await PublicService.getProductByName(productName)
    const productDetails = [...response.data.product_details]
    let product = null
    if (productDetails.length > 0) {
      product = buildProductItem(productDetails[0])
    }
    set({ product })
    set({ loading: false })
  },
  setNewProduct: (newProduct: any) => {
    set({ newProduct })
  },
  setEditProduct: (editProduct: any) => {
    set({ editProduct })
  },
  setEditMode: (editMode: boolean) => {
    set({ editMode })
  },
  setproductServices: async () => {
    const response = await PublicService.getProductServices()
    const serviceDetails = [...response.data.services]
    const productServices = buildServiceRegistrationArray(serviceDetails)
    set({ productServices })
  },
  loading: false
}))

export default useProductV2Store
