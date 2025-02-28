import { render, screen, waitFor } from '@testing-library/react'
import AuthWrapper from '../../../utility/wrapper/AuthWrapper'
import { clearAxiosMock } from '../../mocks/utils'
import { mockUseMsalAuthentication } from '../../../setupTests'
import { mockIntelUser } from '../../mocks/authentication/authHelper'

const TestComponent = () => {
  return <AuthWrapper>Test Child component</AuthWrapper>
}

describe('Auth Wrapper', () => {
  beforeEach(() => {
    clearAxiosMock()
  })

  beforeEach(() => {
    mockIntelUser()
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
})
