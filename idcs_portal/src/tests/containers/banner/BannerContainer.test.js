// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import { Router } from 'react-router-dom'
import { createMemoryHistory } from 'history'
import {
  mockPremiumUser,
  mockEnterpriseUser,
  mockIntelUser,
  mockStandardUser,
  mockEnterprisePendingUser
} from '../../mocks/authentication/authHelper'
import { clearAxiosMock } from '../../mocks/utils'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import BannerContainer from '../../../containers/banner/BannerContainer'
import idcConfig from '../../../config/configurator'
import useBannerStore from '../../../store/bannerStore/BannerStore'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <div className="d-flex flex-column w-100 h-100">
          <BannerContainer />
        </div>
      </Router>
    </AuthWrapper>
  )
}

const mockBannerConfig = (userTypes, openInNewTab) => {
  idcConfig.REACT_APP_SITE_BANNERS = [
    {
      id: 1726168351137,
      type: 'info',
      title: 'Banner title',
      status: 'active',
      message: 'Banner message',
      userTypes,
      routes: ['all'],
      regions: ['all'],
      expirationDatetime: '',
      link: {
        label: 'Link label',
        href: 'https://domain.com',
        openInNewTab
      }
    }
  ]
}

const clearBannerConfig = () => {
  idcConfig.REACT_APP_SITE_BANNERS = []
}

describe('Banner container', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    clearBannerConfig()
    useBannerStore.setState((prevState) => ({ ...prevState, bannerList: [] }))
    history = createMemoryHistory()
  })

  afterAll(() => {
    clearBannerConfig()
  })

  it('Checks title, messsage are displayed in banner', async () => {
    mockBannerConfig(['intel', 'enterprise', 'premium', 'standard'], true)
    mockIntelUser()
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Banner title')).toBeVisible()
    expect(
      screen.getByText('Banner message', {
        exact: false
      })
    ).toBeVisible()
  })

  it('Checks banner link as correct label and href', async () => {
    mockBannerConfig(['intel', 'enterprise', 'premium', 'standard'], true)
    mockIntelUser()
    render(<TestComponent history={history} />)
    const link = await screen.findByLabelText('Link label')
    expect(link).toBeVisible()
    expect(link).toHaveAttribute('href', 'https://domain.com')
  })

  it('Checks banner link open in new tab when openInNewTab is true', async () => {
    mockBannerConfig(['intel', 'enterprise', 'premium', 'standard'], true)
    mockIntelUser()
    render(<TestComponent history={history} />)
    const link = await screen.findByLabelText('Link label')
    expect(link).toBeVisible()
    expect(link).toHaveAttribute('target', '_blank')
  })

  it('Checks banner link open in same window when openInNewTab is false', async () => {
    mockBannerConfig(['intel', 'enterprise', 'premium', 'standard'], false)
    mockIntelUser()
    render(<TestComponent history={history} />)
    const link = await screen.findByLabelText('Link label')
    expect(link).toBeVisible()
    expect(link).toHaveAttribute('target', '_self')
  })

  it('Checks banner for Intel user is displayed', async () => {
    mockBannerConfig(['intel'], true)
    mockIntelUser()
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Banner title')).toBeVisible()
  })

  it('Checks banner for Standard user is displayed', async () => {
    mockBannerConfig(['standard'], true)
    mockStandardUser()
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Banner title')).toBeVisible()
  })

  it('Checks banner for Premium user is displayed', async () => {
    mockBannerConfig(['premium'], true)
    mockPremiumUser()
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Banner title')).toBeVisible()
  })

  it('Checks banner for Enteprise user is displayed', async () => {
    mockBannerConfig(['enterprise'], true)
    mockEnterpriseUser()
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Banner title')).toBeVisible()
  })

  it('Checks banner for Enteprise Pending user is displayed', async () => {
    mockBannerConfig(['enterprise'], true)
    mockEnterprisePendingUser()
    render(<TestComponent history={history} />)
    expect(await screen.findByText('Banner title')).toBeVisible()
  })
})
