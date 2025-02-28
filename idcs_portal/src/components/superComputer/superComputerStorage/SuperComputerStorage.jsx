// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { BsDatabase } from 'react-icons/bs'
import { Button } from 'react-bootstrap'
import LabelValuePair from '../../../utils/labelValuePair/LabelValuePair'

const SuperComputerStorage = ({ storageItems, isClusterInActiveState, goToAddStorage }) => {
  let content = null

  if (storageItems.length > 0) {
    const columnSize = 12 / 3
    content = (
      <div className="section">
        <div className="row">
          {storageItems.map((item, index) => (
            <LabelValuePair className={`col-md-${columnSize}`} label={item.label} key={index}>
              {item.value}
            </LabelValuePair>
          ))}
        </div>
      </div>
    )
  } else {
    content = (
      <div className="section align-self-center align-items-center" intc-id="data-view-empty">
        <h5 className="h4">No Storage found</h5>
        {isClusterInActiveState ? (
          <span className="add-break-line lead">The Cluster is being updated, please wait to request storage</span>
        ) : null}
        <Button
          intc-id="btn-iksMyClusters-addVip"
          disabled={isClusterInActiveState}
          data-wap_ref="btn-iksMyClusters-addVip"
          variant="outline-primary"
          onClick={goToAddStorage}
        >
          <BsDatabase /> Add Storage
        </Button>
      </div>
    )
  }
  return <>{content}</>
}

export default SuperComputerStorage
