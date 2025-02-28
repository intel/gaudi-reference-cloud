// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React from 'react'
import { useNavigate } from 'react-router'
import ProductCatalogApproval from '../../components/productCatalog/ProductCatalogApproval'

const ProductCatalogApprovalContainer = (): JSX.Element => {
  // *****
  // local state
  // *****
  const navigate = useNavigate()
  // *****
  // Hooks
  // *****

  // *****
  // functions
  // *****
  const onCancel = (): void => {
    navigate(-1)
  }
  return <ProductCatalogApproval onCancel={onCancel} />
}

export default ProductCatalogApprovalContainer
