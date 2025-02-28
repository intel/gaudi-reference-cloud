import { React, useState } from 'react'
import CloudCreditCreate from '../../components/cloudCredits/cloudCreditCreate/CloudCreditCreate'
import { useNavigate } from 'react-router-dom'
import {
  UpdateFormHelper,
  isValidForm,
  getFormValue,
  setValidationMessage
} from '../../utility/updateFormHelper/UpdateFormHelper'
import moment from 'moment'
import useUserStore from '../../store/userStore/UserStore'
import BillingService from '../../services/BillingService'
import useToastStore from '../../store/toastStore/ToastStore'

const CloudCreditsContainer = () => {
  // local state
  const initialState = {
    desciption: 'Generate coupons for Intel Tiber AI Cloud Console',
    form: {
      amount: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Amount:',
        placeholder: 'Amount',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true,
          onlyCreditNumeric: true
        },
        validationMessage: '',
        helperMessage: ''
      },
      users: {
        type: 'text', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Number of users:',
        placeholder: 'Number of users',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true,
          onlyCreditNumeric: true
        },
        validationMessage: '',
        helperMessage: ''
      },
      startDt: {
        type: 'radio', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'Start date:',
        placeholder: 'Start date',
        value: '1',
        isValid: true,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        options: [
          {
            name: 'Current Date',
            value: '1'
          },
          {
            name: 'Select from calendar',
            value: '2'
          }
        ],
        validationMessage: '',
        helperMessage: ''
      },
      startDtCld: {
        label: 'Start date:',
        hiddenLabel: true,
        type: 'date', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        placeholder: '',
        value: '',
        isValid: true,
        isTouched: false,
        isReadOnly: true,
        validationRules: {
          isRequired: false
        },
        validationMessage: 'Start date is required.',
        helperMessage: ''
      },
      endDt: {
        type: 'date', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: 'End date:',
        placeholder: 'End date',
        value: '',
        isValid: false,
        isTouched: false,
        isReadOnly: false,
        validationRules: {
          isRequired: true
        },
        customContent: [],
        validationMessage: '',
        helperMessage: ''
      },
      isStandard: {
        type: 'checkbox', // options = 'text ,'textArea'
        fieldSize: 'medium', // options = 'small', 'medium', 'large'
        label: '',
        placeholder: '',
        value: false,
        isValid: true,
        isTouched: true,
        isReadOnly: false,
        validationRules: {
          isRequired: false
        },
        options: [
          {
            name: 'Is Coupon for Standard User',
            value: '0'
          }
        ],
        customContent: [],
        validationMessage: '',
        helperMessage: ''
      }

    },
    isValidForm: false,
    payloadService: {
      amount: null,
      creator: null,
      start: null,
      expires: null,
      numUses: null,
      isStandard: null
    },
    navigationTop: [
      {
        label: '⟵ Back to home',
        buttonVariant: 'link',
        function: () => onCancel()
      }
    ],
    navigationBottom: [
      {
        label: 'Create',
        buttonVariant: 'primary'
      },
      {
        label: 'Cancel',
        buttonVariant: 'link',
        function: () => onCancel()
      }
    ]
  }
  const [state, setState] = useState(initialState)
  const [showModal, setShowModal] = useState(false)
  const [onCreateOkModal, setOnCreateOkModal] = useState({
    show: false,
    message: '',
    redirecTo: '/billing/coupons',
    onClose: () => onResetForm()
  })

  // Global Store
  const user = useUserStore((state) => state.user)
  const showError = useToastStore((state) => state.showError)
  // Navigation
  const navigate = useNavigate()

  async function onSubmit() {
    const updatedState = {
      ...state
    }
    let updateForm = { ...updatedState.form }

    const amount = getFormValue('amount', updateForm)
    const users = getFormValue('users', updateForm)
    const isStandard = getFormValue('isStandard', updateForm)
    const startDtValue = getFormValue('startDt', updateForm)
    let startDt = null
    if (startDtValue === '1') {
      const now = moment()
      startDt = now.add(moment.duration(7, 'seconds'))
    } else if (startDtValue === '2') {
      const dateInput = moment(getFormValue('startDtCld', updateForm))
      startDt = dateInput.add(moment.duration(7, 'seconds'))
    }
    const endDt = moment(getFormValue('endDt', updateForm))
    if (!endDt.isSameOrAfter(startDt)) {
      updateForm = setValidationMessage('endDt', 'End date cannot be greater than Start date', updateForm)
      updatedState.form = updateForm
      setState(updatedState)
      return
    }
    const payload = { ...updatedState.payloadService }
    payload.amount = amount
    payload.start = startDt.format()
    payload.expires = endDt.format()
    payload.numUses = users
    payload.creator = user.email
    payload.isStandard = isStandard

    try {
      setShowModal(true)
      const response = await BillingService.submitCoupons(payload)
      const code = response.data.code
      const modalCopy = { ...onCreateOkModal }
      modalCopy.show = true
      modalCopy.message = code
      setOnCreateOkModal(modalCopy)
      setShowModal(false)
    } catch (error) {
      setShowModal(false)
      let message = ''
      if (error.response) {
        if (error.response.data.message !== '') {
          message = error.response.data.message
        } else {
          message = error.message
        }
      } else {
        message = error.message
      }
      showError(message)
    }
  }

  function onCancel() {
    navigate('/')
  }

  function onChangeInput(event, formInputName) {
    const updatedState = {
      ...state
    }

    let inputValue = event.target.value

    if (formInputName === 'isStandard') {
      inputValue = event.target.checked
    }

    const updatedForm = UpdateFormHelper(
      inputValue,
      formInputName,
      updatedState.form
    )

    if (formInputName === 'startDt') {
      updatedForm.startDtCld.value = ''
      updatedForm.startDtCld.isReadOnly = inputValue === '1'
      updatedForm.startDtCld.isValid = inputValue === '1'
    }

    updatedState.isValidForm = isValidForm(updatedForm)

    updatedState.form = updatedForm

    setState(updatedState)
  }

  function onResetForm() {
    setState(initialState)
    const modalCopy = { ...onCreateOkModal }
    modalCopy.show = false
    modalCopy.message = ''
    setOnCreateOkModal(modalCopy)
  }

  return (
    <CloudCreditCreate
      state={state}
      showModal={showModal}
      onCreateOkModal={onCreateOkModal}
      onResetForm={onResetForm}
      onSubmit={onSubmit}
      onChangeInput={onChangeInput} />
  )
}

export default CloudCreditsContainer
