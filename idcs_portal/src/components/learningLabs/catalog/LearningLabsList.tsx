// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useMemo } from 'react'
import LineDivider from '../../../utils/lineDivider/LineDivider'
import PublicService from '../../../services/PublicService'
import { IDCVendorFamilies } from '../../../utils/Enums'
import ProductCard from '../../../utils/productCard/ProductCard'

interface LearningLabsListProps {
  learningLabsList: any[]
}

const LearningLabsList: React.FC<LearningLabsListProps> = (props): JSX.Element => {
  const learningLabsList = props.learningLabsList

  function compare(a: any, b: any): number {
    if (a.familyDisplayName < b.familyDisplayName) {
      return -1
    }
    if (a.familyDisplayName > b.familyDisplayName) {
      return 1
    }
    return 0
  }

  const learningLabsListByDisplayName = useMemo(() => {
    let learningLabsSortedList: any[] = []
    if (learningLabsList.length > 0) {
      learningLabsSortedList = learningLabsList.sort(compare)
    }
    const familyDisplayNameList = Array.from(new Set(learningLabsSortedList.map((x: any) => x.familyDisplayName)))
    const newLearningLabsList: any[] = []
    familyDisplayNameList.forEach((familyDisplayName) => {
      const learningLabsByFamily = learningLabsSortedList.filter(
        (learningLabsElement: any) => learningLabsElement.familyDisplayName === familyDisplayName
      )
      newLearningLabsList.push(learningLabsByFamily)
    })
    return newLearningLabsList
  }, [learningLabsList])

  // Functions
  function getData(learningLabsList: any): JSX.Element {
    return (
      <div className="d-flex flex-column gap-s8">
        <div className="d-flex flex-xs-column flex-md-row align-items-start gap-s5">
          <div className="d-flex flex-column align-items-start gap-s4 w-100">
            <h3>{learningLabsList[0].familyDisplayName}</h3>
            <p className="mb-0">
              {PublicService.getCatalogDescription(IDCVendorFamilies.labs, learningLabsList[0].familyDisplayName)}
            </p>
          </div>
        </div>
        <div className="row g-s8">
          {learningLabsList.map((learningLabsProduct: any, index: number) => (
            <div key={index} className="col-xs-12 col-md-6 col-xl-4">
              <ProductCard
                intc-id={`btn-training-select ${learningLabsProduct.displayName}`}
                title={learningLabsProduct.displayName}
                description={learningLabsProduct.displayCatalogDesc}
                actionLabel="Launch"
                href={`/learning/labs/${String(learningLabsProduct?.launch)}`}
                learnMoreHref={`/learning/labs/${String(learningLabsProduct?.launch)}`}
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
        {learningLabsListByDisplayName.map((learningLabsList, index) => (
          <React.Fragment key={index}>
            {getData(learningLabsList)}
            {index !== learningLabsListByDisplayName.length - 1 && <LineDivider horizontal />}
          </React.Fragment>
        ))}
      </div>
    </>
  )
}

export default LearningLabsList
