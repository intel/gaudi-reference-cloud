// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import ProductList from './ProductList'
import ProductCatalogueUnavailable from '../../pages/error/ProductCatalogueUnavailable'
import ProductFilter from '../../utils/productFilter/ProductFilter'
import Spinner from '../../utils/spinner/Spinner'

const HardwareCatalog = (props) => {
  // props
  const searchFilter = props.searchFilter
  const productTextFilter = props.productTextFilter
  const loading = props.loading
  const products = props.products
  const isProductServiceAvailable = props.isProductServiceAvailable
  const catalogDescription = props.catalogDescription
  const searchValues = props.searchValues
  // Functions
  const onSelectedProduct = props.onSelectedProduct
  const onChangeProductFilter = props.onChangeProductFilter
  const setTagFilter = props.setTagFilter

  // Variables
  let filterProducts = []
  let recommendedUseFilterList = []

  // Distinct to display in Tags filter
  recommendedUseFilterList = [...new Set(products.map((x) => x.recommendedUseCase))]
  // Tags filters
  for (const fieldToFilter in searchFilter) {
    const filter = searchFilter[fieldToFilter]
    if (filter.length === 0 || filter.some((x) => x === 'All')) {
      Array.prototype.push.apply(filterProducts, products)
    } else {
      const productFilter = products.filter((item) => filter.some((x) => x === item[fieldToFilter]))
      Array.prototype.push.apply(filterProducts, productFilter)
    }
  }

  if (filterProducts.length === 0) {
    filterProducts = products
  }

  if (productTextFilter) {
    filterProducts = filterProducts.filter((item) =>
      item.familyDisplayName.toLowerCase().includes(productTextFilter.toLowerCase())
    )
  }

  let productsFinal = []
  if (filterProducts.length > 0) {
    productsFinal = filterProducts.reduce((acc, current) => {
      if (!acc.find((item) => item.familyDisplayName === current.familyDisplayName)) {
        acc.push(current)
      }

      return acc
    }, [])
  } else {
    productsFinal = filterProducts
  }

  return (
    <>
      <ProductFilter
        searchValues={searchValues}
        setTagFilter={setTagFilter}
        productsCount={products.length}
        onChangeSearchBox={onChangeProductFilter}
        availableFilters={recommendedUseFilterList}
        filterField="recommendedUseCase"
        title="Available hardware"
      />
      {catalogDescription && (
        <div className="section">
          <p>{catalogDescription}</p>
        </div>
      )}
      <>
        {loading ? (
          <Spinner />
        ) : !isProductServiceAvailable || products.length === 0 ? (
          <ProductCatalogueUnavailable />
        ) : (
          <>
            <div className="section flex-xs-column flex-md-row gap-s8">
              <ProductList onSelectedProduct={onSelectedProduct} products={productsFinal} />
            </div>
          </>
        )}
      </>
    </>
  )
}

export default HardwareCatalog
