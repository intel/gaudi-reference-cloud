// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Button } from 'react-bootstrap'
import CustomInput from '../../utility/customInput/CustomInput'
import GridPagination from '../../utility/gridPagination/gridPagination'
import SearchBox from '../../utility/searchBox/SearchBox'
import { NavLink } from 'react-router-dom'

const ProductCatalogView = (props: any): JSX.Element => {
  const onCancel = props.onCancel
  const state = props.state
  const mainTitle = state.mainTitle
  const form = state.form
  const onChangeInput = props.onChangeInput
  const produtColumns = props.produtColumns
  const productItems = props.productItems
  const emptyGrid = props.emptyGrid
  const filterText = props.filterText
  const setFilter = props.setFilter
  const isPageReady = props.isPageReady
  const formElements = []

  // *****
  // Functions
  // *****
  const buildCustomInput = (element: any, key: number): JSX.Element => {
    const response = <CustomInput
      key={key}
      type={element.configInput.type}
      fieldSize={element.configInput.fieldSize}
      placeholder={element.configInput.placeholder}
      isRequired={element.configInput.validationRules.isRequired}
      label={element.configInput.label}
      value={element.configInput.value}
      onChanged={(event: any) => onChangeInput(event, element.id, element.idParent, element.nodeIndex)}
      isValid={element.configInput.isValid}
      isTouched={element.configInput.isTouched}
      helperMessage={element.configInput.helperMessage}
      isReadOnly={element.configInput.isReadOnly}
      options={element.configInput.options}
      validationMessage={element.configInput.validationMessage}
      refreshButton={element.configInput.refreshButton}
      extraButton={element.configInput.extraButton}
      selectAllButton={element.configInput.selectAllButton}
      labelButton={element.configInput.labelButton}
      emptyOptionsMessage={element.configInput.emptyOptionsMessage}
    />
    return response
  }

  const getFilterCheck = (filterValue: string, data: any): boolean => {
    filterValue = filterValue.toLowerCase()
    return (
      data?.name?.toString().toLowerCase().indexOf(filterValue) > -1 ||
      data?.description?.toString().toLowerCase().indexOf(filterValue) > -1
    )
  }

  for (const key in form) {
    const formItem = {
      ...form[key]
    }

    formElements.push({
      id: key,
      configInput: formItem
    })
  }

  let gridItems = productItems

  if (filterText !== '' && productItems) {
    const input = filterText.toLowerCase()
    gridItems = productItems.filter((item: any) => getFilterCheck(input, item))
  } else {
    gridItems = productItems
  }

  let content = <div className="section">
    <div className="col-12 row mt-s2">
      <div className="spinner-border text-primary center"></div>
    </div>
  </div>

  if (isPageReady) {
    content = <>
      <div className="section">
        <Button variant="link" className="p-s0" onClick={() => onCancel()}>
          ⟵ Back to Home
        </Button>
      </div>
      <div className="section">
        <div className="filter flex-wrap p-0">
          <h2 intc-id="maintitle">{mainTitle}</h2>
        </div>
      </div>
      <div className="section">
        <div className="filter flex-wrap p-0">
          <div className='flex-fill'>
            {formElements.map((element, index) => buildCustomInput(element, index))}
          </div>
          <SearchBox
            intc-id="filterProducts"
            value={filterText}
            onChange={setFilter}
            placeholder="Filter products..."
            aria-label="Type to filter products..."
          />
          <NavLink intc-id="btn-create-product" to="/products/create" className="btn btn-primary" aria-label='Go to create product page'>
              Create product
          </NavLink>
          <NavLink intc-id="btn-approvsl-product" to="/products/approvals" className="btn btn-outline-primary" aria-label='Go to product approvals page'>
              Approvals
          </NavLink>
        </div>
      </div>
      <div className="section">
        <GridPagination data={gridItems} columns={produtColumns} loading={false} emptyGrid={emptyGrid} />
      </div>
    </>
  }

  return <>{content}</>
}

export default ProductCatalogView
