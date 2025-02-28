// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation
import type Product from '../Product/Product'

export interface Annotation {
  key: string
  value: string
}

export interface Network {
  clustercidr: string
  clusterdns: string
  enableloadbalancer: boolean
  region: string
  servicecidr: string
}

export interface Node {
  createddate: string
  dnsname: string
  imi: string
  ipaddress: string
  name: string
  state: string
  status: string
}

export interface UpgradeStrategy {
  drainnodes: boolean
  maxunavailablepercentage: number
}

export interface Vnet {
  availabilityzonename: string
  networkinterfacevnetname: string
}

export interface ProvisioningLog {
  logentry: string
  loglevel: string
  logobject: string
  timestamp: string
}

export interface Tag {
  key: string
  value: string
}

export interface VipMember {
  ipaddresses: string[]
}

export interface Vip {
  description: string
  dnsalias: string[]
  members: VipMember[]
  name: string
  poolport: number
  port: number
  vipIp: string
  vipid: number
  vipstate: string
  vipstatus: string
  viptype: string
}

export interface NodeGroup {
  annotations: Annotation[]
  clusteruuid: string
  count: number
  createddate: string
  description: string
  imiid: string
  instancetypeid: string
  name: string
  networkinterfacename: string
  nodegroupstate: string
  nodegroupstatus: string
  nodegroupuuid: string
  nodes: Node[]
  sshkeyname: string[]
  tags: Tag[]
  upgradeavailable: boolean
  upgradeimiid: string[]
  upgradestrategy: UpgradeStrategy
  userdataurl: string
  vnets: Vnet[]
  instanceTypeDetails: Product | null
  nodeGroupType: string
}

export interface Storage {
  storageprovider: string
  size: string
  state: string
  reason: string
  message: string
}

export interface SuperCluster {
  annotations: Annotation[]
  clusterstate: string
  clusterstatus: string
  clustertype: string
  createddate: string
  description: string
  k8sversion: string
  name: string
  network: Network
  nodegroups: NodeGroup[]
  provisioningLog: ProvisioningLog[]
  tags: Tag[]
  upgradeavailable: boolean
  upgradek8sversionavailable: string[]
  uuid: string
  vips: Vip[]
  kubeconfig: string
  storages: Storage[]
  securityRules: SecurityRulesTypes[] | []
}

export interface SecurityRulesTypes {
  destinationip: string
  port: string
  protocol: string[]
  sourceip: string[]
  state: string
  vipid: number
  vipname: string
  viptype: string
}

export interface SuperClusterResourceLimit {
  maxclusterpercloudaccount: number
  maxnodegroupspercluster: number
  maxvipspercluster: number
}

export interface runtimes {
  runtimename: string
  k8sversionname: string[]
}

export interface InstanceTypes {
  cpu: number
  memory: number
  storage: number
  instancetypename: string[]
}
