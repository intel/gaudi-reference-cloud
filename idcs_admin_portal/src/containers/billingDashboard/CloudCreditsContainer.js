import React, { Component } from 'react'
import CloudCreditsForm from '../../components/billingDashboard/cloudcredits/CloudCreditsForm'
import CloudCreditsList from '../../components/billingDashboard/cloudcredits/CloudCreditsList'
import CloudCreditsView from '../../components/billingDashboard/cloudcredits/CloudCreditsView'
import Breadcrumb from '../../components/breadcrumb/Breadcrumb'
import BillingService from '../../services/BillingService'
import Wrapper from '../../utility/wrapper/Wrapper'
const moment = require('moment')
class CloudCreditsContainer extends Component {
  constructor (props) {
    super(props)

    this.state = {
      pageName: 'Cloud Credits',
      cloudList: [],
      loading: true,
      currentPath: this.props.location?.pathname.split('/').pop(),
      page: 'list',
      currentStage: 0,
      createdCCMessage: false,
      form: {
        creditCode: {
          machinename: 'creditcode',
          type: 'text', // options = 'text ,'textArea'
          fieldSize: 'medium', // options = 'small', 'medium', 'large'
          formGroup: 'cloudcredits', // This helps to group the form
          label: 'Credit code',
          placeholder: 'Credit code',
          value: '', // Value enter by the user
          isValid: false, // Flag to validate if the input is ready
          isTouched: false, // Flag to validate if the user has modified the input
          isReadOnly: false, // Input create as read only
          validationRules: {
            isRequired: true
          },
          validationMessage: '', // Errror message to display to the user
          button: {
            status: true,
            label: 'Generate'
          }
        },
        amount: {
          machinename: 'amount',
          type: 'text', // options = 'text ,'textArea'
          fieldSize: 'medium', // options = 'small', 'medium', 'large'
          formGroup: 'cloudcredits', // This helps to group the form
          label: 'Amount',
          placeholder: 'e.g:$2000',
          value: '', // Value enter by the user
          isValid: false, // Flag to validate if the input is ready
          isTouched: false, // Flag to validate if the user has modified the input
          isReadOnly: false, // Input create as read only
          validationRules: {
            isRequired: true
          },
          validationMessage: '' // Errror message to display to the user
        },
        creditDescription: {
          type: 'textarea', // options = 'text ,'textArea'
          fieldSize: 'medium', // options = 'small', 'medium', 'large'
          formGroup: 'cloudcredits', // This helps to group the form
          label: 'Credit Description',
          placeholder: 'Autosize height based on content lines',
          description: 'Provide a description for this Cloud Credit',
          value: '', // Value enter by the user
          isValid: false, // Flag to validate if the input is ready
          isTouched: false, // Flag to validate if the user has modified the input
          isReadOnly: false, // Input create as read only
          maxLength: 100,
          validationRules: {
            isRequired: true
          },
          validationMessage: '' // Errror message to display to the user
        },
        audience: {
          machinename: 'audience',
          type: 'radio', // options = 'text ,'textArea'
          fieldSize: 'medium', // options = 'small', 'medium', 'large'
          formGroup: 'cloudcredits',
          label: 'Audience',
          value: '',
          wildcard: '',
          isValid: false,
          isTouched: false,
          isReadOnly: false,
          validationRules: {
            isRequired: true
          },
          options: [
            {
              key: 'all',
              value: 'All',
              input: { status: false, value: 'all' }
            },
            {
              key: 'wildcard',
              value: 'Wildcard',
              input: {
                status: true,
                value: '',
                placeholder: '*@example.com;*.com'
              }
            },
            {
              key: 'business',
              value: 'Business accounts',
              input: { status: false, value: 'business' }
            },
            {
              key: 'personal',
              value: 'Personal accounts',
              input: { status: false, value: 'personal' }
            }
          ],
          validationMessage: ''
        },
        expireson: {
          type: 'datepicker', // options = 'text ,'textArea'
          fieldSize: 'medium', // options = 'small', 'medium', 'large'
          formGroup: 'cloudcredits', // This helps to group the form
          label: 'Expires on',
          description: "If left blank, code won't expire",
          placeholder: 'Select date',
          value: '', // Value enter by the user
          isValid: false, // Flag to validate if the input is ready
          isTouched: false, // Flag to validate if the user has modified the input
          isReadOnly: false, // Input create as read only
          validationRules: {
            isRequired: true
          },
          validationMessage: '' // Errror message to display to the user
        },
        creationstate: {
          machinename: 'creationstate',
          type: 'radio', // options = 'text ,'textArea'
          fieldSize: 'medium', // options = 'small', 'medium', 'large'
          formGroup: 'cloudcredits',
          label: 'Creation state',
          value: '',
          isValid: false,
          isTouched: false,
          isReadOnly: false,
          validationRules: {
            isRequired: true
          },
          options: [
            {
              key: 'active',
              value: 'Active',
              input: { status: false, value: 'active' }
            },
            {
              key: 'paused',
              value: 'Paused',
              input: { status: false, value: 'paused' }
            }
          ],
          validationMessage: ''
        }
      },
      stages: [
        {
          sectionGroup: 'cloudcredits', // it helps to group elements in the stages
          sectionType: 'form', // available types (card, table, form, review) modify the component based on the type
          title: 'Cloud Credit Details', // Title used in the stepper wizard
          titleStage: 'Cloud Credit Details', // Main Title for the stage
          subTitleStage: null, // Subtitle for the stage
          description: 'Specific the needed information', // Description stage
          stageParam: '', // Param for specific stage
          isTouched: true, // Flag for stage status (Complete, Incomplete, Error)
          data: null, // Data for the stage
          show_nav_buttons: true,
          show_right_details: true // Navigation between stages
        },
        {
          sectionGroup: 'review',
          sectionType: 'review',
          title: 'Review',
          titleStage: 'Review', // Subtitle for the stage
          subTitleStage: null, // subTitleStage for the stage
          description: 'Specific the needed information',
          stageParam: null,
          isTouched: false,
          data: null, // Data for the stage
          show_nav_buttons: true,
          show_right_details: true
        }
      ],
      isValidForm: false,
      instanceSelected: null,
      timeoutMiliseconds: 4000,
      cloudCredits: []
    }
  }

