// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from 'react-query'
import { BrowserRouter } from 'react-router-dom'
import { clearAxiosMock } from '../../mocks/utils'
import { mockStandardUser } from '../../mocks/authentication/authHelper'
import Header from '../../../components/header/Header'
import useUserStore from '../../../store/userStore/UserStore'
import idcConfig from '../../../config/configurator'
import { mockNotifications } from '../../mocks/billing/notifications'

const TestComponent = ({ isMemberOfOtherAccounts = false }) => {
  const user = useUserStore((state) => state.user)
  const mockUser = { ...user }
  mockUser.hasInvitations = isMemberOfOtherAccounts

  return (
    <QueryClientProvider client={new QueryClient()} contextSharing={true}>
      <BrowserRouter>
        <Header userDetails={mockUser} pathname={'/'} />
      </BrowserRouter>
    </QueryClientProvider>
  )
}

describe('Header Component', () => {
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

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_NOTIFICATIONS = 1
  })

  beforeEach(() => {
    mockStandardUser()
    mockNotifications()
  })

  afterAll(() => {
    idcConfig.REACT_APP_FEATURE_NOTIFICATIONS = 0
  })

  it('Should header top bar buttons be visible', () => {
    idcConfig.REACT_APP_FEATURE_NOTIFICATIONS = 1
    render(<TestComponent history={history} />)
    expect(screen.queryByTestId('siteLogo')).toBeVisible()

    expect(screen.queryByTestId('regionIcon')).toBeVisible()
    expect(screen.queryByTestId('regionLabel')).toBeVisible()

    expect(screen.queryByTestId('helpIcon')).toBeVisible()

    expect(screen.queryByTestId('notificationIcon')).toBeVisible()

    expect(screen.queryByTestId('userIcon')).toBeVisible()
  })
})
