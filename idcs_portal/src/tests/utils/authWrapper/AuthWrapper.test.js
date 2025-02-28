// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, waitFor } from '@testing-library/react'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { clearAxiosMock } from '../../mocks/utils'
import { mockNonAcceptedTCsUser, mockStandardUser } from '../../mocks/authentication/authHelper'
import { mockUseMsalAuthentication } from '../../../setupTests'
import { mockUserCloudAccountsList } from '../../mocks/profile/profile'
import idcConfig from '../../../config/configurator'
import { mockVendorsApi } from '../../mocks/shared/vendors'
import { mockConsoleUIList } from '../../mocks/consoleUIs/ConsoleUIs'

const TestComponent = () => {
  return <AuthWrapper>Test Child component</AuthWrapper>
}

describe('Auth Wrapper', () => {
  beforeEach(() => {
    clearAxiosMock()
  })

  beforeEach(() => {
    mockStandardUser()
    mockUserCloudAccountsList()
  })

  it('Render child component is MSAL auth is successful', async () => {
    render(<TestComponent></TestComponent>)
    expect(await screen.findByText('Test Child component')).toBeInTheDocument()
  })

  it('Show error page when azure auth fails', async () => {
    // Set error
    mockUseMsalAuthentication.error = {
      errorCode: 1,
      errorMessage: 'Occurs an Error with MSAL'
    }
    render(<TestComponent></TestComponent>)
    expect(await screen.findByText('Occurs an Error with MSAL')).toBeInTheDocument()
    // Clear error
    mockUseMsalAuthentication.error = null
  })

  it(`Show Account selection page when user is member of an account and
   not default cloudaccount is selected`, async () => {
    clearAxiosMock()
    mockStandardUser(true)
    mockUserCloudAccountsList()
    render(<TestComponent></TestComponent>)
    expect(await screen.findByText('Select a cloud account')).toBeInTheDocument()
  })

  it('Retry login when error code is interaction_required', async () => {
    // Set error
    mockUseMsalAuthentication.error = {
      errorCode: 'interaction_required',
      errorMessage: ''
    }
    render(<TestComponent></TestComponent>)
    await waitFor(() => {
      expect(mockUseMsalAuthentication.login).toHaveBeenCalled()
    })
    // Clear error
    mockUseMsalAuthentication.error = null
  })

  it('Retry login when error code is monitor_window_timeout', async () => {
    // Set error
    mockUseMsalAuthentication.error = {
      errorCode: 'monitor_window_timeout',
      errorMessage: ''
    }
    render(<TestComponent></TestComponent>)
    await waitFor(() => {
      expect(mockUseMsalAuthentication.login).toHaveBeenCalled()
    })
    // Clear error
    mockUseMsalAuthentication.error = null
  })

  it('Retry login when error code is state_not_found', async () => {
    // Set error
    mockUseMsalAuthentication.error = {
      errorCode: 'state_not_found',
      errorMessage: ''
    }
    render(<TestComponent></TestComponent>)
    await waitFor(() => {
      expect(mockUseMsalAuthentication.login).toHaveBeenCalled()
    })
    // Clear error
    mockUseMsalAuthentication.error = null
  })

  it('Retry login when user does not have not an existing session error', async () => {
    // More info at: https://github.com/AzureAD/microsoft-authentication-library-for-js/issues/349
    mockUseMsalAuthentication.error = {
      errorCode: '',
      errorMessage:
        'AADB2C90077: User does not have an existing session and request prompt parameter has a value of "None"'
    }
    render(<TestComponent></TestComponent>)
    await waitFor(() => {
      expect(mockUseMsalAuthentication.login).toHaveBeenCalled()
    })
    // Clear error
    mockUseMsalAuthentication.error = null
  })

  it('Render ConsoleAccessDenied page when whitelist feature branch is on and user is not whitelisted', async () => {
    idcConfig.REACT_APP_FEATURE_UX_WHITELIST = 1
    render(<TestComponent></TestComponent>)
    expect(await screen.findByText('Access Restricted')).toBeInTheDocument()
    idcConfig.REACT_APP_FEATURE_UX_WHITELIST = 0
  })

  it('Render child when whitelist feature branch is on and user is whitelisted', async () => {
    idcConfig.REACT_APP_FEATURE_UX_WHITELIST = 1
    mockVendorsApi()
    mockConsoleUIList()
    render(<TestComponent></TestComponent>)
    expect(await screen.findByText('Test Child component')).toBeInTheDocument()
    idcConfig.REACT_APP_FEATURE_UX_WHITELIST = 0
  })

  it('Render Accept term and conditions when enroll action response is ENROLL_ACTION_TC', async () => {
    clearAxiosMock()
    mockNonAcceptedTCsUser()
    mockUserCloudAccountsList()
    render(<TestComponent></TestComponent>)
    expect(await screen.findByText('Terms and conditions')).toBeInTheDocument()
  })
})
