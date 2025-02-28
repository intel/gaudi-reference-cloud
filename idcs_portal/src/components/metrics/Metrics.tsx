// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import EmptyView from '../../utils/emptyView/EmptyView'
import MetricsGraphsContainer from '../../containers/metrics/MetricsGraphsContainer'
import Spinner from '../../utils/spinner/Spinner'
import ClusterGraphsContainer from '../../containers/metrics/ClusterGraphsContainer'
import InstanceGroupBuildContainer from '../../containers/metrics/InstanceGroupBuildContainer'

const Loader = (): JSX.Element => {
  return <Spinner />
}

const Metrics = (props: any): JSX.Element => {
  const loading = props.loading
  const items = props.items
  const emptyGrid = props.emptyGrid
  const filteredItems = props.filteredItems
  const mainTitle = props.mainTitle
  const itemType = props.itemType

  const emptyView = (): JSX.Element => {
    return (
      <div className="section">
        <EmptyView title={emptyGrid.title} subTitle={emptyGrid.subTitle} action={emptyGrid.action} />
      </div>
    )
  }

  const showData = (): JSX.Element => {
    if (items.length === 0 || filteredItems.length === 0) {
      return emptyView()
    } else if (itemType === 'cluster') {
      return <ClusterGraphsContainer clusters={filteredItems} showClusters={true} />
    } else if (itemType === 'instance-groups') {
      return <InstanceGroupBuildContainer instances={filteredItems} />
    } else {
      return <MetricsGraphsContainer instances={filteredItems} showInstances={true} />
    }
  }

  return (
    <>
      <div className="section">
        <h2 intc-id="metricsTitle">{mainTitle}</h2>
      </div>

      {loading ? <Loader /> : showData()}
    </>
  )
}

export default Metrics
