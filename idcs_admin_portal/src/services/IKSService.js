import { AxiosInstance } from '../utility/axios/AxiosInstance'
import idcConfig from '../config/configurator'

class IKSService {
  // Cluster Service Endpoints
  async getAllClustersDataStatus() {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/clusters`
    return AxiosInstance.get(route)
  }

  async getClusterData(clusterId) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/clusters/${clusterId}`
    return AxiosInstance.get(route)
  }

  async getSSHKeys(clusterId) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/clusters/${clusterId}/sshkeys`
    return AxiosInstance.get(route)
  }

  async getSecurityDetails(clusterId) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/clusters/${clusterId}/security`
    return AxiosInstance.get(route)
  }

  async getIlbDetails(clusterId) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/clusters/${clusterId}/ilbs`
    return AxiosInstance.get(route)
  }

  async upgradeNodeGroup(k8sVersionName, nodegroupId, clusterId) {
    const payload = JSON.stringify({
      k8sversionname: k8sVersionName
    })

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/clusters/${clusterId}/nodegroups/${nodegroupId}/upgrade`
    return AxiosInstance.post(route, payload)
  }

  async createLoadBalancer(form, clusterId) {
    const payload = JSON.stringify({
      name: form.name,
      description: form.description,
      port: form.port,
      viptype: form.vip
    })

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/clusters/${clusterId}/ilbs`
    return AxiosInstance.post(route, payload)
  }

  async deleteLoadBalancer(clusterId, lbid) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/clusters/${clusterId}/ilbs/${lbid}`
    return AxiosInstance.delete(route, { data: {} })
  }

  // IMIS Service Endpoints
  async getIMIS(id) {
    let route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/imis`

    if (id && id !== '') {
      route += `/${id}`
    }

    return AxiosInstance.get(route)
  }

  async createIMIS(data) {
    const payload = JSON.stringify(data)

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/imis`
    return AxiosInstance.post(route, payload)
  }

  async updateIMIS(id, data) {
    const payload = JSON.stringify(data)

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/imis/${id}`
    return AxiosInstance.put(route, payload)
  }

  async deleteIMIS(id, payload) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/imis/${id}?iksadminkey=${payload.iksadminkey}`
    return AxiosInstance.delete(route, { data: {} })
  }

  async getIMISInfo() {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/imis/info`
    return AxiosInstance.get(route)
  }

  async updateIMISK8s(id, data) {
    const payload = JSON.stringify(data)

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/imis/${id}/k8s`
    return AxiosInstance.put(route, payload)
  }

  async authenticateIMIS(data) {
    const payload = JSON.stringify(data)

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/authenticate`
    return AxiosInstance.post(route, payload)
  }

  // Instance Types Service Endpoints
  async getInstanceTypes(id) {
    let route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/instancetypes`

    if (id && id !== '') {
      route += `/${id}`
    }

    return AxiosInstance.get(route)
  }

  async createInstanceTypes(data) {
    const payload = JSON.stringify(data)

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/instancetypes`
    return AxiosInstance.post(route, payload)
  }

  async updateInstanceTypes(id, data) {
    const payload = JSON.stringify(data)

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/instancetypes/${id}`
    return AxiosInstance.put(route, payload)
  }

  async deleteInstanceTypes(id, payload) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/instancetypes/${id}?iksadminkey=${payload.iksadminkey}`

    return AxiosInstance.delete(route, { data: {} })
  }

  async getInstanceTypesInfo() {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/instancetypes/info`

    return AxiosInstance.get(route)
  }

  async updateInstanceTypesK8s(id, data) {
    const payload = JSON.stringify(data)

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/instancetypes/${id}/k8s`
    return AxiosInstance.put(route, payload)
  }

  // K8S Service Endpoints
  async getK8s() {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/k8sversions`
    return AxiosInstance.get(route)
  }

  // Kube score endpoints
  async getInsightVersions(component, versionId) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/securityinsights/component/${component}/releases/${versionId}`
    return AxiosInstance.get(route)
  }

  async getSbomItems(component, versionId) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/securityinsights/component/${component}/releases/${versionId}/sbom`
    return AxiosInstance.get(route)
  }

  async getVulnerabilitiesItems(component, versionId) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/securityinsights/component/${component}/releases/${versionId}/vulnerabilities`
    return AxiosInstance.get(route)
  }

  async getSummaryItems(component, versionId) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/securityinsights/component/${component}/releases/${versionId}/summary`
    return AxiosInstance.get(route)
  }

  async getComponents() {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/securityinsights/components`
    return AxiosInstance.post(route, {})
  }
}

export default new IKSService()
