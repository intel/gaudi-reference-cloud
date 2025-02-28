// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { AxiosInstance } from '../utils/AxiosInstance'
import useUserStore from '../store/userStore/UserStore'
import idcConfig, { isFeatureFlagEnable, appFeatureFlags } from '../config/configurator'

class ClusterService {
  async getAllClustersDataStatus() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters`
    return AxiosInstance.get(route)
  }

  async getClusterData(clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}`
    return AxiosInstance.get(route)
  }

  async createClusterFunc(payload) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters`

    return AxiosInstance.post(route, payload)
  }

  async putClusterMetadata(form, clusterInfo) {
    const uuid = clusterInfo.uuid
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const payload = {
      name: clusterInfo.name,
      tags: form.clusterTags.options,
      annotations: form.clusterAnnotations.options
    }

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${uuid}`
    return AxiosInstance.put(route, payload)
  }

  // Get Kubeconfig cluster
  async getK8sVersion() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/metadata/k8sversions`
    return AxiosInstance.get(route)
  }

  async getRuntime() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/metadata/runtimes`
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

  // Delete a specific cluster
  async deleteCluster(clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}`
    return AxiosInstance.delete(route, { data: {} })
  }

  // Upgrade a specific cluster to a new version
  async upgradeCluster(form, clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const payload = {
      k8sversionname: form.k8sversionname
    }

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/upgrade`
    return AxiosInstance.post(route, payload)
  }

  async upgradeNodeGroup(nodegroupId, clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/nodegroups/${nodegroupId}/upgrade`
    return AxiosInstance.post(route, {})
  }

  async getAllManagedNodeGroupData(clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/nodegroups?nodes=true`
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

  async getManagedNodeGroupData(clusterId, nodegroup) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/nodegroups/${nodegroup}`
    return AxiosInstance.get(route)
  }

  async createStorage(payload, clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/storage`
    return AxiosInstance.post(route, payload)
  }

  async updateStorage(payload, clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/storage`
    return AxiosInstance.put(route, payload)
  }

  async createNodeGroup(form, clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const payload = {
      count: form.count,
      vnets: form.vnets,
      instancetypeid: form.instanceTypeId,
      instanceType: form.instanceType,
      userdataurl: form.userdataurl,
      name: form.name,
      description: '',
      tags: form.tags,
      sshkeyname: form.sshkeyname
    }

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/nodegroups`
    return AxiosInstance.post(route, payload)
  }

  async deleteNodeGroup(clusterId, nodegroupId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/nodegroups/${nodegroupId}`
    return AxiosInstance.delete(route, { data: {} })
  }

  async updateNodeGroupNodeCount(nodeCount, clusterId, nodegroupId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/nodegroups/${nodegroupId}`
    const payload = {
      count: nodeCount
    }
    return AxiosInstance.put(route, payload)
  }

  async getInstanceTypes() {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/metadata/instancetypes`

    return AxiosInstance.get(route)
  }

  // Load Balancer
  async getLoadBalancersData(clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/vips`

    return AxiosInstance.get(route)
  }

  async createLoadBalancer(form, clusterId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const payload = {
      name: form.loadbalancerName.value,
      description: '',
      port: form.loadbalancerPort.value,
      viptype: form.loadbalancerType.value
    }

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/vips`
    return AxiosInstance.post(route, payload)
  }

  async deleteLoadBalancer(clusterId, vipId) {
    const cloudAccountNumber = useUserStore.getState().user.cloudAccountNumber
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/cloudaccounts/${cloudAccountNumber}/iks/clusters/${clusterId}/vips/${vipId}`
    return AxiosInstance.delete(route, { data: {} })
  }
}

export default new ClusterService()
