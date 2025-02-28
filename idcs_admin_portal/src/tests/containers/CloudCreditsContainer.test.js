import { render, screen, waitFor } from '@testing-library/react'
import { Router } from 'react-router-dom'
import { createMemoryHistory } from 'history'
import userEvent from '@testing-library/user-event'
import CloudCreditsContainer from '../../containers/cloudCredits/CloudCreditsContainer'
import { mockIntelUser } from '../mocks/authentication/authHelper'
import { clearAxiosMock, clickButton, selectRadioButton, setValueOnDateTimeInput, typeOnInputElement } from '../mocks/utils'
import { act } from 'react'
import AuthWrapper from '../../utility/wrapper/AuthWrapper'
import { mockBaseCloudCredits, mockCreateCoupon, mockCreateCouponAPI } from '../mocks/cloudCredits'

const TestComponent = ({ history }) => {
  return (
    <AuthWrapper>
      <Router location={history.location} navigator={history}>
        <CloudCreditsContainer></CloudCreditsContainer>
      </Router>
    </AuthWrapper>
  )
}

const getEndDateTimeISO = () => {
  // Date format - 'yyyy-MM-ddT00:00'
  const today = new Date()
  return `${today.getFullYear() + 2}-01-01`
}

const getStartDateTimeISO = () => {
  // Date format - 'yyyy-MM-ddT00:00'
  const today = new Date()
  return `${today.getFullYear() + 1}-01-01`
}

describe('Cloud Credit container - Create form', () => {
  let history = null

  beforeEach(() => {
    clearAxiosMock()
    history = createMemoryHistory()
  })

  beforeEach(() => {
    mockIntelUser()
    mockBaseCloudCredits()
    mockCreateCouponAPI()
  })

  it('Checks whether the Create new coupon component is loading correctly', async () => {
    await act(async () => {
      render(<TestComponent history={history} />)
    })

    expect(
      await screen.findByText('Generate coupons for Intel Tiber AI Cloud Console')
    ).toBeInTheDocument()
  })

  it('Amount input should only receive numbers.', async() => {
    render(<TestComponent history={history} />)
    const amountInput = screen.getByTestId('AmountInput')
    const input = '1000#abcd'
    await act(async () => {
      await userEvent.type(amountInput, input)
    })
    await waitFor(() => {
      expect(amountInput).toHaveValue('1000')
    })
  })

  it('Number of users input should only receive numbers.', async() => {
    render(<TestComponent history={history} />)
    const numberOfUsersInput = screen.getByTestId('NumberofusersInput')
    const input = '100#abcd'
    await act(async () => {
      await userEvent.type(numberOfUsersInput, input)
    })
    await waitFor(() => {
      expect(numberOfUsersInput).toHaveValue('100')
    })
  })

  it('Start date input should only receive numbers.', async() => {
    render(<TestComponent history={history} />)
    const numberOfUsersInput = screen.getByTestId('NumberofusersInput')
    const input = '100#abcd'
    await act(async () => {
      await userEvent.type(numberOfUsersInput, input)
    })
    await waitFor(() => {
      expect(numberOfUsersInput).toHaveValue('100')
    })
  })

  it('Create a coupon with current data as start date and get the coupon code', async() => {
    render(<TestComponent history={history} />)
    const amountInput = screen.getByTestId('AmountInput')
    await typeOnInputElement(amountInput, '1000')

    const numberOfUsersInput = screen.getByTestId('NumberofusersInput')
    await typeOnInputElement(numberOfUsersInput, '10')

    const currentStartDateRadioID = 'Startdate-Radio-option-CurrentDate'
    expect(screen.getByTestId(currentStartDateRadioID)).toBeChecked()

    const enddateInput = screen.getByTestId('EnddateInput')
    await setValueOnDateTimeInput(enddateInput, getEndDateTimeISO())

    const createButton = screen.getByTestId('navigationTopCreate')
    await clickButton(createButton)
    await screen.findByTestId('OnCreateCouponModal')
    const codeExpected = mockCreateCoupon().code
    expect(await screen.findByText(codeExpected)).toBeInTheDocument()
  })

  it('Create a coupon with Select from Calendar input as start date and get the coupon code', async() => {
    render(<TestComponent history={history} />)
    const amountInput = screen.getByTestId('AmountInput')
    await typeOnInputElement(amountInput, '1000')

    const numberOfUsersInput = screen.getByTestId('NumberofusersInput')
    await typeOnInputElement(numberOfUsersInput, '10')

    const selectFromCalendarStartDateRadio = screen.getByTestId('Startdate-Radio-option-Selectfromcalendar')
    await selectRadioButton(selectFromCalendarStartDateRadio)

    const startDateInput = screen.getByTestId('StartdateInput')
    await setValueOnDateTimeInput(startDateInput, getStartDateTimeISO())

    const enddateInput = screen.getByTestId('EnddateInput')
    await setValueOnDateTimeInput(enddateInput, getEndDateTimeISO())

    const createButton = screen.getByTestId('navigationTopCreate')
    await clickButton(createButton)
    await screen.findByTestId('OnCreateCouponModal')
    const codeExpected = mockCreateCoupon().code
    expect(await screen.findByText(codeExpected)).toBeInTheDocument()
  })
})
