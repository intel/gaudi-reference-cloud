// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React from 'react'
import ClusterHomePage from '../../components/cluster/clusterHomePage/ClusterHomePage'
import idcConfig from '../../config/configurator'

const ClusterHomePageContainer = () => {
  const state = {
    steps: [
      {
        title: 'Create Cluster',
        description: `${idcConfig.REACT_APP_COMPANY_SHORT_NAME} Kubernetes Service will deploy a highly available kubernetes control plane`
      },
      {
        title: 'Add worker node groups',
        description: `${idcConfig.REACT_APP_COMPANY_SHORT_NAME} Kubernetes Service managed data plane uses Intel Max GPU and Intel Habana Accelerators in its managed node groups`
      },
      {
        title: `Connect to ${idcConfig.REACT_APP_COMPANY_SHORT_NAME} Kubernetes Service`,
        description: 'Use kubeconfig to connect to your clusters'
      },
      {
        title: 'Deploy your services',
        description: 'Run GPU-accelerated kubernetes workloads at scale'
      }
    ]
  }
  return <ClusterHomePage state={state} />
}

export default ClusterHomePageContainer
