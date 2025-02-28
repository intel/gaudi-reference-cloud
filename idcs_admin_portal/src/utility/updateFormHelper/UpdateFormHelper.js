// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import moment from 'moment'

/**
 * Regex of credit cards taken from https://www.makeuseof.com/regex-validate-credit-card-numbers/
 */
const cardTypes = {
  visa: {
    type: /^4[0-9]{12}(?:[0-9]{3})?$/,
    cardType: /^4[0-9]{2,}$/,
    validLength: 16,
    customClass: 'creditCard-visa',
    pattern: _format444,
    name: 'Visa',
    space: 3
  },
  mastercard: {
    type: /^5[1-5][0-9]{14}|^(222[1-9]|22[3-9]\\d|2[3-6]\\d{2}|27[0-1]\\d|2720)[0-9]{12}$/,
    cardType: /^5[1-5][0-9]{2,}|^222[1-9][0-9]{1,}|^22[3-9][0-9]{2,}|^2[3-6][0-9]{3,}|^27[01][0-9]{2,}|^2720[0-9]{1,}$/,
    validLength: 16,
    customClass: 'creditCard-mastercard',
    pattern: _format444,
    name: 'Mastercard',
    space: 3
  },
  amex: {
    type: /^3[47][0-9]{13}$/,
    cardType: /^3[47][0-9]{1,}$/,
    validLength: 15,
    customClass: 'creditCard-amex',
    pattern: _format465,
    name: 'Amex',
    space: 2
  },
  discover: {
    type: /^6(?:011|5[0-9]{1}|4[4-9]{1})[0-9]{1,}$/,
    cardType: /^6(?:011|5[0-9]{1}|4[4-9]{1})[0-9]{1,}$/,
    validLength: 16,
    customClass: 'creditCard-discover',
    pattern: _format444,
    name: 'Discover',
    space: 3
  }
}

export const UpdateFormHelper = (value, formInputName, form) => {
  const updatedForm = {
    ...form
  }
  const updatedFormElement = updatedForm[formInputName]
  updatedFormElement.value = value
  updatedFormElement.isTouched = true
  validateInput(updatedFormElement)
  updatedForm[formInputName] = updatedFormElement
  return updatedForm
}

export const UpdateBlurFormHelper = (value, formInputName, form) => {
  const updatedForm = {
    ...form
  }
  const updatedFormElement = updatedForm[formInputName]
  updatedFormElement.value = value
  updatedFormElement.isTouched = true
  validateBlurredInput(updatedFormElement, updatedForm)
  updatedForm[formInputName] = updatedFormElement
  return updatedForm
}

const validateBlurredInput = (formElementToValidate, form) => {
  // Initial validation always is false
  let isValid = true

  if (formElementToValidate.value) {
    if (formElementToValidate.validationRules.onlyCreditCard) {
      isValid = false

      formElementToValidate.validationMessage = 'Invalid card.'
      const creditCardValue = formElementToValidate.value.replace(/[^0-9]+/g, '')

      const cc = checkLuhn(creditCardValue)
      if (cc) {
        for (const i in cardTypes) {
          const card = cardTypes[i]
          const regex = card.type

          // Check card type
          if (regex.test(creditCardValue)) {
            isValid = true
            break
          }
        }

        formElementToValidate.validationMessage = ''
      }

      formElementToValidate.isValid = isValid
    }

    if (formElementToValidate.validationRules.onlyCreditMonthYear) {
      const monthElement = form.cardMonth
      const yearElement = form.cardYear

      if (
        (monthElement.isValid && yearElement.isValid) ||
        (!monthElement.isValid && monthElement.validationMessage === 'Invalid date.')
      ) {
        const monthValue = monthElement.value
        const yearValue = yearElement.value

        if (monthValue && yearValue) {
          let isValid = true
          let validationMessage = ''

          if (!validateCreditExpiry(monthValue, yearValue)) {
            isValid = false
            validationMessage = 'Invalid date.'
          }

          monthElement.validationMessage = validationMessage
          monthElement.isValid = isValid

          if (formElementToValidate.label === 'Month') {
            formElementToValidate.validationMessage = validationMessage
            formElementToValidate.isValid = isValid
          }
        }
      }
    }

    if (formElementToValidate.validationRules.onlyCreditCvc) {
      if (formElementToValidate.value.length !== formElementToValidate.maxLength) {
        isValid = false
        formElementToValidate.validationMessage = 'Invalid CVC'
      }

      formElementToValidate.isValid = isValid
    }

    if (formElementToValidate.validationRules.onlyZipCode) {
      if (formElementToValidate.value.length > formElementToValidate.maxLength) {
        isValid = false
        formElementToValidate.validationMessage = 'Invalid ZIP code'
      }

      formElementToValidate.isValid = isValid
    }
  }
}

