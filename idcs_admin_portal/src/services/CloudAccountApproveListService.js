import { AxiosInstance } from '../utility/axios/AxiosInstance'
import idcConfig from '../config/configurator'

class CloudAccountApproveListService {
  async getCloudAccountsApproveList() {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/cloudaccounts/approvelist`
    return AxiosInstance.get(route)
  }

  async createCloudAccountsApproveList(data) {
    const payload = JSON.stringify({
      status: data.status,
      enableStorage: data.enableStorage,
      maxclusterilb_override: data.maxclusterilb_override,
      maxclusterng_override: data.maxclusterng_override,
      maxclusters_override: data.maxclusters_override,
      maxclustervm_override: data.maxclustervm_override,
      maxnodegroupvm_override: data.maxnodegroupvm_override,
      iksadminkey: data.iksadminkey ?? ''
    })

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/cloudaccounts/${data.account}/approvelist`
    return AxiosInstance.post(route, payload)
  }

  async updateCloudAccountsApproveList(data) {
    const payload = JSON.stringify({
      status: data.status,
      enableStorage: data.enableStorage,
      maxclusterilb_override: data.maxclusterilb_override,
      maxclusterng_override: data.maxclusterng_override,
      maxclusters_override: data.maxclusters_override,
      maxclustervm_override: data.maxclustervm_override,
      maxnodegroupvm_override: data.maxnodegroupvm_override,
      iksadminkey: data.iksadminkey ?? ''
    })

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/iks/admin/cloudaccounts/${data.account}/approvelist`
    return AxiosInstance.put(route, payload)
  }
}

export default new CloudAccountApproveListService()
