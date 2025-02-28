import { create } from 'zustand'
import IKSService from '../../services/IKSService'
import moment from 'moment'

interface Clusters {
  account: string
  cpupgradeavailable: boolean
  cpupgradeversions: string[]
  k8supgradeavailable: boolean
  k8supgradeversions: string[]
  k8sversion: string
  name: string
  provider: string
  state: string
  uuid: string
  createddate: string
  clustertype: string
}

export interface Cluster {
  account: string
  addons: Addon[]
  backup: Backup[]
  certsexpiring: Certsexpiring[]
  k8supgradeavailable: boolean
  k8supgradeversions: string[]
  k8sversion: string
  name: string
  network: Network
  nodegroups: Nodegroup[]
  provider: string
  region: string
  snapshot: Snapshot[]
  storages: Storages[]
}

export interface Addon {
  args: Arg[]
  artifact: string
  name: string
  state: string
  tags: Tag[]
  version: string
}

export interface Arg {
  key: string
  value: string
}

export interface Tag {
  key: string
  value: string
}

export interface Backup {
  endpoint: string
  folder: string
  key: string
  name: string
  region: string
}

export interface Certsexpiring {
  certexpirydate: string
  cpname: string
}

export interface Network {
  clustercidr: string
  loadbalancer: Loadbalancer[]
  servicecidr: string
}

export interface Loadbalancer {
  backendports: string[]
  frontendportd: string[]
  lbname: string
  nodegrouptype: string
  status: string
  vip: string
  viptype: string
  createddate: string
}

export interface Security {
  destinationip: string
  internalport: 0
  port: 0
  protocol: string[]
  sourceip: string[]
  state: string
  vipid: 0
  vipname: string
  viptype: string
}

export interface Nodegroup {
  count: number
  id: string
  imi: string
  imiupgradeavailable: boolean
  imiupgradeversions: string[]
  instancetype: string
  name: string
  nodes: Node[]
  nodgroupuuid: string
  nodegrouptype_name: string
  nodegroupsummary: {
    activenodes: number
    deletingnodes: number
    errornodes: number
    provisioningnodes: number
  }
  releaseversion: string
  sshkey: string
  status: string
}

export interface Node {
  createddate: string
  dnsname: string
  imi: string
  ipaddress: string
  name: string
  state: string
  status: string
  wekaStorage: NodeStorage
}

export interface NodeStorage {
  clientId: string
  customStatus: string
  message: string
  status: string
}

export interface Snapshot {
  created: string
  filename: string
  name: string
  state: string
  type: string
}

export interface runtimes {
  runtimename: string
  k8sversionname: string[]
  // hosts: any
}
export interface InstanceTypes {
  cpu: number
  memory: number
  storage: number
  instancetypename: string[]
}

export interface SSHKeys {
  sshPrivateKey: string
  sshPublicKey: string
}

export interface Storages {
  message: string
  reason: string
  size: string
  state: string
  storageprovider: string
}

export interface insightsVersion {
  releaseId: string
  vendor: string
  license: string
  purl: string
  releaseStamp: string
  eosTimeStamp: string
  eolTimeStamp: string
  components: insightsComponent[]
}

export interface insightsComponent {
  releaseId: string
  vendor: string
  name: string
  version: string
  purl: string
  sha256: string
  license: string
  type: string
}

export interface insightsSbom {
  sbom: string
  format: string
  createTimestamp: string
}

export interface insightsVulnerability {
  id: string
  description: string
  componentName: string
  componentVersion: string
  affectedPackage: string
  affectedVersions: string
  fixedVersion: string
  severity: string
  publishedAt: string
  componentSHA256: string
}

export interface insightsSummary {
  componentName: string
  componentVersion?: string
  scanTimestamp?: string
  releaseId?: string
  critical: string
  high: string
  low: string
  medium: string
}

export interface component {
  name: string
  versions: string[]
}

export interface ClusterStore {
  clustersData: Clusters[] | null
  setClustersData: (isBackGround: boolean) => Promise<void>
  clusterData: Cluster | null
  setClusterData: (clusteruuid: string) => Promise<void>
  clusterNodegroups: Nodegroup[] | []
  certExpiration: Certsexpiring[] | []
  snapshot: Snapshot[] | []
  storages: Storages[]
  refreshRate: boolean
  sshKeys: SSHKeys | null
  insightsVersions: insightsVersion | null
  setInsightsVersions: (component: string, versionID: string) => Promise<void>
  insightsComponents: component[] | []
  setInsightsComponents: () => Promise<void>
  versionLoading: boolean
  insightsSboms: insightsSbom[] | []
  sbomLoading: boolean
  setInsightsSboms: (component: string, versionID: string) => Promise<void>
  insightsVulnerabilities: insightsVulnerability[] | []
  setInsightsVulnerabilities: (versionId: string, payload: any) => Promise<void>
  vulnerabilitiesLoading: boolean
  summary: insightsSummary[] | []
  setSummary: (versionId: string, payload: any) => Promise<void>
  summaryLoading: boolean
  setSSHKeys: (clusteruuid: string) => Promise<void>
  loadBalancerDetails: Loadbalancer[] | []
  setLoadBalancerDetails: (clusteruuid: string) => Promise<void>
  securityDetails: Security[] | []
  setSecurityDetails: (clusteruuid: string) => Promise<void>
  resetLoadBalancerDetails: () => void
  setRefreshRate: (value: boolean) => void
  loading: boolean
  setIlbRefresh: () => void
  refreshIlb: boolean
}

