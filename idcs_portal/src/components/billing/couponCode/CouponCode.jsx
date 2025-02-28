// INTEL CONFIDENTIAL
// Copyright (C) 2023 Intel Corporation

import ButtonGroup from 'react-bootstrap/ButtonGroup'
import Button from 'react-bootstrap/Button'
import CustomInput from '../../../utils/customInput/CustomInput'
import SpinnerBackdrop from '../../../utils/spinner/SpinnerBackdrop'

const CouponCode = (props) => {
  // props Variables
  const state = props.state
  const form = state.form
  const showSpinner = state.showSpinner
  const cancelButtonOptions = props.cancelButtonOptions

  // props functions
  const onChangeInput = props.onChangeInput
  const submitForm = props.submitForm
  // functions
  function buildCustomInput(key) {
    const configInput = { ...form[key] }

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
        isValid={configInput.isValid}
        isTouched={configInput.isTouched}
        isReadOnly={configInput.isReadOnly}
        options={configInput.options}
        validationMessage={configInput.validationMessage}
        readOnly={configInput.readOnly}
        minLength={configInput.minLength}
        maxLength={configInput.maxLength}
        customClass={configInput.customClass}
      />
    )
  }

  return (
    <>
      <SpinnerBackdrop showSpinner={showSpinner} />
      {buildCustomInput('couponCode')}
      <ButtonGroup>
        <Button
          intc-id="btn-managecouponcode-Redeem"
          data-wap_ref="btn-managecouponcode-Redeem"
          variant="primary"
          onClick={() => submitForm('coupon')}
        >
          Redeem
        </Button>
        <Button
          intc-id="btn-managecouponcode-RedeemCancel"
          data-wap_ref="btn-managecouponcode-RedeemCancel"
          variant="link"
          onClick={cancelButtonOptions.onClick}
        >
          {cancelButtonOptions.label}
        </Button>
      </ButtonGroup>
    </>
  )
}

export default CouponCode