export const validateInput = (formElementToValidate) => {
  // Initial validation always is true
  let isValid = true

  // Get the rules for the element to validate
  if (formElementToValidate.validationRules.isRequired) {
    if (
      !formElementToValidate.value ||
      ((formElementToValidate.type?.toLowerCase() === 'multi-select' ||
        formElementToValidate.type?.toLowerCase() === 'multi-select-dropdown') &&
        formElementToValidate.value.length === 0)
    ) {
      isValid = false
      formElementToValidate.validationMessage =
        formElementToValidate.label?.replace(':', '').replace(' *', '') + ' is required'
      formElementToValidate.isValid = isValid
      // there is a error return the status
      return
    }
  }

  if (formElementToValidate.validationRules.onlyAlphaNumLower) {
    const word = formElementToValidate.value
    const regEx = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/g

    if (!formElementToValidate.validationRules.isRequired && word === '') {
      formElementToValidate.isValid = isValid
      return
    }

    if (regEx.test(word) && word.length) {
      formElementToValidate.isValid = isValid
    } else {
      isValid = false
      formElementToValidate.validationMessage = `Only lower case alphanumeric and hypen(-) allowed for ${formElementToValidate.label}.`
      formElementToValidate.isValid = isValid
      // there is a error return the status
      return
    }
  }

  if (formElementToValidate.validationRules.onlyAlphaNumSpace) {
    const word = formElementToValidate.value
    const regEx = /^[a-zA-Z0-9]([a-zA-Z0-9- ]*[a-zA-Z0-9])?$/g

    if (!formElementToValidate.validationRules.isRequired && word === '') {
      formElementToValidate.isValid = isValid
      return
    }

    if (regEx.test(word) && word.length) {
      formElementToValidate.isValid = isValid
    } else {
      isValid = false
      formElementToValidate.validationMessage = `Only alphanumeric, space and hyphen(-) allowed for ${formElementToValidate.label}.`
      formElementToValidate.isValid = isValid
      // there is a error return the status
      return
    }
  }

  if (formElementToValidate.validationRules.onlyAlphaNumSpaceHyphen) {
    const word = formElementToValidate.value
    const regEx = /^[a-zA-Z0-9 -]*$/g

    if (!formElementToValidate.validationRules.isRequired && word === '') {
      formElementToValidate.isValid = isValid
      return
    }

    if (regEx.test(word) && word.length) {
      formElementToValidate.isValid = isValid
    } else {
      isValid = false
      formElementToValidate.validationMessage = `Only alphanumeric, space and hypen(-) allowed for ${formElementToValidate.label}.`
      formElementToValidate.isValid = isValid
      // there is a error return the status
      return
    }
  }

  if (formElementToValidate.validationRules.onlyAlphaNumExtendedCharacters) {
    const word = formElementToValidate.value
    const regEx = /^[a-zA-Z0-9 .:/_-]*$/g

    if (!formElementToValidate.validationRules.isRequired && word === '') {
      formElementToValidate.isValid = isValid
      return
    }

    if (regEx.test(word) && word.length) {
      formElementToValidate.isValid = isValid
    } else {
      isValid = false
      formElementToValidate.validationMessage = `Only alphanumeric, space and special characters(- _ . : /) allowed for ${formElementToValidate.label}.`
      formElementToValidate.isValid = isValid
      return
    }
  }

  if (formElementToValidate.validationRules.onlyAlphabets) {
    const word = formElementToValidate.value
    const regEx = /^[a-zA-Z ]([a-zA-Z ]*[a-zA-Z ])?$/g

    if (!formElementToValidate.validationRules.isRequired && word === '') {
      formElementToValidate.isValid = isValid
      return
    }

    if (regEx.test(word) && word.length <= formElementToValidate.maxLength) {
      formElementToValidate.isValid = isValid
    } else {
      isValid = false
      formElementToValidate.validationMessage = 'Only letters from A-Z or a-z are allowed.'
      formElementToValidate.isValid = isValid
      return
    }
  }

  if (formElementToValidate.validationRules.onlyCreditCard) {
    const creditCardValue = formElementToValidate.value.replace(/[^0-9]+/g, '')

    let [maxLength, inputCustomClass, newValue] = [19, 'creditCard-all-images', creditCardValue]

    let validLength = 16
    let isSupportedCard = false
    for (const i in cardTypes) {
      const card = cardTypes[i]
      const regex = card.cardType

      // Check card type
      if (regex.test(creditCardValue)) {
        isSupportedCard = true
        validLength = card.validLength

        // Set custom class
        inputCustomClass = card.customClass

        // Set min and max length
        maxLength = card.validLength + card.space

        // Add space to digit
        newValue = card.pattern(creditCardValue)

        break
      }
    }

    if (!isSupportedCard && creditCardValue.length > 5) {
      isValid = false
      formElementToValidate.validationMessage = 'Card is not allowed.'
    }

    if (creditCardValue.length === validLength && !checkLuhn(creditCardValue)) {
      isValid = false
      formElementToValidate.validationMessage = 'Invalid card.'
    }

    formElementToValidate.customClass = inputCustomClass
    formElementToValidate.value = newValue
    formElementToValidate.maxLength = maxLength
  }

  if (formElementToValidate.validationRules.onlyCreditNumeric) {
    const value = formElementToValidate.value
    const regEx = /^\d+$/

    if (!regEx.test(value) || value.length > formElementToValidate.maxLength) {
      formElementToValidate.value = value.slice(0, -1)
    }
  }

  if (formElementToValidate.validationRules.onlyCreditMonthYear) {
    const value = formElementToValidate.value

    if (value.length <= formElementToValidate.maxLength && !validateCreditMonthYear(formElementToValidate)) {
      isValid = false
      formElementToValidate.validationMessage = `Invalid ${formElementToValidate.label}.`
    }
  }

  if (
    formElementToValidate.validationRules.checkMaxLength &&
    formElementToValidate.value.length > formElementToValidate.maxLength
  ) {
    isValid = false
    formElementToValidate.validationMessage = `Max length ${formElementToValidate.maxLength} characters.`
    formElementToValidate.isValid = isValid
    // there is a error return the status
    return
  }

  if (
    formElementToValidate.validationRules.checkMinLength &&
    formElementToValidate.value.length < formElementToValidate.minLength
  ) {
    isValid = false
    formElementToValidate.validationMessage = `Min length ${formElementToValidate.minLength} characters.`
    formElementToValidate.isValid = isValid
    // there is a error return the status
    return
  }

  if (formElementToValidate.validationRules.emailAddress) {
    const word = formElementToValidate.value
    const regEx = /^[a-zA-Z0-9_.±-]+@[a-zA-Z0-9-]+.[a-zA-Z0-9-.]+$/g

    if (!formElementToValidate.validationRules.isRequired && word === '') {
      formElementToValidate.isValid = isValid
      return
    }

    if (regEx.test(word)) {
      formElementToValidate.isValid = isValid
    } else {
      isValid = false
      formElementToValidate.validationMessage = 'Invalid email address.'
      formElementToValidate.isValid = isValid
      return
    }
  }

  if (formElementToValidate.validationRules.futureDate) {
    const today = moment(new Date()).format('YYYY-MM-DD')
    const value = formElementToValidate.value

    if (moment(value).isSameOrBefore(today)) {
      isValid = false
      formElementToValidate.isValid = isValid
      formElementToValidate.validationMessage = 'Date must be greater than today'
      return
    }
  }

  if (formElementToValidate.validationRules.checkMinValue) {
    const word = formElementToValidate.value
    const regEx = /[0-9]/

    if (!formElementToValidate.validationRules.isRequired && word === '') {
      formElementToValidate.isValid = isValid
      return
    }

    if (!regEx.test(word) || word < formElementToValidate.validationRules.checkMinValue) {
      isValid = false
      formElementToValidate.validationMessage = `Value less than ${formElementToValidate.validationRules.checkMinValue} is not allowed.`
      formElementToValidate.isValid = isValid
      return
    } else {
      formElementToValidate.isValid = isValid
    }
  }

  if (formElementToValidate.validationRules.checkMaxValue) {
    const word = formElementToValidate.value
    const regEx = /[0-9]/

    if (!formElementToValidate.validationRules.isRequired && word === '') {
      formElementToValidate.isValid = isValid
      return
    }

    if (!regEx.test(word) || word > formElementToValidate.validationRules.checkMaxValue) {
      isValid = false
      formElementToValidate.validationMessage = `Value more than ${formElementToValidate.validationRules.checkMaxValue} is not allowed.`
      formElementToValidate.isValid = isValid
      return
    } else {
      formElementToValidate.isValid = isValid
    }
  }

  if (formElementToValidate.validationRules.isLoadBalancerSourceIP) {
    const word = formElementToValidate.value
    const regEx =
      /^(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(?:$|\/(16|24))$/g

    if (!formElementToValidate.validationRules.isRequired && word === '') {
      formElementToValidate.isValid = isValid
      return
    }

    if ((word === 'any' || regEx.test(word)) && word.length) {
      formElementToValidate.isValid = isValid
    } else {
      isValid = false
      formElementToValidate.validationMessage = 'Invalid IP'
      formElementToValidate.isValid = isValid
      // there is a error return the status
      return
    }
  }

  if (formElementToValidate.validationRules.isSecuritySourceIP) {
    const word = formElementToValidate.value
    const regEx =
      /^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$|^((25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9]{2}|[1-9]?[0-9])\/(3[0-2]|[1-2]?[0-9])$|^$/g

    if (!formElementToValidate.validationRules.isRequired && word === '') {
      formElementToValidate.isValid = isValid
      return
    }

    if ((word === 'any' || regEx.test(word)) && word.length) {
      formElementToValidate.isValid = isValid
    } else {
      isValid = false
      formElementToValidate.validationMessage = 'Invalid IP'
      formElementToValidate.isValid = isValid
      // there is a error return the status
      return
    }
  }

  if (formElementToValidate.validationRules.isEmailNote) {
    const word = formElementToValidate.value
    const regEx = /^[a-zA-Z0-9.,;_@:\- ]+$/g

    if (!formElementToValidate.validationRules.isRequired && word === '') {
      formElementToValidate.isValid = isValid
      return
    }

    if (regEx.test(word) && word.length) {
      formElementToValidate.isValid = isValid
    } else {
      isValid = false
      formElementToValidate.validationMessage =
        'Only alphanumeric characters, spaces and symbols (   .   ,   ;   :  @   _   -  ) are allowed.'
      formElementToValidate.isValid = isValid
      // there is a error return the status
      return
    }
  }

  if (formElementToValidate.validationRules.isValidURL) {
    const word = formElementToValidate.value

    if (!formElementToValidate.validationRules.isRequired && word === '') {
      formElementToValidate.isValid = isValid
      return
    }

    if (URL.canParse(word)) {
      formElementToValidate.isValid = isValid
    } else {
      isValid = false
      formElementToValidate.validationMessage = 'Invalid URL'
      formElementToValidate.isValid = isValid
      return
    }
  }

  formElementToValidate.isValid = isValid
}

export const isValidForm = (form) => {
  let isValid = true
  let errorCount = 0

  for (const key in form) {
    if (typeof form[key] !== 'object') {
      continue
    }

    if (!form[key].isValid && !form[key].hidden) {
      errorCount++
    }
  }

  if (errorCount > 0) {
    isValid = false
  }

  return isValid
}

export const isValidFormSubmit = (form) => {
  let isValid = true
  let errorCount = 0

  for (const key in form) {
    if (typeof form[key] !== 'object') {
      continue
    }

    form[key].isTouched = true
    validateInput(form[key])

    if (!form[key].isValid && !form[key].hidden) {
      errorCount++
    }
  }

  if (errorCount > 0) {
    isValid = false
  }

  return isValid
}

export const getFormValue = (key, form) => {
  const updatedFormElement = form[key]

  const value = updatedFormElement.value

  return value
}

export const setFormValue = (key, value, form) => {
  const formUpdated = { ...form }

  const updatedFormElement = formUpdated[key]

  updatedFormElement.value = value
  updatedFormElement.isValid = updatedFormElement?.validationRules.isRequired ? value !== '' && value !== null : true

  return formUpdated
}

export const setSelectOptions = (key, values, form) => {
  const formUpdated = { ...form }

  const updatedFormElement = formUpdated[key]

  updatedFormElement.options = values

  return formUpdated
}

export const updateDictionary = (value, field, row) => {
  const updatedRow = {
    ...row
  }
  updatedRow[field].value = value
  updatedRow[field].isTouched = true
  validateInput(updatedRow[field])
  updatedRow.isValidRow = updatedRow.key.isValid && updatedRow.value.isValid
  return updatedRow
}

export const validateDictionaryItems = (items) => {
  const isValid = items.filter((item) => !item.isValidRow)
  return !isValid.length > 0
}

const validateCreditMonthYear = (formElement) => {
  const label = formElement.label
  const currentYear = moment().year()
  let isValid = false

  if (label === 'Year' && parseInt('20' + formElement.value) >= parseInt(currentYear)) {
    isValid = true
  }

  if (label === 'Month' && parseInt(formElement.value) <= 12 && parseInt(formElement.value) > 0) {
    isValid = true
  }

  return isValid
}

const validateCreditExpiry = (month, year) => {
  const today = new Date()
  const expiryDate = new Date()
  expiryDate.setFullYear('20' + year, month - 1, 1)

  return expiryDate > today
}

function _format444(cc) {
  return cc ? cc.match(/[0-9]{1,4}/g).join(' ') : ''
}

function _format465(cc) {
  return [cc.substring(0, 4), cc.substring(4, 10), cc.substring(10, 15)].join(' ').trim()
}

/**
 * checkLuhn is package to assist checking PANs (credit/debit card number) or any other number that uses
 * the Luhn algorithum to validate. The algorithm is in the public domain and is in wide use today.
 * It is specified in ISO/IEC 7812-1
 * @param {*} cardNumber
 * @returns {Boolean}
 */
function checkLuhn(cardNumber) {
  let nCheck = 0
  let bEven = false
  const cardNumberToValidate = cardNumber.replace(/\D/g, '')

  for (let n = cardNumberToValidate.length - 1; n >= 0; n--) {
    const cDigit = cardNumberToValidate.charAt(n)
    let nDigit = parseInt(cDigit, 10)

    if (bEven && (nDigit *= 2) > 9) nDigit -= 9

    nCheck += nDigit
    bEven = !bEven
  }

  return nCheck % 10 === 0
}

export const hideFormElement = (key, hide, form) => {
  const formUpdated = { ...form }

  const updatedFormElement = formUpdated[key]

  updatedFormElement.hidden = hide

  return formUpdated
}

// Mark required fields as errors
export const showFormRequiredFields = (form) => {
  const updatedForm = {
    ...form
  }
  for (const key in updatedForm) {
    const element = updatedForm[key]
    const { validationRules } = element
    if (validationRules) {
      if (validationRules.isRequired && !element.isValid) {
        element.isTouched = true
        validateInput(element)
        updatedForm[key] = element
      }
    }
  }
  return updatedForm
}

// Mark error on element
export const markErrorOnElement = (form, formInputName, customErrorMessage) => {
  const updatedForm = {
    ...form
  }
  const element = { ...updatedForm[formInputName] }
  element.isTouched = true
  validateInput(element)
  if (customErrorMessage) {
    element.isValid = false
    element.validationMessage = customErrorMessage
  }
  updatedForm[formInputName] = element
  return updatedForm
}

export const setValidationMessage = (key, message, form) => {
  const formUpdated = { ...form }

  const updatedFormElement = { ...formUpdated[key] }

  updatedFormElement.validationMessage = message
  updatedFormElement.isValid = false

  formUpdated[key] = updatedFormElement

  return formUpdated
}