const useClusterStore = create<ClusterStore>()((set, get) => ({
  loading: false,
  clustersData: null,
  setClustersData: async (isBackGround: boolean) => {
    if (!isBackGround) {
      set({ loading: true })
    }
    let socketResponse: Clusters[] = []
    const socket = await IKSService.getAllClustersDataStatus()
    const clusterInfo = socket.data.response

    if (clusterInfo) {
      socketResponse = createClustersJson(clusterInfo)
      const sortedData: Clusters[] = socketResponse.sort((a, b) => new Date(b.createddate).getTime() - new Date(a.createddate).getTime())
      set({ clustersData: sortedData })
    }
    set({ loading: false })
  },
  clusterData: null,
  clusterNodegroups: [],
  certExpiration: [],
  snapshot: [],
  storages: [],
  setClusterData: async (clusteruuid: string) => {
    let socketResponse: Cluster = ({} as any) as Cluster
    const socket = await IKSService.getClusterData(clusteruuid)
    const clusterInfo = socket.data
    if (clusterInfo) {
      socketResponse = createClusterJson(clusterInfo)
      set({ clusterData: socketResponse })
      set({ clusterNodegroups: socketResponse.nodegroups })
      set({ certExpiration: socketResponse.certsexpiring })
      set({ snapshot: socketResponse.snapshot })
      set({ storages: socketResponse.storages })
    }
    set({ loading: false })
  },
  sshKeys: null,
  setSSHKeys: async (clusteruuid: string) => {
    const socket = await IKSService.getSSHKeys(clusteruuid)
    const sshInfo = socket.data
    set({ sshKeys: sshInfo })
  },
  loadBalancerDetails: [],
  setLoadBalancerDetails: async (clusteruuid: string) => {
    let socketResponse: Loadbalancer[] = []
    const socket = await IKSService.getIlbDetails(clusteruuid)
    const ilbResponse = socket.data.lbresponses ?? []
    if (ilbResponse) {
      socketResponse = createLoadBalancerJson(ilbResponse)
      set({ loadBalancerDetails: socketResponse })
    }
  },
  resetLoadBalancerDetails: () => {
    set({ loadBalancerDetails: [] })
  },
  setRefreshRate: (value: boolean) => {
    set({ refreshRate: value })
  },
  securityDetails: [],
  setSecurityDetails: async (clusteruuid: string) => {
    let socketResponse: Security[] = []

    const socket = await IKSService.getSecurityDetails(clusteruuid)
    const securityResponse = socket.data.getfirewallresponse ?? []
    if (securityResponse) {
      socketResponse = createSecurityJson(securityResponse)
      set({ securityDetails: socketResponse })
    }
  },
  resetSecurityDetails: () => {
    set({ securityDetails: [] })
  },
  refreshRate: false,
  setIlbRefresh: () => {
    const value = needToRefreshLoadBalancer()
    set({ refreshIlb: value })
  },
  refreshIlb: false,
  insightsVersions: null,
  versionLoading: false,
  setInsightsVersions: async (component: string, versionID: string) => {
    set({ versionLoading: true })
    const response = await IKSService.getInsightVersions(component, versionID)
    const versionInfo = response.data
    if (versionInfo) {
      const insightsVersions = createInsightVersionJson(versionInfo)
      set({ insightsVersions })
      set({ versionLoading: false })
    }
  },
  insightsSboms: [],
  sbomLoading: false,
  setInsightsSboms: async (component: string, versionID: string) => {
    set({ sbomLoading: true })
    const response = await IKSService.getSbomItems(component, versionID)
    const sbomInfo = response.data
    if (sbomInfo) {
      const insightsSboms = createInsightSbomJson(sbomInfo)
      set({ insightsSboms })
      set({ sbomLoading: false })
    }
  },
  insightsVulnerabilities: [],
  vulnerabilitiesLoading: false,
  setInsightsVulnerabilities: async (versionID: string, payload: any) => {
    set({ vulnerabilitiesLoading: true })
    const response = await IKSService.getVulnerabilitiesItems(versionID, payload)
    const vulnerabilitiesInfo = response.data.report
    if (vulnerabilitiesInfo) {
      const insightsVulnerabilities = createInsightVulnerabilityJson(vulnerabilitiesInfo)
      set({ insightsVulnerabilities })
      set({ vulnerabilitiesLoading: false })
    }
  },
  summary: [],
  summaryLoading: false,
  setSummary: async (versionID: string, payload: any) => {
    set({ summaryLoading: true })
    const response = await IKSService.getSummaryItems(versionID, payload)
    const vulnerabilitiesInfo = response.data.vulnerabilities
    const summary = createSummaryJson(vulnerabilitiesInfo)
    set({ summary })
    set({ summaryLoading: false })
  },
  insightsComponents: [],
  setInsightsComponents: async () => {
    const response = await IKSService.getComponents()
    const componentInfo = response.data.components
    if (componentInfo) {
      const insightsComponents = createComponentJson(componentInfo)
      set({ insightsComponents })
    }
  }
}))

