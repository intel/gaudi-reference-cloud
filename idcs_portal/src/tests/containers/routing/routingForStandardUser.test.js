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

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <Routing />
      </Router>
    </AuthWrapper>
  )
}

describe('Routes Standard User', () => {
  let history = null
  const configBackup = { ...idcConfig }

  beforeAll(() => {
    idcConfig.REACT_APP_GETTING_STARTED_URL = '/home'
    idcConfig.REACT_APP_TUTORIALS_URL = '/home'
    idcConfig.REACT_APP_WHATSNEW_URL = '/home'
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
    mockStandardUser()
  })

  afterAll(() => {
    idcConfig.REACT_APP_GETTING_STARTED_URL = configBackup.REACT_APP_GETTING_STARTED_URL
    idcConfig.REACT_APP_TUTORIALS_URL = configBackup.REACT_APP_TUTORIALS_URL
    idcConfig.REACT_APP_WHATSNEW_URL = configBackup.REACT_APP_WHATSNEW_URL
  })

  // standard user
  it('Check standard users have access to the home page', async () => {
    history.push('/')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('HomePage')).toBeVisible()
  })

  it('Check standard users have access to hardware page', async () => {
    history.push('/hardware')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('/hardware')).toBeVisible()
  })

  it('Check standard users have access to training page', async () => {
    history.push('/learning/notebooks')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('/learning/notebooks')).toBeVisible()
  })

  it('Check standard users have access to public keys page', async () => {
    history.push('/security/publickeys/')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('/security/publickeys/')).toBeVisible()
  })

  it('Check standard users have access to public keys import page', async () => {
    history.push('/security/publickeys/import')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('/security/publickeys/import')).toBeVisible()
  })

  it('Check standard users have access to error page', async () => {
    history.push('/error/notfound')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check standard users have access to access denied page', async () => {
    history.push('/error/accessdenied')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Your account is being verified.')).toBeVisible()
  })

  it('Check standard users have access to something went wrong page', async () => {
    history.push('/error/somethingwentwrong')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Something went wrong')).toBeVisible()
  })

  it('Check standard users have access to compute reserve page', async () => {
    history.push('/compute/reserve')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('/compute/reserve')).toBeVisible()
  })

  it('Check standard users have access to compute edit page', async () => {
    history.push('/compute/d/test/edit')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('/compute/d/test/edit')).toBeVisible()
  })

  it('Check standard users have access to compute page', async () => {
    history.push('/compute')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('/compute')).toBeVisible()
  })

  it('Check standard users have access to compute-groups reserve page', async () => {
    history.push('/compute-groups/reserve')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('/compute-groups/reserve')).toBeVisible()
  })

  it('Check standard users have access to compute-groups edit page', async () => {
    history.push('/compute-groups/d/test/edit')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('/compute-groups/d/test/edit')).toBeVisible()
  })

  it('Check standard users have access to compute-groups my reservations page', async () => {
    history.push('/compute-groups')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('/compute-groups')).toBeVisible()
  })

  it('Check standard users have access to storage reserve page', async () => {
    history.push('/storage/reserve')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('/storage/reserve')).toBeVisible()
  })

  it('Check standard users have access to storage page', async () => {
    history.push('/storage')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('/storage')).toBeVisible()
  })

  it('Check standard users have NOT access to invoices page', async () => {
    history.push('/billing/invoices')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check standard users have NOT access to invoices details page', async () => {
    history.push('/billing/invoiceDetails')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check standard users have access to credits page', async () => {
    history.push('/billing/credits')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('/billing/credits')).toBeVisible()
  })

  it('Check standard users have NOT access to manage payment methods page', async () => {
    history.push('/billing/managePaymentMethods')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check standard users have NOT access to manage credit card page', async () => {
    history.push('/billing/managePaymentMethods/managecreditcard')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check standard users have access to manage coupon code page', async () => {
    history.push('/billing/credits/managecouponcode')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('/billing/credits/managecouponcode')).toBeVisible()
  })

  it('Check standard users have NOT access to credit response page', async () => {
    history.push('/billing/creditResponse')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
  })

  it('Check standard users have access to premium page', async () => {
    history.push('/premium')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('/premium')).toBeVisible()
  })

  it('Check standard users have access to upgrade account when feature flag is enabled', async () => {
    const originalFlagValue = idcConfig.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM
    idcConfig.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM = 1
    history.push('/upgradeaccount')
    render(<TestComponent history={history} />)
    expect(await screen.findByTestId('/upgradeaccount')).toBeVisible()
    idcConfig.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM = originalFlagValue
  })

  it('Check standard users have NOT access to upgrade account when feature flag is not enabled', async () => {
    const originalFlagValue = idcConfig.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM
    idcConfig.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM = 0
    history.push('/upgradeaccount')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
    idcConfig.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM = originalFlagValue
  })

  it('Check standard users have access to KaaS Create Cluster page', async () => {
    history.push('/cluster/reserve')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('/cluster/reserve')).toBeVisible()
  })

  it('Check standard users have access to KaaS My Clusters page in root', async () => {
    history.push('/cluster')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('/cluster')).toBeVisible()
  })

  it('Check standard users have NOT access with future flag disable to profile apikeys page', async () => {
    const originalFlagValue = idcConfig.REACT_APP_FEATURE_API_KEYS
    idcConfig.REACT_APP_FEATURE_API_KEYS = false
    history.push('/profile/apikeys')
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Page not found')).toBeVisible()
    idcConfig.REACT_APP_FEATURE_API_KEYS = originalFlagValue
  })
})
