// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import React, { useEffect, useRef } from 'react'
import CustomInput from '../../../utils/customInput/CustomInput'
import SpinnerBackdrop from '../../../utils/spinner/SpinnerBackdrop'
import './CreditCard.scss'
import { Button, ButtonGroup } from 'react-bootstrap'

const CreditCard = (props) => {
  // props Variables
  const state = props.state
  const form = state.form
  const showSpinner = state.showSpinner
  const cancelButtonOptions = props.cancelButtonOptions
  const directPost = props.directPost

  // props functions
  const onChangeInput = props.onChangeInput
  const onBlurInput = props.onBlurInput
  const submitForm = props.submitForm

  const ref = useRef(null)

  // functions
  function buildCustomInput(key) {
    const configInput = form[key]

    return (
      <CustomInput
        key={key}
        type={configInput.type}
        fieldSize={configInput.fieldSize}
        placeholder={configInput.placeholder}
        isRequired={configInput.validationRules.isRequired}
        label={configInput.validationRules.isRequired ? configInput.label + ' *' : configInput.label}
        value={configInput.value}
        onChanged={(event) => onChangeInput(event, key)}
        onBlur={(event) => onBlurInput(event, key)}
        isValid={configInput.isValid}
        isTouched={configInput.isTouched}
        isReadOnly={configInput.isReadOnly}
        options={configInput.options}
        validationMessage={configInput.validationMessage}
        readOnly={configInput.readOnly}
        minLength={configInput.minLength}
        maxLength={configInput.maxLength}
        maxWidth={configInput.maxWidth}
        customClass={configInput.customClass}
        autocomplete={configInput.autocomplete}
      />
    )
  }

  const validateDirectPost = (directPost) => {
    return Boolean(directPost.cardNumber)
  }

  const shouldUseBillState = () => {
    const useBillState =
      !directPost.cardCountry ||
      directPost.cardCountry === 'US' ||
      directPost.cardCountry === 'CA' ||
      directPost.cardCountry === 'AU'
    return useBillState
  }

  useEffect(() => {
    if (validateDirectPost(directPost)) {
      const submitButton = ref.current
      submitButton.click()
    }
  }, [directPost])

  return (
    <>
      <SpinnerBackdrop showSpinner={showSpinner} />
      <span>* A $500 credit hold will be placed on your credit card for the next 7 to 10 business days.</span>
      <div className="section px-0">
        <h3 className="h5">Payment information</h3>
        {buildCustomInput('cardNumber')}
        <div className="d-flex flex-row gap-s6 w-100">
          {buildCustomInput('cardMonth')}
          {buildCustomInput('cardYear')}
          <div className="d-flex ms-s8">{buildCustomInput('cardCvc')}</div>
        </div>
      </div>
      <div className="section px-0">
        <h3 className="h5">Cardholder information</h3>
        <div className="d-flex flex-xs-column flex-md-row gap-s6 w-100">
          {buildCustomInput('cardFirstName')}
          {buildCustomInput('cardLastName')}
        </div>
      </div>
      {buildCustomInput('cardEmail')}
      <div className="d-flex flex-xs-column flex-md-row gap-s6 w-100">
        {buildCustomInput('cardCompanyName')}
        {buildCustomInput('cardPhone')}
      </div>
      {buildCustomInput('cardCountry')}
      {buildCustomInput('cardAddress1')}
      {buildCustomInput('cardAddress2')}
      {buildCustomInput('cardCity')}
      <div className="d-flex flex-xs-column flex-md-row gap-s6 w-100">
        {buildCustomInput('cardState')}
        {buildCustomInput('cardZip')}
      </div>
      <ButtonGroup>
        <Button
          intc-id="btn-credit-AddCreditPayment"
          data-wap_ref="btn-credit-AddCreditPayment"
          variant="primary"
          aria-label="Add credit card"
          disabled={!form.isValidForm}
          onClick={() => submitForm()}
        >
          Add Card
        </Button>
        <Button
          intc-id="btn-credit-CancelCreditCard"
          data-wap_ref="btn-credit-CancelCreditCard"
          variant="link"
          onClick={cancelButtonOptions.onClick}
        >
          {cancelButtonOptions.label}
        </Button>
      </ButtonGroup>
      <form id="AriaPay" name="payment_info" action={directPost.directPostUrl} method="post">
        <input type="hidden" id="client_no" name="client_no" value={directPost.directPostClientNo}></input>
        <input type="hidden" id="bill_first_name" name="bill_first_name" value={directPost.cardFirstName}></input>
        <input type="hidden" id="bill_last_name" name="bill_last_name" value={directPost.cardLastName}></input>
        <input type="hidden" id="bill_address1" name="bill_address1" value={directPost.cardAddress1}></input>
        <input type="hidden" id="bill_address2" name="bill_address2" value={directPost.cardAddress2}></input>
        <input type="hidden" id="bill_city" name="bill_city" value={directPost.cardCity}></input>
        <input
          type="hidden"
          id={shouldUseBillState() ? 'bill_state_prov' : 'bill_locality'}
          name={shouldUseBillState() ? 'bill_state_prov' : 'bill_locality'}
          value={directPost.cardState}
        ></input>
        <input type="hidden" id="bill_postal_cd" name="bill_postal_cd" value={directPost.cardZip}></input>
        <input type="hidden" id="bill_country" name="bill_country" value={directPost.cardCountry}></input>
        <input type="hidden" id="bill_company_name" name="bill_company_name" value={directPost.cardCompanyName}></input>
        <input type="hidden" id="bill_phone" name="bill_phone" value={directPost.cardPhone}></input>
        <input type="hidden" id="bill_email" name="bill_email" value={directPost.cardEmail}></input>
        <input type="hidden" id="cc_no" name="cc_no" value={directPost.cardNumber}></input>
        <input type="hidden" id="cc_exp_mm" name="cc_exp_mm" value={directPost.cardMonth}></input>
        <input type="hidden" id="cc_exp_yyyy" name="cc_exp_yyyy" value={directPost.cardYear}></input>
        <input type="hidden" name="cvv" value={directPost.cardCvc}></input>
        <input type="hidden" id="formOfPayment" name="formOfPayment" value={directPost.formOfPayment}></input>
        <input type="hidden" id="mode" name="mode" value={directPost.mode}></input>
        <input type="hidden" id="inSessionID" name="inSessionID" value={directPost.inSessionID}></input>
        <button ref={ref} style={{ display: 'none' }} id="submitAria" type="submit">
          Submit ARIA
        </button>
      </form>
    </>
  )
}

export default CreditCard
