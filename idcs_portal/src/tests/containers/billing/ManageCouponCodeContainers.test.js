// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, waitFor } from '@testing-library/react'
import { Router } from 'react-router-dom'
import { createMemoryHistory } from 'history'
import { act } from 'react'
import { mockPremiumUser } from '../../mocks/authentication/authHelper'
import { clearAxiosMock, expectValueForInputElement } from '../../mocks/utils'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'

import ManageCouponCodeContainers from '../../../containers/billing/ManageCouponCodeContainers'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <ManageCouponCodeContainers />
      </Router>
    </AuthWrapper>
  )
}

describe('Coupon Code container - Redeem coupon form', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    mockPremiumUser()
    history = createMemoryHistory()
  })

  it('Checks whether the Coupon Code component is loading correctly', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(await screen.findByText('Redeem coupon')).toBeInTheDocument()
  })

  it('Reedem Button should be enabled after all the correct values', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    await expectValueForInputElement(screen.getByTestId('CouponCodeInput'), 'test-123-test', 'test-123-test')

    const btnAddCreditPaymentNew = screen.getByTestId('btn-managecouponcode-Redeem')
    await waitFor(() => {
      expect(btnAddCreditPaymentNew).toBeEnabled()
    })
  })
})
