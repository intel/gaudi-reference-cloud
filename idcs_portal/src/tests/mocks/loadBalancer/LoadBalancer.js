// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockAxios } from '../../../setupTests'

export const mockBaseLoadBalancerStore = () => {
  return {
    items: [
      {
        metadata: {
          name: 'my-lb-1',
          resourceId: '74e9f11b-487a-4d05-84ca-bb96b4108447',
          cloudAccount: '316585973272',
          creationTimestamp: '2024-05-21T17:28:20.672774821Z'
        },
        spec: {
          listeners: [
            {
              port: '80',
              pool: {
                port: '8080',
                monitor: 'tcp',
                loadBalancingMode: 'round-robin',
                instanceSelectors: {
                  type: 'loadbalancer',
                  app: 'ai-learning'
                },
                instanceResourceIds: []
              }
            },
            {
              port: '443',
              pool: {
                port: '8443',
                monitor: 'tcp',
                loadBalancingMode: 'least-connections-member ',
                instanceResourceIds: ['3d11c2ea-3b74-4891-9034-a5ba6f57f065'],
                instanceSelectors: {}
              }
            }
          ],
          security: {
            sourceips: ['1.2.3.4']
          }
        },
        status: {
          conditions: {
            listeners: [
              {
                port: '80',
                poolCreated: 'true',
                vipPoolLinked: 'true'
              },
              {
                port: '443',
                poolCreated: 'true',
                vipPoolLinked: 'true'
              }
            ],
            firewallRuleCreated: true
          },
          listeners: [
            {
              port: '80',
              name: 'lb-0d7feae4-a2f5-420a-9de7-d9e446f57f27-80',
              vipID: 145606,
              message: 'Provisioning load balancer',
              poolMembers: [],
              poolID: 135357
            },
            {
              port: '443',
              name: 'lb-0d7feae4-a2f5-420a-9de7-d9e446f57f27-443',
              vipID: 145607,
              message: 'Provisioning load balancer',
              poolMembers: [],
              poolID: 135358
            }
          ],
          state: 'Active',
          vip: '146.152.227.44'
        }
      },
      {
        metadata: {
          name: 'my-lb-2',
          resourceId: '74e9f11b-487a-4d05-84ca-bb96b4108447',
          cloudAccount: '316585973272',
          creationTimestamp: '2024-05-21T17:28:20.672774821Z'
        },
        spec: {
          listeners: [
            {
              port: '80',
              pool: {
                port: '8080',
                monitor: 'tcp',
                loadBalancingMode: 'round-robin',
                instanceSelectors: {
                  type: 'loadbalancer',
                  app: 'ai-learning'
                },
                instanceResourceIds: []
              }
            },
            {
              port: '443',
              pool: {
                port: '8443',
                monitor: 'tcp',
                loadBalancingMode: 'least-connections-member ',
                instanceResourceIds: ['3d11c2ea-3b74-4891-9034-a5ba6f57f065'],
                instanceSelectors: {}
              }
            }
          ],
          security: {
            sourceips: ['1.2.3.4']
          }
        },
        status: {
          conditions: {
            listeners: [
              {
                port: '80',
                poolCreated: 'true',
                vipPoolLinked: 'true'
              },
              {
                port: '443',
                poolCreated: 'true',
                vipPoolLinked: 'true'
              }
            ],
            firewallRuleCreated: true
          },
          listeners: [
            {
              port: '80',
              name: 'lb-0d7feae4-a2f5-420a-9de7-d9e446f57f27-80',
              vipID: 145606,
              message: 'Provisioning load balancer',
              poolMembers: [],
              poolID: 135357
            },
            {
              port: '443',
              name: 'lb-0d7feae4-a2f5-420a-9de7-d9e446f57f27-443',
              vipID: 145607,
              message: 'Provisioning load balancer',
              poolMembers: [],
              poolID: 135358
            }
          ],
          state: 'Deleting',
          vip: '146.152.227.44'
        }
      }
    ]
  }
}

export const mockLoadBalancer = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/loadbalancers'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseLoadBalancerStore()
      })
    )
}

export const mockEmptyLoadBalancer = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/loadbalancers'))
    .mockImplementation(() =>
      Promise.resolve({
        data: {
          items: []
        }
      })
    )
}
