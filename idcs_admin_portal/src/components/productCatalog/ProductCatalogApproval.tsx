// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import React from 'react'
import { Button } from 'react-bootstrap'

const ProductCatalogApproval = (props: any): JSX.Element => {
  const onCancel = props.onCancel
  return <>
    <div className="section">
      <Button variant="link" className="p-s0" onClick={() => onCancel()}>
        ⟵ Back to Products
      </Button>
    </div>
    <div className="section">
      <h2 intc-id="maintitle">Approvals</h2>
    </div>
  </>
}

export default ProductCatalogApproval
