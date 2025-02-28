// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import { Button } from 'react-bootstrap'
import LabelValuePair from '../../../utils/labelValuePair/LabelValuePair'

const ClusterStorage = ({
  storageItems = [],
  emptyStorage,
  reserveDetails,
  setAction,
  goToAddNodeGroup,
  nodegroupsInfo
}) => {
  let content = null
  const isClusterInActiveState = reserveDetails?.clusterstate === 'Active'
  let message = null
  if (!isClusterInActiveState) {
    message = 'The Cluster is being updated, please wait before adding a storage'
  } else {
    message = emptyStorage.subTitle
  }

  if (storageItems.length > 0) {
    content = (
      <>
        <Button
          intc-id="btn-iksMyClusters-editStorage"
          variant="outline-primary"
          disabled={!isClusterInActiveState}
          onClick={() => setAction({ id: 'editStorage' })}
        >
          Edit Storage
        </Button>
        <div className="row col-md-6">
          {storageItems.map((storage, index) => (
            <LabelValuePair className="col-md-4" label={storage.label} key={index}>
              <div className="d-flex flex-column text-wrap">{storage.value}</div>
            </LabelValuePair>
          ))}
        </div>
      </>
    )
  } else {
    content = (
      <div className="section align-self-center align-items-center" intc-id="data-view-empty">
        <h4>{emptyStorage.title}</h4>
        <p className="add-break-line lead">{message}</p>
        <Button
          intc-id="btn-iksMyClusters-addWorkerNodeGroup"
          variant="outline-primary"
          disabled={emptyStorage.action.disabled}
          onClick={() => setAction({ id: 'addStorage' })}
        >
          {emptyStorage.action.label}
        </Button>
      </div>
    )
  }

  return <>{content}</>
}

export default ClusterStorage
