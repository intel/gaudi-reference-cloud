// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { AxiosInstance } from '../utils/AxiosInstance'
import useUserStore from '../store/userStore/UserStore'
import idcConfig, { isFeatureFlagEnable, appFeatureFlags } from '../config/configurator'

class SuperComputerService {
  async getClusters() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters`
    return AxiosInstance.get(route)
  }

  async getClusterById(idCluster) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${idCluster}`
    return AxiosInstance.get(route)
  }

  async getClusterByName(clusterName) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters?metadata.name=${clusterName}`
    return AxiosInstance.get(route)
  }

  async getKubeconfigFile(clusterId, readOnly = true) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    let route = ''
    if (isFeatureFlagEnable(appFeatureFlags.REACT_APP_FEATURE_IKS_KUBE_CONFIG)) {
      route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/kubeconfig?readonly=${readOnly}`
    } else {
      route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/kubeconfig`
    }
    return AxiosInstance.get(route)
  }

  async deleteCluster(clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}`
    return AxiosInstance.delete(route, { data: {} })
  }

  async upgradeCluster(form, clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const payload = {
      k8sversionname: form.k8sversionname
    }

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/upgrade`
    return AxiosInstance.post(route, payload)
  }

  async getAllManagedNodeGroupData(clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/nodegroups?nodes=true`
    return AxiosInstance.get(route)
  }

  async updateNodeGroupNodeCount(nodeCount, clusterId, nodegroupId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/nodegroups/${nodegroupId}`
    const payload = {
      count: nodeCount
    }
    return AxiosInstance.put(route, payload)
  }

  async deleteNodeGroup(clusterId, nodegroupId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/nodegroups/${nodegroupId}`
    return AxiosInstance.delete(route, { data: {} })
  }

  async upgradeNodeGroup(clusterId, nodegroupId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/nodegroups/${nodegroupId}/upgrade`
    return AxiosInstance.post(route, {})
  }

  async getInstanceTypes() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/metadata/instancetypes`

    return AxiosInstance.get(route)
  }

  async createNodeGroup(payload, clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/nodegroups`
    return AxiosInstance.post(route, payload)
  }

  async deleteLoadBalancer(clusterId, vipId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/vips/${vipId}`
    return AxiosInstance.delete(route, { data: {} })
  }

  async createLoadBalancer(payload, clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/vips`
    return AxiosInstance.post(route, payload)
  }

  async createStorage(payload, clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/storage`
    return AxiosInstance.post(route, payload)
  }

  async createSuperCluster(payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/supercompute/clusters`

    return AxiosInstance.post(route, payload)
  }

  async getRuntime() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/metadata/runtimes`
    return AxiosInstance.get(route)
  }

  async getSecurityRules(clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/security`
    return AxiosInstance.get(route)
  }

  async putSecurityRules(clusterId, payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/security`
    return AxiosInstance.put(route, payload)
  }

  async deleteSecurityRule(clusterId, vipId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/security/${vipId}`
    return AxiosInstance.delete(route, { data: {} })
  }
}

export default new SuperComputerService()
