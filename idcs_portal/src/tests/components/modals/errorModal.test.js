// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'
import { clearAxiosMock } from '../../mocks/utils'
import {
  mockEnterprisePendingUser,
  mockEnterpriseUser,
  mockIntelUser,
  mockPremiumUser,
  mockStandardUser
} from '../../mocks/authentication/authHelper'
import ErrorModal from '../../../utils/modals/errorModal/ErrorModal'

const TestComponent = ({ hideRetry }) => {
  return (
    <AuthWrapper>
      <BrowserRouter>
        <ErrorModal showModal message="Test error Message" hideRetryMessage={hideRetry}></ErrorModal>
      </BrowserRouter>
    </AuthWrapper>
  )
}

describe('Auth Wrapper', () => {
  beforeEach(() => {
    clearAxiosMock()
  })

  it('Shows contact support retry message and button for Intel users', async () => {
    mockIntelUser()
    render(<TestComponent></TestComponent>)
    expect(await screen.findByTestId('error-modal-contact-support-link')).toBeInTheDocument()
    expect(
      await screen.findByText('Please try again or contact support if the issue continues.', { exact: false })
    ).toBeInTheDocument()
  })

  it('Shows contact support retry message and button for Enterprise users', async () => {
    mockEnterpriseUser()
    render(<TestComponent></TestComponent>)
    expect(await screen.findByTestId('error-modal-contact-support-link')).toBeInTheDocument()
    expect(
      await screen.findByText('Please try again or contact support if the issue continues.', { exact: false })
    ).toBeInTheDocument()
  })

  it('Shows contact support retry message and button for Premium users', async () => {
    mockPremiumUser()
    render(<TestComponent></TestComponent>)
    expect(await screen.findByTestId('error-modal-contact-support-link')).toBeInTheDocument()
    expect(
      await screen.findByText('Please try again or contact support if the issue continues.', { exact: false })
    ).toBeInTheDocument()
  })

  it('Shows community retry message and button for Standard users', async () => {
    mockStandardUser()
    render(<TestComponent></TestComponent>)
    expect(await screen.findByTestId('error-modal-community-link')).toBeInTheDocument()
    expect(
      await screen.findByText('Please try again or check the community for hints.', { exact: false })
    ).toBeInTheDocument()
  })

  it('Shows community retry message and button for Enterprise pending', async () => {
    mockEnterprisePendingUser()
    render(<TestComponent></TestComponent>)
    expect(await screen.findByTestId('error-modal-community-link')).toBeInTheDocument()
    expect(
      await screen.findByText('Please try again or check the community for hints.', { exact: false })
    ).toBeInTheDocument()
  })
})