const createClustersJson = (rows: any): Clusters[] => {
  const clustersData: Clusters[] = []
  const socketResponseJson = rows

  for (let i = 0; i < socketResponseJson.length; i++) {
    const account = socketResponseJson[i].account
    const cpupgradeavailable = socketResponseJson[i].cpupgradeavailable
    const cpupgradeversions = socketResponseJson[i].cpupgradeversions
    const k8supgradeavailable = socketResponseJson[i].k8supgradeavailable
    const k8supgradeversions = socketResponseJson[i].k8supgradeversions
    const k8sversion = socketResponseJson[i].k8sversion
    const name = socketResponseJson[i].name
    const provider = socketResponseJson[i].provider
    const state = socketResponseJson[i].state
    const uuid = socketResponseJson[i].uuid
    const createddate = socketResponseJson[i].createddate
    const clustertype = socketResponseJson[i].clustertype
    const clusterData: Clusters = {
      account,
      cpupgradeavailable,
      cpupgradeversions,
      k8supgradeavailable,
      k8supgradeversions,
      k8sversion,
      name,
      provider,
      state,
      uuid,
      createddate,
      clustertype
    }
    clustersData.push(clusterData)
  }
  return clustersData
}

const createClusterJson = (rows: any): Cluster => {
  const socketResponseJson = rows

  const account = socketResponseJson.account
  const addons = socketResponseJson.addons
  const backup = socketResponseJson.backup
  const certsexpiring = socketResponseJson.certsexpiring
  const k8supgradeavailable = socketResponseJson.k8supgradeavailable
  const k8supgradeversions = socketResponseJson.k8supgradeversions
  const k8sversion = socketResponseJson.k8sversion
  const name = socketResponseJson.name
  const network = socketResponseJson.network
  const nodegroups = socketResponseJson.nodegroups
  const provider = socketResponseJson.provider
  const region = socketResponseJson.region
  const snapshot = socketResponseJson.Snapshot
  const storages = socketResponseJson.storages

  const clusterData: Cluster = {
    account,
    addons,
    backup,
    certsexpiring,
    k8supgradeavailable,
    k8supgradeversions,
    k8sversion,
    name,
    network,
    nodegroups,
    provider,
    region,
    snapshot,
    storages
  }

  return clusterData
}

const createLoadBalancerJson = (rows: any): Loadbalancer[] => {
  const lbsData: Loadbalancer[] = []
  const socketResponseJson = rows
  for (let i = 0; i < socketResponseJson.length; i++) {
    const lbDetails = socketResponseJson[i].lb

    const backendports = lbDetails.backendports
    const frontendportd = lbDetails.frontendportd
    const lbname = lbDetails.lbname
    const nodegrouptype = lbDetails.nodegrouptype
    const status = lbDetails.status
    const vip = lbDetails.vip
    const viptype = lbDetails.viptype
    const createddate = lbDetails.createddate
    const lbData: Loadbalancer = {
      backendports,
      frontendportd,
      lbname,
      nodegrouptype,
      status,
      vip,
      viptype,
      createddate
    }
    lbsData.push(lbData)
  }
  return lbsData
}

const createSecurityJson = (socketResponseJson: any[]): Security[] => {
  const securityData: Security[] = []

  socketResponseJson.forEach((res) => {
    securityData.push({
      destinationip: res.destinationip,
      internalport: res.internalport,
      port: res.port,
      protocol: res.protocol,
      sourceip: res.sourceip,
      state: res.state,
      vipid: res.vipid,
      vipname: res.vipname,
      viptype: res.viptype
    })
  })

  return securityData
}

