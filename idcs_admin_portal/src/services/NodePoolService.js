import { AxiosInstance } from '../utility/axios/AxiosInstance'
import idcConfig from '../config/configurator'

class NodePoolService {
  getNodesStates(nodeId) {
    let query = ''
    if (nodeId && nodeId !== '') {
      query = `?nodeId=${nodeId}`
    }

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/fleetadmin/nodes/instancetypestats${query}`
    return AxiosInstance.get(route)
  }

  getNodes(poolId) {
    let query = ''
    if (poolId && poolId !== '') {
      query = `?poolId=${poolId}`
    }

    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/fleetadmin/nodes${query}`
    return AxiosInstance.get(route)
  }

  editNode(nodeId, payload) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/fleetadmin/nodes/${nodeId}`
    return AxiosInstance.put(route, payload)
  }

  getPools() {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/fleetadmin/computenodepools`
    return AxiosInstance.get(route)
  }

  createEditPool(poolId, payload) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/fleetadmin/computenodepools/${poolId}`
    return AxiosInstance.put(route, payload)
  }

  getPoolCloudAccounts(poolId) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/fleetadmin/computenodepools/${poolId}/cloudaccounts`
    return AxiosInstance.get(route)
  }

  addCloudAccountsToPool(poolId, payload) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/fleetadmin/computenodepools/${poolId}/cloudaccounts`
    return AxiosInstance.post(route, payload)
  }

  removeCloudAccounts(poolId, accountId) {
    const route = `${idcConfig.REACT_APP_API_REGIONAL_SERVICE}/fleetadmin/computenodepools/${poolId}/cloudaccounts/${accountId}`
    return AxiosInstance.delete(route, { data: {} })
  }
}

export default new NodePoolService()
