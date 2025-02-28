// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockBaseEnrollResponse } from '../authentication/authHelper'
import { mockAxios } from '../../../setupTests'

export const mockBaseInstancesStore = () => {
  return {
    items: [
      {
        metadata: {
          cloudAccountId: '583807362652',
          name: 'test-one',
          resourceId: '0113d66f-7064-42e9-bf8a-db025813740c',
          resourceVersion: '572',
          labels: {},
          creationTimestamp: '2023-06-13T17:28:20.672774821Z',
          deletionTimestamp: null
        },
        spec: {
          availabilityZone: 'us-dev-1a',
          instanceGroup: '',
          instanceType: 'vm-spr-tny',
          machineImage: 'ubuntu-2204-jammy-v20230122',
          runStrategy: 'RerunOnFailure',
          sshPublicKeyNames: ['test'],
          interfaces: [
            {
              name: 'eth0',
              vNet: 'us-dev-1a-default'
            }
          ]
        },
        status: {
          phase: 'Provisioning',
          message:
            'Instance specification has been accepted and is being provisioned. Guest VM is not reported as running',
          interfaces: [
            {
              name: 'eth0',
              vNet: 'us-dev-1a-default',
              dnsName: 'test.583807362652.us-dev-1.idcservice.net',
              prefixLength: 22,
              addresses: ['172.16.0.171'],
              subnet: '172.16.0.0',
              gateway: '172.16.0.1'
            }
          ],
          sshProxy: {
            proxyUser: 'guest-dev9',
            proxyAddress: '10.165.62.252',
            proxyPort: 22
          },
          userName: 'ubuntu'
        }
      },
      {
        metadata: {
          cloudAccountId: '583807362633',
          name: 'test-two',
          resourceId: '0113d66f-7064-42e9-bf8a-db02581374ed',
          resourceVersion: '572',
          labels: {},
          creationTimestamp: '2023-06-13T17:28:20.672774821Z',
          deletionTimestamp: null
        },
        spec: {
          availabilityZone: 'us-dev-1a',
          instanceGroup: '',
          instanceType: 'vm-spr-tny',
          machineImage: 'ubuntu-2204-jammy-v20230122',
          runStrategy: 'RerunOnFailure',
          sshPublicKeyNames: ['test'],
          interfaces: [
            {
              name: 'eth0',
              vNet: 'us-dev-1a-default'
            }
          ]
        },
        status: {
          phase: 'Ready',
          message:
            'Instance specification has been accepted and is being provisioned. Guest VM is not reported as running',
          interfaces: [
            {
              name: 'eth0',
              vNet: 'us-dev-1a-default',
              dnsName: 'test.583807362652.us-dev-1.idcservice.net',
              prefixLength: 22,
              addresses: ['172.16.0.172'],
              subnet: '172.16.0.0',
              gateway: '172.16.0.1'
            }
          ],
          sshProxy: {
            proxyUser: 'guest-dev9',
            proxyAddress: '10.165.62.252',
            proxyPort: 22
          },
          userName: 'ubuntu'
        }
      },
      {
        metadata: {
          cloudAccountId: '583807362698',
          name: 'test-three',
          resourceId: '0113d66f-7064-42e9-bf8a-db02581374rw',
          resourceVersion: '572',
          labels: {},
          creationTimestamp: '2023-06-13T17:28:20.672774821Z',
          deletionTimestamp: null
        },
        spec: {
          availabilityZone: 'us-dev-1a',
          instanceGroup: '',
          instanceType: 'vm-spr-tny',
          machineImage: 'ubuntu-2204-jammy-v20230122',
          runStrategy: 'RerunOnFailure',
          sshPublicKeyNames: ['test'],
          interfaces: [
            {
              name: 'eth0',
              vNet: 'us-dev-1a-default'
            }
          ]
        },
        status: {
          phase: 'Stopped',
          message:
            'Instance specification has been accepted and is being provisioned. Guest VM is not reported as running',
          interfaces: [
            {
              name: 'eth0',
              vNet: 'us-dev-1a-default',
              dnsName: 'test.583807362652.us-dev-1.idcservice.net',
              prefixLength: 22,
              addresses: ['172.16.0.173'],
              subnet: '172.16.0.0',
              gateway: '172.16.0.1'
            }
          ],
          sshProxy: {
            proxyUser: 'guest-dev9',
            proxyAddress: '10.165.62.252',
            proxyPort: 22
          },
          userName: 'ubuntu'
        }
      },
      {
        metadata: {
          cloudAccountId: '583807362623',
          name: 'test-four',
          resourceId: '0113d66f-7064-42e9-bf8a-db02581374we',
          resourceVersion: '572',
          labels: {},
          creationTimestamp: '2023-06-13T17:28:20.672774821Z',
          deletionTimestamp: null
        },
        spec: {
          availabilityZone: 'us-dev-1a',
          instanceGroup: '',
          instanceType: 'vm-spr-tny',
          machineImage: 'ubuntu-2204-jammy-v20230122',
          runStrategy: 'RerunOnFailure',
          sshPublicKeyNames: ['test'],
          interfaces: [
            {
              name: 'eth0',
              vNet: 'us-dev-1a-default'
            }
          ]
        },
        status: {
          phase: 'Failed',
          message:
            'Instance specification has been accepted and is being provisioned. Guest VM is not reported as running',
          interfaces: [
            {
              name: 'eth0',
              vNet: 'us-dev-1a-default',
              dnsName: 'test.583807362652.us-dev-1.idcservice.net',
              prefixLength: 22,
              addresses: ['172.16.0.174'],
              subnet: '172.16.0.0',
              gateway: '172.16.0.1'
            }
          ],
          sshProxy: {
            proxyUser: 'guest-dev9',
            proxyAddress: '10.165.62.252',
            proxyPort: 22
          },
          userName: 'ubuntu'
        }
      },
      {
        metadata: {
          cloudAccountId: '583807362634',
          name: 'test-five',
          resourceId: '0113d66f-7064-42e9-bf8a-db02581374yt',
          resourceVersion: '572',
          labels: {},
          creationTimestamp: '2023-06-13T17:28:20.672774821Z',
          deletionTimestamp: null
        },
        spec: {
          availabilityZone: 'us-dev-1a',
          instanceGroup: '',
          instanceType: 'vm-spr-tny',
          machineImage: 'ubuntu-2204-jammy-v20230122',
          runStrategy: 'RerunOnFailure',
          sshPublicKeyNames: ['test'],
          interfaces: [
            {
              name: 'eth0',
              vNet: 'us-dev-1a-default'
            }
          ]
        },
        status: {
          phase: 'Terminating',
          message:
            'Instance specification has been accepted and is being provisioned. Guest VM is not reported as running',
          interfaces: [
            {
              name: 'eth0',
              vNet: 'us-dev-1a-default',
              dnsName: 'test.583807362652.us-dev-1.idcservice.net',
              prefixLength: 22,
              addresses: ['172.16.0.175'],
              subnet: '172.16.0.0',
              gateway: '172.16.0.1'
            }
          ],
          sshProxy: {
            proxyUser: 'guest-dev9',
            proxyAddress: '10.165.62.252',
            proxyPort: 22
          },
          userName: 'ubuntu'
        }
      }
    ]
  }
}

