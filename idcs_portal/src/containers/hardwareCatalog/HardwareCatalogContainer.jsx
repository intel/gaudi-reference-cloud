// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import HardwareCatalog from '../../components/hardwareCatalog/HardwareCatalog'
import useHardwareStore from '../../store/hardwareStore/HardwareStore'
import { useNavigate } from 'react-router'
import useErrorBoundary from '../../hooks/useErrorBoundary'
import { useSearchParams } from 'react-router-dom'

const HardwareCatalogContainer = () => {
  // localState
  const [productTextFilter, setProductTextFilter] = useState('')
  const [searchFilter, setSearchFilter] = useState({})
  const [isProductServiceAvailable, setIsProductServiceAvailable] = useState(true)

  const [searchParams, setSearchParams] = useSearchParams()
  const [searchValues, setSearchValues] = useState([])
  const [isPageReady, setIsPageReady] = useState(false)
  const defaultKeySearch = 'recommendedUseCase'

  // Store
  const products = useHardwareStore((state) => state.products)
  const families = useHardwareStore((state) => state.families)
  const setProducts = useHardwareStore((state) => state.setProducts)
  const loading = useHardwareStore((state) => state.loading)
  const setFamilyIdSelected = useHardwareStore((state) => state.setFamilyIdSelected)

  // Navigation
  const navigate = useNavigate()

  // Error handle
  const throwError = useErrorBoundary()

  // Hooks
  useEffect(() => {
    const fetchProducts = async () => {
      try {
        await setProducts(products.length > 0)
        setIsPageReady(true)
      } catch (error) {
        setIsPageReady(true)
        let errorMessage = ''
        let errorCode = ''
        let errorStatus = -1
        const isApiErrorWithErrorMessage = Boolean(error.response && error.response.data && error.response.data.message)
        if (isApiErrorWithErrorMessage) {
          errorMessage = error.response.data.message
          errorCode = error.response.data.code
          errorStatus = error.response.status
        } else {
          errorMessage = error.toString()
        }

        if (errorStatus === 403 && errorCode === 7 && errorMessage.toLowerCase().indexOf('user is restricted') !== -1) {
          throwError(error)
        }

        setIsProductServiceAvailable(false)
      }
    }
    fetchProducts()
  }, [])

  useEffect(() => {
    const values = searchParams.get('fctg') ? searchParams.get('fctg').split(',') : ['All']
    if (values.length > 0 && isPageReady) {
      setTagFilter(defaultKeySearch, values)
      setSearchValues(values)
    }
  }, [isPageReady])

  async function onSelectedProduct(event, itemSelected) {
    event.preventDefault()

    const familyDisplayName = itemSelected.familyDisplayName
    const category = itemSelected.category.toLowerCase()

    setFamilyIdSelected(familyDisplayName)

    if (category === 'cluster') {
      navigate({
        pathname: '/compute-groups/reserve',
        search: '?backTo=catalog'
      })
    } else {
      navigate({
        pathname: '/compute/reserve',
        search: '?backTo=catalog'
      })
    }
  }

  function setTagFilter(key, values) {
    if (isPageReady) {
      const updatedFilter = { ...searchFilter }
      updatedFilter[key] = values
      setSearchFilter(updatedFilter)
      setSearchParams(
        (params) => {
          params.set('fctg', values)
          return params
        },
        { replace: true }
      )
    }
  }

  return (
    <HardwareCatalog
      products={products}
      families={families}
      loading={loading}
      productTextFilter={productTextFilter}
      searchFilter={searchFilter}
      searchValues={searchValues}
      setTagFilter={setTagFilter}
      onChangeProductFilter={setProductTextFilter}
      onSelectedProduct={onSelectedProduct}
      isProductServiceAvailable={isProductServiceAvailable}
    />
  )
}

export default HardwareCatalogContainer
