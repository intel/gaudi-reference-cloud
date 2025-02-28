import { act } from 'react'
import { render, screen } from '@testing-library/react'
import { Router } from 'react-router-dom'
import CouponsContainer from '../../containers/coupons/CouponsContainer'
import AuthWrapper from '../../utility/wrapper/AuthWrapper'
import { mockIntelUser } from '../mocks/authentication/authHelper'
import { createMemoryHistory } from 'history'
import { mockBaseCloudCreditsStore, mockCouponsAPI } from '../mocks/cloudCredits'
import { clearAxiosMock } from '../mocks/utils'
const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <CouponsContainer />
      </Router>
    </AuthWrapper>
  )
}

describe('Cloud coupon details - View', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
    mockCouponsAPI()
  })

  it('Check if the view contains info', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    expect(await screen.findByTestId('btn-export')).toBeInTheDocument()
  })

  it('Check if the view contains table view', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })
    const disableButtons = await screen.findAllByTestId('DisableButtonTable')
    expect(disableButtons.length).toEqual(mockBaseCloudCreditsStore().coupons.length)
  })
})
