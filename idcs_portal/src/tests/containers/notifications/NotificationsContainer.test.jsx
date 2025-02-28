// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import { clearAxiosMock } from '../../mocks/utils'
import { mockStandardUser } from '../../mocks/authentication/authHelper'
import NotificationsContainer from '../../../containers/notifications/NotificationsContainer'
import { mockNotifications, mockNotificationsNull } from '../../mocks/billing/notifications'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from 'react-query'
import idcConfig from '../../../config/configurator'

describe('Notification Container', () => {
  const TestComponent1 = () => {
    return (
      <QueryClientProvider client={new QueryClient()} contextSharing={true}>
        <BrowserRouter>
          <NotificationsContainer isWidget={false} />
        </BrowserRouter>
      </QueryClientProvider>
    )
  }
  const TestComponent2 = () => {
    return (
      <QueryClientProvider client={new QueryClient()} contextSharing={true}>
        <BrowserRouter>
          <NotificationsContainer isWidget={true} />
        </BrowserRouter>
      </QueryClientProvider>
    )
  }
  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_NOTIFICATIONS = 1
  })

  beforeEach(() => {
    clearAxiosMock()
  })

  beforeEach(() => {
    mockStandardUser()
  })

  beforeEach(() => {
    mockNotifications()
  })

  afterAll(() => {
    idcConfig.REACT_APP_FEATURE_NOTIFICATIONS = 0
  })

  it('Render notification container without a limit (page) is successful', async () => {
    render(<TestComponent1 />)
    const { length } = await screen.findAllByTestId('notificationItem')
    expect(length).toBe(5)
    expect(screen.queryByTestId('goToNotificationsButton')).toBeNull()
  })

  it('Render notification container with a limit (widget) is successful', async () => {
    render(<TestComponent2 />)
    const { length } = await screen.findAllByTestId('notificationItem')
    expect(length).toBe(4)
    expect(await screen.findByTestId('goToNotificationsButton')).toBeVisible()
  })

  it('Render notification container component should show no notifications message is successful', async () => {
    mockNotificationsNull()
    render(<TestComponent2 />)
    expect(await screen.queryByTestId('goToNotificationsButton')).toBeNull()
    expect(await screen.findByTestId('data-view-empty')).toBeVisible()
  })
})
