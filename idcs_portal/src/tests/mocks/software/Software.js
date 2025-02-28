// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockAxios } from '../../../setupTests'
import { mockBaseEnrollResponse } from '../authentication/authHelper'

export const mockBaseSoftwareStore = () => {
  return {
    products: [
      {
        name: 'sw-intc-oneapi-kit-rendering',
        id: '7cfc5a1f-a92d-49be-8cbb-4f627af78fb9',
        created: '2023-11-10T17:16:48Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '368e03ac-7f44-11ee-b962-0242ac120002',
        description: 'Powerful Libraries for High-Fidelity Rendering and Visualization Applications',
        metadata: {
          Documentation:
            '- [Installation Guide](https://www.intel.com/content/www/us/en/developer/articles/guide/installation-guide-for-oneapi-toolkits.html)\n- Get Started Guides: [Linux](https://www.intel.com/content/www/us/en/docs/oneapi-rendering-toolkit/get-started-guide-linux/current/overview.html) | [Windows](https://www.intel.com/content/www/us/en/docs/oneapi-rendering-toolkit/get-started-guide-windows/current/overview.html) | [macOS](https://www.intel.com/content/www/us/en/docs/oneapi-rendering-toolkit/get-started-guide-macos/current/overview.html)\n- [Release Notes](https://www.intel.com/content/www/us/en/developer/articles/release-notes/intel-oneapi-rendering-toolkit-release-notes.html)\n\n[View All Documentation](https://www.intel.com/content/www/us/en/developer/tools/oneapi/rendering-toolkit-documentation.html)\n',
          access: 'open',
          billingEnable: 'false',
          category: 'Software',
          components:
            '- [Intel® Embree](https://www.embree.org/index.html)\n- [Intel® Implicit SPMD Program Compiler (Intel® ISPC)](https://ispc.github.io/)\n- [Intel® Open Image Denoise](https://openimagedenoise.github.io/)\n- [Intel® Open Volume Kernel Library (Intel® Open VKL)](https://www.openvkl.org/)\n- [Intel® Open Path Guiding Library (Intel® Open PGL)](https://github.com/OpenPathGuidingLibrary/openpgl)\n- [Intel® OSPRay](https://www.ospray.org/)\n- [Intel® OSPRay Studio](https://www.ospray.org/ospray_studio/)\n- [Intel® OSPRay for Hydra* (Open Source GitHub*)](https://github.com/ospray/hdospray)\n- **Rendering Toolkit Utilities**: The included Render Kit Superbuild utility automatically downloads the Render Kit source code, Intel® oneAPI Threading Building Blocks (oneTBB) binaries, Intel ISPC binaries, and build binaries for each component.\n',
          'detail.downloadURL':
            'https://www.intel.com/content/www/us/en/developer/tools/oneapi/rendering-toolkit-download.html',
          'detail.features':
            '- Efficient deployment across parallel processing architectures and platforms\n- Access to all system memory space for even the largest datasets\n- Improved visual fidelity via ray tracing with global illumination\n- Cost-efficient, interactive performance for any data size\n- High-performance, deep learning-based denoising\n',
          'detail.helpURL': 'https://www.intel.com/content/www/us/en/developer/tools/oneapi/support.html',
          'detail.jupyterlab': '/Training/AI/GenAI/simple_llm_inference.ipynb',
          'detail.objectives.audience': '',
          'detail.overview':
            '### Photorealistic Rendering That Scales\nThe Intel® oneAPI Rendering Toolkit (Render Kit) is a powerful set of open source rendering, ray tracing, denoising, and path guiding libraries for AI synthetic data generation, digital twins, high-fidelity and high-performance visualization, and immersive content creation. Achieve optimized rendering performance with these libraries and Intel® CPU and GPU hardware, comprising a scalable solutions stack.\n\nWho needs this product?\n\n- **AI, robotics, and autonomous vehicle developers** who rely on synthetic data generation for simulation and perception.\n- **Machine learning and deep learning applications** that rely on synthetic data to represent any situation, validate mathematical models, or train machine learning models.\n- **Research scientists** who use in-situation simulation or require the highest fidelity images in an HPC distributed rendering solution capable of handling multi-terabyte-sized datasets.\n- **Creators, developers, and artists** who need to create immersive experiences and who would benefit from a unified workflow that scales from workstation to rendering farm.\n- **Product designers and engineers** who use accurate digital twins in product life cycle management.\n',
          'detail.productURL': 'https://www.intel.com/content/www/us/en/developer/tools/oneapi/rendering-toolkit.html',
          'detail.shortDesc': 'Powerful Libraries for High-Fidelity Rendering and Visualization Applications\n',
          'detail.useCases': 'ai, gpu',
          displayCatalogDesc: 'Powerful Libraries for High-Fidelity Rendering and Visualization Applications',
          displayInHomepage: 'false',
          displayName: 'test',
          displayPicture: 'oneAPI.svg',
          'family.displayDescription': '',
          'family.displayName': 'oneAPI Kits',
          homepageDisplayGroup: 'Kits',
          platforms: '',
          region: 'us-staging-1',
          service: 'IDC Software'
        },
        eccn: '',
        pcq: '',
        matchExpr: 'serviceType == "SoftwareAsAService" && trainingID == "sw-intc-oneapi-kit-rendering"',
        rates: [
          {
            accountType: 'ACCOUNT_TYPE_INTEL',
            unit: 'RATE_UNIT_DOLLARS_PER_MINUTE',
            rate: '0.00',
            usageExpr: 'time – previous.time'
          }
        ],
        status: 'error'
      }
    ]
  }
}

export const mockSoftwareList = () => {
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/products'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseSoftwareStore()
      })
    )
}

export const mockEnrollTraining = () => {
  const { cloudAccountId } = mockBaseEnrollResponse()
  when(mockAxios.post)
    .calledWith(expect.stringContaining(`/cloudaccounts/${cloudAccountId}/trainings`), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: {
          sshLoginInfo: 'Log successfully',
          message: 'Log ok',
          expiryDate: null
        }
      })
    )
}

export const mockSoftwareEmpty = () => {
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/products'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: { products: [] }
      })
    )
}