export const mockBaseInstanceTypesStore = () => {
  return {
    items: [
      {
        metadata: {
          name: 'bm-spr-hbm'
        },
        spec: {
          name: 'bm-spr-hbm',
          displayName:
            'Intel® Xeon® processors, codenamed Sapphire Rapids with high bandwidth memory (HBM) – HBM-only mode',
          description:
            'Intel® Xeon® processors, codenamed Sapphire Rapids with high bandwidth memory (HBM) – HBM-only mode',
          instanceCategory: 'BareMetalHost',
          cpu: {
            cores: 56,
            id: '0x806F8',
            modelName: '4th Generation Intel® Xeon® Scalable Processors (Sapphire Rapids)',
            sockets: 2,
            threads: 2
          },
          memory: {
            size: '256Gi',
            dimmSize: '32Gi',
            dimmCount: 8,
            speed: 4000
          },
          disks: [
            {
              size: '2000Gi'
            }
          ],
          gpu: {
            modelName: '',
            count: 0
          },
          hbmMode: 'HBM-only'
        }
      },
      {
        metadata: {
          name: 'bm-icx-atsm-170-1'
        },
        spec: {
          name: 'bm-icx-atsm-170-1',
          displayName: 'Intel® Data Center GPU Flex Series on 3rd Gen Intel® Xeon® processors – 170 series (1x)',
          description: 'Intel® Data Center GPU Flex Series on 3rd Gen Intel® Xeon® processors – 170 series (1x)',
          instanceCategory: 'BareMetalHost',
          cpu: {
            cores: 24,
            id: '0x606A6',
            modelName: '3rd Generation Intel® Xeon® Scalable Processors (Ice Lake)',
            sockets: 2,
            threads: 2
          },
          memory: {
            size: '256Gi',
            dimmSize: '32Gi',
            dimmCount: 8,
            speed: 4000
          },
          disks: [
            {
              size: '2000Gi'
            }
          ],
          gpu: {
            modelName: 'GPU-Flex-170',
            count: 1
          },
          hbmMode: ''
        }
      },
      {
        metadata: {
          name: 'bm-icx'
        },
        spec: {
          name: 'bm-icx',
          displayName: '3rd Generation Intel® Xeon® Scalable Processors (Ice Lake)',
          description: '3rd Generation Intel® Xeon® Scalable Processors (Ice Lake)',
          instanceCategory: 'BareMetalHost',
          cpu: {
            cores: 24,
            id: '0x606A6',
            modelName: '3rd Generation Intel® Xeon® Scalable Processors (Ice Lake)',
            sockets: 2,
            threads: 2
          },
          memory: {
            size: '256Gi',
            dimmSize: '32Gi',
            dimmCount: 8,
            speed: 4000
          },
          disks: [
            {
              size: '2000Gi'
            }
          ],
          gpu: {
            modelName: '',
            count: 0
          },
          hbmMode: ''
        }
      },
      {
        metadata: {
          name: 'bm-spr-atsm-170-1'
        },
        spec: {
          name: 'bm-spr-atsm-170-1',
          displayName: 'Intel® Data Center GPU Flex Series on 4rd Gen Intel® Xeon® processors – 170 series (1x)',
          description: 'Intel® Data Center GPU Flex Series on 4rd Gen Intel® Xeon® processors – 170 series (1x)',
          instanceCategory: 'BareMetalHost',
          cpu: {
            cores: 56,
            id: '0x806F8',
            modelName: '4th Generation Intel® Xeon® Scalable Processors (Sapphire Rapids)',
            sockets: 2,
            threads: 2
          },
          memory: {
            size: '256Gi',
            dimmSize: '32Gi',
            dimmCount: 8,
            speed: 4000
          },
          disks: [
            {
              size: '2000Gi'
            }
          ],
          gpu: {
            modelName: 'GPU-Flex-170',
            count: 1
          },
          hbmMode: ''
        }
      },
      {
        metadata: {
          name: 'vm-spr-med'
        },
        spec: {
          name: 'vm-spr-med',
          displayName: 'Medium VM - Intel® Xeon 4th Gen ® Scalable processor',
          description: '4th Generation Intel® Xeon® Scalable processor',
          instanceCategory: 'VirtualMachine',
          cpu: {
            cores: 16,
            id: '0x806F2',
            modelName: '4th Generation Intel® Xeon® Scalable processor',
            sockets: 1,
            threads: 1
          },
          memory: {
            size: '32Gi',
            dimmSize: '32Gi',
            dimmCount: 1,
            speed: 3200
          },
          disks: [
            {
              size: '32Gi'
            }
          ],
          gpu: null,
          hbmMode: ''
        }
      },
      {
        metadata: {
          name: 'vm-spr-sml'
        },
        spec: {
          name: 'vm-spr-sml',
          displayName: 'Small VM - Intel® Xeon 4th Gen ® Scalable processor',
          description: '4th Generation Intel® Xeon® Scalable processor',
          instanceCategory: 'VirtualMachine',
          cpu: {
            cores: 8,
            id: '0x806F2',
            modelName: '4th Generation Intel® Xeon® Scalable processor',
            sockets: 1,
            threads: 1
          },
          memory: {
            size: '16Gi',
            dimmSize: '16Gi',
            dimmCount: 1,
            speed: 3200
          },
          disks: [
            {
              size: '20Gi'
            }
          ],
          gpu: null,
          hbmMode: ''
        }
      },
      {
        metadata: {
          name: 'vm-spr-tny'
        },
        spec: {
          name: 'vm-spr-tny',
          displayName: 'Tiny VM - Intel® Xeon 4th Gen ® Scalable processor',
          description: '4th Generation Intel® Xeon® Scalable processor',
          instanceCategory: 'VirtualMachine',
          cpu: {
            cores: 4,
            id: '0x806F2',
            modelName: '4th Generation Intel® Xeon® Scalable processor',
            sockets: 1,
            threads: 1
          },
          memory: {
            size: '8Gi',
            dimmSize: '8Gi',
            dimmCount: 1,
            speed: 3200
          },
          disks: [
            {
              size: '10Gi'
            }
          ],
          gpu: null,
          hbmMode: ''
        }
      },
      {
        metadata: {
          name: 'bm-spr-pvc-1100-1'
        },
        spec: {
          name: 'bm-spr-pvc-1100-1',
          displayName: 'Intel® Max Series GPU (PVC) on 4th Gen Intel® Xeon® processors – 1100 series (1x)',
          description: 'Intel® Max Series GPU (PVC) on 4th Gen Intel® Xeon® processors – 1100 series (1x)',
          instanceCategory: 'BareMetalHost',
          cpu: {
            cores: 56,
            id: '0x806F8',
            modelName: '4th Generation Intel® Xeon® Scalable Processors (Sapphire Rapids)',
            sockets: 2,
            threads: 2
          },
          memory: {
            size: '256Gi',
            dimmSize: '32Gi',
            dimmCount: 8,
            speed: 4000
          },
          disks: [
            {
              size: '2000Gi'
            }
          ],
          gpu: {
            modelName: 'gpu-max-1100',
            count: 1
          },
          hbmMode: ''
        }
      },
      {
        metadata: {
          name: 'vm-spr-lrg'
        },
        spec: {
          name: 'vm-spr-lrg',
          displayName: 'Large VM - Intel® Xeon 4th Gen ® Scalable processor',
          description: '4th Generation Intel® Xeon® Scalable processor',
          instanceCategory: 'VirtualMachine',
          cpu: {
            cores: 32,
            id: '0x806F2',
            modelName: '4th Generation Intel® Xeon® Scalable processor',
            sockets: 1,
            threads: 1
          },
          memory: {
            size: '64Gi',
            dimmSize: '64Gi',
            dimmCount: 1,
            speed: 3200
          },
          disks: [
            {
              size: '64Gi'
            }
          ],
          gpu: null,
          hbmMode: ''
        }
      },
      {
        metadata: {
          name: 'bm-spr-hbm-f'
        },
        spec: {
          name: 'bm-spr-f',
          displayName:
            'Intel® Xeon® processors, codenamed Sapphire Rapids with high bandwidth memory (HBM) – Flat mode',
          description:
            'Intel® Xeon® processors, codenamed Sapphire Rapids with high bandwidth memory (HBM) – Flat mode',
          instanceCategory: 'BareMetalHost',
          cpu: {
            cores: 56,
            id: '0x806F8',
            modelName: '4th Generation Intel® Xeon® Scalable Processors (Sapphire Rapids)',
            sockets: 2,
            threads: 2
          },
          memory: {
            size: '256Gi',
            dimmSize: '32Gi',
            dimmCount: 8,
            speed: 4000
          },
          disks: [
            {
              size: '2000Gi'
            }
          ],
          gpu: {
            modelName: '',
            count: 0
          },
          hbmMode: 'Flat-1LM'
        }
      },
      {
        metadata: {
          name: 'bm-spr'
        },
        spec: {
          name: 'bm-spr',
          displayName: '4th Generation Intel® Xeon® Scalable Processors (Sapphire Rapids)',
          description: '4th Generation Intel® Xeon® Scalable Processors (Sapphire Rapids)',
          instanceCategory: 'BareMetalHost',
          cpu: {
            cores: 56,
            id: '0x806F8',
            modelName: '4th Generation Intel® Xeon® Scalable Processors (Sapphire Rapids)',
            sockets: 2,
            threads: 2
          },
          memory: {
            size: '256Gi',
            dimmSize: '32Gi',
            dimmCount: 8,
            speed: 4000
          },
          disks: [
            {
              size: '2000Gi'
            }
          ],
          gpu: {
            modelName: '',
            count: 0
          },
          hbmMode: ''
        }
      },
      {
        metadata: {
          name: 'bm-virtual'
        },
        spec: {
          name: 'bm-virtual',
          displayName: 'Bare metal instance for vBMC',
          description: 'Virtual bare metal node for development and testing',
          instanceCategory: 'BareMetalHost',
          cpu: {
            cores: 1,
            id: '0x00000',
            modelName: 'Intel Xeon Processor (Cooperlake)',
            sockets: 2,
            threads: 1
          },
          memory: {
            size: '8Gi',
            dimmSize: '8Gi',
            dimmCount: 1,
            speed: 4800
          },
          disks: [
            {
              size: '32Gi'
            }
          ],
          gpu: {
            modelName: '',
            count: 0
          },
          hbmMode: ''
        }
      },
      {
        metadata: {
          name: 'bm-clx'
        },
        spec: {
          name: 'bm-clx',
          displayName: '2nd Generation Intel® Xeon® Scalable Processors (Cascade Lake)',
          description: '2nd Generation Intel® Xeon® Scalable Processors (Cascade Lake)',
          instanceCategory: 'BareMetalHost',
          cpu: {
            cores: 24,
            id: '0x50656',
            modelName: '2nd Generation Intel® Xeon® Scalable Processors (Cascade Lake)',
            sockets: 2,
            threads: 2
          },
          memory: {
            size: '256Gi',
            dimmSize: '32Gi',
            dimmCount: 8,
            speed: 4000
          },
          disks: [
            {
              size: '2000Gi'
            }
          ],
          gpu: {
            modelName: '',
            count: 0
          },
          hbmMode: ''
        }
      }
    ]
  }
}

export const mockInstances = () => {
  const { cloudAccountId } = mockBaseEnrollResponse()
  when(mockAxios.get)
    .calledWith(expect.stringContaining(`/cloudaccounts/${cloudAccountId}/instances`))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseInstancesStore()
      })
    )
}

export const mockInstanceTypes = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/instancetypes'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseInstanceTypesStore()
      })
    )
}

export const mockEmptyInstances = () => {
  const { cloudAccountId } = mockBaseEnrollResponse()
  when(mockAxios.get)
    .calledWith(expect.stringContaining(`/cloudaccounts/${cloudAccountId}/instances`))
    .mockImplementation(() =>
      Promise.resolve({
        data: {
          items: []
        }
      })
    )
}
