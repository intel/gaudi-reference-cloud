// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useMemo } from 'react'
import ProductCard from '../../../utils/productCard/ProductCard'
import LineDivider from '../../../utils/lineDivider/LineDivider'
import { IDCVendorFamilies } from '../../../utils/Enums'
import PublicService from '../../../services/PublicService'
import './TrainingList.scss'

const TrainingList = (props) => {
  // props
  let trainings = props.trainings
  const onClickLaunch = props.onClickLaunch

  function compare(a, b) {
    return (
      PublicService.getCatalogOrder(IDCVendorFamilies.Training, a.familyDisplayName) -
      PublicService.getCatalogOrder(IDCVendorFamilies.Training, b.familyDisplayName)
    )
  }

  const trainingsByFamilyDisplayName = useMemo(() => {
    if (trainings.length > 0) {
      trainings = trainings.sort(compare)
    }
    const familyDisplayNameList = [...new Set(trainings.map((x) => x.familyDisplayName))]
    const trainingList = []
    familyDisplayNameList.forEach((familyDisplayName) => {
      const softwareByFamily = trainings.filter((training) => training.familyDisplayName === familyDisplayName)
      trainingList.push(softwareByFamily)
    })
    return trainingList
  }, [trainings])

  function getTrainings(products) {
    return (
      <div className="d-flex flex-column gap-s8">
        <div className="d-flex flex-xs-column flex-md-row align-items-start gap-s5">
          <div className="d-flex flex-column align-items-start gap-s4 w-100">
            <h3>{products[0].familyDisplayName}</h3>
            <p className="mb-0">
              {PublicService.getCatalogDescription(IDCVendorFamilies.Training, products[0].familyDisplayName)}
            </p>
          </div>
        </div>
        <div className="row g-s8">
          {products.map((product, index) => (
            <div key={index} className="col-xs-12 col-md-6 col-xl-4">
              <ProductCard
                intc-id={`btn-training-select ${product.displayName}`}
                title={product.displayName}
                description={product.displayCatalogDesc}
                onClick={(e) => onClickLaunch(e, product.id)}
                learnMoreHref={`/learning/notebooks/detail/${product.id}`}
                actionLabel="Launch"
                href="learning"
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
        {trainingsByFamilyDisplayName.map((productList, index) => (
          <React.Fragment key={index}>
            {getTrainings(productList)}
            {index !== trainingsByFamilyDisplayName.length - 1 && <LineDivider horizontal />}
          </React.Fragment>
        ))}
      </div>
    </>
  )
}

export default TrainingList