  componentDidMount () {
    // Call method to Retrives Cloud Credits.
    this.getCloudCredits()
  }

  getCloudCredits = () => {
    BillingService.getCloudCredits()
      .then((res) => {
        this.setState({
          loading: false,
          cloudCredits: res.data
        })
      })
      .catch(() => {
        this.setState({
          loading: false
        })
      })
  }

  whichPageToOpen = (page) => {
    if (page === 'create') {
      this.setState({ currentStage: 0 })
      this.resetFormValues()
    }

    if (page === 'cloudcredits') {
      this.setState({
        createdCCMessage: false
      })
    }
    this.setState({
      currentPath: page
      // createdCCMessage: false,
    })
  }

  // Function to validate inputs
  validateInput = (formElementToValidate) => {
    // Initial validation always is true
    let isValid = true

    // Get the rules for the element to validate
    if (formElementToValidate.validationRules.isRequired) {
      if (!formElementToValidate.value) {
        isValid = false
        formElementToValidate.validationMessage =
          formElementToValidate.label + ' is required'
        formElementToValidate.isValid = isValid
        // there is a error return the status
        return
      }
    }

    if (formElementToValidate.validationRules.onlyAlphaNumLower) {
      const word = formElementToValidate.value

      const regEx = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/g

      if (regEx.test(word) && word.length <= formElementToValidate.maxLength) {
        formElementToValidate.isValid = isValid
      } else {
        isValid = false
        formElementToValidate.validationMessage = `Only lower case alphanumeric and hypen(-) allowed for ${formElementToValidate.label}.`
        formElementToValidate.isValid = isValid
        // there is a error return the status
        return
      }
    }

    formElementToValidate.isValid = isValid
  }

  // Function to update the form elements
  onChangeInput = (event, formInputName) => {
    // Get a form copy to update
    const updatedForm = {
      ...this.state.form
    }

    // Get a element copy to update
    const updatedFormElement = {
      ...updatedForm[formInputName]
    }

    // Assign new value and change to true the onTouch field

    if (event.type === 'wildcard') {
      updatedFormElement.wildcard = event.target.value
    } else {
      updatedFormElement.value = event.target.value
    }

    updatedFormElement.isTouched = true

    // Validate input format
    this.validateInput(updatedFormElement, formInputName)

    // Update copy with the field validated
    updatedForm[formInputName] = updatedFormElement

    // Check the form to see if there are field with errors
    let isValidForm = true

    let errorCount = 0

    for (const element in updatedForm) {
      if (!updatedForm[element].isValid) {
        errorCount++
      }
    }

    if (errorCount > 0) {
      isValidForm = false
    }

    this.setState({
      form: updatedForm,
      isValidForm
    })
  }

