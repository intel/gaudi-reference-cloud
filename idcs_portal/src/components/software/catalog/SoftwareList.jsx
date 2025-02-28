// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useMemo } from 'react'
import ProductCard from '../../../utils/productCard/ProductCard'
import LineDivider from '../../../utils/lineDivider/LineDivider'
import PublicService from '../../../services/PublicService'
import { IDCVendorFamilies } from '../../../utils/Enums'

const SoftwareList = ({ softwareList }) => {
  const softwareByFamilyDisplayName = useMemo(() => {
    const familyDisplayNameList = [...new Set(softwareList.map((x) => x.familyDisplayName))]
    const softwareArrays = []
    familyDisplayNameList.forEach((familyDisplayName) => {
      const softwareByFamily = softwareList.filter((software) => software.familyDisplayName === familyDisplayName)
      softwareArrays.push(softwareByFamily)
    })
    return softwareArrays
  }, [softwareList])

  function getSoftware(products) {
    return (
      <div className="d-flex flex-column gap-s8">
        <div className="d-flex flex-xs-column flex-md-row align-items-start gap-s5">
          <div className="d-flex flex-column align-items-start gap-s4 w-100">
            <h3>{products[0].familyDisplayName}</h3>
            <p className="mb-0">
              {PublicService.getCatalogDescription(IDCVendorFamilies.Software, products[0].familyDisplayName)}
            </p>
          </div>
        </div>
        <div className="row g-s8">
          {products.map((product, index) => (
            <div key={index} className="col-xs-12 col-md-6 col-xl-4">
              <ProductCard
                intc-id={`btn-software-${product.displayName}`}
                title={product.displayName}
                description={product.displayCatalogDesc}
                href={`/software/d/${product.id}`}
                pricing={product.rate > 0 && product.usageUnit ? `$${product.rate} ${product.usageUnit}` : null}
              />
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <>
      <div className="d-flex flex-column w-100 gap-s8">
        {softwareByFamilyDisplayName.map((productList, index) => (
          <React.Fragment key={index}>
            {getSoftware(productList)}
            {index !== softwareByFamilyDisplayName.length - 1 && <LineDivider horizontal />}
          </React.Fragment>
        ))}
      </div>
    </>
  )
}

export default SoftwareList
