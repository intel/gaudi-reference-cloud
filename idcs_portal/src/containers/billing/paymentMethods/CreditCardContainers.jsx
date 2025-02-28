// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import { useEffect, useState } from 'react'
import CreditCard from '../../../components/billing/creditCard/CreditCard'
import usePaymentMethodStore from '../../../store/billingStore/PaymentMethodStore'
import {
  UpdateFormHelper,
  isValidForm,
  UpdateBlurFormHelper,
  setSelectOptions,
  getFormValue
} from '../../../utils/updateFormHelper/UpdateFormHelper'
import PaymentMethodService from '../../../services/PaymentMethodService'
import idcConfig from '../../../config/configurator'
import useUserStore from '../../../store/userStore/UserStore'
import CloudAccountService from '../../../services/CloudAccountService'

const CreditCardContainers = (props) => {
  // Props Variables
  const cancelButtonOptions = props.cancelButtonOptions
  const formActions = props.formActions

  // local state
  const initialState = {
    form: {
      cardNumber: {
        sectionGroup: 'creditCard',
        type: 'text', // options = 'text ,'textArea'
        label: 'Card number',
        placeholder: 'Card number',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        minLength: 17,
        maxLength: 19,
        isRequired: true,
        customClass: 'creditCard-all-images',
        validationRules: {
          isRequired: true,
          onlyCreditCard: true
        },
        isBlur: true,
        autocomplete: 'cc-number',
        validationMessage: '' // Error message to display to the user
      },
      cardMonth: {
        sectionGroup: 'creditCard',
        type: 'text', // options = 'text ,'textArea'
        label: 'Month',
        placeholder: 'MM',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 2,
        maxWidth: '8rem',
        isRequired: true,
        validationRules: {
          isRequired: true,
          onlyCreditNumeric: true,
          onlyCreditMonthYear: true
        },
        isBlur: true,
        autocomplete: 'cc-exp-month',
        validationMessage: '' // Error message to display to the user
      },
      cardYear: {
        sectionGroup: 'creditCard',
        type: 'text', // options = 'text ,'textArea'
        label: 'Year',
        placeholder: 'YY',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 2,
        maxWidth: '8rem',
        isRequired: true,
        validationRules: {
          isRequired: true,
          onlyCreditNumeric: true,
          onlyCreditMonthYear: true
        },
        isBlur: true,
        autocomplete: 'cc-exp-year',
        validationMessage: '' // Error message to display to the user
      },
      cardCvc: {
        sectionGroup: 'creditCard',
        type: 'text', // options = 'text ,'textArea'
        label: 'CVC',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 3,
        maxWidth: '8rem',
        isRequired: true,
        validationRules: {
          isRequired: true,
          onlyCreditNumeric: true,
          onlyCreditCvc: true
        },
        isBlur: true,
        autocomplete: 'cc-csc',
        validationMessage: '' // Error message to display to the user
      },
      cardFirstName: {
        sectionGroup: 'creditCard',
        type: 'text', // options = 'text ,'textArea'
        label: 'First Name',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        isRequired: true,
        validationRules: {
          isRequired: true,
          onlyAlphabets: true,
          checkMaxLength: true
        },
        autocomplete: 'given-name',
        validationMessage: '' // Error message to display to the user
      },
      cardLastName: {
        sectionGroup: 'creditCard',
        type: 'text', // options = 'text ,'textArea'
        label: 'Last Name',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        isRequired: true,
        validationRules: {
          isRequired: true,
          onlyAlphabets: true,
          checkMaxLength: true
        },
        autocomplete: 'family-name',
        validationMessage: '' // Error message to display to the user
      },
      cardEmail: {
        sectionGroup: 'creditCard',
        type: 'text', // options = 'text ,'textArea'
        label: 'Email',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        isRequired: true,
        validationRules: {
          isRequired: true,
          emailAddress: true
        },
        autocomplete: 'email',
        validationMessage: '' // Error message to display to the user
      },
      cardCompanyName: {
        sectionGroup: 'creditCard',
        type: 'text', // options = 'text ,'textArea'
        label: 'Company name',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        isRequired: false,
        validationRules: {
          checkMaxLength: true
        },
        autocomplete: 'organization',
        validationMessage: '' // Error message to display to the user
      },
      cardPhone: {
        sectionGroup: 'creditCard',
        type: 'text', // options = 'text ,'textArea'
        label: 'Phone',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        isRequired: false,
        validationRules: {
          checkMaxLength: true
        },
        autocomplete: 'tel',
        validationMessage: '' // Error message to display to the user
      },
      cardCountry: {
        sectionGroup: 'creditCard',
        type: 'select', // options = 'text ,'textArea'
        label: 'Country',
        placeholder: 'Please select country',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        isRequired: true,
        validationRules: {
          isRequired: true
        },
        options: [],
        autocomplete: 'country',
        validationMessage: '' // Error message to display to the user
      },
      cardAddress1: {
        sectionGroup: 'creditCard',
        type: 'text', // options = 'text ,'textArea'

        label: 'Address line 1',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        isRequired: true,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        autocomplete: 'address-line1',
        validationMessage: '' // Error message to display to the user
      },
      cardAddress2: {
        sectionGroup: 'creditCard',
        type: 'text', // options = 'text ,'textArea'
        label: 'Address line 2',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: true, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        isRequired: false,
        validationRules: {
          isRequired: false,
          checkMaxLength: true
        },
        autocomplete: 'address-line2',
        validationMessage: '' // Error message to display to the user
      },
      cardCity: {
        sectionGroup: 'creditCard',
        type: 'text', // options = 'text ,'textArea'
        label: 'City',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 63,
        isRequired: true,
        validationRules: {
          isRequired: true,
          checkMaxLength: true
        },
        autocomplete: 'address-level2',
        validationMessage: '' // Error message to display to the user
      },
      cardState: {
        sectionGroup: 'creditCard',
        type: 'select', // options = 'text ,'textArea'
        label: 'State',
        placeholder: 'Please select state',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        isRequired: true,
        validationRules: {
          isRequired: true
        },
        options: [],
        autocomplete: 'address-level1',
        validationMessage: '' // Error message to display to the user
      },
      cardZip: {
        sectionGroup: 'creditCard',
        type: 'text', // options = 'text ,'textArea'

        label: 'ZIP code',
        placeholder: '',
        value: '', // Value enter by the user
        isValid: false, // Flag to validate if the input is ready
        isTouched: false, // Flag to validate if the user has modified the input
        isReadOnly: false, // Input create as read only
        maxLength: 20,
        isRequired: true,
        validationRules: {
          isRequired: true,
          onlyAlphaNumSpace: true,
          checkMaxLength: true
        },
        isBlur: false,
        autocomplete: 'postal-code',
        validationMessage: '' // Error message to display to the user
      },
      isValidForm: false
    },
    showSpinner: false
  }

  const initialDirectPost = {
    cardNumber: '',
    cardMonth: '',
    cardYear: '',
    cardCvc: '',
    cardFirstName: '',
    cardLastName: '',
    cardCompanyName: '',
    cardEmail: '',
    cardPhone: '',
    cardCountry: '',
    cardAddress1: '',
    cardAddress2: '',
    cardCity: '',
    cardState: '',
    cardZip: '',
    inSessionID: '',
    formOfPayment: '',
    mode: '',
    directPostUrl: '',
    directPostClientNo: ''
  }

  const [state, setState] = useState(initialState)
  const [directPost, setDirectPost] = useState(initialDirectPost)
  const [country, setCountry] = useState(null)

  // Global State
  const countries = usePaymentMethodStore((state) => state.countries)
  const setCountries = usePaymentMethodStore((state) => state.setCountries)
  const states = usePaymentMethodStore((state) => state.states)
  const setStates = usePaymentMethodStore((state) => state.setStates)
  const isStandardUser = useUserStore((state) => state.isStandardUser)

  useEffect(() => {
    if (countries === null) {
      const fetchCountries = async () => {
        await setCountries()
      }
      fetchCountries()
    }
    setForm('cardCountry')
  }, [countries])

  useEffect(() => {
    const fetchStates = async () => {
      await setStates(country)
    }
    fetchStates()
  }, [country])

  useEffect(() => {
    if (states !== null) {
      setForm('cardState')
    }
  }, [states])

  const setForm = (inputFormName) => {
    const stateUpdated = {
      ...state
    }

    const options = inputFormName === 'cardCountry' ? countries : states

    stateUpdated.form = setSelectOptions(inputFormName, options ? [...options] : [], stateUpdated.form)

    setState(stateUpdated)
  }

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...state
    }

    const updatedForm = UpdateFormHelper(event.target.value, formInputName, updatedState.form)

    if (formInputName === 'cardNumber') {
      updatedForm.cardCvc.maxLength = 3
      const customClass = updatedForm[formInputName].customClass
      let cardType = customClass.split('-')
      cardType = cardType.at(-1)
      if (cardType === 'amex') {
        updatedForm.cardCvc.maxLength = 4
      }

      if (updatedForm.cardCvc.value.length > updatedForm.cardCvc.maxLength) {
        updatedForm.cardCvc.value = updatedForm.cardCvc.value.slice(0, -1)
      }
    }

    if (formInputName === 'cardCountry') {
      const country = event.target.value
      setCountry(country)
    }

    updatedForm.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  function onBlurInput(event, formInputName) {
    if (state.form[formInputName].isBlur) {
      const updatedState = {
        ...state
      }

      const updatedForm = UpdateBlurFormHelper(event.target.value, formInputName, updatedState.form)

      updatedForm.isValidForm = isValidForm(updatedForm)

      updatedState.form = updatedForm

      setState(updatedState)
    }
  }

  async function submitForm() {
    const updatedState = { ...state }
    updatedState.showSpinner = true
    setState(updatedState)

    // Calling Pre Payment
    try {
      if (isStandardUser()) {
        await CloudAccountService.upgradeCloudAccountByCreditCard()
      }
      const { data } = await PaymentMethodService.getPrePayment()
      submitCreditCard(data)
    } catch (error) {
      formActions.afterError('Could not able to establish session. Please try again later.')
    } finally {
      setState((prev) => ({ ...prev, showSpinner: false }))
    }
  }

  async function submitCreditCard(prePayment) {
    const updateDirectPost = { ...directPost }

    updateDirectPost.cardNumber = getFormValue('cardNumber', state.form)
    updateDirectPost.cardMonth = getFormValue('cardMonth', state.form)
    updateDirectPost.cardYear = `${new Date().getFullYear().toString().substring(0, 2)}${getFormValue(
      'cardYear',
      state.form
    )}`
    updateDirectPost.cardCvc = getFormValue('cardCvc', state.form).toString()
    updateDirectPost.cardFirstName = getFormValue('cardFirstName', state.form).trim()
    updateDirectPost.cardLastName = getFormValue('cardLastName', state.form).trim()
    updateDirectPost.cardEmail = getFormValue('cardEmail', state.form).trim()
    updateDirectPost.cardCompanyName = getFormValue('cardCompanyName', state.form)
    updateDirectPost.cardPhone = getFormValue('cardPhone', state.form).trim()
    updateDirectPost.cardCountry = getFormValue('cardCountry', state.form)
    updateDirectPost.cardAddress1 = getFormValue('cardAddress1', state.form).trim()
    updateDirectPost.cardAddress2 = getFormValue('cardAddress2', state.form).trim()
    updateDirectPost.cardCity = getFormValue('cardCity', state.form).trim()
    updateDirectPost.cardState = getFormValue('cardState', state.form)
    updateDirectPost.cardZip = getFormValue('cardZip', state.form)
    updateDirectPost.inSessionID = prePayment.sessionId
    updateDirectPost.formOfPayment = 'CreditCard'
    updateDirectPost.mode = 'reg'
    updateDirectPost.directPostUrl = prePayment.directPostUrl
    updateDirectPost.directPostClientNo = idcConfig.REACT_APP_ARIA_DIRECT_POST_CLIENT_NO
    setDirectPost(updateDirectPost)
  }

  return (
    <CreditCard
      state={state}
      directPost={directPost}
      onChangeInput={onChangeInput}
      onBlurInput={onBlurInput}
      submitForm={submitForm}
      cancelButtonOptions={cancelButtonOptions}
    />
  )
}

export default CreditCardContainers
