// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React from 'react'
import EmptyView from '../../utility/emptyView/EmptyView'
import ProductCatalogCreate from './ProductCatalogCreate'
import { Button } from 'react-bootstrap'

const ProductCatalogEdit = (props: any): JSX.Element => {
  const onCancel = props.onCancel
  const loading = props.loading
  const state = props.state
  const submitModal = props.submitModal
  const onChangeInput = props.onChangeInput
  const onChangeDropdownMultiple = props.onChangeDropdownMultiple
  const selectAllButton = props.selectAllButton
  const onClickAddMetaData = props.onClickAddMetaData
  const onSubmit = props.onSubmit
  const editProduct = props.editProduct
  const isPageReady = props.isPageReady

  let content = <div className="section">
    <div className="col-12 row mt-s2">
      <div className="spinner-border text-primary center"></div>
    </div>
  </div>

  if (isPageReady) {
    if (editProduct) {
      content = <ProductCatalogCreate onCancel={onCancel}
        state={state}
        loading={loading}
        onChangeInput={onChangeInput}
        onChangeDropdownMultiple={onChangeDropdownMultiple}
        selectAllButton={selectAllButton}
        onClickAddMetaData={onClickAddMetaData}
        onSubmit={onSubmit}
        submitModal={submitModal}
        newMetaDataSets={editProduct?.metaDataSets || []} />
    } else {
      content = <>
        <div className="section">
          <Button variant="link" className="p-s0" onClick={() => onCancel()}>
            ⟵ Back to products
          </Button>
        </div>
        <div className='section'>
          <EmptyView title={'No product found'} subTitle={'Product name not found'} />
        </div>
      </>
    }
  }

  return <>
    {content}
  </>
}

export default ProductCatalogEdit
