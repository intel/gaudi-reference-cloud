// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockAxios } from '../../../setupTests'

export const mockProductSuperComputer = () => {
  return {
    products: [
      {
        name: 'sc-storage-file',
        id: 'f0ad8e1c-795d-49f1-9274-49d562a99c58',
        created: '2024-05-22T20:17:14Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '1a700ca4-0e26-4465-abdf-d10934be3803',
        description: 'High speed File storage',
        metadata: {
          access: 'open',
          billingEnable: 'true',
          disableForAccountTypes: 'standard',
          displayName: 'Storage Service - Filesystem',
          'family.displayDescription': 'Filesystem Storage Service',
          'family.displayName': 'Filesystem Storage Service',
          information: 'Filesystem Storage service\n',
          instanceCategories: 'FileStorage',
          instanceType: 'sc-storage-file',
          'product.family.description': 'Super Computing as a Service',
          recommendedUseCase: 'FileStorage',
          region: 'us-staging-3',
          releaseStatus: 'Released',
          service: 'Storage Service - File',
          'usage.quantity.unit': 'TB hour',
          'usage.unit': 'per TB per Hour',
          'volume.size.max': '100',
          'volume.size.min': '1',
          'volume.size.unit': 'TB'
        },
        eccn: 'EAR99',
        pcq: '19621',
        matchExpr: 'serviceType == "FileStorageAsAService-SC"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_TB_PER_HOUR',
            rate: '0.10',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'ready',
        access: 'open'
      },
      {
        name: 'sc-cluster',
        id: '40aec642-5b67-4df9-aa2d-20f56ce59d61',
        created: '2024-06-12T00:08:44Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '1a700ca4-0e26-4465-abdf-d10934be3803',
        description: 'Managed Supercomputing Service',
        metadata: {
          access: 'open',
          billingEnable: 'true',
          category: 'cluster',
          'cpu.cores': '',
          'cpu.sockets': '',
          disableForAccountTypes: 'standard',
          'disks.size': '',
          displayName: 'Intel managed Supercomputing Service',
          'family.displayDescription': 'IKS Control Plane',
          'family.displayName': 'IKS Control Plane',
          highlight: '',
          information:
            'Lean more at the [Xeon® Scalable processor overview page](https://www.intel.com/content/www/us/en/products/details/processors/xeon/scalable/platinum.html). For detailed processor information see the [Intel product documentation catalog](https://www.intel.com/content/www/us/en/search.html?ws=text#q=5th%20gen&sort=relevancy&f:@tabfilter=[Developers]&f:@stm_10385_en=[Processors,Intel%C2%AE%20Xeon%C2%AE%20Processors,5th%20Generation%20Intel%C2%AE%20Xeon%C2%AE%20Scalable%20Processors]&f:@disclosure=[Public]).\n',
          instanceCategories: 'SupercomputingCluster',
          instanceType: 'sc-cluster',
          'memory.size': '',
          processor: 'CPU',
          'product.family.description': 'Super Computing as a Service',
          recommendedUseCase: 'SC cluster',
          region: 'us-staging-3',
          releaseStatus: 'Released',
          service: 'Supercomputing',
          'usage.quantity.unit': 'min',
          'usage.unit': 'per Minute'
        },
        eccn: 'PURE CLOUD RELEASE',
        pcq: 'ECR-711',
        matchExpr: 'serviceType == "SuperComputingAsAService" && clusterType == "sc-cp-cluster"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.03',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'ready',
        access: 'open'
      },
      {
        name: 'bm-spr-pl-sc',
        id: '9423c95d-981a-4e23-a3c8-3daf7cee85f9',
        created: '2024-08-09T05:41:34Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '1a700ca4-0e26-4465-abdf-d10934be3803',
        description: '96 cores, 2 sockets, 1024 GB memory, 2 TB disk',
        metadata: {
          access: 'controlled',
          billingEnable: 'true',
          category: 'singlenode',
          'cpu.cores': '96',
          'cpu.sockets': '2',
          disableForAccountTypes: 'standard',
          'disks.size': '2TB',
          displayName: '4th Generation Intel® Xeon® Scalable processors (8468)',
          'family.displayDescription': 'Virtual machines and Bare Metal servers available',
          'family.displayName': '4th Generation Intel® Xeon® Scalable processors',
          information:
            'For a Xeon® processor overview see the [Technical Overview page](https://www.intel.com/content/www/us/en/developer/articles/technical/fourth-generation-xeon-scalable-family-overview.html?wapkw=4th%20gen%20intel%20xeon%20scalable%20processors#gs.xc9pz8). For detailed processor information see the [Intel product documentation catalog](https://www.intel.com/content/www/us/en/search.html?ws=text#q=4th%20gen&sort=relevancy&f:@disclosure=[Public]&f:@tabfilter=[Developers]&f:@stm_10385_en=[Processors,Intel%C2%AE%20Xeon%C2%AE%20Processors,4th%20Generation%20Intel%C2%AE%20Xeon%C2%AE%20Scalable%20Processors]), [Intel Accelerator Engine page](https://www.intel.com/content/www/us/en/products/docs/accelerator-engines/overview.html#multiasset_copy), and the [Accelerator e-guide](https://www.intel.com/content/www/us/en/now/xeon-accelerated/accelerators-eguide.html).\n',
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-spr-pl-sc',
          'memory.size': '1024GB',
          processor: 'CPU',
          'product.family.description': 'Compute as a Service: Bare Metal and Virtual Machine',
          recommendedUseCase: 'Core compute',
          region: 'us-staging-3',
          releaseStatus: 'Released',
          service: 'Bare Metal',
          'usage.quantity.unit': 'min',
          'usage.unit': 'per Minute'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr:
          '(serviceType == "ComputeAsAService" || serviceType == "SuperComputingAsAService") && instanceType == "bm-spr-pl-sc" && instanceGroupSize == "1"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.0705',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'ready',
        access: 'controlled'
      },
      {
        name: 'bm-icx-gaudi2-sc-cluster-2',
        id: '1202ce68-456b-407b-bcef-7017f9255c7e',
        created: '2024-08-09T05:48:10Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '1a700ca4-0e26-4465-abdf-d10934be3803',
        description:
          '2 Node cluster 8 Gaudi® 2 HL-225H mezz cards, 3rd Gen Xeon® Platinum 8368 CPUs, 1TB RAM, 30TB disk',
        metadata: {
          access: 'controlled',
          billingEnable: 'true',
          category: 'cluster',
          'cpu.cores': '38',
          disableForAccountTypes: 'standard',
          'disks.size': '2TB',
          displayName:
            '2 Node cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8368 Processors',
          'family.displayDescription':
            'Cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk',
          'family.displayName': 'Gaudi® 2 Deep Learning Server Cluster',
          highlight: '2 Node cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum processors',
          information:
            "The Intel® Gaudi® 2 processor is designed to maximize training throughput and efficiency, while providing developers with optimized software and tools that scale to many workloads and systems. For more information see the [Gaudi2 product page](https://habana.ai/products/gaudi2/). You can find more information on Gaudi and SynapseAI® on [Habana's Developer Site](https://developer.habana.ai/).\n",
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-icx-gaudi2-sc',
          'memory.size': '1TB',
          nodesCount: '2',
          processor: 'AI processors',
          'product.family.description': 'Super Computing as a Service',
          recommendedUseCase: 'AI',
          region: 'us-staging-3',
          releaseStatus: 'Released',
          service: 'Bare Metal',
          'usage.quantity.unit': 'min',
          'usage.unit': 'per Minute'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr:
          '(serviceType == "ComputeAsAService" || serviceType == "SuperComputingAsAService") && instanceType == "bm-icx-gaudi2-sc" && instanceGroupSize == "2"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.3821',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'ready',
        access: 'controlled'
      },
      {
        name: 'bm-icx-gaudi2-sc',
        id: '86394ad0-17ed-4973-8865-8819e369b35b',
        created: '2024-08-09T05:41:50Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '1a700ca4-0e26-4465-abdf-d10934be3803',
        description:
          '8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8368 Processors, 1 TB RAM,30 TB disk',
        metadata: {
          access: 'controlled',
          billingEnable: 'false',
          category: 'singlenode',
          'cpu.cores': '2',
          disableForAccountTypes: 'standard',
          'disks.size': '2TB',
          displayName: '8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8368 Processors',
          'family.displayDescription':
            'Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk',
          'family.displayName': 'Gaudi® 2 Deep Learning Server',
          highlight: '8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum processors',
          information:
            "The Intel® Gaudi® 2 processor is designed to maximize training throughput and efficiency, while providing developers with optimized software and tools that scale to many workloads and systems. For more information see the [Gaudi2 product page](https://habana.ai/products/gaudi2/). You can find more information on Gaudi and SynapseAI® on [Habana's Developer Site](https://developer.habana.ai/).\n",
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-icx-gaudi2-sc',
          'memory.size': '1TB',
          processor: 'AI processors',
          'product.family.description': 'Super Computing as a Service',
          recommendedUseCase: 'Core compute',
          region: 'us-staging-3',
          releaseStatus: 'Released',
          service: 'Bare Metal',
          'usage.quantity.unit': 'min',
          'usage.unit': 'per Minute'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr:
          '(serviceType == "ComputeAsAService" || serviceType == "SuperComputingAsAService") && instanceType == "bm-icx-gaudi2-sc" && instanceGroupSize == "1"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.112',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'ready',
        access: 'controlled'
      },
      {
        name: 'bm-icp-gaudi3-sc-cluster-32',
        id: '53eee73e-aaf5-4b70-b237-8284597b6d6f',
        created: '2024-08-09T05:49:28Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '1a700ca4-0e26-4465-abdf-d10934be3803',
        description:
          '32  Node cluster 8 Gaudi® 2 HL-225H mezz cards, 3rd Gen Xeon® Platinum 8380 CPUs, 1TB RAM,30TB disk',
        metadata: {
          access: 'controlled',
          billingEnable: 'true',
          category: 'cluster',
          'cpu.cores': '32',
          disableForAccountTypes: 'standard',
          'disks.size': '30TB',
          displayName:
            '32 Node cluster of 8 Gaudi® 3 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors',
          'family.displayDescription':
            'Cluster of 8 Gaudi® 3 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk',
          'family.displayName': 'Gaudi® 3 Deep Learning Server Cluster',
          highlight: '32 Node cluster of 8 Gaudi® 3 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum processors',
          information:
            "The Intel® Gaudi® 3 processor is designed to maximize training throughput and efficiency, while providing developers with optimized software and tools that scale to many workloads and systems. For more information see the [Gaudi2 product page](https://habana.ai/products/gaudi2/). You can find more information on Gaudi and SynapseAI® on [Habana's Developer Site](https://developer.habana.ai/).\n",
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-icp-gaudi3-sc',
          'memory.size': '1TB',
          nodesCount: '32',
          processor: 'AI processors',
          'product.family.description': 'Super Computing as a Service',
          recommendedUseCase: 'AI',
          region: 'us-staging-3',
          releaseStatus: 'Released',
          service: 'Bare Metal',
          'usage.quantity.unit': 'min',
          'usage.unit': 'per Minute'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr:
          '(serviceType == "ComputeAsAService" || serviceType == "SuperComputingAsAService") && instanceType == "bm-icp-gaudi3-sc" && instanceGroupSize == "32"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.1587',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'ready',
        access: 'controlled'
      },
      {
        name: 'bm-icp-gaudi3-sc-cluster-128',
        id: '2159f299-f832-4bcd-b1ce-223aecd68fb5',
        created: '2024-08-09T05:48:51Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '1a700ca4-0e26-4465-abdf-d10934be3803',
        description:
          '128 Node cluster 8 Gaudi® 3 HL-225H mezz cards, 3rd Gen Xeon® Platinum 8380 CPUs, 1TB RAM,30TB disk',
        metadata: {
          access: 'controlled',
          billingEnable: 'true',
          category: 'cluster',
          'cpu.cores': '128',
          disableForAccountTypes: 'standard',
          'disks.size': '30TB',
          displayName:
            '128 Node cluster of 8 Gaudi® 3 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors',
          'family.displayDescription':
            'Cluster of 8 Gaudi® 3 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk',
          'family.displayName': 'Gaudi® 3 Deep Learning Server Cluster',
          highlight: '128 Node cluster of 8 Gaudi® 3 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum processors',
          information:
            "The Intel® Gaudi® 3 processor is designed to maximize training throughput and efficiency, while providing developers with optimized software and tools that scale to many workloads and systems. For more information see the [Gaudi2 product page](https://habana.ai/products/gaudi2/). You can find more information on Gaudi and SynapseAI® on [Habana's Developer Site](https://developer.habana.ai/).\n",
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-icp-gaudi3-sc',
          'memory.size': '1TB',
          nodesCount: '128',
          processor: 'AI processors',
          'product.family.description': 'Super Computing as a Service',
          recommendedUseCase: 'AI',
          region: 'us-staging-3',
          releaseStatus: 'Released',
          service: 'Bare Metal',
          'usage.quantity.unit': 'min',
          'usage.unit': 'per Minute'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr:
          '(serviceType == "ComputeAsAService" || serviceType == "SuperComputingAsAService") && instanceType == "bm-icp-gaudi3-sc" && instanceGroupSize == "128"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.1587',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'ready',
        access: 'controlled'
      },
      {
        name: 'bm-icp-gaudi3-sc',
        id: '03070ebd-9066-43e5-b832-dff44a34c90a',
        created: '2024-08-23T21:57:05Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '1a700ca4-0e26-4465-abdf-d10934be3803',
        description:
          '8 Gaudi® 3 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors, 1 TB RAM,30 TB disk',
        metadata: {
          access: 'controlled',
          billingEnable: 'false',
          category: 'singlenode',
          'cpu.cores': '2',
          disableForAccountTypes: 'standard',
          'disks.size': '30TB',
          displayName: '8 Gaudi® 3 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors',
          'family.displayDescription':
            'Gaudi® 3 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk',
          'family.displayName': 'Gaudi® 3 Deep Learning Server',
          highlight: '8 Gaudi® 3 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum processors',
          information:
            "The Intel® Gaudi® 3 processor is designed to maximize training throughput and efficiency, while providing developers with optimized software and tools that scale to many workloads and systems. For more information see the [Gaudi3 product page](https://habana.ai/products/Gaudi3/). You can find more information on Gaudi and SynapseAI® on [Habana's Developer Site](https://developer.habana.ai/).\n",
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-icp-Gaudi3-sc',
          'memory.size': '1TB',
          processor: 'AI processors',
          'product.family.description': 'Super Computing as a Service',
          recommendedUseCase: 'Core compute',
          region: 'us-staging-3',
          releaseStatus: 'Released',
          service: 'Bare Metal',
          'usage.quantity.unit': 'min',
          'usage.unit': 'per Minute'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr:
          '(serviceType == "ComputeAsAService" || serviceType == "SuperComputingAsAService") && instanceType == "bm-icp-gaudi3-sc" && instanceGroupSize == "1"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.1493',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'ready',
        access: 'controlled'
      },
      {
        name: 'bm-icp-gaudi2-sc-cluster-64',
        id: '2b03f5b0-0355-4d11-b862-fb9e9fef7ae8',
        created: '2024-08-09T05:47:49Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '1a700ca4-0e26-4465-abdf-d10934be3803',
        description:
          '64 Node cluster 8 Gaudi® 2 HL-225H mezz cards, 3rd Gen Xeon® Platinum 8380 CPUs, 1TB RAM, 30TB disk',
        metadata: {
          access: 'controlled',
          billingEnable: 'true',
          category: 'cluster',
          'cpu.cores': '40',
          disableForAccountTypes: 'standard',
          'disks.size': '2TB',
          displayName:
            '64 Node cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors',
          'family.displayDescription':
            'Cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk',
          'family.displayName': 'Gaudi® 2 Deep Learning Server Cluster',
          highlight: '64 Node cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum processors',
          information:
            "The Intel® Gaudi® 2 processor is designed to maximize training throughput and efficiency, while providing developers with optimized software and tools that scale to many workloads and systems. For more information see the [Gaudi2 product page](https://habana.ai/products/gaudi2/). You can find more information on Gaudi and SynapseAI® on [Habana's Developer Site](https://developer.habana.ai/).\n",
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-icp-gaudi2-sc',
          'memory.size': '1TB',
          nodesCount: '64',
          processor: 'AI processors',
          'product.family.description': 'Super Computing as a Service',
          recommendedUseCase: 'AI',
          region: 'us-staging-3',
          releaseStatus: 'Released',
          service: 'Bare Metal',
          'usage.quantity.unit': 'min',
          'usage.unit': 'per Minute'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr:
          '(serviceType == "ComputeAsAService" || serviceType == "SuperComputingAsAService") && instanceType == "bm-icp-gaudi2-sc" && instanceGroupSize == "64"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.12',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'ready',
        access: 'controlled'
      },
      {
        name: 'bm-icp-gaudi2-sc-cluster-32',
        id: 'ef8b44ab-ee41-4da2-9c95-65888e179e80',
        created: '2024-08-09T05:47:43Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '1a700ca4-0e26-4465-abdf-d10934be3803',
        description:
          '32 Node cluster 8 Gaudi® 2 HL-225H mezz cards, 3rd Gen Xeon® Platinum 8380 CPUs, 1TB RAM, 30TB disk',
        metadata: {
          access: 'controlled',
          billingEnable: 'true',
          category: 'cluster',
          'cpu.cores': '40',
          disableForAccountTypes: 'standard',
          'disks.size': '2TB',
          displayName:
            '32 Node cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors',
          'family.displayDescription':
            'Cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk',
          'family.displayName': 'Gaudi® 2 Deep Learning Server Cluster',
          highlight: '32 Node cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum processors',
          information:
            "The Intel® Gaudi® 2 processor is designed to maximize training throughput and efficiency, while providing developers with optimized software and tools that scale to many workloads and systems. For more information see the [Gaudi2 product page](https://habana.ai/products/gaudi2/). You can find more information on Gaudi and SynapseAI® on [Habana's Developer Site](https://developer.habana.ai/).\n",
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-icp-gaudi2-sc',
          'memory.size': '1TB',
          nodesCount: '32',
          processor: 'AI processors',
          'product.family.description': 'Super Computing as a Service',
          recommendedUseCase: 'AI',
          region: 'us-staging-3',
          releaseStatus: 'Released',
          service: 'Bare Metal',
          'usage.quantity.unit': 'min',
          'usage.unit': 'per Minute'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr:
          '(serviceType == "ComputeAsAService" || serviceType == "SuperComputingAsAService") && instanceType == "bm-icp-gaudi2-sc" && instanceGroupSize == "32"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.12',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'ready',
        access: 'controlled'
      },
      {
        name: 'bm-icp-gaudi2-sc-cluster-2',
        id: '1b19493d-949a-4dc7-8445-e65ecc3f77c9',
        created: '2024-08-09T05:47:37Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '1a700ca4-0e26-4465-abdf-d10934be3803',
        description:
          '2 Node cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors, 1 TB RAM, 30 TB disk',
        metadata: {
          access: 'controlled',
          billingEnable: 'true',
          category: 'cluster',
          'cpu.cores': '40',
          disableForAccountTypes: 'standard',
          'disks.size': '2TB',
          displayName:
            '2 Node cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors',
          'family.displayDescription':
            'Cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk',
          'family.displayName': 'Gaudi® 2 Deep Learning Server Cluster',
          highlight: '2 Node cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum processors',
          information:
            "The Intel® Gaudi® 2 processor is designed to maximize training throughput and efficiency, while providing developers with optimized software and tools that scale to many workloads and systems. For more information see the [Gaudi2 product page](https://habana.ai/products/gaudi2/). You can find more information on Gaudi and SynapseAI® on [Habana's Developer Site](https://developer.habana.ai/).\n",
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-icp-gaudi2-sc',
          'memory.size': '1TB',
          nodesCount: '2',
          processor: 'AI processors',
          'product.family.description': 'Super Computing as a Service',
          recommendedUseCase: 'AI',
          region: 'us-staging-3',
          releaseStatus: 'Released',
          service: 'Bare Metal',
          'usage.quantity.unit': 'min',
          'usage.unit': 'per Minute'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr:
          '(serviceType == "ComputeAsAService" || serviceType == "SuperComputingAsAService") && instanceType == "bm-icp-gaudi2-sc" && instanceGroupSize == "2"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.3821',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'ready',
        access: 'controlled'
      },
      {
        name: 'bm-icp-gaudi2-sc-cluster-128',
        id: '6ab9f221-42c6-4b80-bfbf-224702aeb49b',
        created: '2024-08-09T05:47:56Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '1a700ca4-0e26-4465-abdf-d10934be3803',
        description:
          '128 Node cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors, 1 TB RAM, 30 TB disk',
        metadata: {
          access: 'controlled',
          billingEnable: 'true',
          category: 'cluster',
          'cpu.cores': '40',
          disableForAccountTypes: 'standard',
          'disks.size': '2TB',
          displayName:
            '128 Node cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors',
          'family.displayDescription':
            'Cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk',
          'family.displayName': 'Gaudi® 2 Deep Learning Server Cluster',
          highlight: '128 Node cluster of 8 Gaudi® 2 HL-225H mezzanine cards with 3rd Gen Xeon® Platinum processors',
          information:
            "The Intel® Gaudi® 2 processor is designed to maximize training throughput and efficiency, while providing developers with optimized software and tools that scale to many workloads and systems. For more information see the [Gaudi2 product page](https://habana.ai/products/gaudi2/). You can find more information on Gaudi and SynapseAI® on [Habana's Developer Site](https://developer.habana.ai/).\n",
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-icp-gaudi2-sc',
          'memory.size': '1TB',
          nodesCount: '128',
          processor: 'AI processors',
          'product.family.description': 'Super Computing as a Service',
          recommendedUseCase: 'AI',
          region: 'us-staging-3',
          releaseStatus: 'Released',
          service: 'Bare Metal',
          'usage.quantity.unit': 'min',
          'usage.unit': 'per Minute'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr:
          '(serviceType == "ComputeAsAService" || serviceType == "SuperComputingAsAService") && instanceType == "bm-icp-gaudi2-sc" && instanceGroupSize == "128"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.12',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'ready',
        access: 'controlled'
      }
    ]
  }
}

export const mockInstanceTypeSuperComputer = () => {
  return {
    instancetypes: [
      {
        instancetypename: 'bm-icx-gaudi2',
        memory: 1000,
        cpu: 8,
        storage: 30000,
        displayname:
          '8 Gaudi2® HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8368 Processors, 1 TB RAM, 30 TB disk',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-spr',
        memory: 256,
        cpu: 56,
        storage: 2000,
        displayname: '4th Generation Intel® Xeon® Scalable Processors (Sapphire Rapids)',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-icp-gaudi2',
        memory: 256,
        cpu: 40,
        storage: 2000,
        displayname: '8 Gaudi2 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'vm-spr-med',
        memory: 32,
        cpu: 16,
        storage: 32,
        displayname: 'Medium VM - Intel® Xeon 4th Gen ® Scalable processor',
        description: '',
        instancecategory: 'VirtualMachine'
      },
      {
        instancetypename: 'bm-spr-pvc-1550-8',
        memory: 0,
        cpu: 0,
        storage: 0,
        displayname: 'Intel_ Max Series GPU (PVC) on 4th Gen Intel_ Xeon_ processors _ 1550 series',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-icp-gaudi2-cluster-8',
        memory: 0,
        cpu: 0,
        storage: 0,
        displayname:
          '8 Node cluster of 8 Gaudi2® HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-icp-gaudi2-cluster-2',
        memory: 0,
        cpu: 0,
        storage: 0,
        displayname: '2  Node cluster of 8 Gaudi2_ HL-225H mezzanine cards with 3rd Gen Xeon_ Platinum 8380 Processors',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'vm-spr-tny',
        memory: 8,
        cpu: 4,
        storage: 10,
        displayname: 'Tiny VM - Intel® Xeon 4th Gen ® Scalable processor',
        description: '',
        instancecategory: 'VirtualMachine'
      },
      {
        instancetypename: 'bm-icp-gaudi2-cluster-16',
        memory: 0,
        cpu: 0,
        storage: 0,
        displayname:
          '16 Node cluster of 8 Gaudi2® HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-icp-gaudi2-cluster-32',
        memory: 0,
        cpu: 0,
        storage: 0,
        displayname:
          '32 Node cluster of 8 Gaudi2® HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-icp-gaudi2-cluster-64',
        memory: 0,
        cpu: 0,
        storage: 0,
        displayname:
          '64 Node cluster of 8 Gaudi2® HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-icp-gaudi2-cluster-128',
        memory: 0,
        cpu: 0,
        storage: 0,
        displayname:
          '128 Node cluster of 8 Gaudi2® HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8380 Processors',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-spr-pvc-1100-4-sa',
        memory: 256,
        cpu: 56,
        storage: 2000,
        displayname: 'Intel® Max Series GPU (PVC) on 4th Gen Intel® Xeon® processors – 1100 series (4x)',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-emr',
        memory: 512,
        cpu: 128,
        storage: 1000,
        displayname: '5th Generation Intel® Xeon® Scalable processors',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-spr-pvc-1100-4',
        memory: 256,
        cpu: 56,
        storage: 2000,
        displayname: 'Intel® Max Series GPU (PVC) on 4th Gen Intel® Xeon® processors – 1100 series (4x)',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-spr-pvc-1100-8',
        memory: 1024,
        cpu: 48,
        storage: 2000,
        displayname: 'Intel® Max Series GPU (PVC) on 4th Gen Intel® Xeon® processors – 1100 series (8x)',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-spr-gaudi2',
        memory: 0,
        cpu: 0,
        storage: 0,
        displayname: '8 Gaudi2 HL-225H mezzanine cards with 4th Gen Xeon® processors, 1 TB RAM, 30 TB disk',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-spr-pl',
        memory: 1024,
        cpu: 48,
        storage: 2000,
        displayname: '4th Generation Intel® Xeon® Scalable Processors (Sapphire Rapids)',
        description: '',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'vm-spr-sml',
        memory: 16,
        cpu: 8,
        storage: 20,
        displayname: 'Small VM - Intel® Xeon 4th Gen ® Scalable processor',
        description: '',
        instancecategory: 'VirtualMachine'
      },
      {
        instancetypename: 'vm-iks-tny',
        memory: 16,
        cpu: 4,
        storage: 15,
        displayname: 'Tiny VM - Intel® Xeon® 4th Gen Scalable processor',
        description: '',
        instancecategory: 'VirtualMachine'
      },
      {
        instancetypename: 'bm-icp-gaudi2-sc',
        memory: 256,
        cpu: 40,
        storage: 2000,
        displayname: '8 Gaudi2 HL-225H mezzanine cards with 3rd Gen Xeon® processors, 1 TB RAM, 30 TB disk',
        description: 'Gaudi2® Deep Learning Server',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-icx-gaudi2-sc',
        memory: 1024,
        cpu: 38,
        storage: 30000,
        displayname:
          '8 Gaudi2® HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8368 Processors, 1 TB RAM, 30 TB disk',
        description: 'Gaudi2® Deep Learning Server',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-spr-pl-sc',
        memory: 1024,
        cpu: 48,
        storage: 2000,
        displayname: '4th Generation Intel® Xeon® Scalable Processors (Sapphire Rapids)',
        description: '4th Generation Intel® Xeon® Scalable processors',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-icp-gaudi2-sc-cluster-2',
        memory: 1000,
        cpu: 8,
        storage: 30000,
        displayname: '2 node 8 Gaudi2 HL-225H mezzanine cards with 3rd Gen Xeon_ processors, 1 TB RAM, 30 TB disk',
        description: 'Gaudi2® Deep Learning Server',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-icx-gaudi2-sc-cluster-2',
        memory: 1000,
        cpu: 8,
        storage: 30000,
        displayname:
          '2 Node cluster of 8 Gaudi2® HL-225H mezzanine cards with 3rd Gen Xeon® Platinum 8368 Processors',
        description: '2 Node cluster  Platinum 8368 Processors',
        instancecategory: 'BareMetalHost'
      },
      {
        instancetypename: 'bm-gnr-gaudi3-smc',
        memory: 17000,
        cpu: 72,
        storage: 2000,
        displayname: '8 Gaudi® 3 HL-325H mezzanine cards with 6th Gen Xeon® Platinum 8468+ Processors - SMC',
        description: '8 Gaudi® 3 HL-325H mezzanine cards 8468+ Processors - SMC',
        instancecategory: 'BareMetalHost'
      }
    ]
  }
}

export const mockRuntimes = () => {
  return {
    runtimes: [
      {
        runtimename: 'Containerd',
        k8sversionname: ['1.24', '1.25', '1.26', '1.27', '1.28', '1.29']
      }
    ]
  }
}

export const mockVnets = () => {
  return {
    items: [
      {
        metadata: {
          cloudAccountId: '105270643061',
          name: 'us-staging-3a-default',
          resourceId: 'ae8063c1-5d8a-41f1-813c-89f010fd00f1'
        },
        spec: {
          region: 'us-staging-3',
          availabilityZone: 'us-staging-3a',
          prefixLength: 27
        }
      }
    ]
  }
}

export const mockClusters = () => {
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

export const clusterEmptyDetail = () => {
  return {
    name: 'super-computer-ui-stg-3',
    clusterstate: 'Active',
    uuid: 'cl-zswnpbotse',
    securityRules: [],
    nodegroups: [
      {
        annotations: '',
        clusteruuid: 'cl-zswnpbotse',
        count: '1',
        createddate: 'string',
        description: 'string',
        imiid: 'string',
        instancetypeid: 'string',
        name: 'string',
        networkinterfacename: 'string',
        nodegroupstate: 'string',
        nodegroupstatus: 'Active',
        nodegroupuuid: 'string',
        nodes: [],
        sshkeyname: [],
        tags: [],
        upgradeavailable: true,
        upgradeimiid: [],
        userdataurl: 'string',
        vnets: [],
        instanceTypeDetails: null,
        nodeGroupType: 'supercompute-ai'
      }
    ],
    storages: [],
    vips: []
  }
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

export const mockReservations = () => {
  return {
    clusters: [
      {
        name: 'super-computer-ui-stg-3',
        description: '',
        uuid: 'cl-zswnpbotse',
        clusterstate: 'Active',
        clusterstatus: {
          name: 'super-computer-ui-stg-3',
          clusteruuid: 'cl-zswnpbotse',
          state: 'Active',
          lastupdate: '2024-11-05 15:23:35 +0000 UTC',
          reason: '',
          message: 'Provisioning compute',
          errorcode: 0
        },
        createddate: '2024-10-29T20:43:15.689861Z',
        k8sversion: '1.29',
        upgradeavailable: false,
        upgradek8sversionavailable: [],
        network: {
          enableloadbalancer: false,
          region: 'us-staging-3',
          servicecidr: '100.66.0.0/16',
          clustercidr: '100.68.0.0/16',
          clusterdns: '100.66.0.10'
        },
        tags: [],
        vips: [
          {
            vipid: 1969,
            name: 'newbalancer',
            description: '',
            vipstate: 'Pending',
            port: 80,
            poolport: 80,
            viptype: 'public',
            dnsalias: ['[]'],
            members: [],
            vipstatus: {
              name: '',
              vipstate: '',
              message: '',
              poolid: 0,
              vipid: '0',
              errorcode: 0
            },
            createddate: '2024-10-30T15:52:44.413916Z'
          },
          {
            vipid: 1964,
            name: 'myscbalancer',
            description: '',
            vipstate: 'Pending',
            port: 80,
            poolport: 80,
            viptype: 'public',
            dnsalias: ['[]'],
            members: [],
            vipstatus: {
              name: '',
              vipstate: '',
              message: '',
              poolid: 0,
              vipid: '0',
              errorcode: 0
            },
            createddate: '2024-10-29T21:11:49.043734Z'
          }
        ],
        annotations: [],
        provisioningLog: [],
        nodegroups: [
          {
            nodegroupuuid: 'ng-6sdtsnmlba',
            clusteruuid: 'cl-zswnpbotse',
            name: 'super-computer-ui-stg-3-group-ai',
            description: '',
            instancetypeid: 'bm-icx-gaudi2-sc-cluster-2',
            nodegroupstate: 'Updating',
            createddate: '2024-10-29T20:43:15.936412Z',
            nodegroupstatus: {
              name: 'super-computer-ui-stg-3-group-ai',
              clusteruuid: 'cl-zswnpbotse',
              nodegroupuuid: 'ng-6sdtsnmlba',
              count: 1,
              state: 'Updating',
              reason: '',
              message: '',
              errorcode: 0,
              nodestatus: [],
              nodegroupsummary: {
                activenodes: 0,
                provisioningnodes: 0,
                errornodes: 0,
                deletingnodes: 0
              }
            },
            count: 1,
            vnets: [
              {
                availabilityzonename: 'us-staging-3a',
                networkinterfacevnetname: 'us-staging-3a-default'
              }
            ],
            sshkeyname: [
              {
                sshkey: 'tri'
              },
              {
                sshkey: 'test'
              }
            ],
            upgradestrategy: {
              drainnodes: true,
              maxunavailablepercentage: 10
            },
            networkinterfacename: 'eth0',
            tags: [],
            nodes: [],
            imiid: 'iks-gd-u22-cd-wk-1-29-2-v20240401',
            upgradeimiid: [],
            upgradeavailable: false,
            annotations: [],
            userdataurl: '',
            nodegrouptype: 'supercompute-ai',
            clustertype: 'supercompute'
          }
        ],
        storageenabled: true,
        storages: [
          {
            storageprovider: 'weka',
            size: '1000GB',
            state: 'Updating',
            reason: '',
            message: ''
          }
        ],
        clustertype: 'supercompute'
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

export const mockIksReservations = () => {
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

export const mockSecurityRules = () => {
  return {
    getfirewallresponse: [
      {
        sourceip: ['any'],
        state: 'Pending',
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

export const mockGetProductCatalog = () => {
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/products'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockProductSuperComputer()
      })
    )
}

export const mockGetProductCatalogEmpty = () => {
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/products'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: {
          products: []
        }
      })
    )
}

export const mockGetInstanceTypesSuperComputer = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/instancetypes'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockInstanceTypeSuperComputer()
      })
    )
}

export const mockGetRuntimes = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/runtimes'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockRuntimes()
      })
    )
}

export const mockGetVnets = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/vnets'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockVnets()
      })
    )
}

export const mockGetClusters = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/clusters'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockClusters()
      })
    )
}

export const mockGetIksClusters = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/clusters'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockIksReservations()
      })
    )
}

export const mockGetNoReservations = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/clusters'))
    .mockImplementation(() =>
      Promise.resolve({
        data: {
          clusters: []
        }
      })
    )
}

export const mockGetReservations = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/clusters'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockReservations()
      })
    )
}

export const mockGetSecurityRules = () => {
  when(mockAxios.get)
    .calledWith(expect.stringContaining('/security'))
    .mockImplementation(() =>
      Promise.resolve({
        data: mockSecurityRules()
      })
    )
}
