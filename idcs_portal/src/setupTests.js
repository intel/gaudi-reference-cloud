// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { configure } from '@testing-library/react'
import configMap from '../public/configMap.json'
import siteBannersObj from '../public/banners/siteBannerMessages.json'
import docData from '../public/docs/docData.json'
import '@testing-library/jest-dom'

window._env_ = configMap
window._env_.REACT_APP_SITE_BANNERS = siteBannersObj.REACT_APP_SITE_BANNERS
window._env_.REACT_APP_LEARNING_DOCS = docData
window._env_.REACT_APP_TOAST_DELAY = 500
window._env_.REACT_APP_TOAST_ERROR_DELAY = 500
window.open = jest.fn()

export const mockMsalAccount = {
  environment: 'consumer.intel.com',
  username: '',
  idTokenClaims: {
    email: 'testAccount@intel.com',
    tid: '46c98d88-e344-4ed4-8496-4ed7712e255d',
    firstName: 'TestUserName',
    lastName: 'TestLastName1 TestLastName2',
    displayName: 'TestLastName1 TestLastName2, TestName',
    idp: 'https://login.microsoftonline.com/46c98d88-e344-4ed4-8496-4ed7712e255d/v2.0',
    enterpriseId: '11735859',
    groups: ['DevCloud Console Standard']
  }
}

configure({
  testIdAttribute: 'intc-id'
})

jest.mock('./AuthConfig', () => ({
  msalInstance: {
    setActiveAccount: jest.fn(),
    addEventCallback: jest.fn(),
    removeEventCallback: jest.fn(),
    enableAccountStorageEvents: jest.fn(),
    disableAccountStorageEvents: jest.fn()
  }
}))

export const mockAxios = {
  interceptors: {
    request: { use: jest.fn(), eject: jest.fn() },
    response: { use: jest.fn(), eject: jest.fn() }
  },
  get: jest.fn(),
  post: jest.fn(),
  put: jest.fn(),
  delete: jest.fn()
}

jest.mock('axios', () => {
  return {
    create: () => mockAxios
  }
})

export const mockUseMsalAuthentication = {
  login: jest.fn(),
  error: null
}

export const azureMock = jest.mock('@azure/msal-react', () => ({
  ...jest.requireActual('@azure/msal-react'),
  useMsalAuthentication: () => mockUseMsalAuthentication,
  useMsal: () => ({
    accounts: [mockMsalAccount],
    inProgress: 'none',
    instance: {
      acquireTokenSilent: async () => 'token',
      acquireTokenRedirect: async () => {},
      setActiveAccount: jest.fn(),
      addEventCallback: jest.fn(),
      removeEventCallback: jest.fn(),
      enableAccountStorageEvents: jest.fn(),
      disableAccountStorageEvents: jest.fn()
    }
  })
}))

export const mockUseNavigate = {
  navigate: jest.fn()
}

jest.mock('react-router', () => ({
  ...jest.requireActual('react-router'),
  useNavigate: () => mockUseNavigate
}))

jest.mock('react-router-dom', () => {
  return {
    ...jest.requireActual('react-router-dom'),
    useSearchParams: () => [new URLSearchParams('region=us-dev-1'), jest.fn()]
  }
})

// this is just a little hack to silence a warning that we'll get until we
// upgrade to 16.9. See also: https://github.com/facebook/react/pull/14853
const originalError = console.error
beforeAll(() => {
  console.error = (...args) => {
    if (/Warning.*not wrapped in act/.test(args[0])) {
      return
    }
    originalError.call(console, ...args)
  }
})

afterAll(() => {
  console.error = originalError
})
