// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import ProductCatalogueUnavailable from '../../../pages/error/ProductCatalogueUnavailable'
import LearningLabsList from './LearningLabsList'
import ComingSoonBanner from '../../../utils/comingSoonBanner/ComingSoonBanner'
import ProductFilter from '../../../utils/productFilter/ProductFilter'

import idcConfig from '../../../config/configurator'
import Spinner from '../../../utils/spinner/Spinner'

interface LearningLabsCatalogProps {
  loading: boolean
  isAvailable: boolean
  comingMessage: string
  learningLabsList: any[]
  setTagFilter: (field: string, filter: string[]) => void
  textFilter: string
  setTextFilter: (textFilter: string) => void
  searchFilter: Record<string, string[]>
}

const LearningLabsCatalog: React.FC<LearningLabsCatalogProps> = (props): JSX.Element => {
  const loading = props.loading
  const isAvailable = props.isAvailable
  const comingMessage = props.comingMessage
  const learningLabsList = props.learningLabsList
  const setTagFilter = props.setTagFilter
  const textFilter = props.textFilter
  const setTextFilter = props.setTextFilter
  const searchFilter = props.searchFilter

  // variables
  let filteredLearningLabsList: any[] = []
  let filteredLearningLabsFamilyList = []

  // Distinct to display in Tags filter
  filteredLearningLabsFamilyList = Array.from(new Set(learningLabsList.map((x) => x.familyDisplayName)))

  // Tags filters
  for (const fieldToFilter in searchFilter) {
    const filter = searchFilter[fieldToFilter]
    if (filter.length === 0 || filter.some((x) => x === 'All')) {
      Array.prototype.push.apply(filteredLearningLabsList, learningLabsList)
    } else {
      const productFilter = learningLabsList.filter((item) => filter.some((x) => x === item[fieldToFilter]))
      Array.prototype.push.apply(filteredLearningLabsList, productFilter)
    }
  }

  if (filteredLearningLabsList.length === 0) {
    filteredLearningLabsList = learningLabsList
  }

  if (textFilter) {
    filteredLearningLabsList = filteredLearningLabsList.filter(
      (item: any) =>
        item.familyDisplayName.toLowerCase().includes(textFilter.toLowerCase()) ||
        item.displayCatalogDesc.toLowerCase().includes(textFilter.toLowerCase()) ||
        item.displayName.toLowerCase().includes(textFilter.toLowerCase())
    )
  }

  return (
    <>
      <ProductFilter
        setTagFilter={setTagFilter}
        productsCount={learningLabsList.length}
        onChangeSearchBox={setTextFilter}
        availableFilters={filteredLearningLabsFamilyList}
        filterField="familyDisplayName"
        title="Available Labs"
      />
      <>
        {loading ? (
          <Spinner />
        ) : (
          <>
            {isAvailable ? (
              learningLabsList.length === 0 ? (
                <ProductCatalogueUnavailable />
              ) : (
                <>
                  <div className="section">
                    <a
                      className="link"
                      rel="noreferrer"
                      target="_blank"
                      aria-label="Disclaimer for using models"
                      href={idcConfig.REACT_APP_LEARNING_LABS_DISCLAIMER}
                    ></a>
                  </div>
                  <div className="section flex-xs-column flex-md-row gap-s8">
                    <LearningLabsList learningLabsList={filteredLearningLabsList} />
                  </div>
                </>
              )
            ) : (
              <ComingSoonBanner message={comingMessage} />
            )}
          </>
        )}
      </>
    </>
  )
}

export default LearningLabsCatalog
