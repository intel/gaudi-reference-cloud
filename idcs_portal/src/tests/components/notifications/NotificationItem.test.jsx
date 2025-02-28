// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import { clearAxiosMock } from '../../mocks/utils'
import { mockStandardUser } from '../../mocks/authentication/authHelper'
import NotificationItem from '../../../components/notifications/NotificationItem'

describe('Notification Item', () => {
  beforeEach(() => {
    clearAxiosMock()
  })

  beforeEach(() => {
    mockStandardUser()
  })

  it('Render billing notification component is successful', async () => {
    render(
      <NotificationItem
        notification={{ serviceName: 'billing', message: 'test', creation: '07/18/2023 9:54 am' }}
        popup={false}
        header={false}
      ></NotificationItem>
    )
    expect(await screen.findByTestId('billingNotificationIcon')).toBeVisible()
  })

  it('Render compute notification component is successful', async () => {
    render(
      <NotificationItem
        notification={{ serviceName: 'compute', message: 'test', creation: '07/18/2023 9:54 am' }}
        popup={false}
        header={false}
      ></NotificationItem>
    )
    expect(await screen.findByTestId('computeNotificationIcon')).toBeVisible()
  })

  it('Render warning notification component is successful', async () => {
    render(
      <NotificationItem
        notification={{ serviceName: 'warning-alert', message: 'test', creation: '07/18/2023 9:54 am' }}
        popup={false}
        header={false}
      ></NotificationItem>
    )
    expect(await screen.findByTestId('warningAlertNotificationIcon')).toBeVisible()
  })

  it('Render default notification component is successful', async () => {
    render(
      <NotificationItem
        notification={{ serviceName: '', message: 'test', creation: '07/18/2023 9:54 am' }}
        popup={false}
        header={false}
      ></NotificationItem>
    )
    expect(await screen.findByTestId('defaultNotificationIcon')).toBeVisible()
  })

  it('Render notification component with message is successful', async () => {
    render(
      <NotificationItem
        notification={{ serviceName: 'test billing', message: 'test message', creation: '07/18/2023 9:54 am' }}
        popup={false}
        header={false}
      ></NotificationItem>
    )
    expect(await screen.findByText('test message')).toBeInTheDocument()
    expect(await screen.findByText('test billing')).toBeInTheDocument()
    expect(await screen.findByText('07/18/2023 9:54 am')).toBeInTheDocument()
  })

  it('Render notification as a popup component is successful', () => {
    const { container } = render(
      <NotificationItem
        notification={{ serviceName: 'test billing', message: 'test message', creation: '07/18/2023 9:54 am' }}
        popup={true}
        header={false}
      ></NotificationItem>
    )
    expect(container.querySelector('br')).toBeInTheDocument()
  })

  it('Render notification as a header component is successful', () => {
    const { container } = render(
      <NotificationItem
        notification={{ serviceName: 'test billing', message: 'test message', creation: '07/18/2023 9:54 am' }}
        popup={false}
        header={true}
      ></NotificationItem>
    )
    expect(container.querySelector('br')).toBeInTheDocument()
  })
})
