import { configure } from '@testing-library/react'
import '@testing-library/jest-dom'
import configMap from '../public/configMap.json'

window._env_ = configMap
window.open = jest.fn()

export const mockMsalAccount = {
  environment: 'consumer.intel.com',
  username: '',
  idTokenClaims: {
    email: 'testAccount@intel.com',
    iss: 'https://login.microsoftonline.com/46c98d88-e344-4ed4-8496-4ed7712e255d/v2.0',
    name: 'TestUserName',
    preferred_username: 'TestUserName',
    roles: ['IDC.Admin'],
    tid: '46c98d88-e344-4ed4-8496-4ed7712e255d',
    ver: '2.0'
  }
}

configure({
  testIdAttribute: 'intc-id'
})

jest.mock('./AuthConfig.js', () => ({
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
