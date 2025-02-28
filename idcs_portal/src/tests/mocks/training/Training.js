// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { when } from 'jest-when'
import { mockAxios } from '../../../setupTests'
import { mockBaseEnrollResponse } from '../authentication/authHelper'
import idcConfig from '../../../config/configurator'

export const mockBaseTrainingStore = () => {
  return {
    products: [
      {
        name: 'ai-numba',
        id: '5ba105a9-1490-47fc-a5cc-5a23570de3eb',
        created: '2023-08-16T22:46:16Z',
        vendorId: '4015bb99-0522-4387-b47e-c821596dc735',
        familyId: '07af0540-2fda-11ee-be56-0242ac120002',
        description: 'Add meaningful description here',
        metadata: {
          category: 'Training',
          'detail.gettingStarted': `Sign in to ${idcConfig.REACT_APP_CONSOLE_LONG_NAME}, select One Click Log In for JupyterLab, and then (if needed) select Launch Server.\nOpen the AI_Numba_dpex_Essentials folder, and then select Welcome.ipynb.\n`,
          'detail.objectives.audience':
            'This course is designed for Python developers who want to learn the basics of data parallel Python for data parallel and heterogeneous hardware (such as CPU and GPU) without leaving the Python ecosystem or compromising on performance.\n',
          'detail.objectives.expectations': `Practice the essential concepts and features of Data Parallel Extension for Numba using live sample code on the ${idcConfig.REACT_APP_CONSOLE_LONG_NAME}.\n`,
          'detail.overview':
            'Python* has become a useful tool in advancing scientific research and computation with a rich ecosystem of open source packages for mathematics, science, and engineering. Python is anchored on the performant numerical computation on arrays and matrices, data analysis, and visualization capabilities.\n\nData Parallel Extension for Numba* is a stand-alone extension to the Numba just-in-time (JIT) compiler and adds SYCL* programming capabilities to Numba. This extension is packaged as part of Intel® Distribution for Python*, which is included with the Intel® AI Analytics Toolkit.\n\nThis data parallel Python course demonstrates high-performing code targeting Intel® XPUs using Python. Developers learn how to take advantage of heterogeneous architectures and speed up applications without using low-level proprietary programming APIs.\n',
          'detail.prerrequisites':
            'Complete the first three modules of the Essentials of SYCL course: \n\nIntroduction to SYCL\nSYCL Program Structure\nSYCL Unified Shared Memory\nLearn about the Intel Distribution for Python."\n',
          'detail.shortDesc': '',
          displayCatalogDesc: '',
          displayName: 'Test',
          'family.displayDescription': '',
          'family.displayName': 'AI',
          launch: 'jupyterlab',
          region: 'global',
          service: 'IDC Training'
        },
        eccn: 'EAR99',
        pcq: '19513',
        matchExpr: 'serviceType == "TrainingAsAService" && trainingType == "ai-numba"',
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

export const mockTrainings = () => {
  const { cloudAccountId } = mockBaseEnrollResponse()
  when(mockAxios.post)
    .calledWith(expect.stringContaining(`/cloudaccounts/${cloudAccountId}/instances`), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseTrainingStore()
      })
    )
}

export const mockTraining = () => {
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/products'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: mockBaseTrainingStore()
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

export const mockTrainingEmpty = () => {
  when(mockAxios.post)
    .calledWith(expect.stringContaining('/products'), expect.anything())
    .mockImplementation(() =>
      Promise.resolve({
        data: { products: [] }
      })
    )
}

export const mockExpiry = () => {
  const { cloudAccountId } = mockBaseEnrollResponse()
  when(mockAxios.get)
    .calledWith(expect.stringContaining(`/cloudaccounts/${cloudAccountId}/trainings/expiry`))
    .mockImplementation(() =>
      Promise.resolve({
        data: { expiryDate: '2023-10-12T00:00:00Z' }
      })
    )
}
