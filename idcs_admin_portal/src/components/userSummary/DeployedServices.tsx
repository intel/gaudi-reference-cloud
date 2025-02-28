// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import GridPagination from '../../utility/gridPagination/gridPagination'
import LineDivider from '../../utility/lineDivider/LineDivider'

interface DeployedServicesProps {
  instances: any[]
  instancesColumns: any
  loadingInstances: boolean
  instancesEmptyGrid: any
  instanceGroups: any[]
  instanceGroupsColumns: any
  loadingInstanceGroup: boolean
  instanceGroupsEmptyGrid: any
  clusters: any[]
  scClusters: any[]
  iksColumns: any
  loadingClusters: boolean
  clustersEmptyGrid: any
}

const DeployedServices: React.FC<DeployedServicesProps> = (props): JSX.Element => {
  // instances table
  const instances = props.instances
  const instancesColumns = props.instancesColumns
  const loadingInstances = props.loadingInstances
  const instancesEmptyGrid = props.instancesEmptyGrid

  // instances groups table
  const instanceGroups = props.instanceGroups
  const instanceGroupsColumns = props.instanceGroupsColumns
  const loadingInstanceGroup = props.loadingInstanceGroup
  const instanceGroupsEmptyGrid = props.instanceGroupsEmptyGrid

  // clusters and sc clusters tables
  const clusters = props.clusters
  const scClusters = props.scClusters
  const iksColumns = props.iksColumns
  const loadingClusters = props.loadingClusters
  const clustersEmptyGrid = props.clustersEmptyGrid

  return (
    <>
      <div className="section" intc-id="instancesGrid">
        {<span className="h4">Instances</span>}
        <GridPagination
          data={instances}
          columns={instancesColumns}
          loading={loadingInstances}
          emptyGrid={instancesEmptyGrid}
        />
      </div>
      <LineDivider horizontal/>
      <div className="section" intc-id="instanceGroupsGrid">
        {<span className="h4">Instance Groups</span>}
        <GridPagination
          data={instanceGroups}
          columns={instanceGroupsColumns}
          loading={loadingInstanceGroup}
          emptyGrid={instanceGroupsEmptyGrid}
        />
      </div>
      <LineDivider horizontal/>
      <div className="section" intc-id="clusterssGrid">
        {<span className="h4">IKS Clusters</span>}
        <GridPagination data={clusters} columns={iksColumns} loading={loadingClusters} emptyGrid={clustersEmptyGrid} />
      </div>
      <LineDivider horizontal/>
      <div className="section" intc-id="scClusterssGrid">
        {<span className="h4">Super Computer Clusters</span>}
        <GridPagination
          data={scClusters}
          columns={iksColumns}
          loading={loadingClusters}
          emptyGrid={clustersEmptyGrid}
        />
      </div>
    </>
  )
}
export default DeployedServices
