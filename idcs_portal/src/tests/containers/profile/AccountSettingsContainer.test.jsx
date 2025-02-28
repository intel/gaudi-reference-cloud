// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import '@testing-library/jest-dom'
import { render, screen } from '@testing-library/react'
import { clearAxiosMock } from '../../mocks/utils'
import idcConfig from '../../../config/configurator'
import {
  mockEnterpriseUser,
  mockStandardUser,
  mockBaseEnrollResponse,
  mockPremiumUser,
  mockEnterprisePendingUser,
  mockIntelUser
} from '../../mocks/authentication/authHelper'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { mockMsalAccount } from '../../../setupTests'
import AccountSettingsContainer from '../../../containers/profile/AccountSettingsContainer'
import { mockMultiUserAdminInvitationList } from '../../mocks/profile/profile'

const TestComponent = () => {
  return (
    <AuthWrapper>
      <AccountSettingsContainer />
    </AuthWrapper>
  )
}

describe('Account Settings', () => {
  beforeEach(() => {
    clearAxiosMock()
  })

  beforeEach(() => {
    mockMultiUserAdminInvitationList()
  })

  const expectedCloudAccountId = mockBaseEnrollResponse().cloudAccountId
  const expectedDisplayName = `${mockMsalAccount.idTokenClaims.lastName} ${mockMsalAccount.idTokenClaims.firstName}`
  const expectedEmail = mockMsalAccount.idTokenClaims.email

  it('Show Tier and User information for Standard User', async () => {
    mockStandardUser()
    render(<TestComponent />)
    expect(await screen.findByText(expectedCloudAccountId)).toBeVisible()
    expect(screen.getByText('Standard tier includes:')).toBeVisible()
    expect(screen.getByText(expectedDisplayName)).toBeVisible()
    expect(screen.getByText(expectedEmail)).toBeVisible()
  })

  it('Show Tier and User information for Premium User', async () => {
    mockPremiumUser()
    render(<TestComponent />)
    expect(await screen.findByText(expectedCloudAccountId)).toBeVisible()
    expect(screen.getByText('Premium tier includes:')).toBeVisible()
    expect(screen.getByText(expectedDisplayName)).toBeVisible()
    expect(screen.getByText(expectedEmail)).toBeVisible()
  })

  it('Show Tier and User information for Enterprise User', async () => {
    mockEnterpriseUser()
    render(<TestComponent />)
    expect(await screen.findByText(expectedCloudAccountId)).toBeVisible()
    expect(screen.getByText('Enterprise tier includes:')).toBeVisible()
    expect(screen.getByText(expectedDisplayName)).toBeVisible()
    expect(screen.getByText(expectedEmail)).toBeVisible()
  })

  it('Show Tier and User information for Enterprise Pending User', async () => {
    mockEnterprisePendingUser()
    render(<TestComponent />)
    expect(await screen.findByText(expectedCloudAccountId)).toBeVisible()
    expect(screen.getByText('Enterprise tier includes:')).toBeVisible()
    expect(screen.getByText(expectedDisplayName)).toBeVisible()
    expect(screen.getByText(expectedEmail)).toBeVisible()
  })

  it('Show User information for Intel User', async () => {
    mockIntelUser()
    render(<TestComponent />)
    expect(await screen.findByText(expectedCloudAccountId)).toBeVisible()
    expect(screen.getByText(expectedDisplayName)).toBeVisible()
    expect(screen.getByText(expectedEmail)).toBeVisible()
  })

  it('Show Upgrade to Premium when feature flag is on and user is Standard', async () => {
    mockStandardUser()
    const originalValue = idcConfig.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM
    idcConfig.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM = 1
    render(<TestComponent />)
    expect(await screen.findByText(expectedCloudAccountId)).toBeVisible()
    expect(screen.getByLabelText('Upgrade to premium')).toBeVisible()
    idcConfig.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM = originalValue
  })

  it('Hide Upgrade to Premium when feature flag is off and user is Standard', async () => {
    mockStandardUser()
    const originalValue = idcConfig.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM
    idcConfig.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM = 0
    render(<TestComponent />)
    expect(await screen.findByText(expectedCloudAccountId)).toBeVisible()
    expect(screen.queryByLabelText('Upgrade to premium')).toBeNull()
    idcConfig.REACT_APP_FEATURE_UPGRADE_TO_PREMIUM = originalValue
  })

  it('Show Upgrade to Enterprise when feature flag is on and user is Premium', async () => {
    mockPremiumUser()
    const originalValue = idcConfig.REACT_APP_FEATURE_UPGRADE_TO_ENTERPRISE
    idcConfig.REACT_APP_FEATURE_UPGRADE_TO_ENTERPRISE = 1
    render(<TestComponent />)
    expect(await screen.findByText(expectedCloudAccountId)).toBeVisible()
    expect(screen.getByLabelText('Upgrade to enterprise')).toBeVisible()
    idcConfig.REACT_APP_FEATURE_UPGRADE_TO_ENTERPRISE = originalValue
  })

  it('Hide Upgrade to Enterprise when feature flag is off and user is Premium', async () => {
    mockPremiumUser()
    const originalValue = idcConfig.REACT_APP_FEATURE_UPGRADE_TO_ENTERPRISE
    idcConfig.REACT_APP_FEATURE_UPGRADE_TO_ENTERPRISE = 0
    render(<TestComponent />)
    expect(await screen.findByText(expectedCloudAccountId)).toBeVisible()
    expect(screen.queryByLabelText('Upgrade to enterprise')).toBeNull()
    idcConfig.REACT_APP_FEATURE_UPGRADE_TO_ENTERPRISE = originalValue
  })
})
