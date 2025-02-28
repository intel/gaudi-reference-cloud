// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useState } from 'react'
import { BsCheck } from 'react-icons/bs'
import Button from 'react-bootstrap/Button'
import SearchBox from '../searchBox/SearchBox'

interface ProductFilterProps {
  title?: string
  filterField?: string
  searchValues?: string[]
  productsCount: number
  setTagFilter: (field: string, filter: string[]) => void
  onChangeSearchBox: (textFilter: string) => void
  availableFilters: string[]
  extraButton?: React.ReactNode
}

const ProductFilter: React.FC<ProductFilterProps> = ({
  title = '',
  filterField = '',
  productsCount = 0,
  searchValues,
  setTagFilter,
  onChangeSearchBox,
  availableFilters,
  extraButton
}): JSX.Element => {
  const [filterText, setFilterText] = useState('')
  const [usecaseFilter, setUsecaseFilter] = useState(['All'])
  const selectUseCase = (filter: string): void => {
    const shouldRemove = usecaseFilter.some((x) => x === filter)
    if (shouldRemove) {
      setUsecaseFilter([...usecaseFilter.filter((x) => x !== filter)])
      return
    }
    if (filter === 'All') {
      setUsecaseFilter(['All'])
    } else {
      setUsecaseFilter([...usecaseFilter.filter((x) => x !== 'All'), filter])
    }
  }

  useEffect(() => {
    if (searchValues) {
      setUsecaseFilter(searchValues)
    }
  }, [searchValues])

  useEffect(() => {
    setTagFilter(filterField, usecaseFilter)
  }, [usecaseFilter])

  availableFilters?.sort((a, b) => a.localeCompare(b))

  return (
    <>
      <div className="filter flex-wrap">
        <div className="d-flex flex-xs-column flex-sm-row gap-s6 align-items-sm-center flex-wrap">
          <h2>{`${title} (${productsCount})`}</h2>
          <div className="d-flex flex-row gap-s4 me-auto flex-wrap">
            <Button
              variant="tag"
              size="sm"
              aria-label="Toggle filter all products"
              className={usecaseFilter.some((x) => x === 'All') ? 'active' : ''}
              onClick={() => {
                selectUseCase('All')
              }}
            >
              {usecaseFilter.some((x) => x === 'All') && <BsCheck />}
              All
            </Button>
            {availableFilters.map((useCase, index) => (
              <Button
                variant="tag"
                size="sm"
                aria-label={`Toggle filter ${useCase} products`}
                key={index}
                className={usecaseFilter.some((x) => x === useCase) ? 'active' : ''}
                onClick={() => {
                  selectUseCase(useCase)
                }}
              >
                {usecaseFilter.some((x) => x === useCase) && <BsCheck />}
                {useCase}
              </Button>
            ))}
          </div>
        </div>
        <div className="d-flex flex-xs-column flex-sm-row gap-s6 flex-wrap">
          <div>
            <SearchBox
              intc-id="Filter-Text"
              placeholder="Type to search..."
              value={filterText}
              onChange={(e) => {
                setFilterText(e.target.value)
                onChangeSearchBox(e.target.value)
              }}
            />
          </div>
          {extraButton && <>{extraButton}</>}
        </div>
      </div>
    </>
  )
}

export default ProductFilter
