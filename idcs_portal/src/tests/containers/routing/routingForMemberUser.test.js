// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

// eslint-disable-next-line no-unused-vars
import Empty from '../../mocks/routing/fakeContainers'
import '@testing-library/jest-dom'
import { render, screen } from '@testing-library/react'
import { clearAxiosMock } from '../../mocks/utils'
import idcConfig from '../../../config/configurator'
import { Router } from 'react-router-dom'
import { createMemoryHistory } from 'history'
import { mockStandardUser } from '../../mocks/authentication/authHelper'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import Routing from '../../../containers/routing/Routing'
import { mockUserCloudAccountsList, premiumAdminMock } from '../../mocks/profile/profile'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <Routing />
      </Router>
    </AuthWrapper>
  )
}

describe('Routes Member User', () => {
  let history = null
  const configBackup = { ...idcConfig }

  beforeAll(() => {
    idcConfig.REACT_APP_GETTING_STARTED_URL = '/home'
    idcConfig.REACT_APP_TUTORIALS_URL = '/home'
    idcConfig.REACT_APP_WHATSNEW_URL = '/home'
    idcConfig.REACT_APP_SELECTED_CLOUD_ACCOUNT = premiumAdminMock.cloudAccountId
  })

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
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
    mockStandardUser(true)
    mockUserCloudAccountsList()
  })

  afterAll(() => {
    idcConfig.REACT_APP_GETTING_STARTED_URL = configBackup.REACT_APP_GETTING_STARTED_URL
    idcConfig.REACT_APP_TUTORIALS_URL = configBackup.REACT_APP_TUTORIALS_URL
    idcConfig.REACT_APP_WHATSNEW_URL = configBackup.REACT_APP_WHATSNEW_URL
    idcConfig.REACT_APP_SELECTED_CLOUD_ACCOUNT = ''
  })

  // member user

  it('Check member users have NOT access to credits page', async () => {
    history.push('/billing/credits')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check member users have NOT access to usages page', async () => {
    history.push('/billing/usages')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check member users have NOT access to invoices page', async () => {
    history.push('/billing/invoices')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check member users have NOT access to invoice details page', async () => {
    history.push('/billing/invoiceDetails')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check member users have NOT access to manage Payment Methods page', async () => {
    history.push('/billing/managePaymentMethods')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check member users have NOT access to manage credit card page', async () => {
    history.push('/billing/managePaymentMethods/managecreditcard')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check member users have NOT access to manage coupon code page', async () => {
    history.push('/billing/credits/managecouponcode')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check member users have NOT access to add credit card response page', async () => {
    history.push('/billing/creditResponse')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check member users have NOT access to upgrade account page', async () => {
    history.push('/billing/upgradeaccount')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })
})