  // Change the current stage
  setCurrentStage = (currentStage) => {
    this.setState({ currentStage })
  }

  // function to validate if the form is valid and prepare the payload to be sent to the service
  onSubmitHandler = () => {
    if (this.state.isValidForm) {
      this.createCloudCredits(this.state.form)
    } else {
      this.showErrorMessages()
    }
  }

  createCloudCredits = (formState) => {
    const cloudCreditsList = {
      name: formState.creditCode.value,
      creditCode: formState.creditCode.value,
      creditUnitAmount: Number(formState.amount.value),
      description: formState.creditDescription.value,
      creditStartTime: moment(formState.expireson.value).unix(),
      creditExpirationTime: moment(formState.expireson.value).unix(),
      customerName: '551a15cf-f39e-48bd-ad71-08ad30f739f3',
      customerType: '66c5771a-d962-43c5-9642-0c773caa97a2',
      customerId: '4cdf2036-6ae2-46d5-aa8c-139e9e824e26',
      currency: 'USD'
    }

    BillingService.submitCloudCredits(cloudCreditsList).then(() => {
      // Nothing to do at moment

      this.setState({
        cloudList: cloudCreditsList,
        createdCCMessage: formState.creditCode.value
      })

      this.getCloudCredits()
      this.whichPageToOpen('')
    })
  }

  // Show error messages when user tries to submit invalid form
  showErrorMessages = () => {
    // Get a form copy to update
    const updatedForm = {
      ...this.state.form
    }

    // Get a element copy to update
    for (const formInputName in updatedForm) {
      const updatedFormElement = {
        ...updatedForm[formInputName]
      }

      updatedFormElement.isTouched = true

      updatedForm[formInputName] = updatedFormElement
    }

    this.setState({
      form: updatedForm
    })
  }

  // Cancel handler.
  onCancelHandler = () => {
    this.resetFormValues()
    this.whichPageToOpen('')
    this.setState({
      currentStage: 0
    })
  }

  // Reset the form values to initial state
  resetFormValues = () => {
    const updatedForm = {
      ...this.state.form
    }

    // Get a element copy to update

    for (const formInputName in updatedForm) {
      const updatedFormElement = {
        ...updatedForm[formInputName]
      }

      updatedFormElement.isTouched = false

      if (updatedFormElement.validationRules.isRequired) {
        updatedFormElement.isValid = false
      } else {
        updatedFormElement.isValid = true
      }

      updatedFormElement.value = ''
      updatedFormElement.wildcard = ''

      updatedForm[formInputName] = updatedFormElement
    }

    this.setState({
      form: updatedForm
    })
  }

  render () {
    return (
      <Wrapper>
        {/* <MainHeading /> */}
        <Breadcrumb activePage="billing" />
        {(this.state.currentPath === 'cloudcredits' ||
          this.state.currentPath === '') && (
          <CloudCreditsList
            getCloudCredits={this.getCloudCredits}
            cloudCredits={this.state.cloudCredits}
            whichPageToOpen={this.whichPageToOpen}
            {...this.state}
          />
        )}
        {(this.state.currentPath === 'create' ||
          this.state.currentPath === 'edit') && (
          <CloudCreditsForm
            currentStage={this.state.currentStage}
            getCloudCredits={this.getCloudCredits}
            cloudCredits={this.state.cloudCredits}
            whichPageToOpen={this.whichPageToOpen}
            onChangeInput={this.onChangeInput}
            form={this.state.form}
            stages={this.state.stages}
            setCurrentStage={this.setCurrentStage}
            onSubmitHandler={this.onSubmitHandler}
            instanceSelected={this.state.instanceSelected}
            loading={this.state.loading}
            onCancelHandler={this.onCancelHandler}
            formType={this.state.currentPath}
          />
        )}
        {this.state.currentPath !== 'cloudcredits' &&
          this.state.currentPath !== 'create' &&
          this.state.currentPath !== '' &&
          this.state.currentPath !== 'edit' && (
            <CloudCreditsView
              getCloudCredits={this.getCloudCredits}
              cloudCredits={this.state.cloudCredits}
              whichPageToOpen={this.whichPageToOpen}
              code={this.state.currentPath}
              {...this.state}
            />
        )}
      </Wrapper>
    )
  }
}

export default CloudCreditsContainer
