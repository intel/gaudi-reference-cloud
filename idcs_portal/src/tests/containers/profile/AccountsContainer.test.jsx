// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import '@testing-library/jest-dom'
import { render, screen } from '@testing-library/react'
import { clearAxiosMock } from '../../mocks/utils'
import idcConfig, { setDefaultCloudAccount } from '../../../config/configurator'
import { mockBaseEnrollResponse, mockStandardUser } from '../../mocks/authentication/authHelper'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { mockUserCloudAccountsList } from '../../mocks/profile/profile'
import { EnrollAccountType } from '../../../utils/Enums'

const TestComponent = () => {
  return <AuthWrapper>Children Page</AuthWrapper>
}

describe('Account Selection Page', () => {
  const originalWindowLocation = window.location

  beforeAll(() => {
    Object.defineProperty(window, 'location', {
      configurable: true,
      enumerable: true,
      value: {
        ...originalWindowLocation,
        // Mock as if we reload config because of reload beign called
        reload: () => {
          setDefaultCloudAccount(idcConfig)
        }
      }
    })
  })

  beforeEach(() => {
    idcConfig.REACT_APP_SELECTED_CLOUD_ACCOUNT = ''
    clearAxiosMock()
  })

  beforeEach(() => {
    mockStandardUser(true)
    mockUserCloudAccountsList()
  })

  it('If user is selected continue to current route', async () => {
    // Modify default User in config
    idcConfig.REACT_APP_SELECTED_CLOUD_ACCOUNT = mockBaseEnrollResponse(EnrollAccountType.standard, true).cloudAccountId
    render(<TestComponent />)
    expect(await screen.findByText('Children Page')).toBeInTheDocument()
  })
})
