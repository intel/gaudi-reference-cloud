// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import { clearAxiosMock } from '../../mocks/utils'
import { mockStandardUser } from '../../mocks/authentication/authHelper'
import NotificationsList from '../../../components/notifications/NotificationsList'
import { BrowserRouter } from 'react-router-dom'

describe('Notification Lists', () => {
  const notification = { serviceName: 'billing', message: 'test', creation: '07/18/2023 9:54 am' }
  const notifications = [notification, notification, notification, notification, notification]
  beforeEach(() => {
    clearAxiosMock()
  })

  beforeEach(() => {
    mockStandardUser()
  })

  it('Render notification list component without a limit (page) is successful', async () => {
    render(<NotificationsList notifications={notifications} />)
    const { length } = await screen.findAllByTestId('notificationItem')
    expect(length).toBe(5)
    expect(screen.queryByTestId('goToNotificationsButton')).toBeNull()
  })

  it('Render notification list component with a limit (widget) is successful', async () => {
    render(
      <BrowserRouter>
        <NotificationsList notifications={notifications} isWidget={true} />
      </BrowserRouter>
    )
    const { length } = await screen.getAllByTestId('notificationItem')
    expect(length).toBe(4)
    expect(screen.getByTestId('goToNotificationsButton')).toBeVisible()
  })
  it('Render notification list component should show no notifications message is successful', () => {
    const emptyNotification = {
      title: 'No notifications yet',
      subTitle: (
        <span>
          Stay tuned for exciting updates!
          <br /> No new notifications at the moment.
        </span>
      )
    }
    render(
      <BrowserRouter>
        <NotificationsList notifications={[]} isWidget={true} emptyNotification={emptyNotification} />
      </BrowserRouter>
    )
    expect(screen.queryByTestId('goToNotificationsButton')).toBeNull()
    expect(screen.getByTestId('data-view-empty')).toBeVisible()
  })
})
