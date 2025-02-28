// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockAxios } from '../../../setupTests'

export const mockBaseProductCatalogStore = () => {
  return {
    products: [
      {
        name: 'bm-icx',
        id: '3bc52387-da79-4947-a562-ab7a88c38e14',
        created: '2023-07-11T20:30:20Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '61befbee-0607-47c5-b140-c6b4ea10b9da',
        description: '2 sockets, 256 GB memory, 2 TB disk',
        metadata: {
          category: 'singlenode',
          'cpu.cores': '48',
          'cpu.sockets': '2',
          'disks.size': '2TB',
          displayName: '3rd Generation Intel® Xeon® Scalable Processors',
          'family.displayDescription': 'Bare Metal servers available',
          'family.displayName': '3rd Generation Intel® Xeon® Scalable Processors',
          information:
            'For a Xeon® processor overview see the [Technical Overview](https://www.intel.com/content/www/us/en/developer/articles/technical/fourth-generation-xeon-scalable-family-overview.html?wapkw=4th%20gen%20intel%20xeon%20scalable%20processors#gs.xc9pz8). For detailed processor information see the [Intel product documentation catalog](https://www.intel.com/content/www/us/en/search.html?ws=recent#q=Xeon%203rd%20gen&sort=relevancy&f:@tabfilter=%5BDevelopers%5D).\n',
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-icx',
          'memory.size': '256GB',
          processor: 'CPU',
          recommendedUseCase: 'Core compute',
          region: 'global',
          service: 'Bare Metal'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr: 'serviceType == "ComputeAsAService" && instanceType == "bm-icx"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '.10',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'error'
      },
      {
        name: 'bm-icx-atsm-170-1',
        id: '3bc52387-da79-4947-a562-ab7a88c38e16',
        created: '2023-07-11T20:30:20Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '61befbee-0607-47c5-b140-c6b4ea10b9da',
        description: '1 GPU with 3rd Gen CPU, 2 sockets, 256 GB memory, 2 TB disk',
        metadata: {
          category: 'singlenode',
          'cpu.cores': '48',
          'cpu.sockets': '2',
          'disks.size': '2TB',
          displayName: 'Intel® Data Center GPU Flex Series on 3rd Gen Intel® Xeon® processors – 170 series (1x)',
          'family.displayDescription': 'Intel® Data Center GPU Flex Series on latest Intel® Xeon® processors family',
          'family.displayName': 'Intel® Data Center GPU Flex Series on latest Intel® Xeon® processors',
          highlight: '1 GPU with 3rd Gen CPU',
          information:
            'For details on the Intel® Data Center GPU Flex Series processor, see the [Technical Overview page](https://www.intel.com/content/www/us/en/products/details/discrete-gpus/data-center-gpu/flex-series.html?wapkw=Intel%C2%AE%20Data%20Center%20GPU%20Flex%20Series). For detailed processor information, see the [Intel product documentation catalog page](https://www.intel.com/content/www/us/en/search.html?ws=recent#q=Intel%C2%AE%20Data%20Center%20GPU%20Flex%20Series&sort=relevancy&f:@tabfilter=[Developers]).\n',
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-icx-atsm-170-1',
          'memory.size': '256GB',
          processor: 'GPU',
          recommendedUseCase: 'GPU',
          region: 'global',
          service: 'Bare Metal'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr: 'serviceType == "ComputeAsAService" && instanceType == "bm-icx-atsm-170-1"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '.07',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'error'
      },
      {
        name: 'bm-spr',
        id: '3bc52387-da79-4947-a562-ab7a88c38e1f',
        created: '2023-07-06T17:59:32Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '61befbee-0607-47c5-b140-c6b4ea10b9da',
        description: '112 cores, 2 sockets, 256 GB memory, 2 TB disk',
        metadata: {
          category: 'singlenode',
          'cpu.cores': '112',
          'cpu.sockets': '2',
          'disks.size': '2TB',
          displayName: '4th Generation Intel® Xeon® Scalable processors',
          'family.displayDescription': 'Virtual machines and Bare Metal servers available',
          'family.displayName': '4th Generation Intel® Xeon® Scalable processors',
          information:
            'For a Xeon® processor overview see the [Intel Innovation presentation](https://www.intel.com/content/www/us/en/developer/articles/technical/fourth-generation-xeon-scalable-family-overview.html?wapkw=4th%20gen%20intel%20xeon%20scalable%20processors#gs.xc9pz8) and the [Technical Overview page](https://www.intel.com/content/www/us/en/developer/articles/technical/fourth-generation-xeon-scalable-family-overview.html?wapkw=4th%20gen%20intel%20xeon%20scalable%20processors#gs.xc9pz8). For detailed processor information see the [Intel product documentation catalog](https://www.intel.com/content/www/us/en/search.html?ws=text#q=4th%20gen&sort=relevancy&f:@disclosure=[Public]&f:@tabfilter=[Developers]&f:@stm_10385_en=[Processors,Intel%C2%AE%20Xeon%C2%AE%20Processors,4th%20Generation%20Intel%C2%AE%20Xeon%C2%AE%20Scalable%20Processors]), [Intel Accelerator Engine page](https://www.intel.com/content/www/us/en/products/docs/accelerator-engines/overview.html#multiasset_copy), and the [Accelerator e-guide](https://www.intel.com/content/www/us/en/now/xeon-accelerated/accelerators-eguide.html).\n',
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-spr',
          'memory.size': '256GB',
          processor: 'CPU',
          recommendedUseCase: 'Core compute',
          region: 'global',
          service: 'Bare Metal'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr: 'serviceType == "ComputeAsAService" && instanceType == "bm-spr"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '.18',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'error'
      },
      {
        name: 'bm-spr-atsm-170-1',
        id: '3bc52387-da79-4947-a562-ab7a88c38e13',
        created: '2023-07-11T20:30:20Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '61befbee-0607-47c5-b140-c6b4ea10b9da',
        description: '1 GPU with 4th Gen CPU, 112 cores, 2 sockets, 256 GB memory, 2 TB disk',
        metadata: {
          category: 'singlenode',
          'cpu.cores': '112',
          'cpu.sockets': '2',
          'disks.size': '2TB',
          displayName: 'Intel® Data Center GPU Flex Series on 4th Gen Intel® Xeon® processors – 170 series (1x)',
          'family.displayDescription': 'GPUs with 4th Gen CPU, 2 sockets',
          'family.displayName': 'Intel® Data Center GPU Flex Series on latest Intel® Xeon® processors',
          highlight: '1 GPU with 4th Gen CPU',
          information:
            'For details on the Intel® Data Center GPU Flex Series processor, see the [Technical Overview page](https://www.intel.com/content/www/us/en/products/details/discrete-gpus/data-center-gpu/flex-series.html?wapkw=Intel%C2%AE%20Data%20Center%20GPU%20Flex%20Series). For detailed processor information, see the [Intel product documentation catalog page](https://www.intel.com/content/www/us/en/search.html?ws=recent#q=Intel%C2%AE%20Data%20Center%20GPU%20Flex%20Series&sort=relevancy&f:@tabfilter=[Developers]).\n',
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-spr-atsm-170-1',
          'memory.size': '256GB',
          processor: 'GPU',
          recommendedUseCase: 'GPU',
          region: 'global',
          service: 'Bare Metal'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr: 'serviceType == "ComputeAsAService" && instanceType == "bm-spr-atsm-170-1"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '.07',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'error'
      },
      {
        name: 'bm-spr-hbm',
        id: '3bc52387-da79-4947-a562-ab7a88c38e1b',
        created: '2023-07-11T20:30:20Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '61befbee-0607-47c5-b140-c6b4ea10b9da',
        description: '112 cores, 2 sockets, 2 TB disk (no DDR5 memory), SNC4 enabled',
        metadata: {
          category: 'singlenode',
          'cpu.cores': '112',
          'cpu.sockets': '2',
          'disks.size': '2TB',
          displayName:
            'Intel® Xeon® processors, codenamed Sapphire Rapids with high bandwidth memory (HBM) – HBM-only mode',
          'family.displayDescription': '2 sockets, 2 TB disk, SNC4 enabled',
          'family.displayName': 'Intel® Xeon® processors, codenamed Sapphire Rapids with high bandwidth memory (HBM)',
          highlight: 'SNC4 enabled',
          information:
            'For details on the Intel® Xeon® Max Series processor see the [Technical Overview page](https://www.intel.com/content/www/us/en/products/details/processors/xeon/max-series.html?wapkw=Xeon%20CPU%20Max). For detailed processor information see the [Intel product documentation catalog page](https://www.intel.com/content/www/us/en/search.html?ws=recent#q=Intel%C2%AE%20Data%20Center%20GPU%20Flex%20Series&sort=relevancy&f:@tabfilter=[Developers]).\n',
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-spr-hbm',
          'memory.size': '0GB',
          processor: 'CPU',
          recommendedUseCase: 'HPC',
          region: 'global',
          service: 'Bare Metal'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr: 'serviceType == "ComputeAsAService" && instanceType == "bm-spr-hbm"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '.23',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'error'
      },
      {
        name: 'bm-spr-hbm-f',
        id: '3bc52387-da79-4947-a562-ab7a88c38e1a',
        created: '2023-07-11T20:30:21Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '61befbee-0607-47c5-b140-c6b4ea10b9da',
        description: '112 cores, 2 sockets, 2 TB disk, 256 Gb memory, SNC4 enabled',
        metadata: {
          category: 'singlenode',
          'cpu.cores': '112',
          'cpu.sockets': '2',
          'disks.size': '2TB',
          displayName:
            'Intel® Xeon® processors, codenamed Sapphire Rapids with high bandwidth memory (HBM) – Flat mode',
          'family.displayDescription': '2 sockets, 2 TB disk, SNC4 enabled',
          'family.displayName': 'Intel® Xeon® processors, codenamed Sapphire Rapids with high bandwidth memory (HBM)',
          highlight: 'SNC4 enabled',
          information:
            'For details on the Intel® Xeon® Max Series processor see the [Technical Overview page](https://www.intel.com/content/www/us/en/products/details/processors/xeon/max-series.html?wapkw=Xeon%20CPU%20Max). For detailed processor information see the [Intel product documentation catalog page](https://www.intel.com/content/www/us/en/search.html?ws=recent#q=Intel%C2%AE%20Data%20Center%20GPU%20Flex%20Series&sort=relevancy&f:@tabfilter=[Developers]).\n',
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-spr-hbm-f',
          'memory.size': '256GB',
          processor: 'CPU',
          recommendedUseCase: 'HPC',
          region: 'global',
          service: 'Bare Metal'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr: 'serviceType == "ComputeAsAService" && instanceType == "bm-spr-hbm-f"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '.23',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'error'
      },
      {
        name: 'bm-spr-pvc-1100-1',
        id: '3bc52387-da79-4947-a562-ab7a88c38e1c',
        created: '2023-07-11T20:30:21Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '61befbee-0607-47c5-b140-c6b4ea10b9da',
        description: '1 GPU with 4th Gen CPU, 112 cores, 2 sockets, 256 GB memory, 2 TB disk',
        metadata: {
          category: 'singlenode',
          'cpu.cores': '112',
          'cpu.sockets': '2',
          'disks.size': '2TB',
          displayName: 'Intel® Max Series GPU (PVC) on 4th Gen Intel® Xeon® processors – 1100 series (1x)',
          'family.displayDescription': '4th Gen CPU, 2 sockets, 256 GB memory, 2 TB disk',
          'family.displayName': 'Intel® Max Series GPU (PVC)',
          highlight: '1 GPU with 4th Gen CPU',
          information:
            'For details on the Intel® Data Center Max Series processor, see the [Technical Overview page](https://www.intel.com/content/www/us/en/developer/articles/technical/intel-data-center-gpu-max-series-overview.html?wapkw=Xeon%20Max#gs.xcf9of). For detailed processor information, see the [Intel product documentation catalog page](https://www.intel.com/content/www/us/en/search.html?ws=recent#q=Intel%C2%AE%20Data%20Center%20GPU%20Max%20Series&sort=relevancy&f:@tabfilter=[Developers]).\n',
          instanceCategories: 'BareMetalHost',
          instanceType: 'bm-spr-pvc-1100-1',
          'memory.size': '256GB',
          processor: 'GPU',
          recommendedUseCase: 'GPU',
          region: 'global',
          service: 'Bare Metal'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr: 'serviceType == "ComputeAsAService" && instanceType == "bm-spr-pvc-1100-1"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '.30',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'error'
      },
      {
        name: 'vm-spr-lrg',
        id: '3bc52387-da79-4947-a562-ab7a88c38ef1',
        created: '2023-07-06T17:59:32Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '61befbee-0607-47c5-b140-c6b4ea10b9da',
        description: '32 cores, 64 GB memory, 64 GB disk',
        metadata: {
          category: 'singlenode',
          'cpu.cores': '32',
          'disks.size': '64GB',
          displayName: 'Large VM - Intel® Xeon 4th Gen ® Scalable processor',
          'family.displayDescription': 'Virtual machines and Bare Metal servers available',
          'family.displayName': '4th Generation Intel® Xeon® Scalable processors',
          highlight: '32 cores',
          information:
            'For a Xeon® processor overview see the [Intel Innovation presentation](https://www.intel.com/content/www/us/en/developer/articles/technical/fourth-generation-xeon-scalable-family-overview.html?wapkw=4th%20gen%20intel%20xeon%20scalable%20processors#gs.xc9pz8) and the [Technical Overview page](https://www.intel.com/content/www/us/en/developer/articles/technical/fourth-generation-xeon-scalable-family-overview.html?wapkw=4th%20gen%20intel%20xeon%20scalable%20processors#gs.xc9pz8). For detailed processor information see the [Intel product documentation catalog](https://www.intel.com/content/www/us/en/search.html?ws=text#q=4th%20gen&sort=relevancy&f:@disclosure=[Public]&f:@tabfilter=[Developers]&f:@stm_10385_en=[Processors,Intel%C2%AE%20Xeon%C2%AE%20Processors,4th%20Generation%20Intel%C2%AE%20Xeon%C2%AE%20Scalable%20Processors]), [Intel Accelerator Engine page](https://www.intel.com/content/www/us/en/products/docs/accelerator-engines/overview.html#multiasset_copy), and the [Accelerator e-guide](https://www.intel.com/content/www/us/en/now/xeon-accelerated/accelerators-eguide.html).\n',
          instanceCategories: 'VirtualMachine',
          instanceType: 'vm-spr-lrg',
          'memory.size': '64GB',
          processor: 'CPU',
          recommendedUseCase: 'Core compute',
          region: 'global',
          service: 'Virtual Machine'
        },
        eccn: 'EAR99',
        pcq: '19621',
        matchExpr: 'serviceType == "ComputeAsAService" && instanceType == "vm-spr-lrg"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '.07',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'error'
      },
      {
        name: 'vm-spr-med',
        id: '3bc52387-da79-4947-a562-ab7a88c38ee1',
        created: '2023-07-06T17:59:32Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '61befbee-0607-47c5-b140-c6b4ea10b9da',
        description: '16 cores, 32 GB memory, 32 GB disk',
        metadata: {
          category: 'singlenode',
          'cpu.cores': '16',
          'disks.size': '32GB',
          displayName: 'Medium VM - Intel® Xeon 4th Gen ® Scalable processor',
          'family.displayDescription': 'Virtual machines and Bare Metal servers available',
          'family.displayName': '4th Generation Intel® Xeon® Scalable processors',
          highlight: '16 cores',
          information:
            'For a Xeon® processor overview see the [Intel Innovation presentation](https://www.intel.com/content/www/us/en/developer/articles/technical/fourth-generation-xeon-scalable-family-overview.html?wapkw=4th%20gen%20intel%20xeon%20scalable%20processors#gs.xc9pz8) and the [Technical Overview page](https://www.intel.com/content/www/us/en/developer/articles/technical/fourth-generation-xeon-scalable-family-overview.html?wapkw=4th%20gen%20intel%20xeon%20scalable%20processors#gs.xc9pz8). For detailed processor information see the [Intel product documentation catalog](https://www.intel.com/content/www/us/en/search.html?ws=text#q=4th%20gen&sort=relevancy&f:@disclosure=[Public]&f:@tabfilter=[Developers]&f:@stm_10385_en=[Processors,Intel%C2%AE%20Xeon%C2%AE%20Processors,4th%20Generation%20Intel%C2%AE%20Xeon%C2%AE%20Scalable%20Processors]), [Intel Accelerator Engine page](https://www.intel.com/content/www/us/en/products/docs/accelerator-engines/overview.html#multiasset_copy), and the [Accelerator e-guide](https://www.intel.com/content/www/us/en/now/xeon-accelerated/accelerators-eguide.html).\n',
          instanceCategories: 'VirtualMachine',
          instanceType: 'vm-spr-med',
          'memory.size': '32GB',
          processor: 'CPU',
          recommendedUseCase: 'Core compute',
          region: 'global',
          service: 'Virtual Machine'
        },
        eccn: 'EAR99',
        pcq: '19621',
        matchExpr: 'serviceType == "ComputeAsAService" && instanceType == "vm-spr-med"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '.03',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'error'
      },
      {
        name: 'vm-spr-sml',
        id: '3bc52387-da79-4947-a562-ab7a88c38ea1',
        created: '2023-07-06T17:59:32Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '61befbee-0607-47c5-b140-c6b4ea10b9da',
        description: '8 cores, 16 GB memory, 20 GB disk',
        metadata: {
          category: 'singlenode',
          'cpu.cores': '8',
          'disks.size': '20GB',
          displayName: 'Small VM - Intel® Xeon 4th Gen ® Scalable processor',
          'family.displayDescription': 'Virtual machines and Bare Metal servers available',
          'family.displayName': '4th Generation Intel® Xeon® Scalable processors',
          highlight: '8 cores',
          information:
            'For a Xeon® processor overview see the [Intel Innovation presentation](https://www.intel.com/content/www/us/en/developer/articles/technical/fourth-generation-xeon-scalable-family-overview.html?wapkw=4th%20gen%20intel%20xeon%20scalable%20processors#gs.xc9pz8) and the [Technical Overview page](https://www.intel.com/content/www/us/en/developer/articles/technical/fourth-generation-xeon-scalable-family-overview.html?wapkw=4th%20gen%20intel%20xeon%20scalable%20processors#gs.xc9pz8). For detailed processor information see the [Intel product documentation catalog](https://www.intel.com/content/www/us/en/search.html?ws=text#q=4th%20gen&sort=relevancy&f:@disclosure=[Public]&f:@tabfilter=[Developers]&f:@stm_10385_en=[Processors,Intel%C2%AE%20Xeon%C2%AE%20Processors,4th%20Generation%20Intel%C2%AE%20Xeon%C2%AE%20Scalable%20Processors]), [Intel Accelerator Engine page](https://www.intel.com/content/www/us/en/products/docs/accelerator-engines/overview.html#multiasset_copy), and the [Accelerator e-guide](https://www.intel.com/content/www/us/en/now/xeon-accelerated/accelerators-eguide.html).\n',
          instanceCategories: 'VirtualMachine',
          instanceType: 'vm-spr-sml',
          'memory.size': '16GB',
          processor: 'CPU',
          recommendedUseCase: 'Core compute',
          region: 'global',
          service: 'Virtual Machine'
        },
        eccn: 'EAR99',
        pcq: '19621',
        matchExpr: 'serviceType == "ComputeAsAService" && instanceType == "vm-spr-sml"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '.02',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'error'
      },
      {
        name: 'vm-spr-tny',
        id: '3bc52387-da79-4947-a562-ab7a88c38eb1',
        created: '2023-07-06T17:59:32Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '61befbee-0607-47c5-b140-c6b4ea10b9da',
        description: '4 cores, 8 GB memory, 10 GB disk',
        metadata: {
          category: 'singlenode',
          'cpu.cores': '4',
          'disks.size': '10GB',
          displayName: 'Tiny VM - Intel® Xeon 4th Gen ® Scalable processor',
          'family.displayDescription': 'Virtual machines and Bare Metal servers available',
          'family.displayName': '4th Generation Intel® Xeon® Scalable processors',
          highlight: '4 cores',
          information:
            'For a Xeon® processor overview see the [Intel Innovation presentation](https://www.intel.com/content/www/us/en/developer/articles/technical/fourth-generation-xeon-scalable-family-overview.html?wapkw=4th%20gen%20intel%20xeon%20scalable%20processors#gs.xc9pz8) and the [Technical Overview page](https://www.intel.com/content/www/us/en/developer/articles/technical/fourth-generation-xeon-scalable-family-overview.html?wapkw=4th%20gen%20intel%20xeon%20scalable%20processors#gs.xc9pz8). For detailed processor information see the [Intel product documentation catalog](https://www.intel.com/content/www/us/en/search.html?ws=text#q=4th%20gen&sort=relevancy&f:@disclosure=[Public]&f:@tabfilter=[Developers]&f:@stm_10385_en=[Processors,Intel%C2%AE%20Xeon%C2%AE%20Processors,4th%20Generation%20Intel%C2%AE%20Xeon%C2%AE%20Scalable%20Processors]), [Intel Accelerator Engine page](https://www.intel.com/content/www/us/en/products/docs/accelerator-engines/overview.html#multiasset_copy), and the [Accelerator e-guide](https://www.intel.com/content/www/us/en/now/xeon-accelerated/accelerators-eguide.html).\n',
          instanceCategories: 'VirtualMachine',
          instanceType: 'vm-spr-tny',
          'memory.size': '8GB',
          processor: 'CPU',
          recommendedUseCase: 'Core compute',
          region: 'global',
          service: 'Virtual Machine'
        },
        eccn: 'EAR99',
        pcq: '19621',
        matchExpr: 'serviceType == "ComputeAsAService" && instanceType == "vm-spr-tny"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '.00',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'error'
      }
    ]
  }
}

export const mockProductCatalog = () => {
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/products'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseProductCatalogStore()
      })
    )
}

export const mockEmptyProductCatalog = () => {
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/products'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: []
      })
    )
}
