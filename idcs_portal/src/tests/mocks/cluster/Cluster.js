// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockAxios } from '../../../setupTests'

export const mockClustersEmpty = () => {
  return {
    clusters: [],
    resourcelimits: {
      maxclusterpercloudaccount: 3,
      maxnodegroupspercluster: 5,
      maxvipspercluster: 2,
      maxnodespernodegroup: 10,
      maxclustervm: 0
    }
  }
}

export const mockGetClustersEmpty = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/clusters'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockClustersEmpty()
      })
    )
}

export const mockSuperComputerReservations = () => {
  return {
    clusters: [
      {
        name: 'iks-ui-stg-1',
        description: '',
        uuid: 'cl-uc6ii67bja',
        clusterstate: 'Active',
        clusterstatus: {
          name: 'iks-ui-stg-1',
          clusteruuid: 'cl-uc6ii67bja',
          state: 'Active',
          lastupdate: '2024-11-07 15:43:07 +0000 UTC',
          reason: '',
          message: 'Cluster ready',
          errorcode: 0
        },
        createddate: '2024-10-29T20:49:58.885144Z',
        k8sversion: '1.29',
        upgradeavailable: false,
        upgradek8sversionavailable: [],
        network: {
          enableloadbalancer: false,
          region: 'us-staging-1',
          servicecidr: '100.66.0.0/16',
          clustercidr: '100.68.0.0/16',
          clusterdns: '100.66.0.10'
        },
        tags: [],
        vips: [],
        annotations: [],
        provisioningLog: [],
        nodegroups: [],
        storageenabled: false,
        storages: [],
        clustertype: 'supercompute'
      }
    ],
    resourcelimits: {
      maxclusterpercloudaccount: 3,
      maxnodegroupspercluster: 5,
      maxvipspercluster: 2,
      maxnodespernodegroup: 10,
      maxclustervm: 0
    }
  }
}

export const mockGeSuperComputerReservations = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/clusters'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockSuperComputerReservations()
      })
    )
}

export const mockIKSReservations = () => {
  return {
    clusters: [
      {
        name: 'iks-ui-stg-1',
        description: '',
        uuid: 'cl-uc6ii67bja',
        clusterstate: 'Active',
        clusterstatus: {
          name: 'iks-ui-stg-1',
          clusteruuid: 'cl-uc6ii67bja',
          state: 'Active',
          lastupdate: '2024-11-07 15:43:07 +0000 UTC',
          reason: '',
          message: 'Cluster ready',
          errorcode: 0
        },
        createddate: '2024-10-29T20:49:58.885144Z',
        k8sversion: '1.29',
        upgradeavailable: false,
        upgradek8sversionavailable: [],
        network: {
          enableloadbalancer: false,
          region: 'us-staging-1',
          servicecidr: '100.66.0.0/16',
          clustercidr: '100.68.0.0/16',
          clusterdns: '100.66.0.10'
        },
        tags: [],
        vips: [],
        annotations: [],
        provisioningLog: [],
        nodegroups: [],
        storageenabled: false,
        storages: [],
        clustertype: 'generalpurpose'
      }
    ],
    getfirewallresponse: [
      {
        sourceip: ['any'],
        state: 'Ready',
        destinationip: '146.152.227.182',
        port: 443,
        vipid: 1954,
        vipname: 'public-apiserver',
        viptype: 'public',
        protocol: ['TCP'],
        internalport: 443
      }
    ],
    resourcelimits: {
      maxclusterpercloudaccount: 3,
      maxnodegroupspercluster: 5,
      maxvipspercluster: 2,
      maxnodespernodegroup: 10,
      maxclustervm: 0
    }
  }
}

export const mockGetIksReservations = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/clusters'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockIKSReservations()
      })
    )
}

export const mockSecurityRules = () => {
  return {
    getfirewallresponse: [
      {
        sourceip: ['any'],
        state: 'Not Specified',
        destinationip: '146.152.227.182',
        port: 443,
        vipid: 1954,
        vipname: 'public-apiserver',
        viptype: 'public',
        protocol: ['TCP'],
        internalport: 443
      }
    ]
  }
}

export const mockGetSecurityRules = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/iks/clusters/cl-uc6ii67bja/security'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockSecurityRules()
      })
    )
}

export const clusterDetail = () => {
  return {
    name: 'super-computer-ui-stg-3',
    clusterstate: 'Active',
    vips: [
      {
        vipid: 1964,
        name: 'myscbalancer',
        description: '',
        vipstate: 'Active',
        port: 80,
        poolport: 80,
        viptype: 'public',
        dnsalias: ['[]'],
        members: [],
        vipstatus: {
          name: '',
          vipstate: 'Active',
          message: '',
          poolid: 0,
          vipid: '0',
          errorcode: 0
        },
        createddate: '2024-10-29T21:11:49.043734Z'
      }
    ],
    securityRules: [
      {
        sourceip: ['any'],
        state: 'Active',
        destinationip: '146.152.227.182',
        port: 443,
        vipid: 1954,
        vipname: 'public-apiserver',
        viptype: 'public',
        protocol: ['TCP'],
        internalport: 443
      }
    ]
  }
}
