// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { act } from 'react'
import userEvent from '@testing-library/user-event'
import { iso31662 } from 'iso-3166'

import { mockPremiumUser } from '../../mocks/authentication/authHelper'
import {
  clearAxiosMock,
  expectValueForInputElement,
  expectClassForInputElement,
  expectTextContentForInputElement
} from '../../mocks/utils'
import AuthWrapper from '../../../utils/authWrapper/AuthWrapper'

import ManageCreditCardContainers from '../../../containers/billing/ManageCreditCardContainers'
import idcConfig from '../../../config/configurator'
import { countryCodesForAcceptedCountries } from '../../../utils/Enums'

const TestComponent = () => {
  return (
    <AuthWrapper>
      <ManageCreditCardContainers />
    </AuthWrapper>
  )
}

const getCurrentMonthAndNextYear = () => {
  const today = new Date()
  return { month: today.getMonth() + 1, year: today.getFullYear() + 1 }
}

describe('Credit Card container - Create form', () => {
  beforeEach(() => {
    clearAxiosMock()
    mockPremiumUser()
  })

  beforeAll(() => {
    idcConfig.REACT_APP_FEATURE_DIRECTPOST = 1
  })

  it('Checks whether the credit card component is loading correctly', async () => {
    await act(async () => {
      render(<TestComponent />)
    })

    expect(await screen.findByText('Add a credit card')).toBeInTheDocument()
  })

  describe('Card Number Input validations', () => {
    it('Card Number input should only receive numbers.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardnumberInput = screen.getByTestId('CardnumberInput')
      await expectValueForInputElement(cardnumberInput, '411ab11@11', '4111 111')
    })

    it('Card Number input should display Visa card number with spaces.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardnumberInput = screen.getByTestId('CardnumberInput')
      await expectValueForInputElement(cardnumberInput, '4111111111111111', '4111 1111 1111 1111')
    })

    it('Card Number input should display American Express card number with spaces.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardnumberInput = screen.getByTestId('CardnumberInput')
      await expectValueForInputElement(cardnumberInput, '371449635398431', '3714 496353 98431')
    })

    it('Card Number input should disply Visa card type.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })

      const cardnumberInput = screen.getByTestId('CardnumberInput')
      await expectClassForInputElement(cardnumberInput, '411ab11@11', 'creditCard-visa')
    })

    it('Card Number input should disply MasterCard card type.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })

      const cardnumberInput = screen.getByTestId('CardnumberInput')
      await expectClassForInputElement(cardnumberInput, '510555', 'creditCard-mastercard')
    })

    it('Card Number input should disply American Express card type.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })

      const cardnumberInput = screen.getByTestId('CardnumberInput')
      await expectClassForInputElement(cardnumberInput, '378288', 'creditCard-amex')
    })

    it('Card Number input should disply Discover card type.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })

      const cardnumberInput = screen.getByTestId('CardnumberInput')
      await expectClassForInputElement(cardnumberInput, '60111', 'creditCard-discover')
    })

    it('Card Number input should show card not allowed error.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })

      const cardnumberInput = screen.getByTestId('CardnumberInput')
      await expectClassForInputElement(cardnumberInput, '305678', 'is-invalid')
      expect(await screen.findByText('Card is not allowed.')).toBeInTheDocument()
    })

    it('Card Number input should show invalid card error.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })

      const cardnumberInput = screen.getByTestId('CardnumberInput')
      await expectValueForInputElement(cardnumberInput, '411111111111111', '4111 1111 1111 111')
      fireEvent.blur(cardnumberInput)
      expect(await screen.findByText('Invalid card.')).toBeInTheDocument()
    })
  })

  describe('Card Month Input validations', () => {
    it('Card Month input should only receive numbers.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardMonthInput = screen.getByTestId('MonthInput')
      await expectValueForInputElement(cardMonthInput, '4@asd2', '42')
    })

    it('Card Month input should only receive numbers upto 2 places.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardMonthInput = screen.getByTestId('MonthInput')
      await expectValueForInputElement(cardMonthInput, '4@11ab11@11', '41')
    })

    it('Card Month input should only receive value greater then 0.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardMonthInput = screen.getByTestId('MonthInput')
      await expectClassForInputElement(cardMonthInput, '0', 'is-invalid')
    })

    it('Card Month input should only receive lesser then 13.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardMonthInput = screen.getByTestId('MonthInput')
      await expectClassForInputElement(cardMonthInput, '13', 'is-invalid')
    })
  })

  describe('Card Year Input validations', () => {
    it('Card Year input should only receive numbers.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardYearInput = screen.getByTestId('YearInput')
      await expectValueForInputElement(cardYearInput, '4@asd2', '42')
    })

    it('Card Year input should only receive numbers upto 2 places.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardYearInput = screen.getByTestId('YearInput')
      await expectValueForInputElement(cardYearInput, '4@11ab11@11', '41')
    })

    it('Card Year input should only receive value greater then equal to current Year.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardYearInput = screen.getByTestId('YearInput')
      const { year } = getCurrentMonthAndNextYear()
      await expectClassForInputElement(cardYearInput, year.toString(), 'is-invalid')
    })
  })

  describe('Card Month and Card Year date validation', () => {
    it('Card Month and Year input should display error for past and current month/year value.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardMonthInput = screen.getByTestId('MonthInput')
      const cardYearInput = screen.getByTestId('YearInput')

      const invalidMonth = 13
      const invalidYear = 2021

      await expectValueForInputElement(cardMonthInput, invalidMonth.toString(), invalidMonth.toString())
      await expectValueForInputElement(
        cardYearInput,
        invalidYear.toString().slice(2, 4),
        invalidYear.toString().slice(2, 4)
      )
      fireEvent.blur(cardYearInput)
      const cardMonthInvalidMessage = await screen.getByTestId('MonthInvalidMessage')
      await waitFor(() => {
        expect(cardMonthInvalidMessage).toHaveTextContent('Invalid Month.')
      })
      const cardYearInvalidMessage = await screen.getByTestId('YearInvalidMessage')
      await waitFor(() => {
        expect(cardYearInvalidMessage).toHaveTextContent('Invalid Year.')
      })
    })
  })

  describe('Card CVC Input validations', () => {
    it('Card CVC input should only receive numbers.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardCvcInput = screen.getByTestId('CVCInput')
      await expectValueForInputElement(cardCvcInput, '4@asd2', '42')
    })

    it('Card CVC input should only receive numbers upto 3 places for Visa Card Type.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardnumberInput = screen.getByTestId('CardnumberInput')
      await expectValueForInputElement(cardnumberInput, '4111111111111111', '4111 1111 1111 1111')
      const cardCvcInput = screen.getByTestId('CVCInput')
      await expectValueForInputElement(cardCvcInput, '123', '123')
      await waitFor(() => {
        expect(cardCvcInput).toHaveAttribute('maxlength', '3')
      })
    })

    it('Card CVC input should only receive numbers upto 3 places for Master Card Type.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardnumberInput = screen.getByTestId('CardnumberInput')
      await expectValueForInputElement(cardnumberInput, '5555555555554444', '5555 5555 5555 4444')
      const cardCvcInput = screen.getByTestId('CVCInput')
      await expectValueForInputElement(cardCvcInput, '123', '123')
      await waitFor(() => {
        expect(cardCvcInput).toHaveAttribute('maxlength', '3')
      })
    })

    it('Card CVC input should only receive numbers upto 3 places for Discover Card Type.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardnumberInput = screen.getByTestId('CardnumberInput')
      await expectValueForInputElement(cardnumberInput, '6011111111111117', '6011 1111 1111 1117')
      const cardCvcInput = screen.getByTestId('CVCInput')
      await expectValueForInputElement(cardCvcInput, '123', '123')
      await waitFor(() => {
        expect(cardCvcInput).toHaveAttribute('maxlength', '3')
      })
    })

    it('Card CVC input should only receive numbers upto 4 places for Amex Card Type.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cardnumberInput = screen.getByTestId('CardnumberInput')
      await expectValueForInputElement(cardnumberInput, '378282246310005', '3782 822463 10005')
      const cardCvcInput = screen.getByTestId('CVCInput')
      await expectValueForInputElement(cardCvcInput, '1234', '1234')
      await waitFor(() => {
        expect(cardCvcInput).toHaveAttribute('maxlength', '4')
      })
    })
  })

  describe('First Name Input validations', () => {
    it('First Name input should display error while entering characters except alphabets and space.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const firstNameInput = screen.getByTestId('FirstNameInput')
      await expectValueForInputElement(
        firstNameInput,
        'test123test123 test123testt112',
        'test123test123 test123testt112'
      )
      const firstNameInvalidMessage = await screen.getByTestId('FirstNameInvalidMessage')
      await waitFor(() => {
        expect(firstNameInvalidMessage).toHaveTextContent('Only letters from A-Z or a-z are allowed.')
      })
    })

    it('First Name input should only receive alphabets and space upto 63 characters.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const firstNameInput = screen.getByTestId('FirstNameInput')
      const inputValue = 'test123test123 test123testt112 test123test123test123 test123tes12435653 test2342742 testets' // 91 characters
      const expectValue = 'test123test123 test123testt112 test123test123test123 test123tes' // 63 Characters
      await expectValueForInputElement(firstNameInput, inputValue, expectValue)
      await waitFor(() => {
        expect(firstNameInput).toHaveAttribute('maxlength', '63')
      })
    })
  })

  describe('Last Name Input validations', () => {
    it('Last Name input should display error while entering characters except alphabets and space.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const lastNameInput = screen.getByTestId('LastNameInput')
      await expectValueForInputElement(
        lastNameInput,
        'test123test123 test123testt112',
        'test123test123 test123testt112'
      )
      const lastNameInvalidMessage = await screen.getByTestId('LastNameInvalidMessage')
      await waitFor(() => {
        expect(lastNameInvalidMessage).toHaveTextContent('Only letters from A-Z or a-z are allowed.')
      })
    })

    it('Last Name input should only receive alphabets and space upto 63 characters.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const lastNameInput = screen.getByTestId('LastNameInput')
      const inputValue = 'test123test123 test123testt112 test123test123test123 test123tes12435653 test2342742 testets' // 91 characters
      const expectValue = 'test123test123 test123testt112 test123test123test123 test123tes' // 63 Characters
      await expectValueForInputElement(lastNameInput, inputValue, expectValue)
      await waitFor(() => {
        expect(lastNameInput).toHaveAttribute('maxlength', '63')
      })
    })
  })

  describe('Max Length validations', () => {
    it('Company input should only receive characters upto 63.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const companynameInput = screen.getByTestId('CompanynameInput')
      const inputValue = 'test123test123 test123testt112 test123test123test123 test123tes12435653 test2342742 testets' // 91 characters
      const expectValue = 'test123test123 test123testt112 test123test123test123 test123tes' // 63 Characters
      await expectValueForInputElement(companynameInput, inputValue, expectValue)
      await waitFor(() => {
        expect(companynameInput).toHaveAttribute('maxlength', '63')
      })
    })

    it('Phone Input should only receive characters upto 63.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const phoneInput = screen.getByTestId('PhoneInput')
      const inputValue = 'test123test123 test123testt112 test123test123test123 test123tes12435653 test2342742 testets' // 91 characters
      const expectValue = 'test123test123 test123testt112 test123test123test123 test123tes' // 63 Characters
      await expectValueForInputElement(phoneInput, inputValue, expectValue)
      await waitFor(() => {
        expect(phoneInput).toHaveAttribute('maxlength', '63')
      })
    })

    it('Address line 1 Input should only receive characters upto 63.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const addressline1Input = screen.getByTestId('Addressline1Input')
      const inputValue = 'test123test123 test123testt112 test123test123test123 test123tes12435653 test2342742 testets' // 91 characters
      const expectValue = 'test123test123 test123testt112 test123test123test123 test123tes' // 63 Characters
      await expectValueForInputElement(addressline1Input, inputValue, expectValue)
      await waitFor(() => {
        expect(addressline1Input).toHaveAttribute('maxlength', '63')
      })
    })

    it('Address line 2 Input should only receive characters upto 63.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const addressline2Input = screen.getByTestId('Addressline2Input')
      const inputValue = 'test123test123 test123testt112 test123test123test123 test123tes12435653 test2342742 testets' // 91 characters
      const expectValue = 'test123test123 test123testt112 test123test123test123 test123tes' // 63 Characters
      await expectValueForInputElement(addressline2Input, inputValue, expectValue)
      await waitFor(() => {
        expect(addressline2Input).toHaveAttribute('maxlength', '63')
      })
    })

    it('City Input should only receive characters upto 63.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const cityInput = screen.getByTestId('CityInput')
      const inputValue = 'test123test123 test123testt112 test123test123test123 test123tes12435653 test2342742 testets' // 91 characters
      const expectValue = 'test123test123 test123testt112 test123test123test123 test123tes' // 63 Characters
      await expectValueForInputElement(cityInput, inputValue, expectValue)
      await waitFor(() => {
        expect(cityInput).toHaveAttribute('maxlength', '63')
      })
    })
  })

  describe('ZipCode Input validations', () => {
    it('Validate Error message for ZipCode input', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const zipCodeInput = screen.getByTestId('ZIPcodeInput')
      await userEvent.clear(zipCodeInput)
      await userEvent.type(zipCodeInput, '898989 ')
      await waitFor(() => {
        expect(screen.getByTestId('ZIPcodeInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('ZIPcodeInvalidMessage')).toHaveTextContent(
          'Only alphanumeric, space and hyphen(-) allowed for ZIP code.'
        )
      })

      await userEvent.clear(zipCodeInput)
      await userEvent.type(zipCodeInput, '898989-')
      await waitFor(() => {
        expect(screen.getByTestId('ZIPcodeInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('ZIPcodeInvalidMessage')).toHaveTextContent(
          'Only alphanumeric, space and hyphen(-) allowed for ZIP code.'
        )
      })

      await userEvent.clear(zipCodeInput)
      await userEvent.type(zipCodeInput, '-898989-')
      await waitFor(() => {
        expect(screen.getByTestId('ZIPcodeInvalidMessage')).toBeInTheDocument()
        expect(screen.getByTestId('ZIPcodeInvalidMessage')).toHaveTextContent(
          'Only alphanumeric, space and hyphen(-) allowed for ZIP code.'
        )
      })
    })

    it('ZipCode Input should only receive characters upto 20.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const ZIPcodeInput = screen.getByTestId('ZIPcodeInput')
      const inputValue = '123456789-12345678-99' // 21 characters
      const expectValue = '123456789-12345678-9' // 20 Characters
      await expectValueForInputElement(ZIPcodeInput, inputValue, expectValue)
      await waitFor(() => {
        expect(ZIPcodeInput).toHaveAttribute('maxlength', '20')
      })
    })
  })

  describe('Email Input validations', () => {
    it('Email input should show invalid email address error.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })

      const emailNegativeValues = 'example@firstname+lastname+@email.com'
      const emailInput = screen.getByTestId('EmailInput')

      await expectValueForInputElement(emailInput, emailNegativeValues, emailNegativeValues)
      await expectTextContentForInputElement(screen.getByTestId('EmailInvalidMessage'), 'Invalid email address.')
    })
  })

  describe('Country and State Input validations', () => {
    it('State Input should have one disabled option on page load.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })

      const stateSelect = screen.getByTestId('StateSelect')
      await waitFor(() => {
        expect(stateSelect.childElementCount).toEqual(1)
      })
    })

    it('Country Input should have list of countries along with one disabled option on page load.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })

      const countrySelect = screen.getByTestId('CountrySelect')
      await waitFor(() => {
        expect(countrySelect.childElementCount).toEqual(countryCodesForAcceptedCountries.length + 1)
      })
    })

    it('State Input options should be based on the selected country of Country Input.', async () => {
      await act(async () => {
        render(<TestComponent />)
      })

      const countrySelect = screen.getByTestId('CountrySelect')
      const selectdValue = 'US'
      const selectedOptionName = 'United States of America'
      userEvent.selectOptions(countrySelect, selectdValue)
      await waitFor(() => {
        expect(screen.getByTestId(selectedOptionName.replaceAll(' ', '') + 'Option').selected).toBeTruthy()
        expect(screen.getByTestId('AustriaOption').selected).toBeFalsy()
      })

      const stateSelect = screen.getByTestId('StateSelect')
      const states = iso31662.filter((x) => x.parent === selectdValue)
      await waitFor(() => {
        expect(stateSelect.childElementCount).toEqual(states.length + 1)
      })
    })
  })

  describe('Add Card button validations', () => {
    it('Add Card Button should be enabled after all the correct values', async () => {
      await act(async () => {
        render(<TestComponent />)
      })
      const { month, year } = getCurrentMonthAndNextYear()

      const btnAddCreditPayment = screen.getByTestId('btn-credit-AddCreditPayment')
      await waitFor(() => {
        expect(btnAddCreditPayment).toBeDisabled()
      })

      await expectValueForInputElement(screen.getByTestId('CardnumberInput'), '4111111111111111', '4111 1111 1111 1111')
      await expectValueForInputElement(screen.getByTestId('MonthInput'), month.toString(), month.toString())

      await expectValueForInputElement(
        screen.getByTestId('YearInput'),
        year.toString().slice(2, 4),
        year.toString().slice(2, 4)
      )

      await expectValueForInputElement(screen.getByTestId('CVCInput'), '123', '123')

      await expectValueForInputElement(screen.getByTestId('FirstNameInput'), 'Test', 'Test')

      await expectValueForInputElement(screen.getByTestId('LastNameInput'), 'Test', 'Test')

      await expectValueForInputElement(screen.getByTestId('EmailInput'), 'abc@abc.com', 'abc@abc.com')

      await expectValueForInputElement(screen.getByTestId('CompanynameInput'), 'Test', 'Test')

      await expectValueForInputElement(screen.getByTestId('PhoneInput'), '1234567890', '1234567890')

      await userEvent.selectOptions(screen.getByTestId('CountrySelect'), 'US')

      await expectValueForInputElement(screen.getByTestId('Addressline1Input'), 'Test', 'Test')

      await expectValueForInputElement(screen.getByTestId('Addressline2Input'), 'Test', 'Test')

      await expectValueForInputElement(screen.getByTestId('CityInput'), 'Test', 'Test')

      await userEvent.selectOptions(screen.getByTestId('StateSelect'), 'AL')

      await expectValueForInputElement(screen.getByTestId('ZIPcodeInput'), '12345', '12345')

      const btnAddCreditPaymentNew = screen.getByTestId('btn-credit-AddCreditPayment')
      await waitFor(() => {
        expect(btnAddCreditPaymentNew).toBeEnabled()
      })
    })
  })
})
