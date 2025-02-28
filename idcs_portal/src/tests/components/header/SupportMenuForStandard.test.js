// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { clearAxiosMock } from '../../mocks/utils'
import { mockStandardUser } from '../../mocks/authentication/authHelper'
import Header from '../../../components/header/Header'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import useUserStore from '../../../store/userStore/UserStore'

const TestComponent = () => {
  const user = useUserStore((state) => state.user)

  return (
    <>
      <BrowserRouter>
        <AuthWrapper>
          <Header userDetails={user} pathname={'/'} />
        </AuthWrapper>
      </BrowserRouter>
    </>
  )
}

describe('Support Menu Standard user', () => {
  beforeAll(() => {
    sessionStorage.clear()
    Object.defineProperty(window, 'liveagent', {
      configurable: true,
      enumerable: true,
      value: {
        // Mock as if we reload config because of reload beign called
        init: () => jest.fn()
      }
    })
  })

  beforeEach(() => {
    clearAxiosMock()
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: jest.fn().mockImplementation((query) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: jest.fn(), // deprecated
        removeListener: jest.fn(), // deprecated
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn()
      }))
    })
  })

  beforeEach(() => {
    mockStandardUser()
  })

  afterAll(() => {
    sessionStorage.clear()
    Object.defineProperty(window, 'liveagent', {
      configurable: true,
      enumerable: true,
      value: null
    })
  })

  const waitForSupportMenu = async () => {
    await screen.findByTestId('help-menu')
  }

  it('Should show "Documentation" menu option', async () => {
    render(<TestComponent />)
    await waitForSupportMenu()
    expect(await screen.findByTestId('help-menu-Browse-documentation')).toBeVisible()
  })

  it('Should show "Community" menu  option', async () => {
    render(<TestComponent />)
    await waitForSupportMenu()
    expect(await screen.findByTestId('help-menu-Community')).toBeVisible()
  })

  it('Should show "Submit a ticket" menu option', async () => {
    render(<TestComponent />)
    await waitForSupportMenu()
    expect(await screen.findByTestId('help-menu-Submit-a-ticket')).toBeVisible()
  })

  it('Should NOT show "Contact support" menu option', async () => {
    render(<TestComponent />)
    await waitForSupportMenu()
    expect(screen.queryByTestId('help-menu-Contact-support')).toBeNull()
  })
})
