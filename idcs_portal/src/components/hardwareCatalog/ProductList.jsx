// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useMemo } from 'react'
import ProductCard from '../../utils/productCard/ProductCard'
import LineDivider from '../../utils/lineDivider/LineDivider'
import PublicService from '../../services/PublicService'
import { IDCVendorFamilies } from '../../utils/Enums'

const ProductList = ({ products, onSelectedProduct }) => {
  // Functions
  function compare(a, b) {
    if (a.recommendedUseCase < b.recommendedUseCase) {
      return -1
    }
    if (a.recommendedUseCase > b.recommendedUseCase) {
      return 1
    }
    return 0
  }

  const productsByRecommendedUseCase = useMemo(() => {
    if (products.length > 0) {
      products = products.sort(compare)
    }
    const recommendUseDistinct = [...new Set(products.map((x) => x.recommendedUseCase))]
    const productList = []
    recommendUseDistinct.forEach((recommendUse) => {
      const productsFilter = products.filter((item) => item.recommendedUseCase === recommendUse)
      productList.push(productsFilter)
    })
    return productList
  }, [products])

  function getProducts(products) {
    return (
      <div className="d-flex flex-column gap-s8">
        <div className="d-flex flex-xs-column flex-md-row align-items-start gap-s5">
          <div className="d-flex flex-column align-items-start gap-s4 w-100">
            <h3>{products[0].recommendedUseCase}</h3>
            <p className="mb-0">
              {PublicService.getCatalogDescription(IDCVendorFamilies.Compute, products[0].recommendedUseCase)}
            </p>
          </div>
        </div>
        <div className="row g-s8">
          {products.map((product, index) => (
            <div key={index} className="col-xs-12 col-md-6 col-xl-4">
              <ProductCard
                intc-id={`btn-hardwarecatalog-select ${product.familyDisplayName}`}
                title={product.familyDisplayName}
                description={product.familyDisplayDescription}
                onClick={(e) => onSelectedProduct(e, product)}
              />
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="d-flex flex-column w-100 gap-s8">
      {productsByRecommendedUseCase.map((productList, index) => (
        <React.Fragment key={index}>
          {getProducts(productList)}
          {index !== productsByRecommendedUseCase.length - 1 && <LineDivider horizontal />}
        </React.Fragment>
      ))}
    </div>
  )
}

export default ProductList
