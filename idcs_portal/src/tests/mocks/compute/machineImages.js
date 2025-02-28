// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockAxios } from '../../../setupTests'

export const mockBaseMachineImagesStore = () => {
  return {
    items: [
      {
        metadata: {
          name: 'tenant_gaudi2_baremetal'
        },
        spec: {
          displayName: 'Ubuntu 20.04',
          description: 'Ubuntu 20.04.',
          userName: 'sdp',
          icon: 'https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png',
          instanceCategories: ['BareMetalHost'],
          instanceTypes: ['bm-icp-gaudi2', 'bm-icx-gaudi2'],
          md5sum: 'c6984a370d56b0f9a4920a9447eb1c8a',
          sha256sum: 'd4f21586fca2f559d3fdb318b0b98d852c46da527865307595dd4feddfeb6853',
          sha512sum: '',
          labels: {
            architecture: 'X86_64 (Baremetal only)',
            family: 'ubuntu-2004-lts'
          },
          imageCategories: ['App testing'],
          components: [
            {
              name: 'Ubuntu 20.04 LTS',
              type: 'OS',
              version: '20.04',
              description:
                'Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.',
              infoUrl: 'https://releases.ubuntu.com/focal',
              imageUrl:
                'https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png'
            },
            {
              name: 'SynapseAI SW',
              type: 'Software kit',
              version: '1.11.0',
              description: 'Designed to facilitate high-performance DL training on Habana’s Gaudi accelerators.',
              infoUrl:
                'https://docs.habana.ai/en/latest/SW_Stack_Packages_Installation/Synapse_SW_Stack_Installation.html#sw-stack-packages-installation',
              imageUrl: ''
            }
          ]
        }
      },
      {
        metadata: {
          name: 'ubuntu-22.04-server-cloudimg-amd64-latest'
        },
        spec: {
          displayName: 'Ubuntu 22.04 LTS (Jammy Jellyfish) v20221204',
          description:
            'Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.',
          userName: 'sdp',
          icon: 'https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png',
          instanceCategories: ['BareMetalHost'],
          instanceTypes: [],
          md5sum: '2944673edc979fb95e07dc4f6741ada8',
          sha256sum: '9ba87b73fcc4f9bcfaae23c43071924796a36c3e088777c0087d2628533a7c45',
          sha512sum: '',
          labels: {
            architecture: 'X86_64 (Baremetal only)',
            family: 'ubuntu-2204-lts'
          },
          imageCategories: ['App testing'],
          components: [
            {
              name: 'Ubuntu 22.04 LTS',
              type: 'OS',
              version: '22.04',
              description:
                'Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.',
              infoUrl:
                'https://discourse.ubuntu.com/t/jammy-jellyfish-release-notes/24668?_ga=2.61253994.716223186.1673128475-452578204.1673128475',
              imageUrl:
                'https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png'
            },
            {
              name: 'OneAPI base kit',
              type: 'Software kit',
              version: '20.22.3',
              description:
                'core set of tools and libraries for developing high-performance, data-centric applications across diverse architectures.',
              infoUrl:
                'https://www.intel.com/content/www/us/en/developer/tools/oneapi/base-toolkit.html#gs.ma20sthttps://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png',
              imageUrl: ''
            }
          ]
        }
      },
      {
        metadata: {
          name: 'ubuntu-22.04-server-cloudimg-amd64-latest'
        },
        spec: {
          displayName: 'Ubuntu 22.04 LTS (Jammy Jellyfish) v20221204',
          description:
            'Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.',
          userName: 'sdp',
          icon: 'https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png',
          instanceCategories: ['BareMetalHost'],
          instanceTypes: [],
          md5sum: '2944673edc979fb95e07dc4f6741ada8',
          sha256sum: '9ba87b73fcc4f9bcfaae23c43071924796a36c3e088777c0087d2628533a7c45',
          sha512sum: '',
          labels: {
            architecture: 'X86_64 (Baremetal only)',
            family: 'ubuntu-2204-lts'
          },
          imageCategories: ['App testing'],
          components: [
            {
              name: 'Ubuntu 22.04 LTS',
              type: 'OS',
              version: '22.04',
              description:
                'Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.',
              infoUrl:
                'https://discourse.ubuntu.com/t/jammy-jellyfish-release-notes/24668?_ga=2.61253994.716223186.1673128475-452578204.1673128475',
              imageUrl:
                'https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png'
            },
            {
              name: 'OneAPI base kit',
              type: 'Software kit',
              version: '20.22.3',
              description:
                'core set of tools and libraries for developing high-performance, data-centric applications across diverse architectures.',
              infoUrl:
                'https://www.intel.com/content/www/us/en/developer/tools/oneapi/base-toolkit.html#gs.ma20sthttps://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png',
              imageUrl: ''
            }
          ]
        }
      },
      {
        metadata: {
          name: 'ubuntu-2204-jammy-v20230122'
        },
        spec: {
          displayName: 'Ubuntu 22.04 LTS (Jammy Jellyfish) v20230122',
          description:
            'Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.',
          userName: 'ubuntu',
          icon: 'https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png',
          instanceCategories: ['VirtualMachine'],
          instanceTypes: [],
          md5sum: 'e698c9dc23bebef414545dd728826f63',
          sha256sum: 'aa909db670ae8c9481a6b84e0ad2ee20727ee96b23db6db588f1359866fb3918',
          sha512sum:
            'cb54b1be5789e00cc86c080913994f5d5158fe1274dcfdca883b14018627520a878f5166ec1bcc68d82cd855b60803453229568f3d9feb734374dae45c2237eb',
          labels: {
            architecture: 'X86_64 (VM only)',
            family: 'ubuntu-2204-lts'
          },
          imageCategories: ['AI', 'General Computing', 'Deep Learning', 'App testing'],
          components: [
            {
              name: 'Ubuntu 22.04 LTS',
              type: 'OS',
              version: '22.04',
              description:
                'Ubuntu Server is a version of the Ubuntu operating system designed and engineered as a backbone for the internet.',
              infoUrl:
                'https://discourse.ubuntu.com/t/jammy-jellyfish-release-notes/24668?_ga=2.61253994.716223186.1673128475-452578204.1673128475',
              imageUrl:
                'https://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png'
            },
            {
              name: 'OneAPI base kit',
              type: 'Software kit',
              version: '20.22.3',
              description:
                'core set of tools and libraries for developing high-performance, data-centric applications across diverse architectures.',
              infoUrl:
                'https://www.intel.com/content/www/us/en/developer/tools/oneapi/base-toolkit.html#gs.ma20sthttps://www.intel.com/content/dam/www/central-libraries/us/en/images/3rd-gen-xeon-scalable-processor-badge-left-rwd.png.rendition.intel.web.368.207.png',
              imageUrl: ''
            }
          ]
        }
      }
    ]
  }
}

export const mockMachineImages = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/machineimages'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseMachineImagesStore()
      })
    )
}

export const mockEmptyMachineImages = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/machineimages'))
    .mockImplementation(() =>
      Promise.resolve({
        data: {}
      })
    )
}
