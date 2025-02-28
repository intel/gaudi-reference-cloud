// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { ReactComponent as SeekrLogoText } from '../../../assets/images/SeekrLogoText.svg'
import ProductCatalogueUnavailable from '../../../pages/error/ProductCatalogueUnavailable'
import SoftwareList from './SoftwareList'
import ComingSoonBanner from '../../../utils/comingSoonBanner/ComingSoonBanner'
import ProductFilter from '../../../utils/productFilter/ProductFilter'
import FeaturedProductCard from '../../../utils/productCard/FeaturedProductCard'
import Spinner from '../../../utils/spinner/Spinner'

const SoftwareCatalog = (props) => {
  const loading = props.loading
  const softwareList = props.softwareList
  const comingMessage = props.comingMessage
  const isAvailable = props.isAvailable
  const setTagFilter = props.setTagFilter
  const softwareTextFilter = props.softwareTextFilter
  const setSoftwareTextFilter = props.setSoftwareTextFilter
  const searchFilter = props.searchFilter

  const seeKrProduct = softwareList?.find((x) => x.name === 'sw-seekr-flow')

  // Variables
  let filterSoftware = []
  let familyFilterList = []

  // Distinct to display in Tags filter
  familyFilterList = [...new Set(softwareList.map((x) => x.familyDisplayName))]

  // Tags filters
  for (const fieldToFilter in searchFilter) {
    const filter = searchFilter[fieldToFilter]
    if (filter.length === 0 || filter.some((x) => x === 'All')) {
      Array.prototype.push.apply(filterSoftware, softwareList)
    } else {
      const productFilter = softwareList.filter((item) => filter.some((x) => x === item[fieldToFilter]))
      Array.prototype.push.apply(filterSoftware, productFilter)
    }
  }

  if (filterSoftware.length === 0) {
    filterSoftware = softwareList
  }

  if (softwareTextFilter) {
    filterSoftware = filterSoftware.filter(
      (item) =>
        item.familyDisplayName.toLowerCase().includes(softwareTextFilter.toLowerCase()) ||
        item.displayCatalogDesc.toLowerCase().includes(softwareTextFilter.toLowerCase()) ||
        item.displayName.toLowerCase().includes(softwareTextFilter.toLowerCase())
    )
  }

  let content = null

  if (!isAvailable) {
    return <ComingSoonBanner message={comingMessage} />
  } else {
    content = (
      <>
        {loading ? (
          <Spinner />
        ) : (
          <>
            {softwareList.length === 0 ? (
              <ProductCatalogueUnavailable />
            ) : (
              <>
                {seeKrProduct && (
                  <div className="section">
                    <FeaturedProductCard
                      title={seeKrProduct.displayCatalogDesc}
                      topImage={<SeekrLogoText />}
                      description={seeKrProduct.overview}
                      learnMoreHref={`/software/d/${seeKrProduct.id}`}
                    />
                  </div>
                )}
                <div className="section flex-xs-column flex-md-row gap-s8">
                  <SoftwareList softwareList={filterSoftware} />
                </div>
              </>
            )}
          </>
        )}
      </>
    )
  }

  return (
    <>
      <ProductFilter
        setTagFilter={setTagFilter}
        productsCount={softwareList.length}
        onChangeSearchBox={setSoftwareTextFilter}
        availableFilters={familyFilterList}
        filterField="familyDisplayName"
        title="Available software"
      />
      {content}
    </>
  )
}

export default SoftwareCatalog