const createInsightVersionJson = (responseJson: any): insightsVersion => {
  const components = [...responseJson.components]
  const insightsComponents: insightsComponent[] = []
  components.forEach((component) => {
    insightsComponents.push({
      releaseId: component.releaseId,
      license: component.license,
      vendor: component.vendor,
      name: component.name,
      version: component.version,
      purl: component.purl,
      sha256: component.sha256,
      type: component.type
    })
  })
  const insightsVersionData: insightsVersion = {
    releaseId: responseJson.releaseId,
    vendor: responseJson.vendor,
    purl: responseJson.purl,
    license: responseJson.license,
    eosTimeStamp: moment(new Date(responseJson.eosTimestamp)).format('YYYY-MM-DD HH:mm:ss'),
    eolTimeStamp: moment(new Date(responseJson.eolTimestamp)).format('YYYY-MM-DD HH:mm:ss'),
    releaseStamp: moment(new Date(responseJson.releaseTimestamp)).format('YYYY-MM-DD HH:mm:ss'),
    components: insightsComponents
  }

  return insightsVersionData
}

const createInsightVulnerabilityJson = (responseJson: any): insightsVulnerability[] => {
  const response: insightsVulnerability[] = []
  responseJson.forEach((res: any) => {
    const vulnerabilities = [...res.vulnerabilities]
    vulnerabilities.forEach((vulnerability: any) => {
      response.push({
        componentSHA256: res.componentSHA256,
        componentName: res.componentName,
        componentVersion: res.componentVersion,
        id: vulnerability.Id,
        description: vulnerability.description,
        affectedPackage: vulnerability.affectedPackage,
        affectedVersions: vulnerability.affectedVersions,
        fixedVersion: vulnerability.fixedVersion,
        severity: vulnerability.severity,
        publishedAt: moment(new Date(vulnerability.publishedAt)).format('YYYY-MM-DD')
      })
    })
  })
  return response
}

const createInsightSbomJson = (responseJson: any): insightsSbom[] => {
  const insightsSbomData: insightsSbom[] = []

  insightsSbomData.push({
    sbom: responseJson?.sbom,
    format: responseJson?.format,
    createTimestamp: moment(new Date(responseJson?.createTimestamp)).format('YYYY-MM-DD')
  })

  return insightsSbomData
}

const createComponentJson = (responseJson: any): component[] => {
  const insightsResponse: component[] = []
  const kubernetesItems = responseJson.KUBERNETES
  insightsResponse.push({
    name: 'KUBERNETES',
    versions: [...kubernetesItems.id]
  })
  const calicoItems = responseJson.CALICO
  insightsResponse.push({
    name: 'CALICO',
    versions: [...calicoItems.id]
  })
  return insightsResponse
}

const createSummaryJson = (responseJson: any): insightsSummary[] => {
  const insightsData: insightsSummary[] = []
  responseJson.forEach((res: any) => {
    insightsData.push({
      componentName: res.componentName,
      componentVersion: res.componentVersion,
      releaseId: res.releaseId,
      scanTimestamp: moment(new Date(res.scanTimestamp)).format('YYYY-MM-DD HH:mm:ss'),
      critical: res.vulnerabilityCount.CRITICAL ?? '0',
      high: res.vulnerabilityCount.HIGH ?? '0',
      medium: res.vulnerabilityCount.MEDIUM ?? '0',
      low: res.vulnerabilityCount.LOW ?? '0'
    })
  })

  return insightsData
}

const allowedStates = ['Pending', 'Deleting', 'DeletePending', 'Updating']

const needToRefreshLoadBalancer = (): boolean => {
  const lbDetails = useClusterStore.getState().loadBalancerDetails
  if (lbDetails) {
    if (lbDetails.length < 0) {
      return true
    }
    return lbDetails.some(x => allowedStates.includes(x.status))
  }
  return false
}

const needToRefreshNodeGroup = (): boolean => {
  const clusterNodeGroups = useClusterStore.getState().clusterNodegroups
  if (clusterNodeGroups) {
    return clusterNodeGroups.some(x => allowedStates.includes(x.status) || x.nodes.some(x => allowedStates.includes(x.state)))
  }
  return false
}

const needToRefreshClusters = (): boolean => {
  const clusters = useClusterStore.getState().clustersData
  const shouldRefreshClusters = useClusterStore.getState().refreshRate
  if (shouldRefreshClusters && clusters) {
    return clusters.some((c: Clusters) => allowedStates.includes(c.state) || needToRefreshNodeGroup() || needToRefreshLoadBalancer())
  }
  return false
}

setInterval(() => {
  if (needToRefreshClusters()) {
    void (async () => {
      await useClusterStore.getState().setClustersData(true)
    })()
  }
}, 10000)

export default useClusterStore
